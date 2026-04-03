# ElastiCache Native Migration Spec

## Overview

Migrate all 8 ElastiCache resources from Terraform-bridged controllers to native AWS SDK v2 controllers. This is the first service with **async operations**, **TF business logic translation** (CustomDiff, PasswordGenerator), and **v1beta2 hub/spoke conversion** — making it a critical stepping stone before Tier 4 services (EC2, RDS, Lambda).

## Resources

| # | Terraform Resource | Kind (RAW) | External Name Strategy | Async | v1beta2 | Complexity |
|---|---|---|---|---|---|---|
| 1 | `aws_elasticache_subnet_group` | SubnetGroupRAW | `NameAsIdentifier` | No | No | Low |
| 2 | `aws_elasticache_parameter_group` | ParameterGroupRAW | `IdentifierFromProvider` | No | No | Low |
| 3 | `aws_elasticache_user_group` | UserGroupRAW | `ParameterAsIdentifier("user_group_id")` | No | No | Low |
| 4 | `aws_elasticache_user` | UserRAW | `ParameterAsIdentifier("user_id")` | No | **Yes (hub)** | Medium |
| 5 | `aws_elasticache_cluster` | ClusterRAW | `ParameterAsIdentifier("cluster_id")` | No | No | Medium |
| 6 | `aws_elasticache_global_replication_group` | GlobalReplicationGroupRAW | `IdentifierFromProvider` | No | No | Medium |
| 7 | `aws_elasticache_serverless_cache` | ServerlessCacheRAW | `NameAsIdentifier` | **Yes** | No | Medium-High |
| 8 | `aws_elasticache_replication_group` | ReplicationGroupRAW | `ParameterAsIdentifier("replication_group_id")` | **Yes** | **Yes (hub)** | **High** |

## Implementation Order

Bottom-up by complexity. Each resource's implement ticket depends on the scaffold ticket. Phase dependencies are complexity-ordered, NOT architectural — if resources have no API-level dependency, they can be parallelized.

```
Phase 1 (scaffold):  All 8 resources scaffolded together
Phase 2 (simple):    SubnetGroup → ParameterGroup → UserGroup
Phase 3 (medium):    User (v1beta2) → Cluster → GlobalReplicationGroup
Phase 4 (async):     ServerlessCache → ReplicationGroup
```

## Design Decisions

### 1. Auth Token Generation (ReplicationGroup)

**Decision**: Implement native PasswordGenerator in the controller, writing the token to the K8s Secret at `authTokenSecretRef` before calling AWS.

The TF provider uses a `PasswordGenerator` `InitializerFn` (defined in `config/cluster/common/common.go`) that runs *before* Create: it generates a random token, writes it to the Secret referenced by `spec.forProvider.authTokenSecretRef`, then the controller reads it from there for the `CreateReplicationGroup` call.

**Native approach** — the controller must replicate this two-step process:

```go
func (e *ExternalClient) Create(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalCreation, error) {
    spec := cr.GetForProvider()
    
    var authToken string
    if spec.AutoGenerateAuthToken != nil && *spec.AutoGenerateAuthToken {
        // Step 0: Check if Secret already has a value — reuse if so (idempotent retry)
        // This matches common.PasswordGenerator behavior: common.go:98-101
        existing, _ := e.readSecretValue(ctx, spec.AuthTokenSecretRef)
        if existing != "" {
            authToken = existing
        } else {
            // Step 1: Generate a secure random token
            token, err := password.Generate() // crossplane-runtime/pkg/password
            if err != nil {
                return managed.ExternalCreation{}, errors.Wrap(err, "cannot generate auth token")
            }
            
            // Step 2: Write the token to the K8s Secret at authTokenSecretRef
            // Set OwnerReference so Secret is GC'd when the MR is deleted (common.go:121-122)
            if err := e.writeAuthTokenToSecret(ctx, cr, token); err != nil {
                return managed.ExternalCreation{}, errors.Wrap(err, "cannot write auth token to secret")
            }
            
            authToken = token
        }
    } else if spec.AuthTokenSecretRef != nil {
        // User provided an explicit auth token via secretRef — read it
        token, err := e.readSecretValue(ctx, spec.AuthTokenSecretRef)
        if err != nil {
            return managed.ExternalCreation{}, errors.Wrap(err, "cannot read auth token from secret")
        }
        authToken = token
    }
    
    input := &elasticache.CreateReplicationGroupInput{
        // ... other fields
        AuthToken: nillableString(authToken),
    }
    
    resp, err := e.Client.CreateReplicationGroup(ctx, input)
    // ...
    
    // Also publish as connection detail for consumers
    connDetails := managed.ConnectionDetails{}
    if authToken != "" {
        connDetails["auth_token"] = []byte(authToken)
    }
    return managed.ExternalCreation{ConnectionDetails: connDetails}, nil
}
```

**Key invariants** (must match `config/cluster/common/common.go` PasswordGenerator behavior):
- The token MUST be persisted to the K8s Secret before the AWS call
- If the Secret already has data at the key, skip generation and reuse the existing value (`common.go:98-101`)
- Set `OwnerReference` on the Secret so it is garbage-collected when the MR is deleted (`common.go:121-122`)
- If Create succeeds at AWS but the pod crashes, the token is recoverable from the Secret
- Observe must re-read `auth_token` from `authTokenSecretRef` and include in connection details (TF sensitive mechanism re-publishes on every reconcile)

**Auth token rotation** (`auth_token_update_strategy`): The Update path must handle auth token changes using the `AuthToken` and `AuthTokenUpdateStrategy` fields together in `ModifyReplicationGroup`. Valid strategies are `SET` (replace token), `ROTATE` (add new while keeping old temporarily), and `DELETE` (remove auth). The Update logic must check if the auth token changed and include `AuthTokenUpdateStrategy` in the input when it does.

The `auto_generate_auth_token` field must exist in the RAW type to maintain YAML compatibility with TF examples. It is a write-only field (never returned by AWS API).

### 2. Async Operations

**Decision**: Full async (Create/Update/Delete) for ReplicationGroup and ServerlessCache, matching TF behavior (`UseAsync = true`).

All three CRUD operations return immediately while the resource transitions through intermediate states.

**Status values**:

| Resource | Creating | Available | Modifying | Deleting | Failed |
|---|---|---|---|---|---|
| ReplicationGroup | `creating` | `available` | `modifying` | `deleting` | `create-failed` |
| ServerlessCache | `CREATING` | `AVAILABLE` | `MODIFYING` | `DELETING` | `CREATE-FAILED` |

Note: ReplicationGroup uses lowercase, ServerlessCache uses UPPERCASE.

**Native async pattern** (using `internal/native/async.go`):

```go
// In Create:
func (e *ExternalClient) Create(ctx context.Context, cr CR) (managed.ExternalCreation, error) {
    resp, err := e.Client.CreateReplicationGroup(ctx, input)
    if err != nil { return managed.ExternalCreation{}, err }
    
    meta.SetExternalName(cr, aws.ToString(resp.ReplicationGroup.ReplicationGroupId))
    
    rid, _ := middleware.GetRequestIDMetadata(resp.ResultMetadata)
    native.SetAsyncState(cr, native.AsyncState{
        Operation: "creating",
        StartedAt: time.Now(),
        RequestID: rid,
    })
    return managed.ExternalCreation{ConnectionDetails: connDetails}, nil
}

// In Observe — async-aware polling:
func (e *ExternalClient) Observe(ctx context.Context, cr CR) (managed.ExternalObservation, error) {
    asyncState := native.GetAsyncState(cr)
    
    resp, err := e.Client.DescribeReplicationGroups(ctx, &elasticache.DescribeReplicationGroupsInput{
        ReplicationGroupId: aws.String(native.GetExternalName(cr)),
    })
    if err != nil {
        if native.IsNotFound(err) {
            if asyncState != nil && asyncState.Operation == "deleting" {
                // Delete completed — resource is gone
                native.ClearAsyncState(cr)
                return managed.ExternalObservation{ResourceExists: false}, nil
            }
            return managed.ExternalObservation{ResourceExists: false}, nil
        }
        return managed.ExternalObservation{}, native.Wrap(err, errDescribe)
    }
    
    // Bounds check — empty result means not found
    if len(resp.ReplicationGroups) == 0 {
        return managed.ExternalObservation{ResourceExists: false}, nil
    }
    rg := resp.ReplicationGroups[0]
    status := aws.ToString(rg.Status)
    
    // Handle async in-flight operations
    if asyncState != nil {
        switch status {
        case "available":
            native.ClearAsyncState(cr)
            cr.SetConditions(xpv1.Available())
            // Fall through to normal observation
        case "creating", "modifying", "snapshotting":
            cr.SetConditions(xpv1.Unavailable())
            return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
        case "deleting":
            cr.SetConditions(xpv1.Deleting())
            return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
        case "create-failed":
            native.ClearAsyncState(cr)
            return managed.ExternalObservation{ResourceExists: true},
                fmt.Errorf("replication group %s failed", asyncState.Operation)
        }
    }
    
    // Handle transitional states even without async annotations
    // (e.g., resource was modified outside the controller)
    switch status {
    case "creating", "modifying", "snapshotting":
        cr.SetConditions(xpv1.Unavailable())
        return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
    case "deleting":
        cr.SetConditions(xpv1.Deleting())
        return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
    case "create-failed":
        return managed.ExternalObservation{ResourceExists: true},
            errors.New("replication group creation failed")
    }
    
    // ... normal observe: populate status, check isUpToDate, late-init
}

// In Update:
func (e *ExternalClient) Update(ctx context.Context, cr CR) (managed.ExternalUpdate, error) {
    // ... build ModifyReplicationGroup input
    resp, err := e.Client.ModifyReplicationGroup(ctx, input)
    if err != nil { return managed.ExternalUpdate{}, err }
    
    rid, _ := middleware.GetRequestIDMetadata(resp.ResultMetadata)
    native.SetAsyncState(cr, native.AsyncState{
        Operation: "updating",
        StartedAt: time.Now(),
        RequestID: rid,
    })
    return managed.ExternalUpdate{}, nil
}

// In Delete:
func (e *ExternalClient) Delete(ctx context.Context, cr CR) (managed.ExternalDelete, error) {
    cr.SetConditions(xpv1.Deleting())
    resp, err := e.Client.DeleteReplicationGroup(ctx, input)
    if err != nil {
        if native.IsNotFound(err) { return managed.ExternalDelete{}, nil }
        return managed.ExternalDelete{}, err
    }
    
    rid, _ := middleware.GetRequestIDMetadata(resp.ResultMetadata)
    native.SetAsyncState(cr, native.AsyncState{
        Operation: "deleting",
        StartedAt: time.Now(),
        RequestID: rid,
    })
    return managed.ExternalDelete{}, nil
}
```

### 3. Transitional State Handling for Non-Async Resources

**Cluster**, **GlobalReplicationGroup**, **User**, and **UserGroup** are NOT marked `UseAsync` in the TF config, but AWS operations are still async at the API level (creating a cluster/user takes seconds to minutes). The TF provider uses synchronous waiters internally.

In the native controller, handle transitional states in Observe exactly like Kinesis Stream. All four resources have a `Status` field with transitional values:

| Resource | Status Field | Active | Transitional |
|---|---|---|---|
| Cluster | `CacheClusterStatus` | `available` | `creating`, `modifying`, `rebooting cluster nodes`, `snapshotting`, `deleting` |
| GlobalReplicationGroup | `Status` | `available` | `creating`, `modifying`, `deleting` |
| User | `Status` | `active` | `modifying`, `deleting` |
| UserGroup | `Status` | `active` | `creating`, `modifying`, `deleting` |

```go
// In Observe for Cluster/GlobalReplicationGroup/User/UserGroup:
status := aws.ToString(resource.Status) // or CacheClusterStatus for Cluster
switch status {
case "available", "active":
    cr.SetConditions(xpv1.Available())
    // proceed to isUpToDate
case "creating", "modifying", "rebooting cluster nodes", "snapshotting":
    cr.SetConditions(xpv1.Unavailable())
    // Return UpToDate=true to prevent spurious Update calls during transitions
    return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
case "deleting":
    cr.SetConditions(xpv1.Deleting())
    return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}
```

**Without this**, Observe calls `isUpToDate()` on a "creating" resource where AWS defaults aren't applied yet → returns `false` → triggers `Update()` → AWS returns `InvalidCacheClusterState` / `InvalidUserState` / `InvalidUserGroupState`.

### 4. Schema Target & Multi-Version

**Decision**: Target v1beta2 flat schema for ReplicationGroup and User (both are hub versions). Implement Hub/Spoke conversion for v1beta1 immediately.

**ReplicationGroup v1beta1 → v1beta2 conversion**:
```
v1beta1: spec.forProvider.clusterMode[0].numNodeGroups      → v1beta2: spec.forProvider.numNodeGroups
v1beta1: spec.forProvider.clusterMode[0].replicasPerNodeGroup → v1beta2: spec.forProvider.replicasPerNodeGroup
```

The existing TF conversion in `config/cluster/elasticache/config.go` defines custom converters with `IdentityConversionExpandPaths` for `clusterMode`. The RAW type conversion must replicate this exactly.

**User v1beta1 → v1beta2 conversion**:

The actual structural change is **slice→pointer** for the `authenticationMode` block:
```
v1beta1: AuthenticationMode []AuthenticationModeParameters   (YAML: array)
v1beta2: AuthenticationMode *AuthenticationModeParameters    (YAML: object)
```

The `authenticationMode` block is NOT flattened/promoted in v1beta2 — it remains nested. The `PasswordsSecretRef` field stays inside the `authenticationMode` block in both versions.

**Important**: RAW types do NOT implement the `Terraformed` interface, so the TF roundtrip conversion (`ujconversion.RoundTrip`) cannot be used. RAW types must implement custom `ConvertTo()`/`ConvertFrom()` methods that explicitly copy each field and set `TypeMeta` on the destination.

### 5. Late Initialization

**Decision**: Match TF behavior exactly.

**ReplicationGroup ignored fields** (from `config/cluster/elasticache/config.go`):
- `cluster_mode` — conflicts with `num_node_groups`/`replicas_per_node_group` (v1beta2 flat equivalents)
- `num_node_groups` — conflicts with `cluster_mode`
- `num_cache_clusters` — conflicts with `number_cache_clusters` (legacy alias)
- `number_cache_clusters` — legacy alias for `num_cache_clusters`
- `replication_group_description` — legacy alias for `description`
- `description` — conflicts with `replication_group_description`

All other AWS-defaulted fields (engine_version, maintenance_window, snapshot_window, port, etc.) should be late-initialized normally.

### 6. Custom Diff Suppression

**Decision**: Skip `security_group_names` in both Observe comparison AND Update inputs.

The TF config deletes `security_group_names.#` from the diff because this is an EC2-Classic legacy field. AWS returns it in the API response even for VPC-based clusters.

**Native approach**:
- In `isUpToDate()`: skip the `SecurityGroupNames` field entirely
- In `Update()`: never include `SecurityGroupNames` in `ModifyReplicationGroupInput` — AWS returns "SecurityGroupNames not supported in VPC" for VPC clusters
- Only compare and update `SecurityGroupIds`

```go
// security_group_names is an EC2-Classic legacy field.
// AWS API returns it even for VPC clusters. Never compare or update it.
// Only use security_group_ids.
```

### 7. Connection Details

Three resources publish connection details:

**Cluster** (Memcached):
| Key | SDK Path | Condition |
|-----|----------|----------|
| `cluster_address` | `CacheCluster.ConfigurationEndpoint.Address` | Memcached only (ConfigurationEndpoint is nil for Redis) |
| `port` | `CacheCluster.ConfigurationEndpoint.Port` (Memcached) or `CacheCluster.CacheNodes[0].Endpoint.Port` (Redis) | **All engines** |

**ReplicationGroup**:
| Key | SDK Path | Condition |
|-----|----------|-----------|
| `configuration_endpoint_address` | `ReplicationGroup.ConfigurationEndpoint.Address` | Cluster mode enabled |
| `primary_endpoint_address` | `ReplicationGroup.NodeGroups[0].PrimaryEndpoint.Address` | Cluster mode disabled |
| `reader_endpoint_address` | `ReplicationGroup.NodeGroups[0].ReaderEndpoint.Address` | Cluster mode disabled |
| `port` | `ReplicationGroup.NodeGroups[0].PrimaryEndpoint.Port` | Always |
| `auth_token` | Re-read from `authTokenSecretRef` Secret | When auth is enabled |

Source: `.agents/specs/connection-details-catalog.json`. Note: `auth_token` must be re-published on every Observe (TF sensitive mechanism re-publishes automatically; native must do it explicitly).

**ServerlessCache**:
| Key | SDK Path |
|-----|----------|
| `endpoint_0_address` | `ServerlessCache.Endpoint.Address` |
| `endpoint_0_port` | `ServerlessCache.Endpoint.Port` |
| `reader_endpoint_0_address` | `ServerlessCache.ReaderEndpoint.Address` |
| `reader_endpoint_0_port` | `ServerlessCache.ReaderEndpoint.Port` |

ServerlessCache uses indexed keys (`endpoint_0_*`) for backward compatibility with existing connection secret consumers. The TF config loops over endpoint lists. Do not simplify the key names even though the SDK returns single structs.

### 8. ReplicationGroup Update Decomposition

ReplicationGroup Update is the most complex Update of any migrated resource. `ModifyReplicationGroup` only handles a subset of changes. Some operations require separate API calls:

| Change | API Call | Async |
|--------|----------|-------|
| Most field changes (description, engine_version, node_type, etc.) | `ModifyReplicationGroup` | Yes |
| Scaling node groups (num_node_groups) | `ModifyReplicationGroupShardConfiguration` | Yes (separate async) |
| Auth token rotation | `ModifyReplicationGroup` with `AuthToken` + `AuthTokenUpdateStrategy` | Yes |
| Tag changes | `AddTagsToResource` / `RemoveTagsFromResource` | No |

**Important**: Like Kinesis Stream, multiple sequential AWS API calls in Update may transition the resource to `modifying` state, causing subsequent calls to fail. The controller should either:
- (a) Make one change per reconcile cycle and return (reconciler retries), or
- (b) Check status between calls

Option (a) is simpler and recommended for initial implementation. The reconciler naturally retries on the next cycle.

### 9. References

Includes references from `config.go` configurators AND `overrides.go` KnownReferencers (auto-applied to `subnet_ids`, `security_group_ids`, `kms_key_id`, `vpc_id`, `*_role_arn`).

| Resource | Field | References | Source |
|---|---|---|---|
| SubnetGroup | `subnet_ids` | `aws_subnet` (SubnetIDRefs/Selector) | KnownReferencers |
| Cluster | `parameter_group_name` | `aws_elasticache_parameter_group` | config.go |
| Cluster | `subnet_group_name` | `aws_elasticache_subnet_group` | auto-ref |
| Cluster | `replication_group_id` | `aws_elasticache_replication_group` | auto-ref |
| Cluster | `security_group_ids` | `aws_security_group` (SecurityGroupIDRefs/Selector) | KnownReferencers |
| GlobalReplicationGroup | `primary_replication_group_id` | `aws_elasticache_replication_group` | auto-ref |
| ReplicationGroup | `subnet_group_name` | `aws_elasticache_subnet_group` | config.go |
| ReplicationGroup | `kms_key_id` | `aws_kms_key` | config.go |
| ReplicationGroup | `security_group_ids` | `aws_security_group` (SecurityGroupIDRefs/Selector) | KnownReferencers |
| ReplicationGroup | `global_replication_group_id` | `aws_elasticache_global_replication_group` | auto-ref |
| ServerlessCache | `kms_key_id` | `aws_kms_key` | config.go |
| ServerlessCache | `security_group_ids` | `aws_security_group` (SecurityGroupIDRefs/Selector) | KnownReferencers |
| ServerlessCache | `subnet_ids` | `aws_subnet` (SubnetIDRefs/Selector) | KnownReferencers |
| UserGroup | `user_ids` | `aws_elasticache_user` (list ref, custom field names: `UserIDRefs`, `UserIDSelector`) | config.go |

Note: `log_delivery_configuration.destination` has its auto-ref **deleted** in both Cluster and ReplicationGroup because it can point to either CloudWatch Logs or Kinesis Firehose — TF can't auto-ref polymorphic targets.

### 10. NativeSetupHook Registration

Every resource controller must be registered via `NativeSetupHook_elasticache` in the thin wrapper's `init()` function. Without this, the controllers compile but are never registered with the manager — all 8 resources silently do nothing at runtime.

```go
// internal/controller/cluster/elasticache/subnetgroupraw/controller.go
func init() {
    clustercontroller.NativeSetupHook_elasticache = Setup
}
```

### 11. IdentifierFromProvider External Name Handling

Two resources use `IdentifierFromProvider` — their AWS-assigned ID must be stored as the external name in Create. Without this, `status.atProvider` is lost before the next Observe (pattern spec §17).

**ParameterGroup**: `meta.SetExternalName(cr, aws.ToString(resp.CacheParameterGroup.CacheParameterGroupName))`

### 11b. GlobalReplicationGroup External Name

`GlobalReplicationGroup` uses `IdentifierFromProvider` — the AWS-assigned ID must be stored as the external name in Create:

```go
func (e *ExternalClient) Create(ctx context.Context, cr CR) (managed.ExternalCreation, error) {
    resp, err := e.Client.CreateGlobalReplicationGroup(ctx, input)
    if err != nil { return managed.ExternalCreation{}, err }
    
    // Critical: store provider-assigned ID as external name
    // status.atProvider is lost before the next Observe (reconciler resets it)
    meta.SetExternalName(cr, aws.ToString(resp.GlobalReplicationGroup.GlobalReplicationGroupId))
    return managed.ExternalCreation{}, nil
}
```

### 12. User Sensitive Field Handling

User has two password paths that must both be supported:

1. **Top-level** `spec.forProvider.passwordsSecretRef` (`*[]v1.SecretKeySelector`) — legacy API path
2. **Nested** `spec.forProvider.authenticationMode.passwordsSecretRef` (`*[]v1.SecretKeySelector`) — new API path

Both fields reference K8s Secrets containing password values. The native controller must:
- Read secret values from whichever path is populated
- Pass them to `CreateUser` / `ModifyUser` AWS API calls
- Never expose password values in `status.atProvider` (sensitive)
- Publish password hashes as connection details (matching TF sensitive field mechanism)

### 13. Type Mismatch Fields

Some ReplicationGroup fields use `*string` in the TF types where the AWS SDK uses `*bool`. RAW types MUST use the same Go type as the TF types for CRD schema parity:

| Field | TF Type | AWS SDK Type | RAW Type Must Use |
|---|---|---|---|
| `AtRestEncryptionEnabled` | `*string` | `*bool` | `*string` — convert to `*bool` when calling AWS SDK |
| `AutoMinorVersionUpgrade` | `*string` | `*bool` | `*string` — convert to `*bool` when calling AWS SDK |

If RAW uses `*bool`, existing YAML with `atRestEncryptionEnabled: "true"` (string) would fail CRD validation.

### 14. MoveToStatus Field Audit

**Audit result**: No ElastiCache resources have fields that need relegation to `status.atProvider`-only beyond what's already computed-only in the schema. The following fields are computed-only and belong in `Observation` only:

- ReplicationGroup: `arn`, `cluster_enabled`, `configuration_endpoint_address`, `engine_version_actual`, `member_clusters`, `primary_endpoint_address`, `reader_endpoint_address`
- Cluster: `arn`, `cache_nodes`, `cluster_address`, `configuration_endpoint`
- User: `arn`
- ServerlessCache: `arn`, `endpoint`, `reader_endpoint`, `create_time`, `status`

The `.agents/specs/move-to-status-catalog.json` has no ElastiCache entries. This is confirmed correct — no writable TF fields were reclassified as read-only during schema evolution.

### 13. Scaffold Scope

Single scaffold ticket covering all 8 resources, following the existing `plan-native-migration` skill pattern:
- Stub RAW types in `apis/cluster/elasticache/{version}/native/`
- Stub RAW types in `apis/namespaced/elasticache/{version}/native/`
- CR interface per resource
- Stub `crud.go` per resource in `internal/controller/elasticache/`
- Thin controller wrappers per scope
- Example YAML manifests mirroring TF examples
- `make generate.native` for CRDs
- Scheme registration in `native_register.go` (both scopes)
- `NativeSetupHook_elasticache` assignment in wrapper `init()`
- Namespaced extractor annotations point to `config/namespaced/common.*` (NOT `config/cluster/common.*`)

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| ReplicationGroup schema complexity causes missed drift fields | High | Medium | Systematic field-by-field audit against TF types during implementation |
| Async state gets stuck (annotation corruption) | Low | High | `internal/native/async.go` handles malformed timestamps; add test for annotation edge cases |
| Auth token lost on pod restart | Medium | High | Write token to K8s Secret BEFORE the AWS call; token is recoverable from Secret on retry |
| v1beta1↔v1beta2 conversion data loss | Medium | High | Copy all fields explicitly, set TypeMeta, test with round-trip assertion; do NOT use ujconversion.RoundTrip |
| `security_group_names` sent in Update input | Low | Medium | Skip in both isUpToDate AND Update input building |
| Multi-step Update hits InvalidCacheClusterState | Medium | Low | One change per reconcile cycle; reconciler retries naturally |
| Transitional states cause spurious Updates (Cluster, GlobalRG) | High | Medium | Return UpToDate=true for all non-"available" states |

## Out of Scope

- ElastiCache Global Datastore (not a separate TF resource)
- Memcached-specific cluster management (handled by the same `aws_elasticache_cluster` resource)
- Migration of existing CR instances from TF to RAW (handled at cutover, a global operation)
- Performance optimization of polling intervals for async operations

## Dependencies

- `internal/native/async.go` — async annotation helpers (already exists)
- `internal/native/tags.go` — tag diffing with defaults (already exists)
- `internal/native/policy.go` — not needed (no IAM policy fields)
- `crossplane-runtime/pkg/password` — for auth token generation (already a dependency)
- Phase 0 infrastructure — already complete
- `make generate.native` — already functional
