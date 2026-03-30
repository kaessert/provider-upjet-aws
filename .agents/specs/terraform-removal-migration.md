# Terraform Removal Migration — Design Spec

## Goal

Remove the Terraform layer from provider-aws by replacing all 349 TF-bridged resources (both `cluster` and `namespaced` scopes = ~698 type/controller pairs) with native AWS SDK v2 implementations. During migration, every resource exists twice (`Kind: Bucket` TF-backed, `Kind: BucketRAW` SDK-native). After agent verification confirms parity, the TF kind is removed and the RAW kind is renamed to the original name. Services are cut over one at a time until the entire TF stack (upjet, terraform-provider-aws) can be removed.

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| CRD types | Independent (no `tf` tags) | Cleaner Go code, purpose-built for SDK |
| Scope | Both cluster + namespaced | Full coverage; namespaced mirrors cluster |
| Generation | LLM executor per batch | Flexible, no tooling investment, adapts per resource |
| Verification | Runtime e2e parity testing | True behavioral proof, not just schema analysis |
| Batching | By service | Coherent context, natural grouping |
| Cutover | Rolling per service | Lower risk, incremental progress |
| Phase 0 | Front-load all infrastructure | Solid foundation before any migration |

## Architecture: TF vs Native

```
CURRENT (TF-bridged):
  K8s CR → upjet AsyncConnector → TF PluginSDK (in-process) → AWS SDK v2 → AWS API
  Types: generated from TF schema, have `tf:"..."` struct tags
  Controller options: tjcontroller.Options (upjet-specific)

NATIVE (RAW):
  K8s CR → native Connector → TypedExternalClient → AWS SDK v2 → AWS API
  Types: independent, json tags only, same YAML schema as TF types
  Controller options: controller.Options (crossplane-runtime standard)
```

Reference pattern: `github.com/crossplane/provider-template` — uses `controller.Options` (crossplane-runtime) and `managed.WithTypedExternalConnector` (type-safe). NOT the ClusterAuth controller (which still imports upjet types).

## Critical Constraint: YAML Compatibility

RAW types MUST produce the same CRD schema as TF types. This means:
- Same `json:"..."` struct tags → same field names in YAML
- Same ForProvider/InitProvider/AtProvider structure (Crossplane convention)
- Same field types and nesting
- Same `+crossplane:generate:reference` annotations (for reference resolution code generation)
- NO `tf:"..."` tags (the whole point)
- Respect `MoveToStatus` — fields moved to status in TF types must be status-only in RAW types

This ensures:
1. Existing YAML manifests work with both kinds (just change `kind:`)
2. Cross-service references continue working after cutover
3. E2e tests are truly shared (same YAML, different kind)
4. `GetAWSConfigWithTracking` credential resolution works (requires `spec.forProvider.region` at the correct path)

---

## Phase 0: Infrastructure (front-loaded, ~25-30 tickets)

All infrastructure work must be complete and validated before any service migration begins.

### 0.1 — Native Controller Options Struct

**Problem**: The current provider uses `tjcontroller.Options` (upjet) which carries `Provider`, `SetupFn`, `OperationTrackerStore` — all TF-specific. The ClusterAuth controller (our only "native" example) also imports upjet types (`tjcontroller.Options`, `ujresource.SetUpToDateCondition`). Neither can serve as a clean template.

**Solution**: Create a native options struct following the `provider-template` pattern:

```go
// internal/native/options.go
package native

import (
    "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
)

// Options extends crossplane-runtime's controller.Options with AWS-specific fields.
type Options struct {
    controller.Options
    // PollInterval for async resources
    PollInterval time.Duration
}
```

Use `controller.Options` (from `crossplane-runtime/v2/pkg/controller`) — NOT `tjcontroller.Options`.

**Acceptance**: Options struct compiles, can be used in a `Setup()` function that wires a controller without any upjet imports.

### 0.2 — Native Connector Framework

Create the reusable connector at `internal/native/`:

```
internal/native/
  options.go       — Native controller options (no upjet)
  connector.go     — ProviderConfig → aws.Config resolution
  errors.go        — Standard error wrapping (NotFound, AccessDenied, etc.)
  tags.go          — Tag diffing utilities
  lateinit.go      — Late initialization helpers with ignore lists
  externalname.go  — External name get/set utilities
  async.go         — Async operation tracking (for long-running resources)
```

**Critical**: The connector MUST call `clients.GetAWSConfigWithTracking(ctx, c.kube, mg)` which requires:
1. The managed resource implements `resource.Managed` interface
2. The CR has `spec.forProvider.region` at the correct JSON path
3. The CR has a valid `providerConfigRef`

**Integration test gate**: Write a test that instantiates a RAW type, calls `GetAWSConfigWithTracking`, and verifies credentials resolve correctly. This MUST pass before any resource migration starts.

**Global resource handling**: `internal/clients/aws.go` has `globalResources` and `globalGroups` maps keyed by API group names (e.g., `iam.aws.upbound.io`). During the parallel phase, RAW types may have different API group suffixes. The connector must:
- Use the same API group as TF types during parallel phase, OR
- Add RAW API group entries to the global maps

### 0.3 — Async Operation System

**Problem**: 76+ resources use `UseAsync = true` in upjet. Upjet handles this with `OperationTrackerStore`, `OperationTrackerFinalizer`, `CallbackProvider`, and `MetricRecorder`. Long-running operations (EKS cluster ~15min, RDS ~10min) don't return immediately — they track in-flight ops and poll for completion.

**Solution**: Build a poll-based async pattern using Crossplane's reconciler:

```go
// internal/native/async.go

// AsyncState tracks in-flight operations via CR annotations
type AsyncState struct {
    Operation  string    // "creating", "updating", "deleting"
    StartedAt time.Time
    RequestID  string    // AWS request ID for idempotent retries
}

// In Observe():
// 1. Check if there's an in-flight operation annotation
// 2. If yes, poll AWS for completion
// 3. If still running → return ResourceExists:true, ResourceUpToDate:true (let reconciler poll again)
// 4. If complete → clear annotation, return actual state
// 5. If failed → clear annotation, return error

// In Create()/Update()/Delete():
// 1. Call AWS API
// 2. If response is async, set annotation with operation state
// 3. Return (reconciler will call Observe on next poll)
```

The managed.Reconciler's `WithPollInterval` handles re-polling. No need for a separate tracker store — use CR annotations.

**Acceptance**: Demonstrate the async pattern with a mock resource that simulates a 30-second creation delay. The reconciler must correctly poll, detect completion, and report Ready.

### 0.4 — Reference Resolution Pipeline

**Problem**: Cross-resource references (e.g., `bucketRef` → resolves to bucket ARN) are currently generated by upjet. The generated `ResolveReferences()` and `ResolveMultipleReferences()` methods use `+crossplane:generate:reference` annotations processed by `crossplane-tools` code generator.

**Solution**: RAW types MUST include the same `+crossplane:generate:reference` annotations. Then run `crossplane-tools` (already available: `make generate` runs it) to generate the reference resolution code. The annotations look like:

```go
// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
// +crossplane:generate:reference:extractor=github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("arn",true)
RoleARN *string `json:"roleArn,omitempty"`

// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
RoleARNRef *v1.Reference `json:"roleArnRef,omitempty"`

// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
RoleARNSelector *v1.Selector `json:"roleArnSelector,omitempty"`
```

**Critical note on extractors**: The TF types use `resource.ExtractParamPath("arn",true)` which traverses `status.atProvider.arn` via fieldpath. This works if the RAW types have the same field path for `arn` in the Observation struct. The TF types also use `resource.TerraformID()` extractor which casts to `resource.Terraformed` (an upjet interface). **RAW types cannot use `TerraformID()` extractor** — these references must be rewritten to use `resource.ExtractResourceID()` or custom extractors.

**Acceptance**: Create a test RAW type with reference annotations, run code generation, and verify reference resolution compiles and resolves correctly.

### 0.5 — External Name Extraction

**Problem**: `config/externalname.go` has **1,008 entries** with complex per-resource logic: `TemplatedStringAsIdentifier`, `IdentifierFromProvider`, custom `GetExternalNameFn`/`SetIdentifierArgumentFn`/`GetIDFn`. Examples:
- S3 Bucket: `RandRFC1123Subdomain` for name generation
- ECS Cluster: ARN parsing to extract cluster name
- Many resources: composite identifiers (`agentRuntimeId:name`)

**Solution**: Extract external name config into a machine-readable JSON catalog:

```json
{
  "aws_s3_bucket": {
    "strategy": "name_as_external_name",
    "name_generator": "rfc1123_subdomain",
    "id_field": "id"
  },
  "aws_ecs_cluster": {
    "strategy": "identifier_from_provider",
    "get_external_name": "parse_arn_last_segment",
    "id_field": "arn"
  },
  "aws_bedrock_agent": {
    "strategy": "templated",
    "template": "{{ .parameters.agent_id }}",
    "id_fields": ["agent_id"]
  }
}
```

Each implementation ticket must specify the exact external name pattern for the resource. The native controller uses `meta.SetExternalName(cr, ...)` and `meta.GetExternalName(cr)` from crossplane-runtime.

**Acceptance**: External name catalog covers all 349 resources; validated against actual `externalname.go` entries.

### 0.6 — Late Initialization Framework

**Problem**: 25+ resources define `LateInitializer` with specific `IgnoredFields` and `ConditionalIgnoredFields`. Late initialization copies server-set defaults from AWS responses into the CR spec so the next reconcile doesn't see a diff.

**Solution**: Generic late-initialization helper:

```go
// internal/native/lateinit.go

type LateInitConfig struct {
    IgnoredFields       []string
    ConditionalIgnored  map[string]func(cr resource.Managed) bool
}

// LateInitialize compares AWS response fields with CR spec fields,
// filling in any that are nil in the CR but set in AWS.
// Respects the ignore list.
func LateInitialize(spec, observed interface{}, config LateInitConfig) bool {
    // Returns true if any field was initialized
}
```

**Acceptance**: Late init helper correctly populates missing fields, respects ignore lists, handles nested structs.

### 0.7 — TF Business Logic Catalog

**Problem**: Resources with `TerraformConfigurationInjector` (9 resources) and `TerraformCustomDiff` (20+ resources) contain critical business logic for drift avoidance. Examples:
- S3 Bucket: Default `force_destroy = false` to prevent reconciliation loops
- ECS Service: Suppress task_definition revision diffs
- RDS: Various diff suppression logic

**Solution**: Create a catalog at `.agents/specs/tf-business-logic-catalog.md` listing every resource with injectors/custom diffs, what the logic does, and how to translate it to native Observe/Update behavior. Each implementation ticket for these resources must explicitly reference this catalog.

**Acceptance**: Catalog covers all resources with injectors/diffs; each entry explains the native equivalent.

### 0.8 — CRD Versioning Strategy

**Problem**: Several resources have multiple API versions (v1beta1, v1beta2) with conversion webhooks. Resources like `s3.BucketLifecycleConfiguration` set `PreviousVersions`, `SetCRDStorageVersion`, `ControllerReconcileVersion`.

**Solution**: During the parallel phase, RAW types only need ONE version (simplest). At cutover:
1. The native type MUST register all versions the TF type served
2. Conversion webhooks must be implemented (or an in-place migration of stored objects)
3. Storage version must match the TF type's storage version

Document this per-resource in the cutover ticket.

### 0.9 — Build System Isolation

**Problem**: Native `*_raw_types.go` in the same package as `zz_*` generated files risks `make generate` clobbering native code.

**Solution**: RAW types live in a separate sub-package during the parallel phase:

```
apis/cluster/s3/v1beta1/          — TF types (zz_* generated)
apis/cluster/s3/v1beta1/native/   — RAW types (hand-written)
```

At cutover, native types move up to replace the TF types.

### 0.10 — MoveToStatus Field Catalog

**Problem**: 12+ resources use `config.MoveToStatus()` to move TF writable fields to the status subresource. If someone mechanically copies TF schema fields to RAW types, they'll put these fields in the wrong place.

**Solution**: Extract `MoveToStatus` calls into a per-resource field list. The scaffold step for each resource uses this to correctly place fields in Parameters vs Observation.

### 0.11 — Connection Details Catalog

**Problem**: 14+ resources define `AdditionalConnectionDetailsFn` to publish secrets (IAM AccessKey publishes username/secret, RDS publishes endpoint/port). Native controllers must replicate this via `managed.ExternalCreation{ConnectionDetails: ...}`.

**Solution**: Catalog each resource's connection detail keys and their source fields in the AWS SDK response.

### 0.12 — Migration Plan Skill

New skill at `.claude/skills/plan-native-migration/SKILL.md` (detailed design in dedicated section below).

### 0.13 — Executor Spec

Write a comprehensive spec at `.agents/specs/native-controller-pattern.md` (detailed design in dedicated section below).

### 0.14 — Provider Binary Hook (`hack/main.go.tmpl`)

**Problem**: `cmd/provider/*/zz_main.go` is fully generated by upjet. There is no way to register native controllers in `main()` without modifying the template. Every `make generate` overwrites `zz_main.go`.

**Solution**: One-time modification to `hack/main.go.tmpl` adding a nil-safe extension hook:

```go
// In hack/main.go.tmpl — after TF controller setup:
if setupFn := clustercontroller.NativeSetupHook_{{ .Group }}; setupFn != nil {
    kingpin.FatalIfError(setupFn(mgr, clusterOptions.Options), "Cannot setup native cluster controllers")
}
if setupFn := namespacedcontroller.NativeSetupHook_{{ .Group }}; setupFn != nil {
    kingpin.FatalIfError(setupFn(mgr, namespacedOptions.Options), "Cannot setup native namespaced controllers")
}
```

The hook variable is defined in a hand-written (non-`zz_`) file in `internal/controller/cluster/` and populated via `init()` in the native controllers package. Providers without native controllers are unaffected (nil check).

**Acceptance**: `make generate` produces `zz_main.go` with the native hook. Native `Setup*` functions are callable.

### 0.15 — Resolver Exclusion for Native Types

**Problem**: The upjet resolver (`go run cmd/resolver -p ../apis/cluster/...`) recurses into `native/` sub-packages and expects `resource.Terraformed`. Native types don't implement this interface → resolver panics or generates broken code.

**Solution**: Either:
a) Exclude `native/` from resolver invocation in `generate/generate.go`, OR
b) Restructure native types to `apis/cluster/<service>native/<version>/` (avoids recursion entirely — cleaner architectural choice)

**Acceptance**: `make generate` completes without errors when native type packages exist.

### 0.16 — Scheme Registration Pattern

**Problem**: `apis/cluster/zz_register.go` is generated and won't include native types. Native types must be registered in the manager scheme for controllers to work.

**Solution**: Create hand-written `apis/cluster/native_register.go` (no `zz_` prefix → survives `make generate`):

```go
package cluster

import natives3 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"

func init() {
    AddToSchemes = append(AddToSchemes, natives3.SchemeBuilder.AddToScheme)
}
```

Go processes `init()` in files alphabetically within a package. `native_register.go` (n) runs before `zz_register.go` (z), so native entries are appended first, then TF entries append on top. Both scheme builders end up in the same `AddToSchemes` slice.

**Acceptance**: Native types are discoverable by the controller manager scheme.

### 0.17 — Conversion Webhook Strategy

**Problem**: Generated `zz_generated.conversion_spokes.go` files contain `dstRaw.(resource.Terraformed)` type assertions. Native types don't implement `resource.Terraformed`. When a multi-version CRD serves v1beta1 and v1beta2, Kubernetes calls the conversion webhook on every request. The type assertion will **panic**, blocking ALL access to the resource kind.

Additionally, `package/kustomize/kustomization.yaml` applies `strategy: Webhook` to ALL CRDs globally, including native ones.

**Solution**: For the parallel phase (RAW types), use a single API version only → no conversion needed, webhook patch is harmless. At cutover:
1. Rewrite conversion spoke functions to use plain JSON round-trip instead of `ujconversion.RoundTrip` (which requires `resource.Terraformed`)
2. Or: implement lightweight `resource.Terraformed` stubs on native types (just enough to satisfy the cast — `GetTerraformResourceType() → ""`, etc.)
3. Ensure the kustomize webhook patch either excludes native CRDs or native types register their own conversion handler

**Acceptance**: Multi-version native CRDs can be read/written through all served versions without panics.

### 0.18 — IAM Policy Semantic Equivalence Library

**Problem**: AWS normalizes IAM policy JSON (reorders keys, adds defaults). 10+ resources (SQS Queue, SNS Topic, KMS Key, S3 Bucket Policy, OpenSearch access policy) need semantic comparison. A naive string comparison causes infinite reconcile loops.

**Solution**: Create `internal/native/policy.go` wrapping `awspolicyequivalence.PoliciesAreEquivalent()`. Every native controller with a `policy` field must use this instead of `reflect.DeepEqual`.

**Acceptance**: Policy comparison returns equal for semantically equivalent but syntactically different IAM JSON.

### 0.19 — Provider Default Tags Awareness

**Problem**: TF ProviderConfig supports `default_tags` that are automatically applied to all resources and tracked in `status.atProvider.tagsAll`. After cutover, the native controller sees AWS tags that aren't in `spec.forProvider.tags` and would remove them — silently deleting production tags.

**Solution**: Native controllers must:
1. Read the ProviderConfig's `default_tags` (if present)
2. Merge them with `spec.forProvider.tags` before comparing against AWS state
3. Populate `status.atProvider.tagsAll` with the merged tag set

**Acceptance**: Tags applied via `default_tags` in ProviderConfig are not removed by native controllers.

### 0.20 — IRSA Credential Cache for Native Path

**Problem**: `GetAWSConfigWithTracking` never calls the `AWSCredentialsProviderCache.RetrieveCredentials()` cache. Native controllers using IRSA hit AWS STS credential provider on every reconcile. At scale this causes excessive STS token refresh calls and potential throttling.

**Solution**: Extend `GetAWSConfigWithTracking` (or create a wrapper `GetAWSConfigWithTrackingAndCache`) that checks the IRSA credential cache before loading fresh credentials. The cache is already TF-dependency-free (`creds_cache.go` imports only `aws-sdk-go-v2` and `crossplane-runtime`).

**Acceptance**: Native controllers using IRSA share the same credential cache as TF controllers.

---

## Per-Service Migration Cycle

For each service, the migration follows this exact sequence. Both `cluster` and `namespaced` scopes are handled in parallel within the same service.

### Step 0: Baseline E2E (1 ticket per resource — HARD GATE)

Before any migration work begins for a resource, run the existing e2e test for the **TF-backed resource** using its current example manifest.

```
For each resource in the service:
  1. Run e2e test: examples/<service>/cluster/<version>/<resource>.yaml
  2. The TF resource must reach Ready condition
  3. The TF resource must delete cleanly
  4. No AWS permission errors
  5. No AWS quota/limit errors
```

**This is an absolute gate with zero tolerance for workarounds:**

| Failure | Action |
|---------|--------|
| E2E test fails for any reason | Mark ticket **Failed** — do NOT proceed to scaffold/implement |
| AWS permission denied | Mark ticket **Failed** — do not attempt IAM workarounds |
| AWS quota/limit exceeded | Mark ticket **Failed** — do not request increases or reduce test scope |
| Example manifest is broken | Mark ticket **Failed** — do not fix the manifest |
| Timeout / infrastructure issue | Mark ticket **Failed** — do not retry with extended timeouts |

**Rationale**: If we cannot verify the existing TF resource works end-to-end, we have no valid baseline to compare the native implementation against. Migrating a broken resource produces a broken native resource. The failure is recorded in the ticket with the exact error, enabling triage as a separate concern.

**Failure propagation**: When a baseline ticket fails, ALL downstream tickets for that resource (scaffold, implement, e2e-RAW, verify, cutover) are blocked. The service can still proceed with its other resources — a single failed baseline does not block the entire service, but the cutover ticket cannot complete until ALL resources in the service pass.

### Step 1: Scaffold (1 ticket per scope = 2 tickets)

Create the RAW type definitions and empty controller stubs for ALL resources in the service:

```
apis/cluster/<service>/<version>/native/<resource>_raw_types.go     — RAW CRD types (cluster)
apis/namespaced/<service>/<version>/native/<resource>_raw_types.go  — RAW CRD types (namespaced)
internal/controller/cluster/<service>/<resource>raw/controller.go   — Empty controller (cluster)
internal/controller/namespaced/<service>/<resource>raw/controller.go — Empty controller (namespaced)
examples/<service>/cluster/<version>/<resource>raw.yaml             — Example manifest
examples/<service>/namespaced/<version>/<resource>raw.yaml          — Example manifest
```

Wire into provider binary at `cmd/provider/<service>/zz_main.go`.

Field placement rules:
- Check `MoveToStatus` catalog → those fields go in Observation only
- Check external name catalog → set the name strategy
- Include `+crossplane:generate:reference` annotations (rewrite any `TerraformID()` extractors)
- Include `spec.forProvider.region` (required for credential resolution)

### Step 2: Implement CRUD (1 ticket per resource per scope)

For each resource, implement the native controller. Each ticket covers ONE resource in ONE scope:
- `Observe()` — Call AWS Describe/Get API, map response to status, handle late initialization
- `Create()` — Call AWS Create API, set external name, publish connection details
- `Update()` — Call AWS Update/Modify API (respect business logic from TF catalog)
- `Delete()` — Call AWS Delete API, handle async deletion

Reference material for executor:
1. TF CRUD code: `vendor/github.com/upbound/terraform-provider-aws/internal/service/<service>/`
2. AWS SDK v2 client: `vendor/github.com/aws/aws-sdk-go-v2/service/<service>/`
3. Native pattern spec: `.agents/specs/native-controller-pattern.md`
4. External name catalog: `.agents/specs/external-name-catalog.json`
5. TF business logic catalog: `.agents/specs/tf-business-logic-catalog.md`
6. MoveToStatus catalog: `.agents/specs/move-to-status-catalog.json`
7. Connection details catalog: `.agents/specs/connection-details-catalog.json`

### Step 3: E2E Test RAW (1 ticket per resource, cluster scope only)

Run the e2e test with the RAW kind:
- Copy example YAML, change `kind: Bucket` → `kind: BucketRAW`
- Run through uptest/chainsaw (create → wait Ready → delete → wait gone)
- Capture test results
- Same hard failure rules as baseline: permission errors, quota errors, or any infrastructure issue → mark Failed, no workarounds

### Step 4: Agent Verification (1 ticket per service)

Deploy resources (both TF and RAW variants) and run the parity test:
- For each resource: deploy TF + RAW side by side
- Mutate each mutable field on both
- Compare AWS state
- Generate parity report
- Gate: ALL resources must pass parity

**Cost mitigation**: Tier verification by complexity:
- Tier 1 (single-resource services): Lifecycle-only testing (create/ready/delete)
- Tier 2-3: Full parity for resources with >3 mutable fields; lifecycle-only for simple ones
- Tier 4: Full parity for all resources (complex services need thorough verification)

### Step 5: Cutover (1 ticket per service)

Once agent verification passes:
1. Move native types from `native/` sub-package to parent package
2. Rename RAW types → original type names (e.g., `BucketRAW` → `Bucket`)
3. Register all API versions the TF type served (handle v1beta1/v1beta2 duality)
4. Implement conversion webhooks if multiple versions existed
5. Replace TF controller registrations with native ones
6. Remove TF-specific config (`config/cluster/<service>/config.go`)
7. Delete `zz_*` generated files for this service
8. Run e2e tests with original kind names
9. Commit: `refactor: migrate <service> from terraform to native SDK`

---

## Batch Ordering

### Priority Tiers

Services ordered by: (1) resource count ascending, (2) cross-reference dependency.

**Tier 1 — Single resource services (38 services, 38 resources)**
Start here to prove the pattern. Each service is a focused test of the framework.

```
Priority (most referenced / simplest first):
  1. secretsmanager    2. sfn    3. firehose    4. dsql    5. ebs
  ... remaining 33 single-resource services
```

**Tier 2 — Small services (2-4 resources, 28 services)**

```
Priority:
  1. kms (5 resources, 35 cross-refs — CRITICAL dependency, promote to Tier 2)
  2. sns (2 resources, 4 cross-refs)
  3. sqs (4 resources, 5 cross-refs)
  4. kinesis (2 resources, 9 cross-refs)
  5. dynamodb (4 resources)
  ... remaining
```

**Tier 3 — Medium services (5-12 resources)**

```
Priority:
  1. iam (12 resources, 29 cross-refs — promote for dependency clearance)
  2. s3 (11 resources, 18 cross-refs)
  3. lambda (9 resources, 18 cross-refs)
  4. eks (7 resources, 8 cross-refs)
  5. rds (15 resources)
  ... remaining
```

**Tier 4 — Large services**

```
  1. directconnect (13 resources)
  2. ec2 (40 resources — largest, most complex, LAST)
```

### Cross-Reference Safety During Parallel Phase

**Why ordering works despite dependency complexity**:

1. During the parallel phase, both TF and RAW resources exist. Cross-references in RAW types point to TF types (same API group, same kind for referenced resources).
2. RAW e2e tests create their own dependency resources (separate instances).
3. At cutover, the RAW type is renamed to the original kind name. The CRD schema is identical (same json tags, same reference annotations). Already-resolved reference values (stored as ARN/ID strings in CR spec) remain valid.
4. IAM and KMS are promoted to Tier 2-3 (earlier migration) because they're the most-referenced. This means they're cut over early, and subsequent services reference the native (renamed) types.

### Ticket Naming Convention

```
native-<service>-<resource>-baseline  — Baseline e2e (TF) — HARD GATE
native-<service>-scaffold-cluster     — Scaffold (cluster scope)
native-<service>-scaffold-namespaced  — Scaffold (namespaced scope)
native-<service>-<resource>-cluster   — Implement CRUD (cluster)
native-<service>-<resource>-ns        — Implement CRUD (namespaced)
native-<service>-<resource>-e2e       — E2E test RAW
native-<service>-verify               — Agent verification
native-<service>-cutover              — Cutover
```

### Dependency Chain Per Resource

```
baseline (TF e2e) ──┐
                     ├──→ implement-cluster ──→ implement-ns
scaffold-cluster ────┘           │
                                 ↓
                          e2e-RAW ──→ verify ──→ cutover
```

If baseline **fails**, the entire downstream chain for that resource is blocked. Other resources in the same service proceed independently.

---

## Plan Skill Design: `plan-native-migration`

```yaml
---
name: plan-native-migration
description: "Create pheromone tickets to migrate one AWS service from Terraform to native SDK."
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Write
  - AskUserQuestion
argument-hint: "[service-name] e.g. 's3', 'dynamodb', 'ec2'"
---
```

### Skill Workflow

1. **Parse service name** from `$ARGUMENTS`
2. **Discover resources**: Read `config/cluster/<service>/config.go`, extract all `AddResourceConfigurator("aws_...")` calls
3. **For each resource, gather metadata**:
   a. Read TF types in `apis/cluster/<service>/<version>/zz_<resource>_types.go`
   b. Read TF CRUD location in `vendor/.../internal/service/<service>/`
   c. Check external name strategy in `config/externalname.go`
   d. Check for `MoveToStatus` calls in config
   e. Check for `TerraformConfigurationInjector` or `TerraformCustomDiff`
   f. Check for `AdditionalConnectionDetailsFn`
   g. Check for `LateInitializer` config
   h. Check for `UseAsync = true`
   i. Check for reference annotations
   j. Check for multiple API versions
4. **Generate field metadata**: Write `internal/verification/testdata/<service>.json`
5. **Create tickets** (all with `stage:executor,plan:native-<service>`):

```
[BLOCKING] Scaffold RAW types for <service> (cluster)
  depends_on: []

[BLOCKING] Scaffold RAW types for <service> (namespaced)
  depends_on: []

Per resource: Implement native <Resource>RAW controller (cluster)
  depends_on: [scaffold-cluster]
  Special notes: {async?, injector?, custom_diff?, connection_details?}

Per resource: Implement native <Resource>RAW controller (namespaced)
  depends_on: [scaffold-namespaced, <resource>-cluster]

Per resource: E2E test <Resource>RAW
  depends_on: [<resource>-cluster impl]

Agent verification for <service>
  depends_on: [all e2e test tickets]

Cutover <service> from TF to native
  depends_on: [agent verification]
  Special notes: {versions to support, conversion webhooks needed?}
```

6. **Report ticket count** and key warnings (async resources, business logic, multi-version resources)

### Per-Ticket Description Template

Each implementation ticket MUST include:
```
## Resource: <TF name> → <RAW kind>

### Files to create/modify
- apis/cluster/<service>/<version>/native/<resource>_raw_types.go
- internal/controller/cluster/<service>/<resource>raw/controller.go

### AWS SDK v2 Operations
- Create: <service>.Create<Resource>
- Observe: <service>.Describe<Resource> / Get<Resource>
- Update: <service>.Update<Resource> / Modify<Resource>
- Delete: <service>.Delete<Resource>

### External Name
Strategy: <from catalog>
Get: <how to extract from AWS response>
Set: <how to pass to AWS API>

### Special Handling
- [ ] Async: {yes/no} — if yes, use async pattern
- [ ] Business logic: {description from TF catalog}
- [ ] Late init ignored fields: {list}
- [ ] Connection details: {keys and sources}
- [ ] References: {fields that reference other resources}
- [ ] MoveToStatus fields: {list}

### Acceptance Criteria
- [ ] Controller compiles without upjet imports
- [ ] go test ./internal/controller/cluster/<service>/<resource>raw/... passes
- [ ] External name is correctly set on Create
- [ ] Observe returns ResourceExists:false for missing resources
- [ ] Observe returns ResourceUpToDate:false when spec != observed state
```

---

## Executor Ant Adaptations

The executor ant directive graph does NOT need structural changes. The current loop (claim → read specs → implement TDD → commit → mark Done) works. What changes:

### 1. Executor Spec Update

Add to `.agents/specs/`:
- `native-controller-pattern.md` — comprehensive guide for building native controllers
- `external-name-catalog.json` — per-resource external name strategies
- `tf-business-logic-catalog.md` — injectors and custom diffs to port
- `move-to-status-catalog.json` — fields that must be observation-only
- `connection-details-catalog.json` — connection detail keys per resource

The executor's Step 3 ("Read specs") already reads `.agents/specs/INDEX.md` and relevant specs. Just create an index entry.

### 2. Verification Ticket Handling

Agent verification tickets require the nohup+poll pattern (real AWS resources, >600s runtime). The executor spec must document this:

```
For tickets labeled "verification":
1. Write a test script to /tmp/run-verification.sh
2. Launch with nohup
3. Poll for completion
4. Report results in ticket
```

### 3. Complexity-Aware Ticket Sizing

For complex resources (EC2, RDS, IAM), tickets must include more detail. The plan skill must generate richer descriptions for these resources, including:
- Specific API call sequences (EC2 has `ModifyVpcAttribute`, `AuthorizeSecurityGroupIngress`, etc.)
- Error handling nuances (eventual consistency, throttling)
- Multi-step creation workflows

---

## Agent Verification Detail

### Per-Resource Parity Test

```go
func TestParity_S3Bucket(t *testing.T) {
    // 1. Create TF resource
    tfBucket := applyManifest(t, "examples/s3/cluster/v1beta2/bucket.yaml")
    // 2. Create RAW resource (same spec, different kind)
    rawBucket := applyManifest(t, "examples/s3/cluster/v1beta2/bucketraw.yaml")

    // 3. Wait for both Ready
    waitForReady(t, tfBucket, 5*time.Minute)
    waitForReady(t, rawBucket, 5*time.Minute)

    // 4. Mutable field tests
    for _, field := range mutableFields("aws_s3_bucket") {
        t.Run("mutable/"+field.Name, func(t *testing.T) {
            patchField(t, tfBucket, field.Name, field.TestValue)
            patchField(t, rawBucket, field.Name, field.TestValue)
            waitForSynced(t, tfBucket)
            waitForSynced(t, rawBucket)
            tfState := describeAWSResource(t, tfBucket)
            rawState := describeAWSResource(t, rawBucket)
            assertStateEqual(t, field.Name, tfState, rawState)
        })
    }

    // 5. Immutable field tests
    for _, field := range immutableFields("aws_s3_bucket") {
        t.Run("immutable/"+field.Name, func(t *testing.T) {
            tfErr := patchField(t, tfBucket, field.Name, field.TestValue)
            rawErr := patchField(t, rawBucket, field.Name, field.TestValue)
            assertSameErrorBehavior(t, tfErr, rawErr)
        })
    }

    // 6. Cleanup
    deleteAndWait(t, tfBucket)
    deleteAndWait(t, rawBucket)
}
```

### Verification Tiering (Cost Mitigation)

| Service Tier | Resources | Verification Level | Estimated AWS Cost |
|-------------|-----------|-------------------|-------------------|
| Tier 1 | 38 single-resource | Lifecycle only (create/ready/delete) | Low |
| Tier 2 | ~80 resources | Full parity for >3 mutable fields | Medium |
| Tier 3 | ~88 resources | Full parity for all resources | Medium-High |
| Tier 4 | ~143 resources | Full parity + stress test | High |

### Parity Report Format

```
=== Agent Verification Report: S3 ===
Service: s3
Resources: 11
Date: 2025-01-15T10:30:00Z

Resource: Bucket
  Status: PASS
  Mutable fields tested: 2/2
    ✓ tags: TF={"Env":"test"} RAW={"Env":"test"}
    ✓ force_destroy: TF=true RAW=true
  Immutable fields tested: 1/1
    ✓ bucket: Both reject (ForceNew)
  Lifecycle: create ✓ | observe ✓ | update ✓ | delete ✓

Overall: 11/11 PASS → READY FOR CUTOVER
```

---

## Cutover Process (per service)

### Critical Finding: CRD Schemas Are NOT Identical

Investigation revealed six concrete divergences between TF-generated and native CRDs:
1. `spec.deletionPolicy` — present in TF (cluster), absent in native (namespaced) schema
2. `spec.providerConfigRef` — TF requires only `name`; native requires `kind` + `name`
3. Storage version — TF stores as `v1beta2`; naive native would serve only `v1beta1`
4. `status.atProvider.tagsAll` — populated by TF (from provider `default_tags`), never by naive native
5. `description` fields — TF descriptions come from TF schema JSON; native from Go doc comments
6. Conversion webhooks — use `.(resource.Terraformed)` type assertion that panics on native types

**These findings mean "just rename RAW → original" is invalid.** Full CRD replacement with pre-migration jobs is required.

### Pre-Cutover Migration Jobs (MANDATORY — run before deploying native controller)

#### Job 1: Storage Version Migration
```bash
# Migrate all stored objects from v1beta2 → v1beta1 (or whichever the native stored version is)
kubectl-migrate --resource buckets.s3.aws.upbound.io
```
**Why**: etcd stores objects in the storage version. If TF CRD stores as v1beta2 and native CRD only serves v1beta1, all existing objects become unreadable.

**Alternative**: Native types MUST serve v1beta2 (even if the internal representation matches v1beta1). This avoids storage migration entirely.

#### Job 2: Patch `providerConfigRef.kind` on All Existing CRs
```bash
# For cluster-scoped resources:
kubectl get buckets.s3.aws.upbound.io -o name | xargs -I{} \
  kubectl patch {} --type=merge -p '{"spec":{"providerConfigRef":{"kind":"ProviderConfig"}}}'
```
**Why**: TF CRD defaults `providerConfigRef` to `{name: "default"}` (no `kind`). If the native CRD schema requires `kind`, existing CRs fail validation on any write.

#### Job 3: Map `deletionPolicy` → `managementPolicies`
```bash
# For any CR with deletionPolicy: Orphan, set managementPolicies: ["Observe"]
# This preserves the user's intent to not delete the AWS resource
```
**Why**: If the native CRD drops `deletionPolicy`, Kubernetes structural schema pruning silently removes it from all CRs on first write. Users lose deletion protection.

#### Job 4: Drain Async Operations
```bash
# Verify no resources are in Creating/Updating/Deleting state
kubectl get managed -o jsonpath='{range .items[?(@.status.conditions)]}{.metadata.name}{"\t"}{range .status.conditions[*]}{.type}={.reason}{" "}{end}{"\n"}{end}' | grep -E 'Creating|Updating|Deleting'
```
**Why**: `OperationTrackerStore` is in-memory. Pod restart (to deploy native controller) loses all tracked operations → potential double-create/delete for resources mid-operation.

#### Job 5: Rewrite Conversion Spokes
Replace `ujconversion.RoundTrip(dstRaw.(resource.Terraformed), tr)` in `zz_generated.conversion_spokes.go` with a non-upjet conversion path (plain JSON round-trip or hand-written hub-spoke converter).

**Why**: The `.(resource.Terraformed)` type assertion panics when the hub type is native. This blocks ALL API access to the resource in the old version.

### Pre-Cutover Checklist
- [ ] All resources in service pass e2e tests as RAW (both scopes)
- [ ] Agent verification report shows parity pass
- [ ] No other service is mid-cutover (one at a time)
- [ ] **Migration Job 1**: Storage version migrated (or native serves all TF versions)
- [ ] **Migration Job 2**: All CRs patched with `providerConfigRef.kind`
- [ ] **Migration Job 3**: `deletionPolicy: Orphan` mapped to `managementPolicies`
- [ ] **Migration Job 4**: No async operations in flight
- [ ] **Migration Job 5**: Conversion spokes rewritten
- [ ] Native types registered under SAME GVKs as TF types (including v1beta2)
- [ ] `status.atProvider.tagsAll` populated by native controller (or field preserved)

### Cutover Steps

1. **Run pre-migration jobs** (1-5 above)
2. **Move native types**: From `native/` sub-package to parent package
3. **Rename types**: `BucketRAW` → `Bucket` (json tags stay identical)
4. **Register all API versions**: Native type must serve v1beta1 AND v1beta2 (matching TF)
5. **Register conversion webhooks**: Using non-upjet conversion path (from Job 5)
6. **Register in scheme**: Under same GVK so cross-resource resolvers (`GetManagedResource("v1beta2")`) still work
7. **Replace controller**: Point provider binary at native controller
8. **Handle `default_tags`**: Native controller reads ProviderConfig's `default_tags` and merges with resource tags
9. **Remove TF code**: Delete `zz_*` files, TF config for this service
10. **Run full e2e**: With original kind names
11. **Commit**: `refactor: migrate <service> from terraform to native SDK`

### Post-Cutover Verification
- E2e tests pass with original kind names
- No upjet imports in the service's native code
- Cross-service references still resolve (resolvers find types in scheme under same GVK)
- `make generate` doesn't clobber native types
- Existing CRs (created by TF controller) are reconciled correctly by native controller
- Tags (including `default_tags`) are not spuriously modified
- Resources with `deletionPolicy: Orphan` (now `managementPolicies: [Observe]`) are not deleted

---

## Final Phase: TF Stack Removal

After ALL services are cut over:

1. **Remove upjet dependency**: `go.mod` — remove `github.com/crossplane/upjet/v2`
2. **Remove TF provider**: `go.mod` — remove `github.com/hashicorp/terraform-provider-aws`
3. **Remove TF SDK**: Remove `github.com/hashicorp/terraform-plugin-sdk`
4. **Remove code gen**: Delete `cmd/generator/main.go` and upjet pipeline
5. **Remove TF clients**: Rewrite `internal/clients/aws.go` — remove `SelectTerraformSetup`, `configureNoForkAWSClient`
6. **Remove config layer**: `config/` directory (resource configs, registry, externalname.go)
7. **Clean vendor**: `go mod tidy && go mod vendor`
8. **Full regression**: Run e2e tests for representative sample across all tiers
9. **Commit**: `refactor: remove terraform stack — fully native SDK implementation`

---

## Estimated Scale

| Phase | Per service (avg) | Total (both scopes) |
|-------|-------------------|---------------------|
| Phase 0 Infrastructure | — | ~25-30 tickets |
| Baseline E2E (TF) | ~3.5 | ~349 |
| Scaffold | 2 (cluster + ns) | ~200 |
| Implement CRUD | ~7 (3.5 × 2 scopes) | ~698 |
| E2E Test RAW | ~3.5 | ~349 |
| Agent Verify | 1 | ~100 |
| Cutover | 1 | ~100 |
| Final TF Removal | — | ~10 |
| **Grand Total** | | **~1,840 tickets** |

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| RAW types drift from TF types | Agent verification catches behavioral differences |
| Cross-service references break during cutover | CRD schema identical; IAM/KMS migrated early |
| AWS SDK v2 behavior differs from TF wrapper | Agent verification compares actual AWS state |
| Executor generates incorrect CRUD | E2E tests catch failures before verification |
| Large services (EC2: 40) overwhelm executor | One resource per ticket; proper dependencies |
| Credential resolution breaks | Integration test in Phase 0 validates path |
| Global resource region fails | Global maps updated for RAW API groups |
| `make generate` clobbers native files | Separate `native/` sub-package during parallel |
| Reference extractors use TerraformID() | Rewrite to native extractors in scaffold step |
| Async resources miss completion | Poll-based pattern tested in Phase 0 |
| CRD version mismatch at cutover | Multi-version + conversion webhooks documented per resource |
| LLM executor fails on complex resources | Rich ticket descriptions with API sequences |
| AWS cost during verification | Tiered verification; lifecycle-only for simple resources |
| Migrating a broken TF resource | Baseline e2e gate — must pass before any migration work |
| AWS permissions/quotas block testing | Hard fail, no workarounds — recorded for separate triage |
| Partial service migration (some resources fail baseline) | Per-resource gating; service cutover blocked until all pass |

---

## Appendix: Service Inventory

### Tier 1 — Single Resource (38 services, 38 resources)
apprunner, athena, batch, bedrockagent, budgets, cloudformation, cognitoidentity, cur, datasync, dax, dms, ds, dsql, ebs, ecrpublic, emrcontainers, firehose, identitystore, iot, kendra, kinesisanalytics, kinesisanalyticsv2, licensemanager, medialive, memorydb, mwaa, osis, qldb, ram, redshift, rolesanywhere, route53profiles, secretsmanager, servicediscovery, sfn, transfer, verifiedaccess, vpclattice

### Tier 2 — Small + Promoted Dependencies (28+ services, ~85 resources)
**Promoted for dependency clearance**: kms(5)

Standard: acm(2), apigateway(2), appstream(2), cloudsearch(2), cloudwatch(2), devicefarm(2), ecr(2), elb(2), fsx(2), globalaccelerator(2), kinesis(2), lakeformation(2), networkfirewall(2), opensearch(2), redshiftserverless(2), route53resolver(2), sns(2), acmpca(3), amp(3), autoscaling(3), cloudwatchevents(3), cloudwatchlogs(3), docdb(3), kafkaconnect(3), mq(3), organization(3), route53recoverycontrolconfig(3), sagemaker(3), wafv2(3)

### Tier 3 — Medium + Promoted Dependencies (~14 services, ~88 resources)
**Promoted for dependency clearance**: iam(12)

Standard: codeartifact(4), dynamodb(4), ecs(4), elasticache(4), grafana(4), neptune(4), networkmanager(4), opensearchserverless(4), sqs(4), ssoadmin(4), efs(5), elbv2(5), gamelift(5), kafka(5), route53(5)

### Tier 4 — Large (5+ services, ~143 resources)
cloudfront(6), servicecatalog(6), backup(7), cognitoidp(7), eks(7), glue(8), connect(9), lambda(9), bedrockagentcore(10), apigatewayv2(11), s3(11), directconnect(13), rds(15), ec2(40)
