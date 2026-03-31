# Architecture: provider-upjet-aws

A Crossplane provider that manages AWS resources as Kubernetes Custom Resources.
Built using [Upjet](https://github.com/crossplane/upjet) code generation from the
[Terraform AWS provider](https://github.com/hashicorp/terraform-provider-aws) (v6.34.0) schema.

---

## Table of Contents

- [Overview](#overview)
- [Key Metrics](#key-metrics)
- [Provider Family Architecture](#provider-family-architecture)
- [Directory Structure](#directory-structure)
- [Dual Scope Architecture](#dual-scope-architecture)
- [Resource Lifecycle](#resource-lifecycle)
- [Authentication](#authentication)
- [Code Generation Pipeline](#code-generation-pipeline)
- [Generated File Types](#generated-file-types)
- [Three-Struct Pattern](#three-struct-pattern)
- [Hand-Written Configuration Layer](#hand-written-configuration-layer)
- [Service Provider Binary Structure](#service-provider-binary-structure)
- [API Version Conversion](#api-version-conversion)
- [Build System](#build-system)

---

## Overview

This is a **family-style Crossplane provider** for AWS. Rather than shipping a single
monolithic binary, it decomposes into ~175 service-specific provider binaries
(e.g., `provider-aws-ec2`, `provider-aws-s3`) plus a shared `provider-aws-config`
for credential management. Each can be installed independently.

The provider translates Kubernetes Custom Resources into AWS API calls by embedding
the Terraform AWS provider in-process (no external Terraform CLI). 99%+ of the Go
source is auto-generated from Terraform schemas; the hand-written code lives
almost entirely under `config/` and `internal/clients/`.

```
                   Kubernetes API Server
                          |
            +-------------+-------------+
            |             |             |
    provider-aws-s3  provider-aws-ec2  provider-aws-rds  ... (175 binaries)
            |             |             |
            +------+------+------+------+
                   |
           provider-aws-config  (shared ProviderConfig / credentials)
                   |
            Terraform AWS Provider (in-process, v6.34.0)
                   |
               AWS APIs (SDK)
```

---

## Key Metrics

| Metric | Value |
|---|---|
| AWS services covered | **177** |
| Service provider binaries | **175** (`cmd/provider/`) |
| Resource types (CRDs) | **~2,015** (in `package/crds/`) |
| Total Go files | **~10,300** |
| Generated Go files (`zz_*`) | **~10,050** (97.4%) |
| Hand-written config services | **101** (under `config/cluster/`) |
| Terraform provider version | **6.34.0** |
| API groups | `aws.upbound.io` (cluster), `aws.m.upbound.io` (namespaced) |

---

## Provider Family Architecture

The provider ships as a **family** of independently installable packages:

```
provider-aws-config       # Shared ProviderConfig CRD + credential management
provider-aws-ec2          # EC2 resources (VPC, Instance, SecurityGroup, ...)
provider-aws-s3           # S3 resources (Bucket, BucketPolicy, Object, ...)
provider-aws-rds          # RDS resources (Instance, Cluster, Proxy, ...)
provider-aws-iam          # IAM resources (Role, Policy, User, ...)
...                       # 175 service-specific providers total
```

**Why family architecture?**
- Install only the AWS services you use → smaller memory footprint
- Independent upgrade cycles per service
- Shared credential management through `provider-aws-config`
- Each binary registers controllers for both cluster-scoped and namespace-scoped variants

Each service provider binary (`cmd/provider/{service}/zz_main.go`) is a generated
but complete Go program that:
1. Runs a boot check (`internal/bootcheck/`)
2. Creates a controller-runtime manager
3. Registers both cluster-scoped and namespace-scoped API schemes
4. Initializes the Terraform AWS provider in-process
5. Sets up controllers for its service's resources
6. Starts the manager with leader election, webhooks, and metrics

---

## Directory Structure

```
provider-upjet-aws/
├── apis/                           # Generated API types (CRD Go definitions)
│   ├── cluster/                    # Cluster-scoped APIs: aws.upbound.io
│   │   ├── {service}/              # 177 service directories
│   │   │   └── v1beta1/            # Version directory (some have v1beta2)
│   │   │       ├── zz_*_types.go           # Resource type definitions
│   │   │       ├── zz_*_terraformed.go     # Terraform bridge methods
│   │   │       ├── zz_generated.deepcopy.go
│   │   │       ├── zz_generated.managed.go
│   │   │       ├── zz_generated.managedlist.go
│   │   │       ├── zz_generated.conversion_hubs.go
│   │   │       ├── zz_generated.conversion_spokes.go
│   │   │       ├── zz_generated.resolvers.go
│   │   │       └── zz_groupversion_info.go
│   │   └── ...
│   └── namespaced/                 # Namespace-scoped APIs: aws.m.upbound.io
│       └── {service}/              # 177 service directories (mirrors cluster/)
│
├── cmd/                            # Binary entrypoints
│   ├── generator/main.go           # Code generation pipeline (hand-written)
│   ├── partitiongen/main.go        # AWS partition data generator (hand-written)
│   └── provider/                   # 175 service-specific provider binaries
│       ├── config/zz_main.go       # Shared credential/config provider
│       ├── ec2/zz_main.go          # EC2 service provider
│       ├── s3/zz_main.go           # S3 service provider
│       └── .../zz_main.go          # One per service
│
├── config/                         # Hand-written resource configuration *** CRITICAL ***
│   ├── externalname.go             # External name mappings (~3,676 lines)
│   ├── externalnamenottested.go    # External names not yet tested
│   ├── groups.go                   # Group/Kind mapping (~320 lines)
│   ├── overrides.go                # Global resource overrides (~253 lines)
│   ├── registry_cluster.go         # Cluster-scoped provider factory (~300 lines)
│   ├── registry_namespaced.go      # Namespace-scoped provider factory
│   ├── registry_common.go          # Shared registry code (skip list, schema loading)
│   ├── schema.json                 # Embedded Terraform provider schema (large)
│   ├── provider-metadata.yaml      # Embedded provider metadata
│   ├── generated.lst               # List of all generated Terraform resource names
│   ├── field-rename.yaml           # Field rename mappings (embedded at build)
│   ├── old-singleton-list-apis.txt # Resources with old singleton-list API versions
│   ├── cluster/                    # Cluster-scoped service configs
│   │   ├── provider.go             # Service registry init (imports + registers ~101 services)
│   │   ├── common/common.go        # Shared extractors (ARN, TerraformID, password gen)
│   │   └── {service}/config.go     # Per-service resource configuration (~101 services)
│   └── namespaced/                 # Namespace-scoped service configs
│       ├── provider.go             # Namespace-scoped service registry init
│       ├── common/common.go
│       └── {service}/config.go     # Mirrors cluster/ configs
│
├── internal/                       # Controllers and internal code
│   ├── controller/
│   │   ├── cluster/                # Cluster-scoped controllers (350 files)
│   │   └── namespaced/             # Namespace-scoped controllers (350 files)
│   ├── clients/                    # AWS client management
│   │   ├── aws.go                  # SetupConfig, global resource maps, region resolution
│   │   ├── provider_config.go      # Auth methods (IRSA, WebIdentity, PodIdentity, Secret, Upbound)
│   │   ├── cache.go                # AWS client caching
│   │   ├── creds_cache.go          # Credential caching
│   │   ├── partitions.go           # AWS partition utilities
│   │   ├── pc_resolver.go          # ProviderConfig resolution
│   │   └── zz_partitions_gen.go    # Generated partition data
│   ├── apis/                       # Internal API resolver scheme
│   ├── features/features.go        # Feature flags
│   ├── bootcheck/default.go        # Provider startup health checks
│   └── version/                    # Build version info
│
├── package/                        # Crossplane package definitions
│   ├── crds/                       # ~2,015 CRD YAML files
│   ├── crossplane.yaml.tmpl        # Package metadata template
│   ├── auth.yaml                   # RBAC rules
│   └── kustomize/                  # Kustomize overlays
│
├── examples/                       # Hand-written example manifests
│   └── {service}/                  # Per-service examples (177 services)
│
├── examples-generated/             # Auto-generated example manifests
│   ├── cluster/                    # Cluster-scoped examples
│   └── namespaced/                 # Namespace-scoped examples
│
├── cluster/                        # Cluster-level assets
│   ├── images/                     # Container image configs
│   └── test/                       # Cluster test fixtures
│
├── e2e/                            # End-to-end test infrastructure
│   └── providerconfig-aws-e2e-test/  # ProviderConfig e2e test suite
│
├── generate/generate.go            # Code generation go:generate entrypoint
├── build/                          # Build system (git submodule, crossplane/build)
├── hack/                           # Build scripts and templates
│   ├── main.go.tmpl                # Template for generated zz_main.go binaries
│   ├── embed.go                    # Go embed directives
│   ├── boilerplate.go.txt          # Go file header boilerplate
│   ├── boilerplate.yaml.txt        # YAML file header boilerplate
│   └── check-duplicate.sh          # Duplicate resource checker
│
├── scripts/                        # CI/CD scripts
│   ├── check-examples.py           # Example manifest validation
│   ├── family-test.py              # Family provider test runner
│   ├── tag.sh                      # Git tagging for releases
│   └── version_diff.py             # Version diff tooling
│
├── Makefile                        # GNU Make build orchestration (~440 lines)
└── go.mod / go.sum                 # Go module definition
```

---

## Dual Scope Architecture

Every resource type exists in **two variants**: cluster-scoped and namespace-scoped.
This is a first-class architectural decision that runs through the entire codebase.

```
                    +---------------------------+
                    |    Terraform AWS Schema    |
                    +---------------------------+
                               |
              +----------------+----------------+
              |                                 |
   config/registry_cluster.go       config/registry_namespaced.go
   GetProvider()                    GetProviderNamespaced()
              |                                 |
   Root group: aws.upbound.io       Root group: aws.m.upbound.io
              |                                 |
   apis/cluster/{service}/          apis/namespaced/{service}/
              |                                 |
   internal/controller/cluster/     internal/controller/namespaced/
              |                                 |
   Scope=Cluster CRDs               Scope=Namespaced CRDs
```

### Differences Between Scopes

| Aspect | Cluster-Scoped | Namespace-Scoped |
|---|---|---|
| API group | `aws.upbound.io` | `aws.m.upbound.io` |
| CRD scope | `scope=Cluster` | `scope=Namespaced` |
| Source config | `config/cluster/` | `config/namespaced/` |
| API path | `apis/cluster/` | `apis/namespaced/` |
| Controllers | `internal/controller/cluster/` | `internal/controller/namespaced/` |
| Field rename conversions | Yes (for backwards compat) | No |
| Singleton list API bumping | Yes (`bumpVersionsWithEmbeddedLists`) | No (uses `registerTFSingletonListConversions`) |

Both scopes share the same Terraform provider schema, the same resource include/skip
lists, and the same default resource options. They are built from the same code
generation pipeline (`pipeline.Run(pc, pns, absRootDir)` in `cmd/generator/main.go`).

Each service provider binary registers controllers for **both** scopes:

```go
// From cmd/provider/s3/zz_main.go
clustercontroller.SetupGated_s3(mgr, clusterOptions)
namespacedcontroller.SetupGated_s3(mgr, namespacedOptions)
```

---

## Resource Lifecycle

```
  User applies YAML manifest
         |
         v
  +----------------------------------------------+
  |  Kubernetes API Server                        |
  |  stores CR (e.g., s3.aws.upbound.io/Bucket)  |
  +----------------------------------------------+
         |
         v
  +----------------------------------------------+
  |  Controller watches for CR changes            |
  |  (poll interval: 10m, jitter: 5%)             |
  +----------------------------------------------+
         |
         v
  +----------------------------------------------+
  |  Resolve cross-resource references            |
  |  (zz_generated.resolvers.go)                  |
  |  e.g., BucketPolicy -> Bucket ARN             |
  +----------------------------------------------+
         |
         v
  +----------------------------------------------+
  |  Load credentials from ProviderConfig         |
  |  (internal/clients/provider_config.go)        |
  |  IRSA / WebIdentity / PodIdentity / Secret    |
  +----------------------------------------------+
         |
         v
  +----------------------------------------------+
  |  Initialize Terraform provider (in-process)   |
  |  (hashicorp/terraform-provider-aws/xpprovider)|
  |  No external Terraform CLI or state file      |
  +----------------------------------------------+
         |
         v
  +----------------------------------------------+
  |  Observe: Read current AWS state              |
  |  Compare: Desired (spec.forProvider) vs       |
  |           Actual  (status.atProvider)          |
  +----------------------------------------------+
         |
    +----+----+
    |         |
    v         v
  In sync   Drift detected
    |         |
    |         v
    |    +--------------------------------------------+
    |    |  Create / Update / Delete via TF provider   |
    |    |  Merges InitProvider + ForProvider params    |
    |    +--------------------------------------------+
    |         |
    v         v
  +----------------------------------------------+
  |  Update status.atProvider with observed state |
  |  Set conditions (Synced, Ready)               |
  |  Late-initialize unset spec fields            |
  +----------------------------------------------+
         |
         v
  Sleep until next poll interval (10m ± 5% jitter)
```

### Key Runtime Parameters

From `cmd/provider/{service}/zz_main.go`:

| Parameter | Default | Description |
|---|---|---|
| `--sync` | `1h` | Full sync interval (all resources) |
| `--poll` | `10m` | Per-resource drift check interval |
| `--poll-state-metric` | `5s` | State metric recording interval |
| `--max-reconcile-rate` | `100` | Global max reconciliations/second |
| `--leader-election` | `false` | Enable leader election for HA |
| `--enable-management-policies` | `true` | Enable beta management policies |
| `--enable-changelogs` | `false` | Enable alpha change log capture |

---

## Authentication

Authentication is configured via `ProviderConfig` CRDs and handled in
`internal/clients/provider_config.go`. Five methods are supported:

### 1. Secret-Based (Long-Term IAM Credentials)

```yaml
apiVersion: aws.upbound.io/v1beta1
kind: ProviderConfig
spec:
  credentials:
    source: Secret
    secretRef:
      name: aws-creds
      namespace: crossplane-system
      key: credentials   # Standard AWS credentials INI file
```

Handled by `UseProviderSecret()` — parses AWS credentials file format (INI),
extracts `aws_access_key_id`, `aws_secret_access_key`, and optionally
`aws_session_token` from the specified profile.

### 2. IRSA (IAM Roles for Service Accounts)

```yaml
spec:
  credentials:
    source: IRSA
```

Handled by `UseDefault()` — uses the default AWS SDK credential chain,
which automatically discovers the IRSA-injected web identity token
from the pod's projected volume (`AWS_WEB_IDENTITY_TOKEN_FILE` and `AWS_ROLE_ARN`).

### 3. Web Identity (OIDC Tokens)

```yaml
spec:
  credentials:
    source: WebIdentity
    webIdentity:
      roleARN: arn:aws:iam::123456789:role/my-role
      tokenConfig:
        source: Secret  # or Filesystem
```

Handled by `UseWebIdentityToken()` — performs `AssumeRoleWithWebIdentity` using
an OIDC token from a Kubernetes Secret or filesystem path.

### 4. EKS Pod Identity

```yaml
spec:
  credentials:
    source: PodIdentity
```

Handled by `UseDefault()` — same as IRSA, but relies on the EKS Pod Identity
agent injecting credentials via the container credential endpoint.

### 5. Upbound Identity

```yaml
spec:
  credentials:
    source: Upbound
```

Handled by `UseUpbound()` — uses the Upbound-injected identity token
from `/var/run/secrets/upbound.io/provider/token`.

### IAM Role Chaining (All Methods)

All methods support role chaining via `spec.assumeRoleChain`:

```yaml
spec:
  credentials:
    source: IRSA
  assumeRoleChain:
    - roleARN: arn:aws:iam::111111111111:role/role-a
      externalID: external-id-a
    - roleARN: arn:aws:iam::222222222222:role/role-b
```

Handled by `GetRoleChainConfig()` — iterates through the chain, performing
`sts:AssumeRole` for each hop. Enables cross-account access patterns.

---

## Code Generation Pipeline

The generation process transforms Terraform schemas into Crossplane CRDs,
controllers, and example manifests.

```
  Terraform AWS Provider (v6.34.0)
         |
         v
  config/schema.json                  (embedded Terraform JSON schema)
  config/provider-metadata.yaml       (embedded provider metadata)
         |
         v
  +----------------------------------------------+
  |  cmd/generator/main.go                        |
  |                                               |
  |  1. Load Terraform schema (JSON + Go SDK)     |
  |  2. Sync MaxItems constraints                 |
  |  3. Build cluster provider config             |
  |     (config.GetProvider)                      |
  |  4. Build namespaced provider config          |
  |     (config.GetProviderNamespaced)            |
  |  5. Run Upjet pipeline                        |
  |     (pipeline.Run(pc, pns, rootDir))          |
  |  6. Apply API converters to examples          |
  |  7. Dump generated resource list + coverage   |
  +----------------------------------------------+
         |
         v
  +-----------------+  +-------------------+  +------------------+
  | apis/cluster/   |  | apis/namespaced/  |  | internal/        |
  | {svc}/v1beta1/  |  | {svc}/v1beta1/    |  | controller/      |
  | zz_*.go         |  | zz_*.go           |  | {scope}/{svc}/   |
  +-----------------+  +-------------------+  +------------------+
         |                      |                      |
         v                      v                      v
  +----------------------------------------------+
  |  package/crds/*.yaml                          |
  |  examples-generated/{scope}/{svc}/            |
  |  cmd/provider/{svc}/zz_main.go                |
  +----------------------------------------------+
```

### Configuration Loading Sequence

When `GetProvider()` or `GetProviderNamespaced()` is called, the following
resource options are applied globally to every resource (in order):

```go
// From config/registry_cluster.go
defaultResourceOptions := []config.ResourceOption{
    GroupKindOverrides(),                        // Override API group/kind names
    KindOverrides(),                             // Override individual Kind names
    RegionRequired(),                            // Make region field required
    TagsAllRemoval(),                            // Remove tags_all computed field
    IdentifierAssignedByAWS(),                   // Use provider-assigned IDs
    KnownReferencers(),                          // Auto-add common references (role_arn, etc.)
    ResourceConfigurator(),                      // Apply per-resource configs
    NamePrefixRemoval(),                         // Remove name_prefix TF-specific fields
    DocumentationForTags(),                      // Add tag documentation
    injectFieldRenamingConversionFunctions(),     // [cluster only] Field rename conversions
    injectPluginFrameworkCustomStateEmptyCheck(), // Plugin framework state checks
}
```

After global options, each service's `Configure()` function is called via the
registry pattern.

### Skip List

Some Terraform resources are excluded from generation (`config/registry_common.go`):

```go
var skipList = []string{
    "aws_waf_rule_group$",              // CRD schema too large
    "aws_wafregional_rule_group$",      // CRD schema too large
    "aws_ecs_tag$",                     // Tags managed by ECS resources
    "aws_alb$",                         // Identical with aws_lb
    "aws_alb_listener$",               // Identical with aws_lb_listener
    "aws_alb_target_group$",           // Identical with aws_lb_target_group
    "aws_alb_target_group_attachment$", // Identical with aws_lb_target_group_attachment
    "aws_iam_policy_attachment$",      // Identical with aws_iam_*_policy_attachment
    "aws_iam_group_policy$",           // Identical with aws_iam_*_policy_attachment
    "aws_iam_user_policy$",            // Identical with aws_iam_*_policy_attachment
    "aws_location_map$",               // Failure with unknown reason
    "aws_appflow_connector_profile$",  // Failure with unknown reason
    "aws_rds_reserved_instance",       // Expense of testing
}
```

---

## Generated File Types

For each resource type (e.g., S3 Bucket), the pipeline generates the following
files inside `apis/{scope}/{service}/v1beta1/`:

| File | Generator | Purpose |
|---|---|---|
| `zz_{resource}_types.go` | Upjet | API type definitions (InitParameters, Observation, Parameters, Spec, Status) |
| `zz_{resource}_terraformed.go` | Upjet | Terraform bridge methods (GetTerraformResourceType, Get/SetObservation, Get/SetParameters, LateInitialize, GetMergedParameters) |
| `zz_generated.deepcopy.go` | controller-gen | `DeepCopyInto()` / `DeepCopy()` for all structs |
| `zz_generated.managed.go` | angryjet | Crossplane managed resource interface (GetCondition, SetConditions, GetDeletionPolicy, GetProviderConfigReference, etc.) |
| `zz_generated.managedlist.go` | angryjet | `GetItems()` for List types → `[]resource.Managed` |
| `zz_generated.conversion_hubs.go` | Upjet | `Hub()` marker methods for API version conversion |
| `zz_generated.conversion_spokes.go` | Upjet | `ConvertTo()` / `ConvertFrom()` for multi-version API conversion |
| `zz_generated.resolvers.go` | Upjet | Cross-resource reference resolvers |
| `zz_groupversion_info.go` | Upjet | Package metadata: CRDGroup, CRDVersion, SchemeBuilder, AddToScheme |

### Example: S3 v1beta1 Directory

`apis/cluster/s3/v1beta1/` contains 55 files:
- 24 resource types × 2 files each (types + terraformed) = 48 files
- 7 generated support files (deepcopy, managed, managedlist, hubs, spokes, resolvers, groupversion)

---

## Three-Struct Pattern

Every generated resource follows a consistent three-struct pattern that maps to
Crossplane's resource model. Using S3 Bucket as an example
(`apis/cluster/s3/v1beta1/zz_bucket_types.go`):

```go
// InitParameters — fields set ONLY at creation time (immutable after)
// Merged into ForProvider during the first reconciliation
type BucketInitParameters struct {
    ForceDestroy    *bool             `json:"forceDestroy,omitempty" tf:"force_destroy,omitempty"`
    ObjectLockEnabled *bool           `json:"objectLockEnabled,omitempty" tf:"object_lock_enabled,omitempty"`
    Tags            map[string]*string `json:"tags,omitempty" tf:"tags,omitempty"`
}

// Observation — read-only state observed from AWS (computed fields)
// Populated by the controller after each Observe cycle
type BucketObservation struct {
    Arn              *string            `json:"arn,omitempty" tf:"arn,omitempty"`
    BucketDomainName *string            `json:"bucketDomainName,omitempty" tf:"bucket_domain_name,omitempty"`
    ID               *string            `json:"id,omitempty" tf:"id,omitempty"`
    TagsAll          map[string]*string `json:"tagsAll,omitempty" tf:"tags_all,omitempty"`
    // ... all observed fields
}

// Parameters — user-settable desired state (superset of Init + references)
// +kubebuilder:validation markers control CRD schema validation
type BucketParameters struct {
    // +kubebuilder:validation:Required
    Region          *string            `json:"region" tf:"region,omitempty"`
    // +kubebuilder:validation:Optional
    ForceDestroy    *bool              `json:"forceDestroy,omitempty" tf:"force_destroy,omitempty"`
    // +mapType=granular
    Tags            map[string]*string `json:"tags,omitempty" tf:"tags,omitempty"`
}

// Spec wraps Parameters into Crossplane's ForProvider/InitProvider structure
type BucketSpec struct {
    v1.ResourceSpec `json:",inline"`
    ForProvider     BucketParameters     `json:"forProvider"`
    InitProvider    BucketInitParameters `json:"initProvider,omitempty"`
}

// Status wraps Observation into Crossplane's AtProvider structure
type BucketStatus struct {
    v1.ResourceStatus `json:",inline"`
    AtProvider        BucketObservation `json:"atProvider,omitempty"`
}

// The CRD root type
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Bucket struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec              BucketSpec   `json:"spec"`
    Status            BucketStatus `json:"status,omitempty"`
}
```

### How the Structs Map to YAML

```yaml
apiVersion: s3.aws.upbound.io/v1beta1
kind: Bucket
metadata:
  name: my-bucket
spec:
  forProvider:          # → BucketParameters
    region: us-west-2
    tags:
      Environment: dev
  initProvider:         # → BucketInitParameters (creation-only)
    objectLockEnabled: true
  providerConfigRef:    # → v1.ResourceSpec (inherited)
    name: default
status:
  atProvider:           # → BucketObservation (read-only, from AWS)
    arn: "arn:aws:s3:::my-bucket"
    bucketDomainName: "my-bucket.s3.amazonaws.com"
    id: "my-bucket"
  conditions:           # → v1.ResourceStatus (inherited)
    - type: Ready
      status: "True"
    - type: Synced
      status: "True"
```

### Terraform Bridge Methods

Each resource also gets a `zz_{resource}_terraformed.go` file implementing the
`resource.Terraformed` interface:

```go
func (tr *Bucket) GetTerraformResourceType() string    // → "aws_s3_bucket"
func (tr *Bucket) GetConnectionDetailsMapping() map[string]string
func (tr *Bucket) GetObservation() (map[string]any, error)
func (tr *Bucket) SetObservation(obs map[string]any) error
func (tr *Bucket) GetParameters() (map[string]any, error)
func (tr *Bucket) SetParameters(params map[string]any) error
func (tr *Bucket) GetInitParameters() (map[string]any, error)
func (tr *Bucket) GetMergedParameters(shouldMerge bool) (map[string]any, error)
func (tr *Bucket) LateInitialize(attrs []byte) (bool, error)
func (tr *Bucket) GetTerraformSchemaVersion() int
func (tr *Bucket) GetID() string
```

These methods marshal/unmarshal between Go structs and `map[string]any` using
`json.TFParser`, bridging the Kubernetes API types to Terraform's internal
representation.

---

## Hand-Written Configuration Layer

While 97%+ of the code is generated, the hand-written configuration in `config/`
is the most critical code for correctness. It tells Upjet *how* to generate each
resource.

### External Name Mapping (`config/externalname.go` — 3,676 lines)

Maps every Terraform resource to its Kubernetes external name strategy. This is
the single largest hand-written file and determines how resources are identified:

```go
var TerraformPluginFrameworkExternalNameConfigs = map[string]config.ExternalName{
    // Use a parameter as the identifier
    "aws_s3_bucket": config.ParameterAsIdentifier("bucket"),

    // Use the provider-assigned ID directly
    "aws_vpc": config.IdentifierFromProvider,

    // Combine multiple fields with a separator
    "aws_s3_bucket_analytics_configuration": FormattedIdentifierFromProvider(":", "bucket", "name"),

    // Template-based ID
    "aws_vpc_endpoint_connection_accepter": config.TemplatedStringAsIdentifier("",
        "{{ .parameters.vpc_endpoint_service_id }}_{{ .parameters.vpc_endpoint_id }}"),

    // Custom function for complex logic
    "aws_vpc_security_group_egress_rule": vpcSecurityGroupRule(),
    // ... ~1,000+ entries
}
```

### Global Overrides (`config/overrides.go` — 253 lines)

Applied to every resource during configuration:

| Override | Purpose |
|---|---|
| `RegionRequired()` | Makes `region` a required field on all region-aware resources |
| `TagsAllRemoval()` | Marks `tags_all` as computed-only (not user-settable) |
| `IdentifierAssignedByAWS()` | Sets default external name to provider-assigned |
| `KnownReferencers()` | Auto-detects common reference patterns (`*_role_arn` → IAM Role) |
| `NamePrefixRemoval()` | Removes Terraform's `name_prefix` field |
| `AddExternalTagsField()` | Adds Crossplane external tags support |

### Per-Service Configuration (`config/cluster/{service}/config.go`)

Each service has a hand-written `Configure()` function. Example from
`config/cluster/s3/config.go`:

```go
func Configure(p *config.Provider) {
    p.AddResourceConfigurator("aws_s3_bucket", func(r *config.Resource) {
        // Extract connection details (sensitive outputs → K8s Secret)
        r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
            conn := map[string][]byte{}
            if a, ok := attr["arn"].(string); ok { conn["arn"] = []byte(a) }
            return conn, nil
        }

        // Move read-only fields to status (not user-settable)
        config.MoveToStatus(r.TerraformResource, "arn", "acl", "grant", ...)

        // Cross-resource references
        r.References["...bucket_arn"] = config.Reference{
            TerraformName: "aws_s3_bucket",
            Extractor:     `resource.ExtractParamPath("arn",true)`,
        }

        // Late initialization exclusions
        r.LateInitializer = config.LateInitializer{
            IgnoredFields: []string{"acl", "access_control_policy"},
        }

        // API version management
        r.Version = "v1beta2"
        r.PreviousVersions = []string{"v1beta1"}
    })
}
```

### Service Registry Pattern (`config/cluster/provider.go`)

Services are registered via a collector pattern:

```go
// config/cluster/provider.go
type Configure func(provider *config.Provider)
type Configurator []Configure

var ProviderConfiguration = Configurator{}

func init() {
    ProviderConfiguration.AddConfig(s3.Configure)
    ProviderConfiguration.AddConfig(ec2.Configure)
    ProviderConfiguration.AddConfig(rds.Configure)
    // ... 101 service Configure functions
}
```

The registry is consumed by `GetProvider()` / `GetProviderNamespaced()`:

```go
// config/registry_cluster.go
for _, configure := range cluster.ProviderConfiguration {
    configure(pc)
}
pc.ConfigureResources()
```

---

## Service Provider Binary Structure

Every service provider binary (`cmd/provider/{service}/zz_main.go`) follows an
identical generated template (`hack/main.go.tmpl`). The S3 provider
(`cmd/provider/s3/zz_main.go`, 311 lines) is representative:

```
  main()
    |
    +-- bootcheck.CheckEnv()           // Validate environment at startup
    |
    +-- Parse CLI flags                // kingpin-based flag parsing
    |
    +-- ctrl.NewManager()              // Create controller-runtime manager
    |     - Leader election ID: "crossplane-leader-election-provider-aws-{service}"
    |     - Cache sync: 1h
    |     - Webhook server on port 9443
    |     - Metrics on :8080
    |     - Health probe on :8081
    |
    +-- Register API schemes           // Both cluster + namespaced
    |     - clusterapis.AddToScheme()
    |     - namespacedapis.AddToScheme()
    |     - apiextensionsv1.AddToScheme()
    |
    +-- xpprovider.GetProvider()       // Get Terraform SDK + Framework providers
    |
    +-- config.GetProvider()           // Cluster-scoped provider config
    +-- config.GetProviderNamespaced() // Namespace-scoped provider config
    |
    +-- Build controller options       // For both scopes
    |     - SetupFn: clients.SelectTerraformSetup(setupConfig)
    |     - OperationTrackerStore: track in-flight TF operations
    |     - GlobalRateLimiter: 100/sec
    |     - PollInterval: 10m
    |     - PollJitter: 5%
    |
    +-- Enable feature flags
    |     - EnableBetaManagementPolicies (default: true)
    |     - EnableAlphaChangeLogs (default: false, gRPC-based)
    |
    +-- canWatchCRD() check            // RBAC pre-check for CRD watching
    |     |
    |     +-- [allowed] SetupGated_{service}()  // CRD-gated controller setup
    |     +-- [denied]  Setup_{service}()       // Ungated fallback
    |
    +-- conversion.RegisterConversions()  // Webhook API version converters
    |
    +-- mgr.Start()                    // Start the controller manager
```

### CRD-Gated Controller Setup

The `canWatchCRD()` function performs a `SelfSubjectAccessReview` to check if
the provider has RBAC permissions to watch CRDs (`apiextensions.k8s.io`). If
allowed, controllers use **gated startup** — they wait until their CRD exists
before starting, preventing errors when not all CRDs are installed. This is
critical for the family architecture where only a subset of CRDs may be present.

---

## API Version Conversion

The provider supports multiple API versions per resource (e.g., v1beta1, v1beta2)
with automatic conversion. This is primarily driven by the
**singleton list → embedded object** migration.

### Background

Early CRD versions used Terraform's singleton lists (`maxItems: 1`) for nested
objects. Later versions converted these to proper embedded objects. The conversion
system handles this transparently.

### How It Works

1. **Hub version** — marked with `Hub()` method (typically v1beta1 or v1beta2)
2. **Spoke versions** — implement `ConvertTo()` / `ConvertFrom()` via
   `ujconversion.RoundTrip()`
3. **Storage version** — set to the older version for downgrade safety
4. **Controller reconcile version** — the version the controller operates on

```
  v1beta1 (storage, singleton lists)
     |
     | ConvertTo() / ConvertFrom()
     | (zz_generated.conversion_spokes.go)
     |
  v1beta2 (hub, embedded objects)
```

### Configuration in `config/registry_cluster.go`

```go
func bumpVersionsWithEmbeddedLists(pc *config.Provider) error {
    for name, r := range pc.Resources {
        if len(r.CRDListConversionPaths()) == 0 {
            continue  // No singleton lists converted
        }
        if _, ok := oldSLAPIs[name]; ok {
            // Old resource: needs full API conversion (v1beta1 ↔ v1beta2)
            configureSingletonListAPIConverters(r)
        } else {
            // New resource: only needs Terraform-level conversion
            r.TerraformConversions = []config.TerraformConversion{
                config.NewTFSingletonConversion(),
            }
        }
    }
}
```

The namespaced scope (`config/registry_namespaced.go`) is simpler — all
namespace-scoped APIs were created after the singleton list migration, so
they only need Terraform-level conversions.

---

## Build System

The build uses GNU Make with the Crossplane build submodule (`build/`):

```makefile
# Key variables from Makefile
PROVIDER_NAME := aws
PROJECT_REPO  := github.com/upbound/provider-aws/v2
TERRAFORM_VERSION          := 1.5.5
TERRAFORM_PROVIDER_VERSION := 6.34.0
PLATFORMS ?= linux_amd64 linux_arm64
```

### Key Build Targets

| Target | Purpose |
|---|---|
| `make generate` | Run full code generation pipeline |
| `make build` | Build all provider binaries |
| `make test` | Run unit tests |
| `make e2e` | Run end-to-end tests |
| `make check-examples` | Validate example manifests |

### Code Generation Entry Point

`generate/generate.go` contains the `go:generate` directive that triggers
`cmd/generator/main.go`:

```go
//go:generate go run ../cmd/generator/main.go ..
```

### Partition Data Generation

`cmd/partitiongen/main.go` generates AWS partition metadata
(`internal/clients/zz_partitions_gen.go`) from the AWS SDK Go v2 endpoints
JSON document. This data includes:

- Partition IDs (aws, aws-cn, aws-us-gov, etc.)
- Region lists and regex patterns
- DNS suffixes per partition
- Global service signing regions
- IAM region mappings

---

## Global Resource Handling

Some AWS services are global (not region-specific) but still require a region
for Terraform compatibility. The provider maintains explicit maps in
`internal/clients/aws.go`:

```go
// Individual global resources
var globalResources = map[string]string{
    "backup.aws.upbound.io/GlobalSettings":              "backup",
    "directconnect.aws.upbound.io/Gateway":              "directconnect",
    "s3control.aws.upbound.io/AccountPublicAccessBlock": "s3control",
}

// Entire API groups that are global
var globalGroups = map[string]string{
    "iam.aws.upbound.io":              "iam",
    "route53.aws.upbound.io":          "route53",
    "cloudfront.aws.upbound.io":       "cloudfront",
    "organizations.aws.upbound.io":    "organizations",
    // ... and their namespaced counterparts
}
```

---

## Key Dependencies

| Dependency | Role |
|---|---|
| `github.com/crossplane/upjet/v2` | Code generation framework and runtime |
| `github.com/crossplane/crossplane-runtime/v2` | Crossplane core abstractions (managed resources, conditions, etc.) |
| `github.com/hashicorp/terraform-provider-aws` | AWS resource schemas and CRUD operations |
| `github.com/hashicorp/terraform-plugin-sdk/v2` | Terraform Plugin SDK (most resources) |
| `github.com/hashicorp/terraform-plugin-framework` | Terraform Plugin Framework (newer resources) |
| `github.com/aws/aws-sdk-go-v2` | AWS SDK for authentication and STS operations |
| `sigs.k8s.io/controller-runtime` | Kubernetes controller framework |

---

## Where to Find Things

| "I need to..." | Look at... |
|---|---|
| Add a new AWS resource | `config/externalname.go` + `config/cluster/{service}/config.go` |
| Fix a resource configuration | `config/cluster/{service}/config.go` |
| Understand auth | `internal/clients/provider_config.go` |
| Debug a controller | `internal/controller/cluster/{service}/{resource}/` |
| Read API types for a resource | `apis/cluster/{service}/v1beta1/zz_{resource}_types.go` |
| Check if a resource is supported | `config/generated.lst` |
| See why a resource is skipped | `config/registry_common.go` (`skipList`) |
| Modify generation globally | `config/overrides.go` |
| Change the provider binary template | `hack/main.go.tmpl` |
| Add a cross-resource reference | `config/cluster/{service}/config.go` → `r.References[...]` |
| Write an example manifest | `examples/{service}/` |
| Run the generation pipeline | `cmd/generator/main.go` |
