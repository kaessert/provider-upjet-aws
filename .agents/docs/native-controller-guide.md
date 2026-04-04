# Native Controller Guide

> **Audience**: Agents and developers building native (non-Terraform) controllers for provider-upjet-aws.
> **Status**: Reference architecture — derived from `crossplane/provider-template` and adapted for the AWS provider migration.

---

## Overview

This guide documents the pattern for building native AWS controllers that bypass the Terraform/Upjet layer entirely. Each native controller implements the Crossplane `ExternalClient` interface using the AWS SDK v2 directly, replacing the current TF-bridged approach.

### Why Native Controllers?

The current architecture routes every AWS API call through the Terraform Plugin SDK:

```
K8s CR → upjet AsyncConnector → TF PluginSDK (in-process) → AWS SDK v2 → AWS API
```

Native controllers eliminate two layers:

```
K8s CR → native Connector → TypedExternalClient → AWS SDK v2 → AWS API
```

**Benefits**: faster reconciliation, no TF state management, cleaner error handling, purpose-built types without `tf:"..."` struct tags, direct control over AWS SDK behavior.

---

## 1. Provider Template Architecture

The `crossplane/provider-template` repository is the canonical reference for native Crossplane controllers. It demonstrates every pattern used in the migration.

### 1.1 Key Interfaces

Native controllers are built around two generic interfaces from crossplane-runtime:

```go
// TypedExternalClient — implements CRUD for a single resource type.
// The reconciler calls these methods based on the result of Observe.
type TypedExternalClient[T resource.Managed] interface {
    Observe(ctx context.Context, cr T) (ExternalObservation, error)
    Create(ctx context.Context, cr T) (ExternalCreation, error)
    Update(ctx context.Context, cr T) (ExternalUpdate, error)
    Delete(ctx context.Context, cr T) (ExternalDelete, error)
    Disconnect(ctx context.Context) error
}

// TypedExternalConnector — creates TypedExternalClient instances.
// Called once per reconciliation to establish a connection.
type TypedExternalConnector[T resource.Managed] interface {
    Connect(ctx context.Context, cr T) (TypedExternalClient[T], error)
}
```

These are generics-based (Go 1.18+). The type parameter `T` is the concrete CR pointer type (e.g., `*BucketRAW`).

### 1.2 Reconciliation Flow

The managed reconciler drives a simple state machine:

```
Watch/Poll event triggers reconciliation
    ↓
managed.Reconciler fetches the CR from the API server
    ↓
connector.Connect(ctx, cr)
  → Resolves credentials from ProviderConfig
  → Creates AWS SDK client
  → Returns TypedExternalClient
    ↓
external.Observe(ctx, cr)
  → Calls AWS Describe/Get/Head API
  → Returns ExternalObservation{ResourceExists, ResourceUpToDate, ...}
    ↓
Decision:
  ResourceExists=false              → Create(ctx, cr)
  ResourceExists=true, UpToDate=false → Update(ctx, cr)
  ResourceExists=true, UpToDate=true  → Nothing (re-polls at PollInterval)
  CR has DeletionTimestamp            → Delete(ctx, cr)
    ↓
Reconciler updates status conditions, connection secret, external name
    ↓
PollInterval (default 1m) triggers next reconciliation
```

### 1.3 Credential Flow

```
CR.Spec.ProviderConfigReference (name + kind)
    ↓
Reconciler fetches ProviderConfig by name
    ↓
ProviderConfig.Spec.Credentials
    ↓
CommonCredentialExtractor resolves credentials
  (Secret | Environment | Filesystem | InjectedIdentity)
    ↓
aws.Config is constructed with resolved credentials + region
    ↓
newServiceFn(cfg) → AWS SDK service client (e.g., *s3.Client)
```

In `provider-template`, this is explicit in the connector's `Connect()` method. In provider-aws, the `native.NewTypedConnector` helper handles it via `clients.GetAWSConfigWithTracking`.

### 1.4 Controller Setup Pattern

From `provider-template/internal/controller/mytype/mytype.go`:

```go
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
    o.Gate.Register(func() {
        if err := Setup(mgr, o); err != nil {
            panic(errors.Wrap(err, "cannot setup MyType controller"))
        }
    }, v1alpha1.MyTypeGroupVersionKind)
    return nil
}

func Setup(mgr ctrl.Manager, o controller.Options) error {
    name := managed.ControllerName(v1alpha1.MyTypeGroupKind)

    r := managed.NewReconciler(mgr,
        resource.ManagedKind(v1alpha1.MyTypeGroupVersionKind),
        managed.WithTypedExternalConnector[*v1alpha1.MyType](&connector{...}),
        managed.WithLogger(o.Logger.WithValues("controller", name)),
        managed.WithPollInterval(o.PollInterval),
        managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
    )

    return ctrl.NewControllerManagedBy(mgr).
        Named(name).
        WithOptions(o.ForControllerRuntime()).
        WithEventFilter(resource.DesiredStateChanged()).
        For(&v1alpha1.MyType{}).
        Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}
```

**Critical**: Uses `controller.Options` from crossplane-runtime, **NOT** `tjcontroller.Options` from upjet. The `SetupGated` pattern ensures the controller only starts after its CRD is registered (safe startup).

### 1.5 Type Definition Pattern

From `provider-template/apis/sample/v1alpha1/mytype_types.go`:

```go
type MyTypeParameters struct {
    ConfigurableField string `json:"configurableField"`
}

type MyTypeObservation struct {
    ConfigurableField string `json:"configurableField"`
    ObservableField   string `json:"observableField,omitempty"`
}

type MyTypeSpec struct {
    xpv2.ManagedResourceSpec `json:",inline"`    // namespaced
    ForProvider              MyTypeParameters `json:"forProvider"`
}

type MyTypeStatus struct {
    xpv1.ResourceStatus `json:",inline"`
    AtProvider          MyTypeObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,template}
type MyType struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   MyTypeSpec   `json:"spec"`
    Status MyTypeStatus `json:"status,omitempty"`
}
```

**Key detail**: Namespaced types embed `xpv2.ManagedResourceSpec`, cluster-scoped types embed `xpv1.ResourceSpec`. This determines which `resource.Managed` interface variant the type satisfies.

### 1.6 Registration Pattern

Controllers are registered into a central `SetupGated` function:

```go
// internal/controller/register.go
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
    for _, setup := range []func(ctrl.Manager, controller.Options) error{
        config.Setup,
        mytype.SetupGated,
    } {
        if err := setup(mgr, o); err != nil {
            return err
        }
    }
    return nil
}
```

### 1.7 Key Dependencies

From `provider-template/go.mod`:

| Dependency | Version | Purpose |
|-----------|---------|---------|
| `crossplane-runtime/v2` | `v2.0.0` | Reconciler, types, controller framework |
| `controller-runtime` | `v0.21.0` | Kubernetes controller infrastructure |
| `kingpin/v2` | `v2.4.0` | CLI flags |

For AWS controllers, add:

| Dependency | Purpose |
|-----------|---------|
| `aws-sdk-go-v2` | AWS SDK core |
| `aws-sdk-go-v2/service/<svc>` | Per-service SDK client |
| `smithy-go` | AWS error types |

---

## 2. Migration Context

### 2.1 Current vs Target Architecture

**Current (TF-bridged)**:
```
K8s CR → upjet AsyncConnector → TF PluginSDK (in-process) → AWS SDK v2 → AWS API
Types: generated from TF schema, have tf:"..." struct tags
Controller options: tjcontroller.Options (upjet-specific)
```

**Target (Native/RAW)**:
```
K8s CR → native Connector → TypedExternalClient → AWS SDK v2 → AWS API
Types: independent, json tags only, same YAML schema
Controller options: controller.Options (crossplane-runtime standard)
```

### 2.2 Parallel Coexistence Strategy

During migration, both versions coexist:

- `Kind: Bucket` — TF-backed (existing)
- `Kind: BucketRAW` — SDK-native (new)

Both manage the same AWS resource type. After verification confirms parity, the TF kind is removed and `BucketRAW` is renamed to `Bucket`.

### 2.3 YAML Compatibility Constraint

RAW types **MUST** produce the same CRD schema as TF types:

| Property | Requirement |
|----------|-------------|
| Field names | Same `json:"..."` tags → identical YAML field names |
| Structure | Same `ForProvider`/`InitProvider`/`AtProvider` layout |
| Field types | Same Go types, same nesting depth |
| Struct tags | `json` only — NO `tf:"..."` tags |
| Validation | Same kubebuilder validation markers where applicable |

This ensures users can migrate CRs by changing only `kind: Bucket` → `kind: BucketRAW` (and eventually back when the RAW suffix is dropped).

---

## 3. File Layout

### 3.1 During Migration (RAW Suffix Phase)

```
apis/cluster/<service>/<version>/native/
  <resource>_raw_types.go           # Cluster-scoped RAW CRD types

apis/namespaced/<service>/<version>/native/
  <resource>_raw_types.go           # Namespaced RAW CRD types

internal/controller/cluster/<service>/<resource>raw/
  controller.go                     # Cluster-scoped native controller

internal/controller/namespaced/<service>/<resource>raw/
  controller.go                     # Namespaced native controller

examples/<service>/cluster/<version>/
  <resource>raw.yaml                # Example CR

examples/<service>/namespaced/<version>/
  <resource>raw.yaml                # Example CR
```

**Why `native/` subpackage?** RAW types live in `native/` so that `make generate` (upjet/crossplane-tools) does NOT clobber hand-written code. At cutover, native types move up to the parent package and the `RAW` suffix is dropped.

### 3.2 After Cutover

```
apis/cluster/<service>/<version>/
  <resource>_types.go               # Now the primary type (no RAW suffix)

internal/controller/cluster/<service>/<resource>/
  controller.go                     # Primary controller
```

---

## 4. Implementing a Native Controller

### 4.1 Step-by-Step Checklist

1. **Define RAW types** in `apis/cluster/<service>/<version>/native/`
2. **Implement controller** in `internal/controller/cluster/<service>/<resource>raw/`
3. **Register scheme** in `apis/cluster/native_register.go`
4. **Register controller** via `NativeSetupHook_<service>` in `init()`
5. **Write example CR** in `examples/<service>/cluster/<version>/`
6. **Repeat for namespaced scope** if needed
7. **Test**: `go test`, `golangci-lint`, manual verification

### 4.2 CRD Types

#### Cluster-Scoped Type

```go
// apis/cluster/<service>/<version>/native/<resource>_raw_types.go
package native

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"

    xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// Parameters — user-settable fields in spec.forProvider.
type BucketRAWParameters struct {
    // Region is required for credential resolution.
    // +kubebuilder:validation:Required
    Region string `json:"region"`

    // +optional
    Tags map[string]*string `json:"tags,omitempty"`
}

// Observation — read-only fields in status.atProvider.
type BucketRAWObservation struct {
    // +optional
    ARN *string `json:"arn,omitempty"`
}

// Spec embeds ResourceSpec (cluster-scoped → v1).
type BucketRAWSpec struct {
    xpv1.ResourceSpec `json:",inline"`
    ForProvider       BucketRAWParameters `json:"forProvider"`
}

// Status embeds ConditionedStatus.
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

#### Namespaced Type

For namespaced scope, swap the embedded spec:

```go
import xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

type BucketRAWSpec struct {
    xpv2.ManagedResourceSpec `json:",inline"`    // ← v2 for namespaced
    ForProvider              BucketRAWParameters `json:"forProvider"`
}

// +kubebuilder:resource:scope=Namespaced
type BucketRAW struct { ... }
```

#### Field Placement Rules

| Rule | Location |
|------|----------|
| User-configurable fields | `Parameters` struct (spec.forProvider) |
| Read-only / computed fields | `Observation` struct (status.atProvider) |
| Fields in `move-to-status-catalog.json` | `Observation` only — NOT `Parameters` |
| `region` | Always in `Parameters`, always required |

> **Always check** `.agents/specs/move-to-status-catalog.json` before placing fields. Some fields that are writable in TF have been moved to status-only in the RAW types.

### 4.3 Controller Setup

```go
// internal/controller/cluster/<service>/<resource>raw/controller.go
package bucketraw

import (
    xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
    "github.com/crossplane/crossplane-runtime/v2/pkg/event"
    "github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
    "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"

    awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

    clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
    native "github.com/upbound/provider-aws/v2/internal/native"
    nativev1beta1 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"
)

const (
    errDescribe = "cannot describe bucket"
    errCreate   = "cannot create bucket"
    errUpdate   = "cannot update bucket"
    errDelete   = "cannot delete bucket"
)

// init registers this controller into the NativeSetupHook for the service.
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
                    awss3.NewFromConfig,
                    func(c *awss3.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta1.BucketRAW] {
                        return &external{client: c, kube: kube}
                    },
                ),
            ),
            managed.WithLogger(o.Logger.WithValues("controller", name)),
            managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
            managed.WithPollInterval(o.PollInterval),
        ))
}
```

### 4.4 The `NewTypedConnector` Helper

`native.NewTypedConnector` handles all credential resolution:

```go
connector := native.NewTypedConnector[*MyResourceRAW, *awssvc.Client](
    mgr.GetClient(),                          // kube client
    awssvc.NewFromConfig,                     // aws.Config → *svc.Client
    func(c *awssvc.Client, kube client.Client) managed.TypedExternalClient[*MyResourceRAW] {
        return &external{client: c, kube: kube}
    },
)
```

Internally, `NewTypedConnector.Connect()` calls `clients.GetAWSConfigWithTracking(ctx, kube, mg)`, which:
1. Reads `spec.forProvider.region` and `spec.providerConfigRef` from the CR
2. Fetches the referenced `ProviderConfig`
3. Extracts credentials via `CommonCredentialExtractor`
4. Constructs an `aws.Config` with the resolved credentials and region

Requirements for the CR:
- Must implement `resource.Managed` (satisfied by embedding `ResourceSpec` or `ManagedResourceSpec`)
- Must have `spec.forProvider.region` at the expected JSON path
- Must have a valid `providerConfigRef`

### 4.5 NativeSetupHook Registration

The `NativeSetupHook_<service>` variable is declared in `internal/controller/cluster/native_hook.go`. Register by setting it from `init()`:

```go
func init() {
    clustercontroller.NativeSetupHook_s3 = Setup
}
```

For services with **multiple RAW resources**, create a shared setup file:

```go
// internal/controller/cluster/s3/setup.go
package s3

import (
    clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
    "github.com/upbound/provider-aws/v2/internal/controller/cluster/s3/bucketraw"
    "github.com/upbound/provider-aws/v2/internal/controller/cluster/s3/lifecycleraw"
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

### 4.6 Scheme Registration

New native types must be added to the scheme:

```go
// apis/cluster/native_register.go  (or per-service: apis/cluster/s3native_register.go)
package cluster

import natives3 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"

func init() {
    AddToSchemes = append(AddToSchemes, natives3.SchemeBuilder.AddToScheme)
}
```

---

## 5. CRUD Implementation

### 5.1 External Client Structure

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

### 5.2 Observe

Observe checks whether the external resource exists and whether it matches the desired state.

```go
func (e *external) Observe(ctx context.Context, cr *MyResourceRAW) (managed.ExternalObservation, error) {
    // 1. Get external name — if empty, resource hasn't been created yet
    externalName := native.GetExternalName(cr)
    if externalName == "" {
        return managed.ExternalObservation{ResourceExists: false}, nil
    }

    // 2. Describe the resource
    resp, err := e.client.DescribeMyResource(ctx, &awssvc.DescribeMyResourceInput{
        ResourceId: aws.String(externalName),
    })
    if err != nil {
        if native.IsNotFound(err) {
            return managed.ExternalObservation{ResourceExists: false}, nil
        }
        return managed.ExternalObservation{}, native.Wrap(err, errDescribe)
    }

    // 3. Map observed state to status
    cr.Status.AtProvider.ARN = resp.Resource.Arn
    cr.Status.AtProvider.State = aws.ToString(resp.Resource.State)

    // 4. Late initialization — fill nil spec fields from AWS response.
    //    ⚠️  Always use GetForProvider() + SetForProvider() — do NOT mutate
    //    cr.Spec.ForProvider directly. On namespaced types, GetForProvider()
    //    returns a copy; without SetForProvider() the changes are lost,
    //    causing an infinite late-init loop (ResourceLateInitialized: true
    //    every reconcile, Ready condition never becomes Available).
    spec := cr.GetForProvider()
    lateInit := false
    lateInit = native.LateInitializeStringPtr(
        &spec.Description, resp.Resource.Description,
    ) || lateInit
    if lateInit {
        cr.SetForProvider(*spec) // Write back — required for namespaced types
    }

    // 5. Set availability condition
    if aws.ToString(resp.Resource.State) == "ACTIVE" {
        cr.Status.SetConditions(xpv1.Available())
    } else {
        cr.Status.SetConditions(xpv1.Unavailable())
    }

    // 6. Detect drift
    upToDate := isUpToDate(cr, resp)

    return managed.ExternalObservation{
        ResourceExists:          true,
        ResourceUpToDate:        upToDate,
        ResourceLateInitialized: lateInit,
    }, nil
}
```

**Observe rules**:

| Scenario | Return | Effect |
|----------|--------|--------|
| External name empty | `ResourceExists: false` | Triggers Create |
| AWS returns NotFound | `ResourceExists: false`, nil error | Triggers Create |
| Resource exists, spec drifted | `ResourceUpToDate: false` | Triggers Update |
| Resource exists, in sync | `ResourceUpToDate: true` | No action, re-polls |
| Late-init populated a field | `ResourceLateInitialized: true` | Reconciler writes spec back |
| Other AWS error | Return the error | Reconciler retries |

> **⚠️ SetForProvider is mandatory for late init**: When `ResourceLateInitialized: true` is
> returned, the reconciler writes the CR spec back to the API server. However, this only works
> if `SetForProvider()` was called first — `GetForProvider()` on namespaced types returns a
> **copy** (to convert reference fields). Mutating the copy without calling `SetForProvider()`
> means the late-initialized fields are never stored in the CR. The result is an infinite
> reconcile loop where `ResourceLateInitialized: true` is returned every reconcile and the
> `Ready` condition never becomes `Available`.

### 5.3 Create

```go
func (e *external) Create(ctx context.Context, cr *MyResourceRAW) (managed.ExternalCreation, error) {
    cr.Status.SetConditions(xpv1.Creating())

    input := &awssvc.CreateMyResourceInput{
        ResourceName: aws.String(cr.Name),
        Description:  cr.Spec.ForProvider.Description,
        Tags:         buildTags(ctx, e.kube, cr),
    }

    resp, err := e.client.CreateMyResource(ctx, input)
    if err != nil {
        return managed.ExternalCreation{}, native.Wrap(err, errCreate)
    }

    // CRITICAL: Set external name from provider-assigned ID
    native.SetExternalName(cr, aws.ToString(resp.Resource.ResourceId))

    // Return connection details for sensitive outputs
    conn := managed.ConnectionDetails{
        "endpoint": []byte(aws.ToString(resp.Resource.Endpoint)),
    }
    return managed.ExternalCreation{ConnectionDetails: conn}, nil
}
```

**Create rules**:
- Always call `cr.Status.SetConditions(xpv1.Creating())` first
- Always call `native.SetExternalName(cr, id)` after successful creation
- Return `ConnectionDetails` for any sensitive outputs (check `connection-details-catalog.json`)

### 5.4 Update

```go
func (e *external) Update(ctx context.Context, cr *MyResourceRAW) (managed.ExternalUpdate, error) {
    externalName := native.GetExternalName(cr)

    _, err := e.client.UpdateMyResource(ctx, &awssvc.UpdateMyResourceInput{
        ResourceId:  aws.String(externalName),
        Description: cr.Spec.ForProvider.Description,
    })
    if err != nil {
        return managed.ExternalUpdate{}, native.Wrap(err, errUpdate)
    }

    // Tags are typically a separate API call
    if err := e.reconcileTags(ctx, cr); err != nil {
        return managed.ExternalUpdate{}, err
    }

    return managed.ExternalUpdate{}, nil
}
```

**Update rules**:
- Tags are usually a **separate** API call — do NOT include in the main update input
- Use `native.DiffTagsWithDefaults` to handle `default_tags` from ProviderConfig
- Never remove tags that belong to `defaultTags`

### 5.5 Delete

```go
func (e *external) Delete(ctx context.Context, cr *MyResourceRAW) (managed.ExternalDelete, error) {
    cr.Status.SetConditions(xpv1.Deleting())

    externalName := native.GetExternalName(cr)

    _, err := e.client.DeleteMyResource(ctx, &awssvc.DeleteMyResourceInput{
        ResourceId: aws.String(externalName),
    })
    if err != nil {
        if native.IsNotFound(err) {
            return managed.ExternalDelete{}, nil  // Already gone — idempotent
        }
        return managed.ExternalDelete{}, native.Wrap(err, errDelete)
    }

    return managed.ExternalDelete{}, nil
}
```

**Delete rules**:
- Always call `cr.Status.SetConditions(xpv1.Deleting())` first
- `NotFound` during delete → return nil error (idempotent)
- For async deletes: set `AsyncState{Operation: "deleting"}` and let Observe poll

### 5.6 Disconnect

Always a no-op:

```go
func (e *external) Disconnect(_ context.Context) error { return nil }
```

---

## 6. Helper Packages

All native helpers live in `internal/native/`. Import as:

```go
import native "github.com/upbound/provider-aws/v2/internal/native"
```

### 6.1 Error Handling (`native/errors.go`)

```go
// Check if an AWS error means "resource not found"
// Covers: NotFound, ResourceNotFoundException, NoSuchKey, NoSuchBucket, etc.
native.IsNotFound(err) bool

// Check if an AWS error is access denied
native.IsAccessDenied(err) bool

// Wrap errors with context (nil-safe)
native.Wrap(err, "cannot describe bucket")
native.Wrapf(err, "cannot describe bucket %q", name)
```

**Error handling patterns**:
- In Observe: `IsNotFound` → `ResourceExists: false`, no error
- In Delete: `IsNotFound` → return nil (idempotent)
- `IsAccessDenied` → return the error (do NOT swallow; reconciler retries)
- Always wrap errors with context using `native.Wrap`

Define error constants at the top of each controller:

```go
const (
    errDescribe      = "cannot describe myresource"
    errCreate        = "cannot create myresource"
    errUpdate        = "cannot update myresource"
    errDelete        = "cannot delete myresource"
    errListTags      = "cannot list tags for myresource"
    errTagResource   = "cannot tag myresource"
    errUntagResource = "cannot untag myresource"
)
```

### 6.2 External Name (`native/externalname.go`)

```go
// Read the crossplane.io/external-name annotation
externalName := native.GetExternalName(cr)  // returns "" if not set

// Set the annotation (call after successful Create)
native.SetExternalName(cr, providerAssignedID)
```

**External name strategies** — check `.agents/specs/external-name-catalog.json`:

| Strategy | Meaning | Create Implementation |
|----------|---------|----------------------|
| `name_as_identifier` | Use `cr.Name` | `SetExternalName(cr, cr.Name)` |
| `parameter_as_identifier` | Use a spec field | `SetExternalName(cr, *cr.Spec.ForProvider.FieldName)` |
| `identifier_from_provider` | AWS assigns the ID | `SetExternalName(cr, resp.ResourceId)` |
| `templated` | Composite ID | Encode components in Create, decode in Observe |

### 6.3 Tag Management (`native/tags.go`)

```go
// Get provider-level default tags from ProviderConfig
defaultTags, err := native.GetProviderDefaultTags(ctx, kube, cr)

// Diff tags accounting for default_tags (recommended for ALL resources)
add, update, remove := native.DiffTagsWithDefaults(
    cr.Spec.ForProvider.Tags,  // map[string]*string — desired
    defaultTags,               // map[string]*string — from ProviderConfig (may be nil)
    actualTags,                // map[string]*string — from AWS
)

// Simple diff without default_tags
diff := native.DiffTags(desired, actual)

// Semantic equality check
if native.AreSame(desired, actual) { ... }
```

**Always use `DiffTagsWithDefaults`** over `DiffTags` — it correctly prevents removing tags that belong to the ProviderConfig's `default_tags`.

### 6.4 Late Initialization (`native/lateinit.go`)

Called in Observe to populate nil spec fields from the AWS response:

> **⚠️ Never mutate `cr.Spec.ForProvider` directly** during late initialization.
> `GetForProvider()` on namespaced types returns a **copy** (the conversion from namespaced
> reference fields to cluster-scoped fields produces a new value). Changes to the copy are
> silently discarded unless written back with `SetForProvider()`. Forgetting this causes an
> infinite late-init loop: `ResourceLateInitialized: true` is returned every reconcile, the
> spec is never updated, and the `Ready` condition never becomes `Available`.

```go
// CORRECT — use GetForProvider / SetForProvider pattern
spec := cr.GetForProvider()
lateInit := false
lateInit = native.LateInitializeStringPtr(&spec.Description, resp.Description) || lateInit
lateInit = native.LateInitializeBoolPtr(&spec.EnableDNS, resp.EnableDns) || lateInit
lateInit = native.LateInitializeInt64Ptr(&spec.Timeout, resp.TimeoutSeconds) || lateInit
lateInit = native.LateInitializeMapStringPtr(&spec.Tags, convertTags(resp.Tags)) || lateInit
if lateInit {
    cr.SetForProvider(*spec) // Write back — required for namespaced types
}
```

Each helper returns `true` if the field was populated (destination was nil, source was non-nil).

```go
// WRONG — do NOT do this (changes are lost on namespaced types)
lateInit = native.LateInitializeStringPtr(&cr.Spec.ForProvider.Description, resp.Description) || lateInit
```

### 6.5 Async Operations (`native/async.go`)

For resources that take time to create/update/delete (marked `UseAsync = true` in the TF catalog):

```go
// Set async state (in Create/Update/Delete)
native.SetAsyncOperation(cr, native.AsyncState{
    Operation: "creating",
    StartedAt: time.Now(),
    RequestID: aws.ToString(resp.RequestMetadata.RequestId),
})

// Check in Observe
if native.IsAsyncInProgress(cr) {
    state, _ := native.GetAsyncOperation(cr)
    // Poll AWS for completion status
    // If complete: native.ClearAsyncOperation(cr)
    // If still running: return ResourceExists=true, UpToDate=true (re-poll)
}
```

### 6.6 IAM Policy Comparison (`native/policy.go`)

For resources with IAM policy fields:

```go
equal, err := native.PoliciesAreEquivalent(
    cr.Spec.ForProvider.Policy,
    aws.ToString(resp.Policy),
)
```

**Always use `PoliciesAreEquivalent`** for policy comparisons. AWS normalizes JSON key order and whitespace — naive string comparison causes infinite reconcile loops.

---

## 7. Cross-Resource References

RAW types use the same `+crossplane:generate:reference` annotations as TF types, but with a different extractor:

```go
// CORRECT — works with RAW types
// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
// +crossplane:generate:reference:extractor=github.com/crossplane/crossplane-runtime/v2/pkg/resource.ExtractResourceID()
KMSKeyID *string `json:"kmsKeyId,omitempty"`

// +optional
KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIdRef,omitempty"`

// +optional
KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIdSelector,omitempty"`
```

**WRONG** — do not use the upjet extractor (requires Terraformed interface, will panic):
```go
// +crossplane:generate:reference:extractor=github.com/crossplane/upjet/v2/pkg/resource.TerraformID()
```

Run `make generate` after adding reference annotations to produce the `ResolveReferences()` method.

---

## 8. Connection Details

Check `.agents/specs/connection-details-catalog.json` for which keys each resource must publish:

```go
conn := managed.ConnectionDetails{
    "endpoint": []byte(aws.ToString(resp.Endpoint.Address)),
    "port":     []byte(fmt.Sprint(aws.ToInt32(resp.Endpoint.Port))),
    "username": []byte(aws.ToString(resp.MasterUsername)),
}
return managed.ExternalCreation{ConnectionDetails: conn}, nil
```

For sensitive fields (marked `sensitive:*` in the catalog): read from the spec's SecretRef — most AWS services do NOT return secrets on read.

---

## 9. Complete Example: S3 BucketRAW

### Types

```go
// apis/cluster/s3/v1beta1/native/bucket_raw_types.go
package native

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

type BucketRAWParameters struct {
    // +kubebuilder:validation:Required
    Region string `json:"region"`

    // +optional
    Tags map[string]*string `json:"tags,omitempty"`

    // +optional
    // +kubebuilder:default=false
    ForceDestroy *bool `json:"forceDestroy,omitempty"`
}

type BucketRAWObservation struct {
    ARN          *string `json:"arn,omitempty"`
    BucketRegion *string `json:"bucketRegion,omitempty"`
}

type BucketRAWSpec struct {
    xpv1.ResourceSpec `json:",inline"`
    ForProvider       BucketRAWParameters `json:"forProvider"`
}

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

### Controller

```go
// internal/controller/cluster/s3/bucketraw/controller.go
package bucketraw

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/aws"
    awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
    awss3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
    xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
    xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
    "github.com/crossplane/crossplane-runtime/v2/pkg/event"
    "github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
    "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"

    clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
    native "github.com/upbound/provider-aws/v2/internal/native"
    nativev1beta1 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"
)

const (
    errDescribeBucket = "cannot describe s3 bucket"
    errCreateBucket   = "cannot create s3 bucket"
    errDeleteBucket   = "cannot delete s3 bucket"
    errListTags       = "cannot list tags for s3 bucket"
    errTagBucket      = "cannot tag s3 bucket"
)

func init() {
    clustercontroller.NativeSetupHook_s3 = Setup
}

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
                    awss3.NewFromConfig,
                    func(c *awss3.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta1.BucketRAW] {
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

func (e *external) Observe(ctx context.Context, cr *nativev1beta1.BucketRAW) (managed.ExternalObservation, error) {
    bucketName := native.GetExternalName(cr)
    if bucketName == "" {
        return managed.ExternalObservation{ResourceExists: false}, nil
    }

    // HeadBucket is the idiomatic S3 existence check
    _, err := e.client.HeadBucket(ctx, &awss3.HeadBucketInput{
        Bucket: aws.String(bucketName),
    })
    if err != nil {
        if native.IsNotFound(err) {
            return managed.ExternalObservation{ResourceExists: false}, nil
        }
        return managed.ExternalObservation{}, native.Wrap(err, errDescribeBucket)
    }

    // Populate status
    cr.Status.AtProvider.ARN = aws.String("arn:aws:s3:::" + bucketName)
    cr.Status.SetConditions(xpv1.Available())

    // Check tag drift
    upToDate := true
    tagsResp, err := e.client.GetBucketTagging(ctx, &awss3.GetBucketTaggingInput{
        Bucket: aws.String(bucketName),
    })
    if err == nil {
        defaultTags, _ := native.GetProviderDefaultTags(ctx, e.kube, cr)
        actual := s3TagsToMap(tagsResp.TagSet)
        add, upd, rem := native.DiffTagsWithDefaults(
            cr.Spec.ForProvider.Tags, defaultTags, actual,
        )
        upToDate = len(add)+len(upd)+len(rem) == 0
    }

    return managed.ExternalObservation{
        ResourceExists:   true,
        ResourceUpToDate: upToDate,
    }, nil
}

func (e *external) Create(ctx context.Context, cr *nativev1beta1.BucketRAW) (managed.ExternalCreation, error) {
    cr.Status.SetConditions(xpv1.Creating())

    bucketName := cr.Name // S3 uses name_as_identifier strategy

    input := &awss3.CreateBucketInput{Bucket: aws.String(bucketName)}
    if cr.Spec.ForProvider.Region != "us-east-1" {
        input.CreateBucketConfiguration = &awss3types.CreateBucketConfiguration{
            LocationConstraint: awss3types.BucketLocationConstraint(cr.Spec.ForProvider.Region),
        }
    }

    if _, err := e.client.CreateBucket(ctx, input); err != nil {
        return managed.ExternalCreation{}, native.Wrap(err, errCreateBucket)
    }

    native.SetExternalName(cr, bucketName)

    return managed.ExternalCreation{
        ConnectionDetails: managed.ConnectionDetails{
            "id":     []byte(bucketName),
            "arn":    []byte("arn:aws:s3:::" + bucketName),
            "region": []byte(cr.Spec.ForProvider.Region),
        },
    }, nil
}

func (e *external) Update(ctx context.Context, cr *nativev1beta1.BucketRAW) (managed.ExternalUpdate, error) {
    // S3 bucket metadata is immutable — only tags are mutable
    return managed.ExternalUpdate{}, e.reconcileTags(ctx, cr)
}

func (e *external) Delete(ctx context.Context, cr *nativev1beta1.BucketRAW) (managed.ExternalDelete, error) {
    cr.Status.SetConditions(xpv1.Deleting())

    _, err := e.client.DeleteBucket(ctx, &awss3.DeleteBucketInput{
        Bucket: aws.String(native.GetExternalName(cr)),
    })
    if err != nil {
        if native.IsNotFound(err) {
            return managed.ExternalDelete{}, nil
        }
        return managed.ExternalDelete{}, native.Wrap(err, errDeleteBucket)
    }
    return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(_ context.Context) error { return nil }

func (e *external) reconcileTags(ctx context.Context, cr *nativev1beta1.BucketRAW) error {
    bucketName := native.GetExternalName(cr)

    tagsResp, err := e.client.GetBucketTagging(ctx, &awss3.GetBucketTaggingInput{
        Bucket: aws.String(bucketName),
    })
    if err != nil && !native.IsNotFound(err) {
        return native.Wrap(err, errListTags)
    }
    var actual map[string]*string
    if err == nil {
        actual = s3TagsToMap(tagsResp.TagSet)
    }

    defaultTags, _ := native.GetProviderDefaultTags(ctx, e.kube, cr)
    add, upd, rem := native.DiffTagsWithDefaults(
        cr.Spec.ForProvider.Tags, defaultTags, actual,
    )

    desired := mergePtrMaps(add, upd)
    // S3 PutBucketTagging replaces all tags — so merge desired with defaults
    if len(desired) > 0 || len(rem) > 0 {
        allTags := mergePtrMaps(defaultTags, desired)
        _, err = e.client.PutBucketTagging(ctx, &awss3.PutBucketTaggingInput{
            Bucket:  aws.String(bucketName),
            Tagging: &awss3types.Tagging{TagSet: mapToS3Tags(allTags)},
        })
        if err != nil {
            return native.Wrap(err, errTagBucket)
        }
    }
    return nil
}

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

## 10. Common Pitfalls

| Pitfall | Correct Approach |
|---------|-----------------|
| Using `tjcontroller.Options` in Setup | Use `xpcontroller.Options` (crossplane-runtime) |
| Using `resource.TerraformID()` extractor | Use `resource.ExtractResourceID()` |
| Returning error on NotFound in Observe | Return `ResourceExists: false`, nil error |
| Returning error on NotFound in Delete | Return `ExternalDelete{}`, nil error |
| Comparing IAM policies with `==` | Use `native.PoliciesAreEquivalent()` |
| Removing ProviderConfig default_tags | Use `native.DiffTagsWithDefaults()` |
| Placing MoveToStatus fields in spec | Check `move-to-status-catalog.json`; put in Observation |
| Calling `SetExternalName` in Observe | Only call in Create (after successful provider response) |
| Using upjet imports | Avoid — native controllers should have zero upjet dependencies |
| Missing build tag | Add `//go:build <service> || all` to match service partitioning |
| Mutating `cr.Spec.ForProvider` directly in late init | Use `spec := cr.GetForProvider()`, mutate `spec`, then `cr.SetForProvider(*spec)` |

---

## 11. Pre-Commit Checklist

- [ ] `go test ./internal/controller/cluster/<service>/<resource>raw/...` passes
- [ ] `go test ./apis/cluster/<service>/...` passes (generated code compiles)
- [ ] `golangci-lint run ./internal/controller/cluster/<service>/...` passes
- [ ] `NativeSetupHook_<service> = Setup` is set in an `init()` function
- [ ] Scheme registration added in `apis/cluster/native_register.go`
- [ ] All error strings are lowercase (Go convention)
- [ ] `Disconnect` returns nil
- [ ] Tags use `native.DiffTagsWithDefaults` (not bare `DiffTags`)
- [ ] IAM policy comparisons use `native.PoliciesAreEquivalent`
- [ ] MoveToStatus fields are in `Observation`, not `Parameters`
- [ ] External name strategy matches `external-name-catalog.json` entry
- [ ] `TerraformID()` extractor NOT used anywhere
- [ ] Build tag matches service name (`//go:build s3 || all`)
- [ ] No upjet imports in controller or type files
- [ ] Late init uses `spec := cr.GetForProvider()` + `cr.SetForProvider(*spec)` (not `cr.Spec.ForProvider` direct mutation)

---

## 12. Reference Catalogs

| Catalog | Path | Use For |
|---------|------|---------|
| External name strategies | `.agents/specs/external-name-catalog.json` | How to set/get external name in Create/Observe |
| Fields moved to status | `.agents/specs/move-to-status-catalog.json` | Ensuring TF-writable fields that became read-only are in Observation |
| Connection details | `.agents/specs/connection-details-catalog.json` | Keys to publish in ConnectionDetails |
| TF business logic | `.agents/specs/tf-business-logic-catalog.md` | Replicating TF injector/custom-diff behavior |

---

## 13. Comparison: provider-template vs provider-aws Native

| Aspect | provider-template | provider-aws Native |
|--------|-------------------|---------------------|
| Connector | Hand-written (explicit credential resolution) | `native.NewTypedConnector` (encapsulates credential logic) |
| Controller options | `controller.Options` | `xpcontroller.Options` (same type, aliased import) |
| Registration | Explicit in `register.go` | Via `NativeSetupHook_<service>` + `init()` |
| Types | Single scope (namespaced) | Dual scope (cluster + namespaced) |
| ProviderConfig | Custom `ProviderCredentials` type | `clients.GetAWSConfigWithTracking` |
| Error handling | `github.com/pkg/errors` | `native.Wrap` / `native.IsNotFound` |
| Service client | Generic `interface{}` | Concrete SDK client (e.g., `*s3.Client`) |
| Features | Management policies, change logs, metrics | Same + async ops, tag management, policy comparison |
