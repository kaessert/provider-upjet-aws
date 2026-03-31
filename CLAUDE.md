# CLAUDE.md — Agent Guide for provider-upjet-aws

> **What this is**: Primary reference for AI agents working in this codebase.
> For deep-dives, see linked docs in [`.agents/docs/`](.agents/docs/).
> For migration specs, see [`.agents/specs/`](.agents/specs/).

---

## Project Identity

**Repository**: `provider-upjet-aws` (fork of `crossplane-contrib/provider-upjet-aws`)
**Purpose**: Crossplane provider that manages **1000+ AWS resource types** across **170+ services** as Kubernetes CRDs
**Technology**: Go, Upjet code generation from Terraform AWS provider (v6.34.0)
**Scale**: ~10,318 Go files — **99% auto-generated** (prefixed `zz_`)

### Active Migration
This repo is **preparing to migrate** from Upjet/Terraform-based controllers to **native AWS SDK v2 controllers**. See [Migration Specs](#migration-context) below.

---

## Critical Rules

1. **Never edit `zz_` files** — they are auto-generated and will be overwritten by `make generate`
2. **Hand-written code lives in `config/`** — this is where resource configuration happens
3. **Always run `make check-diff`** after changing config — this is a CI gate
4. **Dual scope**: every resource has cluster-scoped (`aws.upbound.io`) AND namespace-scoped (`aws.m.upbound.io`) variants
5. **go.mod/go.sum may be empty** — dependencies are managed externally
6. **Build needs 16GB+ RAM** — the codebase is enormous

---

## Repository Map

```
├── apis/                          # Generated API types (CRD definitions)
│   ├── cluster/{service}/{ver}/   #   Cluster-scoped (aws.upbound.io)
│   └── [root-level registration]  #   Scheme registration files
├── cmd/                           # Binary entrypoints
│   ├── generator/main.go         #   Code generation orchestrator
│   ├── partitiongen/main.go      #   AWS partition data generator
│   └── provider/{service}/       #   170+ service provider binaries (generated)
├── config/                        # ⭐ HAND-WRITTEN resource configuration
│   ├── externalname.go           #   ID mapping — 3,676 lines (CRITICAL)
│   ├── groups.go                 #   Group/Kind mapping — 320 lines
│   ├── overrides.go              #   Global overrides — 253 lines
│   ├── cluster/{service}/config.go   # Cluster-scoped service configs
│   ├── cluster/common/common.go      # Shared utilities
│   ├── namespaced/{service}/config.go # Namespace-scoped configs
│   └── namespaced/common/common.go
├── internal/                      # Controllers & internal code
│   ├── controller/cluster/       #   Generated cluster controllers
│   ├── controller/namespaced/    #   Generated namespaced controllers
│   ├── clients/                  #   ⭐ Hand-written: AWS auth, caching, credentials
│   ├── features/                 #   Feature flags
│   └── bootcheck/                #   Health checks
├── package/                       # Crossplane package definitions + CRDs
│   └── crds/                     #   2000+ CRD YAML files
├── examples/                      # Hand-written example manifests
├── examples-generated/            # Auto-generated examples
├── e2e/                           # E2E test infrastructure
├── build/                         # Build system (git submodule: crossplane/build)
├── hack/                          # Build scripts, templates
├── scripts/                       # CI scripts (check-examples.py, version_diff.py)
├── generate/                      # Code generation entry point
├── .agents/                       # Agent documentation & specs
│   ├── docs/                     #   Deep-dive reference docs
│   └── specs/                    #   Migration specifications
└── .hive/                         # Hive ant configuration
    └── ants/executor.yaml        #   Executor ant (stage:executor tickets)
```

### Key Files to Know

| File | Lines | Purpose |
|------|-------|---------|
| `config/externalname.go` | 3,676 | Maps Terraform resource IDs ↔ Crossplane external names |
| `config/overrides.go` | 253 | Global overrides (region, tags, known referencers) |
| `config/groups.go` | 320 | Terraform→Crossplane group/kind mapping |
| `config/cluster/common/common.go` | 173 | ARN extractor, password generator, diff helpers |
| `config/cluster/provider.go` | — | Registers all cluster service configurations |
| `cmd/generator/main.go` | — | Code generation pipeline orchestrator |
| `internal/clients/` | — | AWS client init, credential caching, ProviderConfig resolver |
| `AUTHENTICATION.md` | — | Full auth documentation (Secret, IRSA, WebIdentity, PodIdentity) |
| `Makefile` | 440 | Build orchestration with submodule includes |

---

## Architecture

### How It Works

```
User writes YAML (Kubernetes CR)
         ↓
Controller watches → resolves cross-resource references
         ↓
Loads AWS credentials from ProviderConfig
         ↓
Terraform provider (in-process) compares desired vs actual
         ↓
Creates/Updates/Deletes AWS resource
         ↓
Updates .status.atProvider with observed state
         ↓
Polls every ~10min for drift detection
```

### The Three-Struct Pattern (every resource)

```go
type BucketInitParameters struct { ... }  // Creation-only fields
type BucketObservation struct { ... }     // Read-only state from AWS
type BucketParameters struct { ... }      // User-settable fields + references

type BucketSpec struct {
    ForProvider  BucketParameters       `json:"forProvider"`
    InitProvider BucketInitParameters   `json:"initProvider,omitempty"`
}
type BucketStatus struct {
    AtProvider BucketObservation `json:"atProvider,omitempty"`
}
```

### Provider Family

- **`config` provider**: Shared ProviderConfig and authentication
- **Service providers** (ec2, s3, rds, ...): Individual AWS service controllers
- Each installable independently as Crossplane packages
- Published to `xpkg.upbound.io/upbound`

### Dual Scope

| Scope | API Group | Config | Controllers |
|-------|-----------|--------|-------------|
| Cluster | `aws.upbound.io` | `config/cluster/` | `internal/controller/cluster/` |
| Namespaced | `aws.m.upbound.io` | `config/namespaced/` | `internal/controller/namespaced/` |

→ **Deep dive**: [`.agents/docs/architecture.md`](.agents/docs/architecture.md)

---

## Build System

### Essential Commands

```bash
# Code generation
make generate.init          # Download TF provider schema → config/schema.json
make generate               # Run Upjet code generation pipeline

# Building
make build                  # Compile provider binaries
SUBPACKAGES="config ec2" make build  # Build specific services

# Testing
make test                   # Unit tests
make uptest UPTEST_EXAMPLE_LIST="examples/s3/cluster/v1beta1/bucket.yaml"  # E2E
make e2e                    # Full family e2e

# Validation (CI gates)
make check-diff             # Verify generated files match source
make lint                   # golangci-lint with buildtagger
make crddiff                # Detect breaking CRD changes

# Local development
make run                    # Run provider locally
make local-deploy           # Build & deploy to local Crossplane
```

### CI Pipeline (`.github/workflows/ci.yml`)

```
PR/Push → detect-noop → lint → check-diff → unit-tests → local-deploy → check-examples
```

E2E triggered via PR comment: `/test-examples="examples/s3/cluster/v1beta1/bucket.yaml"`

→ **Deep dive**: [`.agents/docs/build-system.md`](.agents/docs/build-system.md)

---

## Configuration System (How Resources Are Configured)

The `config/` directory is the **only hand-written code that matters** for resource behavior. Everything else is generated.

### External Name Patterns (`config/externalname.go`)

```go
// Pattern 1: AWS assigns ID (most common)
"aws_sqs_queue": config.IdentifierFromProvider,

// Pattern 2: Use name field as ID
"aws_iam_role": config.NameAsIdentifier,

// Pattern 3: Specific field is ID
"aws_key_pair": config.ParameterAsIdentifier("key_name"),

// Pattern 4: Composite template ID
"aws_route53_record": config.TemplatedStringAsIdentifier("name",
    "{{.parameters.zone_id}}_{{.external_name}}_{{.parameters.type}}"),

// Pattern 5: Custom function for complex cases
"aws_cognito_user_pool_client": cognitoUserPoolClient(),
```

### Resource Configuration (`config/cluster/{service}/config.go`)

```go
func Configure(p *config.Provider) {
    p.AddResourceConfigurator("aws_instance", func(r *config.Resource) {
        r.References["subnet_id"] = config.Reference{TerraformName: "aws_subnet"}
        r.LateInitializer = config.LateInitializer{
            IgnoredFields: []string{"subnet_id", "network_interface"},
        }
        r.UseAsync = true
    })
}
```

### Adding a New Resource

1. Add external name in `config/externalname.go`
2. Create/update `config/cluster/{service}/config.go` with `Configure()` function
3. Register in `config/cluster/provider.go`
4. Repeat for `config/namespaced/` if namespace-scoped variant needed
5. Run `make generate` then `make check-diff`

### Global Overrides (`config/overrides.go`)

Auto-applied to all resources:
- `KnownReferencers()` — auto-adds refs for `*_role_arn`, `vpc_id`, `subnet_ids`, `security_group_ids`, `kms_key_id`
- `RegionRequired()` — marks region field as required
- `TagsAllRemoval()` — removes TF-specific `tags_all`

→ **Deep dive**: [`.agents/docs/config-patterns.md`](.agents/docs/config-patterns.md)

---

## Testing

### Test Layers

| Layer | Coverage | Command |
|-------|----------|---------|
| Unit tests | 8 files — config converters, client utilities | `make test` |
| Example validation | YAML syntax/schema checking | `scripts/check-examples.py` |
| E2E (Uptest) | Full resource lifecycle against real AWS | `make uptest` |

### Unit Test Locations

```
config/cluster/autoscaling/config_test.go     # Version conversion
config/cluster/common/common_test.go           # Password generation
config/cluster/rds/utils/engine_version_test.go # Engine version parsing
internal/clients/cache_test.go                  # Client caching
internal/clients/partitions_test.go             # AWS partitions
```

### Test Patterns

- Standard Go `testing` package with table-driven tests
- `google/go-cmp` for value comparison
- `crossplane-runtime/test` for mocks
- Coverage filtering: `zz_*` files excluded from reports

→ **Deep dive**: [`.agents/docs/testing.md`](.agents/docs/testing.md)

---

## Authentication

Five methods supported (see `AUTHENTICATION.md` for full docs):

1. **Secret** — AWS credentials in K8s Secret (dev only)
2. **IRSA** — IAM Roles for Service Accounts (EKS)
3. **WebIdentity** — OIDC token-based
4. **PodIdentity** — EKS Pod Identity
5. **Role Chaining** — Assume roles for multi-account (works with all above)

```yaml
apiVersion: aws.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret  # or IRSA, WebIdentity, PodIdentity
    secretRef:
      name: aws-creds
      namespace: upbound-system
      key: credentials
```

---

## Migration Context

### Current State: TF-Bridged

```
K8s CR → upjet AsyncConnector → TF PluginSDK (in-process) → AWS SDK v2 → AWS API
```

### Target State: Native

```
K8s CR → native Connector → TypedExternalClient → AWS SDK v2 → AWS API
```

### Key Specs

| Spec | Path | Summary |
|------|------|---------|
| Migration Design | [`.agents/specs/terraform-removal-migration.md`](.agents/specs/terraform-removal-migration.md) | Master plan: 349 resources, 5 phases, ~1,840 tickets |
| Native Controller Pattern | [`.agents/specs/native-controller-pattern.md`](.agents/specs/native-controller-pattern.md) | Implementation guide for RAW controllers |

### Migration Strategy

- **Parallel coexistence**: `Kind: Bucket` (TF) + `Kind: BucketRAW` (native) run simultaneously
- **YAML compatibility**: RAW types produce identical CRD schema (same json tags, no `tf:` tags)
- **Rolling cutover**: Per-service, 4 tiers from simple to complex (EC2 last)
- **Phase 0**: Infrastructure tickets (helpers, catalogs, build hooks) — must complete first

### RAW Controller File Layout

```
apis/cluster/<service>/<version>/native/    # RAW types (native/ subpackage)
internal/controller/cluster/<service>/<resource>raw/  # Native controller
```

→ **Deep dive**: [`.agents/docs/native-controller-guide.md`](.agents/docs/native-controller-guide.md)

---

## Hive Integration

### Executor Ant

The executor ant (`.hive/ants/executor.yaml`) watches for `stage:executor` tickets:
- Claims one ticket at a time
- Reads specs from `.agents/specs/`
- Implements via TDD (test first, then code)
- Commits with conventional commits (`feat:`, `fix:`, etc.)
- Stages specific files only (never `git add -A`)

### Creating Work

When creating tickets for the executor:
- Label with `stage:executor`
- Include clear acceptance criteria
- Reference relevant specs in description
- Set dependencies when ordering matters

---

## Deep-Dive Documentation Index

| Document | Path | Topics |
|----------|------|--------|
| Architecture | [`.agents/docs/architecture.md`](.agents/docs/architecture.md) | Provider family, dual scope, lifecycle, generated patterns |
| Build System | [`.agents/docs/build-system.md`](.agents/docs/build-system.md) | Makefile targets, CI/CD, tool versions, code generation flow |
| Config Patterns | [`.agents/docs/config-patterns.md`](.agents/docs/config-patterns.md) | External names, references, late init, custom diff, overrides |
| Testing | [`.agents/docs/testing.md`](.agents/docs/testing.md) | Unit tests, E2E (Uptest), CI pipeline, test patterns |
| Native Controllers | [`.agents/docs/native-controller-guide.md`](.agents/docs/native-controller-guide.md) | Provider-template pattern, ExternalClient, migration |
| Migration Spec | [`.agents/specs/terraform-removal-migration.md`](.agents/specs/terraform-removal-migration.md) | Full migration plan, phases, design decisions |
| Native Pattern Spec | [`.agents/specs/native-controller-pattern.md`](.agents/specs/native-controller-pattern.md) | Implementation guide for RAW controllers |

---

## Quick Reference

### Common Tasks

| Task | Command/Location |
|------|-------------------|
| Add a new AWS resource | Edit `config/externalname.go` + `config/cluster/{svc}/config.go` + `make generate` |
| Run unit tests | `make test` |
| Run E2E for a resource | `make uptest UPTEST_EXAMPLE_LIST="examples/{svc}/cluster/v1beta1/{res}.yaml"` |
| Check if generated files are stale | `make check-diff` |
| Find config for a service | `config/cluster/{service}/config.go` |
| Find generated types | `apis/cluster/{service}/{version}/zz_{resource}_types.go` |
| Find controller setup | `internal/controller/cluster/zz_{service}_setup.go` |
| Understand a resource's ID strategy | Search `config/externalname.go` for the Terraform resource name |
| View auth options | `AUTHENTICATION.md` |

### Coding Conventions

- **Go**: Follow `.golangci.yml` — gofmt, goimports with local prefix `github.com/upbound/provider-aws/v2`
- **Commits**: Conventional commits (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`)
- **Generated code**: Never edit `zz_` prefixed files
- **Config changes**: Always pair with `make generate` + `make check-diff`
- **Tests**: Table-driven, use `google/go-cmp` for assertions
- **Files**: Stage specific files (`git add <file>`) — never `git add -A`
