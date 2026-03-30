# Native Controller Pattern — Implementation Guide

> **Primary reference** for every executor agent implementing a native (RAW) AWS controller.
> Read this **before** writing any code. Follow every pattern exactly — deviations cause
> reconcile loops, data loss, or compilation failures.

---

## 0. Quick Reference — Package Map

| Purpose | Import path |
|---------|-------------|
| Framework helpers (errors, tags, async, late-init, external name) | `github.com/upbound/provider-aws/v2/internal/native` |
| AWS credential resolution | `github.com/upbound/provider-aws/v2/internal/clients` |
| crossplane-runtime controller options | `github.com/crossplane/crossplane-runtime/v2/pkg/controller` |
| crossplane-runtime reconciler | `github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed` |
| crossplane-runtime resource types | `github.com/crossplane/crossplane-runtime/v2/pkg/resource` |
| crossplane-runtime external name | `github.com/crossplane/crossplane-runtime/v2/pkg/meta` |
| crossplane-runtime conditions | `github.com/crossplane/crossplane-runtime/v2/apis/common/v1` |
| AWS SDK core | `github.com/aws/aws-sdk-go-v2/aws` |
| AWS SDK smithy errors | `github.com/aws/smithy-go` |

---

## 1. File Layout

During the parallel migration phase (RAW type coexists with TF type):

```
apis/cluster/<service>/<version>/native/
  <resource>_raw_types.go          — cluster-scoped RAW CRD types

apis/namespaced/<service>/<version>/native/
  <resource>_raw_types.go          — namespaced RAW CRD types

internal/controller/cluster/<service>/<resource>raw/
  controller.go                    — cluster-scoped controller

internal/controller/namespaced/<service>/<resource>raw/
  controller.go                    — namespaced controller

examples/<service>/cluster/<version>/
  <resource>raw.yaml               — example CR for cluster scope

examples/<service>/namespaced/<version>/
  <resource>raw.yaml               — example CR for namespaced scope
```

> RAW types live in a `native/` sub-package (not the parent `v1beta1/` package) so that
> `make generate` (upjet/crossplane-tools) does NOT clobber hand-written code.
> At cutover, native types move up to the parent package and the `RAW` suffix is dropped.

---

## 2. Controller Setup Pattern

### 2.1 Setup function signature

Native controllers use **`xpcontroller.Options`** (crossplane-runtime), NOT `tjcontroller.Options`
(upjet). The `NativeSetupHook_<service>` variable in
`internal/controller/cluster/native_hook.go` has type `func(ctrl.Manager, xpcontroller.Options) error`.

```go
// internal/controller/cluster/<service>/<resource>raw/controller.go
package <resource>raw

import (
    xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
    "github.com/crossplane/crossplane-runtime/v2/pkg/event"
    "github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
    "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
    ctrl "sigs.k8s.io/controller-runtime"

    awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

    clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
    native "github.com/upbound/provider-aws/v2/internal/native"
    nativev1beta1 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"
)

// init registers the Setup function into the NativeSetupHook for this service.
// This runs automatically when the package is imported by the provider binary.
func init() {
    clustercontroller.NativeSetupHook_s3 = Setup
}

// Setup adds a controller that reconciles BucketRAW managed resources.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
    name := managed.ControllerName(nativev1beta1.BucketRAW_GroupKind)

    return ctrl.NewControllerManagedBy(mgr).
        Named(name).
        WithOptions(o.ForControllerRuntime()).
        For(&nativev1beta1.BucketRAW{}).
        Complete(managed.NewReconciler(mgr,
            resource.ManagedKind(nativev1beta1.BucketRAW_GroupVersionKind),
            managed.WithTypedExternalConnector[*nativev1beta1.BucketRAW](
                native.NewTypedConnector(
                    mgr.GetClient(),
                    func(cfg awss3.Options) *awss3.Client { // NOTE: use aws.Config overload for real services
                        return awss3.NewFromConfig(cfg)
                    },
                    func(client *awss3.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta1.BucketRAW] {
                        return &external{client: client, kube: kube}
                    },
                ),
            ),
            managed.WithLogger(o.Logger.WithValues("controller", name)),
            managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
            managed.WithPollInterval(o.PollInterval),
        ))
}
```

### 2.2 `NewTypedConnector` — correct call signature

```go
// TypedConnector is generic over the CR type T and the AWS SDK client type C.
// clientFactory receives an aws.Config and returns the service SDK client.
// newExternal receives the SDK client and kube.Client, returns TypedExternalClient[T].
connector := native.NewTypedConnector[*MyResourceRAW, *awssvc.Client](
    mgr.GetClient(),
    awssvc.NewFromConfig,                           // ConnectorFn: aws.Config → *svc.Client
    func(c *awssvc.Client, kube client.Client) managed.TypedExternalClient[*MyResourceRAW] {
        return &external{client: c, kube: kube}
    },
)
```

> `native.NewTypedConnector` calls `clients.GetAWSConfigWithTracking(ctx, kube, mg)` internally.
> This requires the CR to have `spec.forProvider.region` and a valid `providerConfigRef`.

---

## 3. CRD Type Pattern

### 3.1 Cluster-scoped type (embeds `v1.ResourceSpec`)

```go
// apis/cluster/<service>/<version>/native/<resource>_raw_types.go
package native

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"

    xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

const (
    BucketRAW_Kind              = "BucketRAW"
    BucketRAW_GroupKind         = "s3.aws.upbound.io/BucketRAW"   // built at init time
    // Populated from CRDGroup/CRDGroupVersion constants in the package
)

// BucketRAWSpec defines the desired state.
type BucketRAWSpec struct {
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Enum=us-east-1;us-west-2;...
    Region string `json:"region"`

    ForProvider BucketRAWParameters `json:"forProvider"`

    // Embedded ResourceSpec gives providerConfigRef, managementPolicies, etc.
    xpv1.ResourceSpec `json:",inline"`
}

type BucketRAWParameters struct {
    // Tags to apply to the bucket.
    // +optional
    Tags map[string]*string `json:"tags,omitempty"`

    // ... other spec fields ...
}

type BucketRAWObservation struct {
    // ARN of the bucket.
    // +optional
    ARN *string `json:"arn,omitempty"`

    // ... other observed (read-only) fields ...
}

// BucketRAWStatus is the observed state.
type BucketRAWStatus struct {
    xpv1.ConditionedStatus `json:",inline"`
    AtProvider BucketRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type BucketRAW struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   BucketRAWSpec   `json:"spec"`
    Status BucketRAWStatus `json:"status,omitempty"`
}
```

**Cluster types embed `xpv1.ResourceSpec` (v1)** — this makes them implement `resource.LegacyManaged`.

### 3.2 Namespaced type (embeds `v2.ManagedResourceSpec`)

```go
import (
    xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
)

type BucketRAWSpec struct {
    Region      string              `json:"region"`
    ForProvider BucketRAWParameters `json:"forProvider"`

    // ManagedResourceSpec gives providerConfigRef (with Kind field), etc.
    xpv2.ManagedResourceSpec `json:",inline"`
}

// +kubebuilder:resource:scope=Namespaced
type BucketRAW struct { ... }
```

**Namespaced types embed `xpv2.ManagedResourceSpec` (v2)** — this makes them implement
`resource.ModernManaged`.

### 3.3 Field placement rules

Always check **`.agents/specs/move-to-status-catalog.json`** before placing fields:

- Fields listed in the catalog for a resource → **`BucketRAWObservation`** (status.atProvider) only,
  NOT in `BucketRAWParameters` (spec.forProvider)
- Fields NOT in the catalog → spec.forProvider (writable by user)
- `spec.forProvider.region` is always required (mandatory for credential resolution)

### 3.4 External name in the type

Always include `// +crossplane:generate:reference` annotations for cross-resource references.
Use `resource.ExtractResourceID()` — **never** `resource.TerraformID()` (requires upjet interface).

```go
// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
// +crossplane:generate:reference:extractor=github.com/crossplane/crossplane-runtime/v2/pkg/resource.ExtractResourceID()
RoleARN *string `json:"roleArn,omitempty"`

// +optional
RoleARNRef *xpv1.Reference `json:"roleArnRef,omitempty"`

// +optional
RoleARNSelector *xpv1.Selector `json:"roleArnSelector,omitempty"`
```

---

## 4. Connector Pattern

The `native.TypedConnector` handles all credential resolution. Your controller only needs to
provide a `clientFactory` and a `newExternal` factory.

```go
// internal/native/connector.go — TypedConnector[T, C]
//
// func (c *TypedConnector[T, C]) Connect(ctx context.Context, mg T) (managed.TypedExternalClient[T], error) {
//     cfg, err := clients.GetAWSConfigWithTracking(ctx, c.kube, mg)
//     ...
//     svcClient := c.clientFactory(*cfg)
//     return c.newExternal(svcClient, c.kube), nil
// }
```

### 4.1 AWS credential requirements

`clients.GetAWSConfigWithTracking` requires:
1. The managed resource implements `resource.Managed` (both `LegacyManaged` and `ModernManaged` qualify)
2. The CR has `spec.forProvider.region` at the JSON path `spec.forProvider.region`
3. The CR has a valid `providerConfigRef`

For **cluster-scoped** resources: the ProviderConfig is cluster-scoped → resolved via `LegacyManaged`.
For **namespaced** resources: the ProviderConfig lookup uses `ModernManaged` path (supports
`ClusterProviderConfig` kind via `providerConfigRef.kind`).

---

## 5. CRUD Method Templates

The `external` struct implements `managed.TypedExternalClient[T]` with four methods:

```go
type external struct {
    client *awssvc.Client
    kube   client.Client
}

func (e *external) Observe(ctx context.Context, cr *MyResourceRAW) (managed.ExternalObservation, error) { ... }
func (e *external) Create(ctx context.Context, cr *MyResourceRAW) (managed.ExternalCreation, error) { ... }
func (e *external) Update(ctx context.Context, cr *MyResourceRAW) (managed.ExternalUpdate, error) { ... }
func (e *external) Delete(ctx context.Context, cr *MyResourceRAW) (managed.ExternalDelete, error) { ... }
func (e *external) Disconnect(_ context.Context) error { return nil }
```

### 5.1 Observe

```go
func (e *external) Observe(ctx context.Context, cr *MyResourceRAW) (managed.ExternalObservation, error) {
    // Step 1: Check for in-flight async operation (UseAsync resources only)
    if native.IsAsyncInProgress(cr) {
        asyncState, err := native.GetAsyncOperation(cr)
        if err != nil {
            return managed.ExternalObservation{}, native.Wrap(err, errGetAsyncState)
        }
        // Poll AWS for completion
        completed, err := e.pollAsyncOperation(ctx, cr, asyncState)
        if err != nil {
            return managed.ExternalObservation{}, err
        }
        if !completed {
            // Still in progress — tell reconciler "up to date" so it re-polls at PollInterval
            return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
        }
        native.ClearAsyncOperation(cr)
        // fall through to normal observe
    }

    // Step 2: Describe the resource
    externalName := native.GetExternalName(cr)
    if externalName == "" {
        // Not yet created
        return managed.ExternalObservation{ResourceExists: false}, nil
    }

    resp, err := e.client.DescribeMyResource(ctx, &awssvc.DescribeMyResourceInput{
        ResourceId: aws.String(externalName),
    })
    if err != nil {
        if native.IsNotFound(err) {
            return managed.ExternalObservation{ResourceExists: false}, nil
        }
        return managed.ExternalObservation{}, native.Wrap(err, errDescribe)
    }

    // Step 3: Map observed state to CR status
    cr.Status.AtProvider.ARN = resp.Resource.Arn
    cr.Status.AtProvider.State = aws.ToString(resp.Resource.State)

    // Step 4: Late initialization — populate nil spec fields from AWS response
    lateInitialized := false
    lateInitialized = native.LateInitializeStringPtr(&cr.Spec.ForProvider.Description, resp.Resource.Description) || lateInitialized

    // Step 5: Set condition
    if aws.ToString(resp.Resource.State) == "ACTIVE" {
        cr.Status.SetConditions(xpv1.Available())
    } else {
        cr.Status.SetConditions(xpv1.Unavailable())
    }

    // Step 6: Check for drift
    upToDate := isUpToDate(cr, resp)
    if !upToDate {
        cr.Status.SetConditions(xpv1.ReconcileSuccess())
    }

    return managed.ExternalObservation{
        ResourceExists:          true,
        ResourceUpToDate:        upToDate,
        ResourceLateInitialized: lateInitialized,
    }, nil
}
```

**Important Observe rules:**
- `ResourceExists: false` → reconciler calls `Create`
- `ResourceExists: true, ResourceUpToDate: false` → reconciler calls `Update`
- `ResourceExists: true, ResourceUpToDate: true` → nothing to do (re-polls at PollInterval)
- `ResourceLateInitialized: true` → reconciler writes the CR spec back to the API server
- On `NotFound` → return `ResourceExists: false`, **no error**
- On any other AWS error → return the error (reconciler retries)

### 5.2 Create

```go
func (e *external) Create(ctx context.Context, cr *MyResourceRAW) (managed.ExternalCreation, error) {
    cr.Status.SetConditions(xpv1.Creating())

    input := &awssvc.CreateMyResourceInput{
        ResourceName: aws.String(cr.Name), // or use native.GetExternalName(cr) if pre-set
        Description:  cr.Spec.ForProvider.Description,
        Tags:         buildTags(ctx, e.kube, cr),
    }

    resp, err := e.client.CreateMyResource(ctx, input)
    if err != nil {
        return managed.ExternalCreation{}, native.Wrap(err, errCreate)
    }

    // Set the external name from the provider-assigned ID
    native.SetExternalName(cr, aws.ToString(resp.Resource.ResourceId))

    // For async resources: record the in-flight operation
    // if native.UseAsync(cr) { // check catalog UseAsync flag
    //     return managed.ExternalCreation{}, native.SetAsyncOperation(cr, native.AsyncState{
    //         Operation: "creating",
    //         StartedAt: time.Now(),
    //         RequestID: aws.ToString(resp.ResponseMetadata.RequestID),
    //     })
    // }

    // Publish connection details (see connection-details-catalog.json)
    conn := managed.ConnectionDetails{
        "endpoint": []byte(aws.ToString(resp.Resource.Endpoint)),
    }
    return managed.ExternalCreation{ConnectionDetails: conn}, nil
}
```

**Important Create rules:**
- Set `cr.Status.SetConditions(xpv1.Creating())` at the start
- **Always** call `native.SetExternalName(cr, id)` after a successful create
- For async resources: call `native.SetAsyncOperation`, return immediately (Observe will poll)
- Return `ConnectionDetails` for any sensitive outputs (see connection-details-catalog.json)

### 5.3 Update

```go
func (e *external) Update(ctx context.Context, cr *MyResourceRAW) (managed.ExternalUpdate, error) {
    externalName := native.GetExternalName(cr)

    // Build the update input — map spec fields to AWS SDK input
    _, err := e.client.UpdateMyResource(ctx, &awssvc.UpdateMyResourceInput{
        ResourceId:  aws.String(externalName),
        Description: cr.Spec.ForProvider.Description,
    })
    if err != nil {
        return managed.ExternalUpdate{}, native.Wrap(err, errUpdate)
    }

    // Tags are usually a separate API call
    if err := e.reconcileTags(ctx, cr); err != nil {
        return managed.ExternalUpdate{}, err
    }

    return managed.ExternalUpdate{}, nil
}

// reconcileTags handles tag add/update/remove as a separate AWS API call.
func (e *external) reconcileTags(ctx context.Context, cr *MyResourceRAW) error {
    externalName := native.GetExternalName(cr)

    // Fetch current tags from AWS
    tagsResp, err := e.client.ListTagsForResource(ctx, &awssvc.ListTagsForResourceInput{
        ResourceArn: aws.String(aws.ToString(cr.Status.AtProvider.ARN)),
    })
    if err != nil {
        return native.Wrap(err, errListTags)
    }

    // Get provider default_tags from ProviderConfig
    defaultTags, err := native.GetProviderDefaultTags(ctx, e.kube, cr)
    if err != nil {
        return native.Wrap(err, errGetDefaultTags)
    }

    add, update, remove := native.DiffTagsWithDefaults(
        cr.Spec.ForProvider.Tags,  // resource-level desired tags (map[string]*string)
        defaultTags,               // ProviderConfig default_tags
        convertAWSTags(tagsResp.Tags), // actual tags on AWS resource
    )

    if len(add)+len(update) > 0 {
        toTag := mergeMaps(add, update)
        if _, err := e.client.TagResource(ctx, &awssvc.TagResourceInput{
            ResourceArn: aws.String(aws.ToString(cr.Status.AtProvider.ARN)),
            Tags:        toTagSlice(toTag),
        }); err != nil {
            return native.Wrap(err, errTagResource)
        }
    }
    if len(remove) > 0 {
        if _, err := e.client.UntagResource(ctx, &awssvc.UntagResourceInput{
            ResourceArn: aws.String(aws.ToString(cr.Status.AtProvider.ARN)),
            TagKeys:     remove,
        }); err != nil {
            return native.Wrap(err, errUntagResource)
        }
    }
    return nil
}
```

**Important Update rules:**
- Tags are usually a **separate** API call — do NOT include them in the main UpdateMyResource input
- Use `native.DiffTagsWithDefaults` (not `native.DiffTags`) to correctly handle `default_tags`
- Never remove tags that belong to `defaultTags` — the function handles this automatically

### 5.4 Delete

```go
func (e *external) Delete(ctx context.Context, cr *MyResourceRAW) (managed.ExternalDelete, error) {
    cr.Status.SetConditions(xpv1.Deleting())

    externalName := native.GetExternalName(cr)

    _, err := e.client.DeleteMyResource(ctx, &awssvc.DeleteMyResourceInput{
        ResourceId: aws.String(externalName),
    })
    if err != nil {
        if native.IsNotFound(err) {
            // Already gone — not an error
            return managed.ExternalDelete{}, nil
        }
        return managed.ExternalDelete{}, native.Wrap(err, errDelete)
    }

    // For async deletes: record in-flight operation
    // native.SetAsyncOperation(cr, native.AsyncState{Operation: "deleting", StartedAt: time.Now()})

    return managed.ExternalDelete{}, nil
}
```

**Important Delete rules:**
- Set `cr.Status.SetConditions(xpv1.Deleting())` at the start
- `NotFound` during delete → return `nil` error (idempotent)
- For async deletes (UseAsync=true in TF catalog): set `AsyncState{Operation: "deleting"}`

### 5.5 Disconnect

Always implement `Disconnect` as a no-op:

```go
func (e *external) Disconnect(_ context.Context) error { return nil }
```

---

## 6. Error Handling

All error functions are in `internal/native/errors.go`:

```go
// Test if an AWS API error is a "not found" condition.
// Covers: NotFound, ResourceNotFoundException, NoSuchKey, NoSuchBucket
native.IsNotFound(err) bool

// Test if an AWS API error is an access-denied / permission error.
// Covers: AccessDenied, UnauthorizedException, Forbidden
native.IsAccessDenied(err) bool

// Wrap an error with a context message (returns nil if err is nil)
native.Wrap(err, "cannot describe bucket")

// Wrap with a format string
native.Wrapf(err, "cannot describe bucket %q", name)
```

**Rules:**
- In `Observe`: `IsNotFound` → return `ResourceExists: false`, no error
- In `Delete`: `IsNotFound` → return `nil` error (idempotent)
- `IsAccessDenied` → return the wrapped error (reconciler will retry; do not swallow)
- Always wrap errors with context using `native.Wrap` / `native.Wrapf`
- Define per-controller error constants at the top of the file:

```go
const (
    errDescribe      = "cannot describe MyResource"
    errCreate        = "cannot create MyResource"
    errUpdate        = "cannot update MyResource"
    errDelete        = "cannot delete MyResource"
    errListTags      = "cannot list tags for MyResource"
    errTagResource   = "cannot tag MyResource"
    errUntagResource = "cannot untag MyResource"
)
```

---

## 7. External Name

All external name helpers are in `internal/native/externalname.go` (wrappers around
`github.com/crossplane/crossplane-runtime/v2/pkg/meta`):

```go
// Read the crossplane.io/external-name annotation (returns "" if not set)
externalName := native.GetExternalName(cr)

// Set the crossplane.io/external-name annotation (call after successful Create)
native.SetExternalName(cr, providerAssignedID)
```

**How to determine the external name strategy — check `.agents/specs/external-name-catalog.json`:**

| Strategy | `external-name-catalog.json` field | How to implement |
|----------|-------------------------------------|-----------------|
| `name_as_identifier` | `strategy: "name_as_identifier"` | Use `cr.Name` as the resource name; AWS stores it; set external name = `cr.Name` on Create |
| `parameter_as_identifier` | `strategy: "parameter_as_identifier"`, `parameter: "analyzer_name"` | Use `cr.Spec.ForProvider.<parameter>` as the name; set external name = that value on Create |
| `identifier_from_provider` | `strategy: "identifier_from_provider"` | AWS assigns the ID; read `resp.Resource.Id`; call `native.SetExternalName(cr, id)` in Create |
| `templated` | `strategy: "templated"`, `template: "{{ .parameters.X }}/{{ .external_name }}"` | Composite ID; encode in Create, decode in Observe |
| `custom` | `strategy: "custom"`, `description: "..."` | Read the description field; implement accordingly |

---

## 8. Tag Management

All tag helpers are in `internal/native/tags.go`.

### 8.1 Simple tag diff (no ProviderConfig default_tags)

```go
// DiffTags returns { ToAdd, ToUpdate, ToRemove }
diff := native.DiffTags(desired, actual) // both are map[string]string
if len(diff.ToAdd)+len(diff.ToUpdate) > 0 { ... }
if len(diff.ToRemove) > 0 { ... }
```

### 8.2 Tag diff with ProviderConfig default_tags (recommended for all resources)

```go
// GetProviderDefaultTags reads ProviderConfig.spec.defaultTags for LegacyManaged resources.
// Returns nil (no error) for ModernManaged resources (not yet supported).
defaultTags, err := native.GetProviderDefaultTags(ctx, kube, cr)

// DiffTagsWithDefaults: returns add, update (map[string]*string) and remove ([]string).
// Tags in defaultTags are NEVER added to the remove list.
add, update, remove := native.DiffTagsWithDefaults(
    cr.Spec.ForProvider.Tags,  // map[string]*string
    defaultTags,               // map[string]*string (may be nil)
    actualTags,                // map[string]*string (from AWS)
)
```

### 8.3 Equality check

```go
// AreSame returns true if the two maps are semantically identical.
if native.AreSame(desired, actual) { ... }
```

---

## 9. Late Initialization

All late-init helpers are in `internal/native/lateinit.go`.

Call late initialization in `Observe` **after** successfully describing the resource.
Return `ResourceLateInitialized: true` if any field was populated.

```go
// Each helper returns true if the field was populated (dst was nil and src was non-nil).
lateInit := false
lateInit = native.LateInitializeStringPtr(&cr.Spec.ForProvider.Description, resp.Description) || lateInit
lateInit = native.LateInitializeBoolPtr(&cr.Spec.ForProvider.EnableDNS, resp.EnableDns) || lateInit
lateInit = native.LateInitializeInt64Ptr(&cr.Spec.ForProvider.Timeout, resp.TimeoutSeconds) || lateInit
lateInit = native.LateInitializeMapStringPtr(&cr.Spec.ForProvider.Tags, convertTags(resp.Tags)) || lateInit

return managed.ExternalObservation{
    ...
    ResourceLateInitialized: lateInit,
}, nil
```

### 9.1 Ignored fields

Use `native.LateInitConfig` and `native.IsIgnored` to respect per-resource ignore lists:

```go
config := native.LateInitConfig{
    IgnoredFields: []string{"subnet_id", "network_interface"},
}
if !native.IsIgnored("subnet_id", config) {
    native.LateInitializeStringPtr(&cr.Spec.ForProvider.SubnetID, resp.SubnetId)
}
```

Ignored field lists come from the TF resource's `LateInitializer.IgnoredFields` configuration.

---

## 10. Async Operations

Use async when the TF business-logic catalog marks the resource with `UseAsync = true`.
All helpers are in `internal/native/async.go`.

```go
const AnnotationKeyAsyncOperation = "native.aws.upbound.io/async-operation"

// AsyncState persisted as JSON in the CR annotation.
type AsyncState struct {
    Operation string    `json:"operation"` // "creating", "updating", "deleting"
    StartedAt time.Time `json:"startedAt"`
    RequestID string    `json:"requestId,omitempty"`
}
```

### 10.1 Setting async state (in Create/Update/Delete)

```go
if err := native.SetAsyncOperation(cr, native.AsyncState{
    Operation: "creating",
    StartedAt: time.Now(),
    RequestID: aws.ToString(resp.RequestMetadata.RequestId),
}); err != nil {
    return managed.ExternalCreation{}, err
}
```

### 10.2 Polling in Observe

```go
if native.IsAsyncInProgress(cr) {
    state, err := native.GetAsyncOperation(cr)
    if err != nil {
        return managed.ExternalObservation{}, native.Wrap(err, errGetAsyncState)
    }

    // Describe the resource to check if the operation completed
    resp, err := e.client.DescribeMyResource(ctx, &awssvc.DescribeMyResourceInput{
        ResourceId: aws.String(native.GetExternalName(cr)),
    })
    if err != nil {
        if native.IsNotFound(err) && state.Operation == "deleting" {
            native.ClearAsyncOperation(cr)
            return managed.ExternalObservation{ResourceExists: false}, nil
        }
        return managed.ExternalObservation{}, native.Wrap(err, errDescribe)
    }

    switch aws.ToString(resp.Resource.Status) {
    case "ACTIVE":
        native.ClearAsyncOperation(cr)
        // fall through to normal observe
    case "CREATING", "UPDATING", "DELETING":
        // Still running — do not call Update; let reconciler re-poll
        return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
    case "FAILED":
        native.ClearAsyncOperation(cr)
        return managed.ExternalObservation{}, fmt.Errorf("async %s failed for %s", state.Operation, native.GetExternalName(cr))
    }
}
```

### 10.3 Clearing async state

```go
native.ClearAsyncOperation(cr) // removes the annotation
```

---

## 11. Policy / IAM Semantic Comparison

For resources with IAM policy fields (e.g., `spec.forProvider.policy`):

```go
import "github.com/upbound/provider-aws/v2/internal/native"

// PoliciesAreEquivalent compares two IAM policy JSON strings semantically.
// Returns true even if key order, whitespace, or default fields differ.
equal, err := native.PoliciesAreEquivalent(cr.Spec.ForProvider.Policy, aws.ToString(resp.Policy))
if err != nil {
    // JSON parse error — treat as not-equal (triggers update)
    equal = false
}
```

**Always use `native.PoliciesAreEquivalent` instead of string equality for policy fields.**
Naive string comparison causes infinite reconcile loops because AWS normalizes JSON.

See `.agents/specs/tf-business-logic-catalog.md` Pattern A for affected resources.

---

## 12. Reference Resolution

RAW types use `+crossplane:generate:reference` annotations like TF types.
Run `make generate` after adding annotations to produce the `ResolveReferences()` method.

### 12.1 Correct extractor

```go
// CORRECT — works with RAW types
// +crossplane:generate:reference:extractor=github.com/crossplane/crossplane-runtime/v2/pkg/resource.ExtractResourceID()

// WRONG — requires upjet Terraformed interface, will panic
// +crossplane:generate:reference:extractor=github.com/crossplane/upjet/v2/pkg/resource.TerraformID()
```

### 12.2 Full annotation block example

```go
// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
// +crossplane:generate:reference:extractor=github.com/crossplane/crossplane-runtime/v2/pkg/resource.ExtractResourceID()
KMSKeyID *string `json:"kmsKeyId,omitempty"`

// +optional
KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIdRef,omitempty"`

// +optional
KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIdSelector,omitempty"`
```

After running `make generate`, a `zz_<resource>_terraformed.go`-equivalent file is created in
the native sub-package with `ResolveReferences()`. This must NOT be in the `native/` package
(resolver excludes it) — see `.agents/specs/terraform-removal-migration.md` section 0.15.

---

## 13. Scheme Registration

New native types must be registered in the scheme. Create a `native_register.go` file
(no `zz_` prefix — survives `make generate`) in the apis package:

```go
// apis/cluster/native_register.go  (or per-service: apis/cluster/s3native_register.go)
package cluster

import natives3 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"

func init() {
    AddToSchemes = append(AddToSchemes, natives3.SchemeBuilder.AddToScheme)
}
```

---

## 14. NativeSetupHook Registration

The `NativeSetupHook_<service>` variable is declared in
`internal/controller/cluster/native_hook.go` (cluster) and
`internal/controller/namespaced/native_hook.go` (namespaced).

Register by setting the variable from an `init()` function in the controller package:

```go
// internal/controller/cluster/<service>/<resource>raw/controller.go

package <resource>raw  // e.g., bucketraw

import (
    clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
    xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
    ctrl "sigs.k8s.io/controller-runtime"
)

func init() {
    // NOTE: if multiple RAW resources in one service all set this hook,
    // they must cooperate: last init() wins. Instead, use a single
    // setup.go per service that calls each resource's individual Setup().
    clustercontroller.NativeSetupHook_s3 = Setup
}

func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
    return SetupBucketRAW(mgr, o)
}
```

For services with **multiple** RAW resources, create a single `setup.go`:

```go
// internal/controller/cluster/<service>/setup.go
package <service>  // e.g., package s3

import (
    clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
    bucketraw "github.com/upbound/provider-aws/v2/internal/controller/cluster/s3/bucketraw"
    lifecycleraw "github.com/upbound/provider-aws/v2/internal/controller/cluster/s3/lifecycleraw"
    xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
    ctrl "sigs.k8s.io/controller-runtime"
)

func init() {
    clustercontroller.NativeSetupHook_s3 = Setup
}

func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
    for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
        bucketraw.Setup,
        lifecycleraw.Setup,
    } {
        if err := setup(mgr, o); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 15. Complete Example — Simplified S3 BucketRAW

> **Note**: This is a simplified teaching example. A production S3 controller
> requires additional logic (versioning, lifecycle, CORS, etc.) and must consult
> `.agents/specs/move-to-status-catalog.json` (S3 has 14 moved fields).

### 15.1 Types: `apis/cluster/s3/v1beta1/native/bucket_raw_types.go`

```go
//go:build s3 || all

// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"

    xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

const (
    BucketRAW_Kind             = "BucketRAW"
    BucketRAW_GroupVersionKind = "s3.aws.upbound.io/v1beta2/BucketRAW"
)

var (
    BucketRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: BucketRAW_Kind}.String()
    BucketRAW_GroupVersionKind_ = CRDGroupVersion.WithKind(BucketRAW_Kind)
)

// BucketRAWParameters are the configurable fields.
type BucketRAWParameters struct {
    // Region where the bucket is created. Required.
    // +kubebuilder:validation:Required
    Region string `json:"region"`

    // Tags to apply to the bucket.
    // +optional
    Tags map[string]*string `json:"tags,omitempty"`

    // ForceDestroy allows deletion of non-empty buckets.
    // +optional
    // +kubebuilder:default=false
    ForceDestroy *bool `json:"forceDestroy,omitempty"`
}

// BucketRAWObservation are the observed fields.
// NOTE: per move-to-status-catalog.json, S3 bucket has 14 fields that are
// status-only (acceleration_status, acl, grant, cors_rule, etc.).
type BucketRAWObservation struct {
    // ARN of the bucket.
    ARN *string `json:"arn,omitempty"`

    // Region where the bucket was created.
    BucketRegion *string `json:"bucketRegion,omitempty"`
}

// BucketRAWSpec defines desired state.
type BucketRAWSpec struct {
    xpv1.ResourceSpec `json:",inline"`
    ForProvider       BucketRAWParameters `json:"forProvider"`
}

// BucketRAWStatus defines observed state.
type BucketRAWStatus struct {
    xpv1.ConditionedStatus `json:",inline"`
    AtProvider             BucketRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories=crossplane
type BucketRAW struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   BucketRAWSpec   `json:"spec"`
    Status BucketRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type BucketRAWList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []BucketRAW `json:"items"`
}
```

### 15.2 Controller: `internal/controller/cluster/s3/bucketraw/controller.go`

```go
//go:build s3 || all

// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package bucketraw

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/aws"
    awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
    xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
    xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
    "github.com/crossplane/crossplane-runtime/v2/pkg/event"
    "github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
    "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"

    clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
    native "github.com/upbound/provider-aws/v2/internal/native"
    nativev1beta2 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta2/native"
)

const (
    errDescribeBucket = "cannot describe S3 BucketRAW"
    errCreateBucket   = "cannot create S3 BucketRAW"
    errUpdateBucket   = "cannot update S3 BucketRAW"
    errDeleteBucket   = "cannot delete S3 BucketRAW"
    errListTags       = "cannot list tags for S3 BucketRAW"
    errTagBucket      = "cannot tag S3 BucketRAW"
    errUntagBucket    = "cannot untag S3 BucketRAW"
)

func init() {
    clustercontroller.NativeSetupHook_s3 = Setup
}

func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
    name := managed.ControllerName(nativev1beta2.BucketRAW_GroupKind)

    return ctrl.NewControllerManagedBy(mgr).
        Named(name).
        WithOptions(o.ForControllerRuntime()).
        For(&nativev1beta2.BucketRAW{}).
        Complete(managed.NewReconciler(mgr,
            resource.ManagedKind(nativev1beta2.BucketRAW_GroupVersionKind),
            managed.WithTypedExternalConnector[*nativev1beta2.BucketRAW](
                native.NewTypedConnector(
                    mgr.GetClient(),
                    awss3.NewFromConfig,
                    func(c *awss3.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta2.BucketRAW] {
                        return &external{client: c, kube: kube}
                    },
                ),
            ),
            managed.WithLogger(o.Logger.WithValues("controller", name)),
            managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
            managed.WithPollInterval(o.PollInterval),
        ))
}

type external struct {
    client *awss3.Client
    kube   client.Client
}

func (e *external) Observe(ctx context.Context, cr *nativev1beta2.BucketRAW) (managed.ExternalObservation, error) {
    bucketName := native.GetExternalName(cr)
    if bucketName == "" {
        return managed.ExternalObservation{ResourceExists: false}, nil
    }

    // HeadBucket is the idiomatic "exists?" check for S3
    _, err := e.client.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: aws.String(bucketName)})
    if err != nil {
        if native.IsNotFound(err) {
            return managed.ExternalObservation{ResourceExists: false}, nil
        }
        return managed.ExternalObservation{}, native.Wrap(err, errDescribeBucket)
    }

    // Get bucket location
    loc, err := e.client.GetBucketLocation(ctx, &awss3.GetBucketLocationInput{Bucket: aws.String(bucketName)})
    if err != nil {
        return managed.ExternalObservation{}, native.Wrap(err, errDescribeBucket)
    }

    region := string(loc.LocationConstraint)
    cr.Status.AtProvider.BucketRegion = aws.String(region)
    cr.Status.AtProvider.ARN = aws.String("arn:aws:s3:::" + bucketName)
    cr.Status.SetConditions(xpv1.Available())

    // Check tag drift
    tagsResp, err := e.client.GetBucketTagging(ctx, &awss3.GetBucketTaggingInput{Bucket: aws.String(bucketName)})
    upToDate := true
    if err == nil {
        defaultTags, _ := native.GetProviderDefaultTags(ctx, e.kube, cr)
        actual := s3TagsToMap(tagsResp.TagSet)
        add, upd, rem := native.DiffTagsWithDefaults(cr.Spec.ForProvider.Tags, defaultTags, actual)
        upToDate = len(add)+len(upd)+len(rem) == 0
    }

    return managed.ExternalObservation{
        ResourceExists:   true,
        ResourceUpToDate: upToDate,
    }, nil
}

func (e *external) Create(ctx context.Context, cr *nativev1beta2.BucketRAW) (managed.ExternalCreation, error) {
    cr.Status.SetConditions(xpv1.Creating())

    bucketName := cr.Name  // S3: name_as_identifier strategy (from external-name-catalog.json)

    input := &awss3.CreateBucketInput{Bucket: aws.String(bucketName)}
    // us-east-1 does NOT accept a LocationConstraint
    if cr.Spec.ForProvider.Region != "us-east-1" {
        input.CreateBucketConfiguration = &awss3types.CreateBucketConfiguration{
            LocationConstraint: awss3types.BucketLocationConstraint(cr.Spec.ForProvider.Region),
        }
    }

    if _, err := e.client.CreateBucket(ctx, input); err != nil {
        return managed.ExternalCreation{}, native.Wrap(err, errCreateBucket)
    }

    // S3 strategy: name_as_identifier → external name = bucket name
    native.SetExternalName(cr, bucketName)

    conn := managed.ConnectionDetails{
        "id":     []byte(bucketName),
        "arn":    []byte("arn:aws:s3:::" + bucketName),
        "region": []byte(cr.Spec.ForProvider.Region),
    }
    return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, cr *nativev1beta2.BucketRAW) (managed.ExternalUpdate, error) {
    // S3 bucket metadata is immutable; only tags are mutable
    if err := e.reconcileTags(ctx, cr); err != nil {
        return managed.ExternalUpdate{}, err
    }
    return managed.ExternalUpdate{}, nil
}

func (e *external) reconcileTags(ctx context.Context, cr *nativev1beta2.BucketRAW) error {
    bucketName := native.GetExternalName(cr)

    tagsResp, err := e.client.GetBucketTagging(ctx, &awss3.GetBucketTaggingInput{Bucket: aws.String(bucketName)})
    if err != nil && !native.IsNotFound(err) {
        return native.Wrap(err, errListTags)
    }
    var actual map[string]*string
    if err == nil {
        actual = s3TagsToMap(tagsResp.TagSet)
    }

    defaultTags, err := native.GetProviderDefaultTags(ctx, e.kube, cr)
    if err != nil {
        return native.Wrap(err, "cannot get provider default tags")
    }

    add, upd, rem := native.DiffTagsWithDefaults(cr.Spec.ForProvider.Tags, defaultTags, actual)

    desired := mergePtrMaps(add, upd)
    if len(desired) > 0 {
        _, err = e.client.PutBucketTagging(ctx, &awss3.PutBucketTaggingInput{
            Bucket:  aws.String(bucketName),
            Tagging: &awss3types.Tagging{TagSet: mapToS3Tags(desired)},
        })
        if err != nil {
            return native.Wrap(err, errTagBucket)
        }
    }
    if len(rem) > 0 {
        // For S3, delete then re-add is safer than selective removal
        _ = rem // handled by PutBucketTagging (replaces all tags)
    }
    return nil
}

func (e *external) Delete(ctx context.Context, cr *nativev1beta2.BucketRAW) (managed.ExternalDelete, error) {
    cr.Status.SetConditions(xpv1.Deleting())

    bucketName := native.GetExternalName(cr)
    _, err := e.client.DeleteBucket(ctx, &awss3.DeleteBucketInput{Bucket: aws.String(bucketName)})
    if err != nil {
        if native.IsNotFound(err) {
            return managed.ExternalDelete{}, nil
        }
        return managed.ExternalDelete{}, native.Wrap(err, errDeleteBucket)
    }
    return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(_ context.Context) error { return nil }

// s3TagsToMap converts S3 SDK Tag slice to map[string]*string.
func s3TagsToMap(tags []awss3types.Tag) map[string]*string {
    result := make(map[string]*string, len(tags))
    for _, t := range tags {
        v := aws.ToString(t.Value)
        result[aws.ToString(t.Key)] = &v
    }
    return result
}
```

---

## 16. Connection Details

Check **`.agents/specs/connection-details-catalog.json`** for the keys your resource must publish.

```go
// Example: for aws_db_instance, publish these keys in Create and Update:
conn := managed.ConnectionDetails{
    "endpoint": []byte(aws.ToString(resp.Endpoint.Address) + ":" + fmt.Sprint(aws.ToInt32(resp.Endpoint.Port))),
    "address":  []byte(aws.ToString(resp.Endpoint.Address)),
    "host":     []byte(aws.ToString(resp.Endpoint.Address)), // same field, two keys
    "port":     []byte(fmt.Sprint(aws.ToInt32(resp.Endpoint.Port))),
    "username": []byte(aws.ToString(resp.MasterUsername)),
    // "password" is sensitive: obtained from spec.forProvider.passwordSecretRef
}
```

For fields marked `sensitive:*` in the catalog: read from the SecretRef in the spec
(do NOT try to read back from AWS — most services do not return secrets).

---

## 17. Common Pitfalls

| Pitfall | Correct approach |
|---------|-----------------|
| Using `tjcontroller.Options` in Setup | Use `xpcontroller.Options` (crossplane-runtime) |
| Using `resource.TerraformID()` extractor | Use `resource.ExtractResourceID()` |
| Returning error on NotFound in Observe | Return `ResourceExists: false`, nil error |
| Returning error on NotFound in Delete | Return empty `ExternalDelete{}`, nil error |
| Comparing IAM policies with string equality | Use `native.PoliciesAreEquivalent()` |
| Removing default_tags from AWS resource | Use `native.DiffTagsWithDefaults()` |
| Placing MoveToStatus fields in spec | Check `move-to-status-catalog.json`; put in `Observation` struct |
| Initializing async polling before describe | Check `native.IsAsyncInProgress(cr)` first in Observe |
| Using `upjet/v2/pkg/resource.SetUpToDateCondition` | Avoid upjet imports; use crossplane-runtime conditions |
| Calling `SetExternalName` in Observe | Only call in Create (after successful provider response) |

---

## 18. Checklist Before Committing

- [ ] `go test ./internal/controller/cluster/<service>/<resource>raw/...` passes
- [ ] `go test ./apis/cluster/<service>/...` passes (generated code compiles)
- [ ] `golangci-lint run ./internal/controller/cluster/<service>/...` passes
- [ ] `NativeSetupHook_<service> = Setup` is set in an `init()` function
- [ ] All error strings are lowercase (Go convention)
- [ ] `Disconnect` returns `nil`
- [ ] Tags use `native.DiffTagsWithDefaults` (not bare `DiffTags`)
- [ ] IAM policy comparisons use `native.PoliciesAreEquivalent`
- [ ] MoveToStatus fields are in `Observation`, not `Parameters`
- [ ] External name strategy matches the external-name-catalog.json entry
- [ ] `TerraformID()` extractor NOT used anywhere in native types
- [ ] Build tag matches service name (e.g., `//go:build s3 || all`)

---

## 19. Catalog References

| Catalog | Location | Use for |
|---------|----------|---------|
| External name strategies | `.agents/specs/external-name-catalog.json` | Deciding how to set/get external name in Create/Observe |
| Fields moved to status | `.agents/specs/move-to-status-catalog.json` | Ensuring writable TF fields that became read-only are in `Observation` |
| Connection details | `.agents/specs/connection-details-catalog.json` | Keys to publish in `Create`/`Update` ConnectionDetails |
| TF business logic | `.agents/specs/tf-business-logic-catalog.md` | Replicating TF injector/custom-diff behavior in native Observe/Update |
