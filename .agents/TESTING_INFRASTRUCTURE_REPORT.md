# Comprehensive Testing Infrastructure Report
## Upbound Crossplane Provider for AWS (provider-aws)

**Last Updated**: 2024
**Repository**: `upbound/provider-aws` (v2 - multi-package monolith)
**Provider Type**: Upjet-based Crossplane provider

---

## Table of Contents

1. [Overview](#overview)
2. [Test Types & Frameworks](#test-types--frameworks)
3. [Unit Tests](#unit-tests)
4. [End-to-End (E2E) Tests](#end-to-end-e2e-tests)
5. [Examples & Fixtures](#examples--fixtures)
6. [CI/CD Testing Pipeline](#cicd-testing-pipeline)
7. [Testing Commands & Targets](#testing-commands--targets)
8. [Test Infrastructure Components](#test-infrastructure-components)
9. [Testing Patterns & Conventions](#testing-patterns--conventions)
10. [Gaps & Observations](#gaps--observations)

---

## Overview

The provider-aws testing infrastructure is sophisticated and multi-layered, supporting:
- **Unit tests** for Go configuration converters and helpers
- **Integration tests** via local deployment and validation
- **End-to-End (E2E) tests** using the Uptest framework
- **ProviderConfig-specific E2E tests** for authentication scenarios
- **Example manifest validation** via Python scripts
- **CI automation** via GitHub Actions with multiple test jobs

The codebase demonstrates a mature Crossplane provider testing approach, leveraging Upjet's code generation and Uptest's e2e testing capabilities.

---

## Test Types & Frameworks

### Testing Frameworks Used

| Framework/Tool | Purpose | Location | Version/Reference |
|---|---|---|---|
| **Go testing** | Unit tests for config converters, utilities | `config/**/*_test.go`, `internal/**/*_test.go` | Built-in `testing` package |
| **Uptest** | E2E test orchestration, example validation | Makefile targets: `uptest`, `e2e` | v2.2.0 (main), v0.13.0 (e2e test config) |
| **Kind** | Local Kubernetes clusters for testing | CI/CD & local development | v0.30.0 (main), v0.22.0 (e2e config) |
| **Kubectl** | Kubernetes resource management | Test setup & validation | Via build tools |
| **Crossplane CLI** | Provider package operations | Test setup | v2.2.0 |
| **Python (pytest)** | Example manifest validation | `scripts/check-examples.py` | Python 3 |
| **KUTTL** | Kubernetes test orchestration (if used) | e2e tests | Referenced in e2e Makefile |
| **google/go-cmp** | Go value comparison testing | Unit tests | Via go.mod |
| **crossplane-runtime/test** | Mock clients and test utilities | Unit tests | `crossplane-runtime/v2` |
| **Upjet fake resources** | Test resource mocks | Unit tests | `upjet/v2/pkg/resource/fake` |

### Test Coverage Categories

1. **Unit Tests**: Configuration conversions, password generation, parsing utilities
2. **Integration Tests**: Local provider deployment and validation
3. **E2E Tests**: Full resource lifecycle management (create, read, update, delete) against AWS
4. **ProviderConfig Tests**: Authentication methods (IAM role, credentials, IRSA, WebIdentity)
5. **Example Validation**: YAML manifest syntax and structure validation

---

## Unit Tests

### Location & Structure

```
config/
├── cluster/
│   ├── autoscaling/
│   │   └── config_test.go          # AutoscalingGroup v1beta1↔v1beta2 conversion tests
│   ├── common/
│   │   └── common_test.go          # Password generation for RDS/DatabaseInstance
│   ├── elasticache/
│   │   └── config_test.go          # ElastiCache configuration converter
│   └── rds/
│       └── utils/
│           └── engine_version_test.go  # RDS engine version parsing
├── namespaced/
│   ├── common/
│   │   └── common_test.go          # Namespaced password generation
│   └── rds/
│       └── utils/
│           └── engine_version_test.go  # Namespaced RDS utilities

internal/
├── clients/
│   ├── cache_test.go               # Client caching utilities
│   └── partitions_test.go          # AWS partition handling
```

### Test Files Found (8 unit test files)

1. **config/cluster/autoscaling/config_test.go**
   - Tests: `TestAutoScalingGroupConverterFromv1beta1Tov1beta2`, `TestAutoScalingGroupConverterFromv1beta2Tov1beta1`
   - Validates tag structure conversion between API versions
   - Tests successful conversions, error handling, missing key cases

2. **config/cluster/common/common_test.go**
   - Tests: `TestPasswordGenerator`
   - 16 test cases covering:
     - Secret retrieval errors
     - Password auto-generation toggle logic
     - Secret creation vs. patching
     - Both namespaced and cluster variants (RDS, Elasticache)
   - Uses mocks: `test.MockClient`, `ujfake.Terraformed`

3. **config/cluster/elasticache/config_test.go**
   - ElastiCache-specific configuration converters

4. **config/cluster/rds/utils/engine_version_test.go**
   - RDS engine version parsing and validation

5. **internal/clients/cache_test.go**
   - Client-side caching tests

6. **internal/clients/partitions_test.go**
   - AWS partition resolution tests

### Unit Test Patterns

**Pattern: Table-Driven Tests**

```go
// From config/cluster/autoscaling/config_test.go
func TestAutoScalingGroupConverterFromv1beta1Tov1beta2(t *testing.T) {
    type args struct { src xpresource.Managed; target xpresource.Managed }
    type want struct { target xpresource.Managed; err error }
    cases := map[string]struct {
        args args
        want want
    }{
        "Successful": { /* test case */ },
        "Unsuccessful": { /* test case */ },
        "MissingKey": { /* test case */ },
    }
    for name, tc := range cases {
        t.Run(name, func(t *testing.T) {
            err := /* function under test */
            if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
                t.Errorf("...: -want error, +got error:\n%s", diff)
            }
        })
    }
}
```

**Pattern: Mock Clients**

```go
kube: &test.MockClient{
    MockGet: test.NewMockGetFn(errBoom),  // Inject errors
    MockCreate: func(ctx context.Context, obj client.Object, ...) error { /* assertions */ },
    MockPatch: func(ctx context.Context, obj client.Object, ...) error { /* assertions */ },
}
```

### Running Unit Tests

```bash
# Run all unit tests (with controlled parallelism due to kube-apiserver startup)
make test

# Run with custom parallelism (default: NPROCS/2)
make -j2 test

# Run specific unit test package
go test ./config/cluster/common/...
```

**CI Configuration** (from `.github/workflows/ci.yml`):
- Job: `unit-tests`
- Runs on: Ubuntu-Jumbo-Runner (large-machine)
- Step: `Run Unit Tests` executes `make -j2 test`
- Parallelism: `-j2` (configured because each test suite starts kube-apiserver)

---

## End-to-End (E2E) Tests

### E2E Test Categories

#### 1. **Standard Resource E2E Tests (Uptest-based)**

**Purpose**: Validate full resource lifecycle (CRUD operations) against AWS

**Framework**: Uptest v2.2.0

**Location**: `examples/` and `examples-generated/` directories

**Mechanism**:
1. Reads example YAML manifests for AWS resources
2. Deploys provider to local Kind cluster via Crossplane
3. Applies resource manifests to cluster
4. Observes resource status until ready (`Test` condition)
5. Verifies AWS backend state
6. Deletes resources and verifies cleanup

**Example Structure**:

```
examples/
├── ec2/
│   ├── cluster/           # Cluster-scoped resources
│   │   └── v1beta1/
│   │       ├── instance.yaml
│   │       ├── vpc.yaml
│   │       ├── securitygroup.yaml
│   │       └── [90+ more EC2 resources]
│   └── namespaced/        # Namespace-scoped resources
│       └── v1beta1/
│           └── [same resources]
├── rds/
│   ├── cluster/
│   │   └── v1beta1/
│   │       ├── cluster.yaml
│   │       ├── instance.yaml
│   │       └── [20+ RDS resources]
│   └── namespaced/
├── s3/
│   └── [bucket.yaml, bucketacl.yaml, etc.]
├── [100+ AWS services]
└── examples-generated/    # Auto-generated from Terraform schema
    ├── cluster/
    ├── namespaced/
    └── [mirrors examples/ but with TF defaults]
```

**Total Examples**: ~2,364 files in `examples/`, ~2,011 in `examples-generated/`

**Running Uptest E2E Tests**:

```bash
# Standard: test specific examples with cloud credentials
export UPTEST_CLOUD_CREDENTIALS='[default]
aws_access_key_id = YOUR_KEY
aws_secret_access_key = YOUR_SECRET'

# Test comma-separated list of examples
export UPTEST_EXAMPLE_LIST="examples/s3/cluster/v1beta1/bucket.yaml,examples/ec2/cluster/v1beta1/instance.yaml"

# Optional: provide dynamic values via datasource
export UPTEST_DATASOURCE_PATH=".work/uptest-datasource.yaml"

make uptest

# With family provider (dynamically determines which subpackage to test)
make family-e2e

# Full e2e with local deployment
make e2e
```

**Uptest Flow** (from `Makefile` line 230-232):

```makefile
uptest: $(UPTEST) $(KUBECTL) $(CHAINSAW) $(CROSSPLANE_CLI)
	KUBECTL=$(KUBECTL) CHAINSAW=$(CHAINSAW) CROSSPLANE_CLI=$(CROSSPLANE_CLI) \
	CROSSPLANE_NAMESPACE=$(CROSSPLANE_NAMESPACE) $(UPTEST) e2e \
	"${UPTEST_EXAMPLE_LIST}" --data-source="${UPTEST_DATASOURCE_PATH}" \
	--setup-script=cluster/test/setup.sh --default-conditions="Test"
```

**Setup Script** (`cluster/test/setup.sh`):
- Creates AWS credential secrets from `UPTEST_CLOUD_CREDENTIALS`
- Deploys default (and peer for cross-account) ProviderConfigs
- Supports multiple credential sets (`DEFAULT`, `PEER`)

---

#### 2. **ProviderConfig E2E Tests**

**Purpose**: Test various authentication methods for ProviderConfigs (IAM role, credentials, IRSA, WebIdentity)

**Framework**: Uptest v0.13.0 + Crossplane Configuration Package + Kuttl

**Location**: `e2e/providerconfig-aws-e2e-test/`

**Architecture**: Two-level Crossplane cluster setup (Crossplane manages Crossplane)

```
┌─ Local Kind Cluster (control plane 1) ──┐
│  Runs local Crossplane instance         │
│  Manages EKS cluster creation           │
│  Deploys providers to remote EKS        │
└──────────────────────────────────────────┘
                    ↓
    ┌─ AWS EKS Cluster (remote) ─┐
    │ Second Crossplane instance   │
    │ Tests authentication methods │
    │ (IAM role, IRSA, WebId)     │
    └──────────────────────────────┘
```

**Components**:

```
e2e/providerconfig-aws-e2e-test/
├── Makefile                      # e2e test orchestration
├── README.md                     # Detailed setup instructions
├── test/
│   └── setup.sh                 # Test environment setup
├── package/
│   ├── crossplane.yaml          # Configuration package metadata
│   ├── apis/
│   │   └── e2etestcluster/
│   │       ├── definition.yaml   # XRD: E2ETestCluster
│   │       ├── composition.yaml  # Composite resource implementation
│   │       └── [functions.yaml]  # Crossplane functions
│   └── examples/
│       └── e2etestcluster-claim.yaml  # Test resource claim
```

**E2ETestCluster Composition** creates:
- VPC, Subnets, Security Groups (networking)
- EKS Cluster with IAM roles
- IRSA-enabled service accounts
- WebIdentity provider configuration
- Provider packages (ec2, rds, kafka, config)
- Example MRs testing different auth methods

**Running ProviderConfig E2E Tests**:

```bash
# Option 1: With pre-built provider images
export AWS_FAMILY_PACKAGE_IMAGE="xpkg.upbound.io/upbound/provider-family-aws:v1.16.0"
export AWS_EC2_PACKAGE_IMAGE="xpkg.upbound.io/upbound/provider-aws-ec2:v1.16.0"
export AWS_RDS_PACKAGE_IMAGE="xpkg.upbound.io/upbound/provider-aws-rds:v1.16.0"
export AWS_KAFKA_PACKAGE_IMAGE="xpkg.upbound.io/upbound/provider-aws-kafka:v1.16.0"
export AWS_EKS_IAM_DEFAULT_ADMIN_ROLE="arn:aws:iam::123456789012:role/mydefaulteksadminrole"
export TARGET_CROSSPLANE_VERSION="1.17.2"
export UPTEST_CLOUD_CREDENTIALS="$(cat my-aws-creds.txt)"

make -C e2e/providerconfig-aws-e2e-test e2e

# Option 2: Build and publish providers first
export XPKG_REG_ORGS="index.docker.io/myrepo"
export VERSION="v1.4.0-test"
make VERSION=${VERSION} XPKG_REG_ORGS=${XPKG_REG_ORGS} providerconfig-e2e
```

**E2E Targets** (from `Makefile` lines 252-267):

```makefile
# Full pipeline: build → publish → test
providerconfig-e2e:
	$(MAKE) SUBPACKAGES="ec2 rds kafka config" build.all publish
	# ... runs e2e tests with published images

# Without publishing (if images already exist in registry)
providerconfig-e2e-nopublish:
	# ... runs e2e tests with existing images
```

**E2ETestCluster Lifecycle**:
1. Configuration package deployed to local Kind cluster
2. Claim `E2ETestCluster` created with parameters (region, EKS version, IAM roles, nodes)
3. Composition creates EKS cluster in AWS
4. Example managed resources deployed to remote cluster
5. Uptest validates resource status via conditions
6. Resources cleaned up via dependent deletion ordering (handled via Crossplane Usages)

**Test Datasources**: UPTEST_DATASOURCE_PATH allows injecting dynamic values
- Example: VPC CIDR, subnet ranges, ARNs, etc.
- Format: YAML with `name: value` pairs

---

### E2E Test Configuration in CI

**Job: `uptest` trigger** (`.github/workflows/uptest-trigger.yaml`)
- Triggered by: `@uptest /test-examples="path/to/examples/*.yaml"`
- Runs on: Ubuntu-Jumbo-Runner
- Permissions: Admin or Write on repository
- Steps:
  1. Parse comment to extract example list
  2. Create pending GitHub status check
  3. Run `make e2e` with example list
  4. Upload cluster dump on failure
  5. Cleanup managed resources
  6. Update GitHub status (success/failure)

**Example Invocation** (from CI):
```yaml
env:
  UPTEST_CLOUD_CREDENTIALS: ${{ secrets.UPTEST_CLOUD_CREDENTIALS }}
  UPTEST_EXAMPLE_LIST: ${{ needs.get-example-list.outputs.example_list }}
  UPTEST_TEST_DIR: ./_output/controlplane-dump
  UPTEST_DATASOURCE_PATH: .work/uptest-datasource.yaml
run: make e2e
```

---

## Examples & Fixtures

### Examples Organization

**Primary Examples** (`examples/`):
- Hand-crafted, curated resource examples
- Grouped by service/API group, then scope (cluster/namespaced), then API version
- Pattern: `examples/{service}/{scope}/{version}/{resource}.yaml`
- **2,364 files** covering 100+ AWS services

**Generated Examples** (`examples-generated/`):
- Auto-generated from Terraform AWS provider schema
- Mirrors structure of `examples/` for consistency
- **2,011 files**
- Used as fallback/reference when manual examples don't exist

### Example Manifest Structure

Each example YAML contains:
1. **apiVersion**: Service-specific (e.g., `ec2.aws.upbound.io/v1beta1`)
2. **kind**: Resource type (e.g., `Instance`, `Bucket`, `DBInstance`)
3. **metadata**: Name and labels
4. **spec**:
   - **forProvider**: AWS resource parameters
   - **providerConfigRef** (optional): Which ProviderConfig to use
5. **status**: Observed state (populated by provider)

**Example Services** (sampling):
- EC2: instance, vpc, securitygroup, subnet, routetable, etc.
- RDS: dbinstance, cluster, parametergroup, subnetgroup, etc.
- S3: bucket, object, bucketacl, bucketpolicy, etc.
- ACM: certificate, etc.
- API Gateway: api, stage, authorizer, etc.
- Lambda: function, layer, permission, etc.
- DynamoDB: table, etc.
- And 90+ more AWS services

### Examples Used in Testing

**Uptest Mechanism**:
1. Selects examples from `examples/` or `examples-generated/` directories
2. Applies to Crossplane-enabled cluster
3. Observes status until `Test` condition = `True` (default-conditions="Test")
4. Verifies against AWS backend
5. Cleans up resources

**Example Validation Script** (`scripts/check-examples.py`):
```bash
./scripts/check-examples.py package/crds examples
```
- Validates that example manifests match CRD schemas
- Checks for required fields
- Ensures examples use valid API versions

---

## CI/CD Testing Pipeline

### GitHub Actions Workflows

#### **ci.yml** - Main CI Pipeline

**Triggers**:
- Push to `main` or `release-*` branches
- Pull requests
- Manual workflow dispatch

**Jobs**:

| Job | Purpose | Condition | Notes |
|---|---|---|---|
| `detect-noop` | Skip if only docs/images changed | Always | Output: `noop` flag |
| `report-breaking-changes` | Check CRD schema changes | On modified CRDs | Uses `crddiff` tool |
| `lint` | Go linting with golangci-lint | `noop != true` | Large machine; cache analysis |
| `check-diff` | Verify generated code matches source | `noop != true` | Runs `make check-diff` |
| `unit-tests` | Run unit test suite | `noop != true` | `make -j2 test` |
| `local-deploy` | Test local provider deployment | `noop != true` | `make local-deploy` |
| `check-examples` | Validate example manifests | `noop != true` | `scripts/check-examples.py` |

**Key Targets**:

```makefile
# Line 145: Code coverage report (for Cobertura integration)
cobertura:
	@cat $(GO_TEST_OUTPUT)/coverage.txt | \
		grep -v zz_ | \
		$(GOCOVER_COBERTURA) > $(GO_TEST_OUTPUT)/cobertura-coverage.xml

# Line 326-340: CRD breaking changes detection
crddiff:
	@$(INFO) Checking breaking CRD schema changes
	@for crd in $${MODIFIED_CRD_LIST}; do \
		# ... uses crddiff to compare against base branch
```

#### **uptest-trigger.yaml** - E2E Test Trigger

**Triggers**: Issue comments on PRs with `/test-examples`

**Steps**:
1. **check-permissions**: Verify commenter is admin/write
2. **get-example-list**: Parse `/test-examples="path"` from comment
3. **uptest**: Run `make e2e` with example list
4. **Artifacts**: Upload cluster dump on failure

**Status Checks**: Creates GitHub status for each test run (unique hash per example set)

---

### Local Testing Workflow

**Development Flow**:

```bash
# 1. Setup submodules
make submodules

# 2. Lint and test locally (same as CI)
make vendor vendor.check
make lint
make -j2 test

# 3. Check generated code
make check-diff

# 4. Validate examples
./scripts/check-examples.py package/crds examples

# 5. Run e2e tests (requires AWS credentials)
export UPTEST_CLOUD_CREDENTIALS="$(cat ~/.aws/credentials)"
export UPTEST_EXAMPLE_LIST="examples/s3/cluster/v1beta1/bucket.yaml"
export UPTEST_DATASOURCE_PATH=".work/uptest-datasource.yaml"
make e2e

# 6. Test ProviderConfig scenarios (requires more setup)
cd e2e/providerconfig-aws-e2e-test
export AWS_FAMILY_PACKAGE_IMAGE="..."
export AWS_EKS_IAM_DEFAULT_ADMIN_ROLE="..."
make e2e
```

---

## Testing Commands & Targets

### Makefile Test Targets

| Target | Command | Purpose | Requirements |
|---|---|---|---|
| `test` | `go test ./...` | Run all unit tests | Go, vendor deps |
| `lint` | `golangci-lint run` | Lint Go code | golangci-lint, buildtagger |
| `check-diff` | Compare generated code | Validate `make generate` | Go, goimports |
| `cobertura` | Generate coverage XML | Cobertura reports | Test output files |
| `crddiff` | Check CRD breaking changes | Pre-CI check | crddiff tool, git |
| `schema-version-diff` | Check schema version changes | Native state versioning | Python script |
| `uptest` | `uptest e2e` with examples | E2E resource tests | Uptest, Kind, K8s cluster, AWS creds |
| `e2e` | `family-e2e` + `uptest` | Full E2E pipeline | All uptest deps + dynamic subpackage loading |
| `family-e2e` | Build & deploy per-API family | Smart E2E | Dynamic provider loading |
| `local-deploy` | Deploy locally built provider | Integration test | Kind, Crossplane, provider binary |
| `local-deploy.{service}` | Deploy specific service provider | Selective testing | Kind, Crossplane |
| `providerconfig-e2e` | Test auth scenarios | ProviderConfig validation | All uptest deps + EKS access + published images |
| `providerconfig-e2e-nopublish` | Test auth without publishing | ProviderConfig validation | All uptest deps + EKS access |

### Command Examples

```bash
# Unit tests only
make -j2 test

# Lint with build tags
make lint RUN_BUILDTAGGER=true GOLANGCI_LINT_VERSION=2.11.4

# Generate code and validate
make generate check-diff

# Single example E2E
UPTEST_EXAMPLE_LIST="examples/s3/cluster/v1beta1/bucket.yaml" \
UPTEST_CLOUD_CREDENTIALS="$(cat ~/.aws/credentials)" \
make uptest

# Multiple examples
UPTEST_EXAMPLE_LIST="examples/s3/cluster/v1beta1/bucket.yaml,examples/ec2/cluster/v1beta1/instance.yaml" \
UPTEST_CLOUD_CREDENTIALS="$(cat ~/.aws/credentials)" \
make uptest

# With dynamic datasource
UPTEST_DATASOURCE_PATH=".work/uptest-datasource.yaml" \
UPTEST_EXAMPLE_LIST="..." \
UPTEST_CLOUD_CREDENTIALS="..." \
make uptest

# E2E with automatic family detection
UPTEST_EXAMPLE_LIST="examples/ec2/cluster/v1beta1/instance.yaml,examples/rds/cluster/v1beta1/instance.yaml" \
UPTEST_CLOUD_CREDENTIALS="..." \
make e2e  # Automatically builds & deploys ec2 + rds providers

# ProviderConfig E2E
make VERSION=v1.4.0 providerconfig-e2e

# Local deployment testing
make local-deploy  # Deploys config provider
make local-deploy.ec2  # Deploys ec2 provider
```

---

## Test Infrastructure Components

### Test Tools & Versions

**Installed via Build System**:

```makefile
# Kubernetes Tools (from build/makelib/k8s_tools.mk)
KIND_VERSION = v0.30.0              # Local Kubernetes
UPTEST_VERSION = v2.2.0             # E2E test orchestration
KUSTOMIZE_VERSION = v5.3.0          # Kustomize
YQ_VERSION = v4.40.5                # YAML manipulation
CROSSPLANE_VERSION = 2.2.0          # Crossplane core
CROSSPLANE_CLI_VERSION = v2.2.0     # Crossplane CLI
CRDDIFF_VERSION = v0.12.1           # CRD diff tool

# Go tools
GOLANGCILINT_VERSION = 2.11.4       # Linter
BUILDTAGGER_VERSION = v0.12.0-rc.0.28.gdc5d6f3  # Build constraint tagging
TERRAFORM_VERSION = 1.5.5           # Schema generation
TERRAFORM_PROVIDER_VERSION = 6.34.0 # AWS provider schema
```

**Test Utilities** (Go packages):

```go
// From unit test imports
"github.com/crossplane/crossplane-runtime/v2/pkg/test"      // MockClient, fake resources
"github.com/crossplane/crossplane-runtime/v2/pkg/resource"  // Managed resource interfaces
"github.com/crossplane/upjet/v2/pkg/resource/fake"         // Upjet fake resources
"github.com/google/go-cmp/cmp"                               // Value comparison
```

### Test Environment Setup

**cluster/test/setup.sh** - Executed before E2E tests:

```bash
#!/usr/bin/env bash
set -aeuo pipefail

# Creates AWS credential secret from UPTEST_CLOUD_CREDENTIALS
# Supports multiple credential sets (DEFAULT, PEER for cross-account testing)

if [[ -n "${UPTEST_CLOUD_CREDENTIALS:-}" ]]; then
  eval "${UPTEST_CLOUD_CREDENTIALS}"

  # DEFAULT credentials
  if [[ -n "${DEFAULT:-}" ]]; then
    kubectl -n upbound-system create secret generic provider-secret \
      --from-literal=credentials="${DEFAULT}" --dry-run=client -o yaml | kubectl apply -f -
    
    # Create ProviderConfig (namespaced)
    kubectl apply -f - <<EOF
apiVersion: aws.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: provider-secret
      namespace: upbound-system
      key: credentials
EOF
    
    # Create ClusterProviderConfig
    kubectl apply -f - <<EOF
apiVersion: aws.m.upbound.io/v1beta1
kind: ClusterProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: provider-secret
      namespace: upbound-system
      key: credentials
EOF
  fi

  # PEER credentials (for cross-account testing)
  if [[ -n "${PEER:-}" ]]; then
    kubectl -n upbound-system create secret generic provider-secret-peer \
      --from-literal=credentials="${PEER}" --dry-run=client -o yaml | kubectl apply -f -
    
    # Create peer ProviderConfig and ClusterProviderConfig
  fi
fi
```

**e2e/providerconfig-aws-e2e-test/test/setup.sh** - E2E test setup:

```bash
#!/usr/bin/env bash
set -aeuo pipefail

# Waits for packages to be installed
kubectl wait configuration.pkg --all --for=condition=Healthy --timeout 5m
kubectl wait provider.pkg --all --for condition=Healthy --timeout 5m

# Creates credential secret
kubectl -n upbound-system create secret generic aws-creds \
  --from-literal=credentials="${UPTEST_CLOUD_CREDENTIALS}" \
  --dry-run=client -o yaml | kubectl apply -f -

# Waits for deployments
kubectl -n upbound-system wait --for=condition=Available deployment --all --timeout=5m

# Waits for XRDs to be established
kubectl wait xrd --all --for condition=Established

# Creates default ProviderConfig
kubectl apply -f - <<EOF
apiVersion: aws.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    secretRef:
      key: credentials
      name: aws-creds
      namespace: upbound-system
    source: Secret
EOF
```

---

## Testing Patterns & Conventions

### Unit Test Pattern: Table-Driven Tests

The codebase exclusively uses **table-driven tests** for unit testing:

```go
func TestSomething(t *testing.T) {
    type args struct { /* input fields */ }
    type want struct { /* expected output */ }
    
    cases := map[string]struct {
        reason string  // Optional: explain the test case
        args   args
        want   want
    }{
        "SuccessCase": {
            reason: "Should succeed when conditions are met.",
            args: args{ /* ... */ },
            want: want{ /* ... */ },
        },
        "ErrorCase": {
            reason: "Should return error when input is invalid.",
            args: args{ /* ... */ },
            want: want{
                err: expected.Error(),
            },
        },
    }
    
    for name, tc := range cases {
        t.Run(name, func(t *testing.T) {
            // Test execution
            got := FunctionUnderTest(tc.args)
            
            // Comparison using go-cmp
            if diff := cmp.Diff(tc.want, got, test.EquateErrors()); diff != "" {
                t.Errorf("FunctionUnderTest(...): -want, +got:\n%s", diff)
            }
        })
    }
}
```

**Benefits**:
- Clear, parameterized test cases
- Easy to add new cases
- Consistent error reporting with `cmp.Diff`
- `test.EquateErrors()` for idiomatic error comparison

### E2E Test Pattern: Example-Driven

```yaml
# examples/{service}/{scope}/{version}/{resource}.yaml
apiVersion: ec2.aws.upbound.io/v1beta1
kind: Instance
metadata:
  name: example-instance
spec:
  forProvider:
    ami: ami-0c55b159cbfafe1f0
    instanceType: t2.micro
    tags:
      Name: example
  providerConfigRef:
    name: default
  deletionPolicy: Delete
```

Uptest automatically:
1. Applies the manifest
2. Observes status.conditions["Test"]
3. Validates observed state matches AWS
4. Deletes on cleanup

### Testing Authentication Methods

**ProviderConfig Tests** cover:

1. **Secret-based credentials**
   ```yaml
   spec:
     credentials:
       source: Secret
       secretRef:
         name: aws-creds
         key: credentials
   ```

2. **IAM Role (IRSA)**
   ```yaml
   spec:
     credentials:
       source: IRSA
   ```

3. **WebIdentity**
   ```yaml
   spec:
     credentials:
       source: WebIdentity
   ```

4. **Cross-account role assumption**
   - Multiple ProviderConfigs with different IAM roles
   - Tested via `PEER` credentials in setup scripts

---

## Gaps & Observations

### Current Gaps

1. **Limited Unit Test Coverage**
   - Only 8 `_test.go` files found
   - Most testing deferred to E2E layer
   - Configuration converters and utilities receive tests, but controllers do not
   - Recommendation: Expand unit tests for controller logic, retry mechanisms, error handling

2. **No Visible Test Utilities Library**
   - Relying heavily on `crossplane-runtime/test` package
   - Could benefit from provider-specific test builders/factories
   - Limited test data fixtures (no testdata directories found)

3. **Limited Documentation of Test Data Injection**
   - `UPTEST_DATASOURCE_PATH` mechanism exists but underutilized
   - Dynamic resource dependencies (VPC IDs, subnet IDs, ARNs) require manual setup or environment variables
   - No examples of datasource YAML in repository

4. **E2E Test Timing**
   - Default timeout: 5400 seconds (90 minutes) for some E2E tests
   - No visible test timeouts per-resource type
   - Complex resource dependencies can cause cascading failures

5. **Cross-Account Testing**
   - `PEER` credentials support exists in setup scripts
   - But limited guidance on actual cross-account scenarios
   - ProviderConfig e2e tests reference cross-account but implementation details sparse

6. **Controller-Level Testing**
   - No visible integration tests for reconciliation loops
   - No mock AWS client testing at controller level
   - Relying on e2e for validation of controller behavior

### Notable Patterns & Strengths

1. **Comprehensive Example Coverage**
   - 2,364+ curated examples covering 100+ AWS services
   - Both cluster and namespaced scopes
   - Multiple API versions per resource (v1beta1, v1beta2)
   - Generated examples as fallback reference

2. **Sophisticated E2E Infrastructure**
   - Two-level Crossplane setup for ProviderConfig testing (Crossplane managing Crossplane)
   - Explicit dependency ordering via Crossplane Usages
   - Clean separation of concerns (networking, EKS, providers, test resources)

3. **Monolith Architecture Testing**
   - Dynamic subpackage selection (`SUBPACKAGES` variable)
   - Smart build system that deploys only needed providers
   - `family-e2e` target intelligently scans examples and builds necessary providers

4. **CI/CD Best Practices**
   - No-op detection to skip unnecessary jobs on doc-only changes
   - Breaking change detection for CRDs
   - Native schema version diffing
   - Efficient GitHub Actions integration with PR comments for e2e triggering

5. **Multi-Version Support**
   - Tests validate conversion between API versions (v1beta1 ↔ v1beta2)
   - Schema evolution testing via `crddiff`

### Recommendations

1. **Expand Unit Test Coverage**
   ```bash
   # Current: 8 test files
   # Target: 20-30 test files covering:
   - Controller reconciliation logic
   - Error retry mechanisms
   - Middleware/interceptor chains
   - Custom validators
   ```

2. **Create Test Utilities Library**
   ```go
   // e.g., internal/testing/builders.go
   type ResourceBuilder struct { /* ... */ }
   func NewEC2InstanceBuilder() *ResourceBuilder { /* ... */ }
   
   // Use in multiple tests
   ```

3. **Document Datasource YAML Schema**
   ```yaml
   # .work/uptest-datasource.yaml
   vpcId: vpc-xxxxx
   subnetIds:
     - subnet-xxxxx
     - subnet-yyyyy
   securityGroupId: sg-xxxxx
   ```

4. **Add Controller Integration Tests**
   ```go
   // internal/controller/{service}/{resource}_test.go
   // Test reconciliation loops without full AWS
   ```

5. **Expand ProviderConfig Test Scenarios**
   - Document IRSA setup
   - WebIdentity configuration
   - AssumeRole cross-account flows
   - Session tags and transitive permissions

---

## Summary

The provider-aws testing infrastructure is **mature and comprehensive**, featuring:

| Aspect | Coverage | Status |
|--------|----------|--------|
| **Unit Tests** | Config converters, utilities | ⭐⭐⭐ (limited scope) |
| **E2E Tests** | Full resource lifecycle | ⭐⭐⭐⭐⭐ (excellent) |
| **Example Coverage** | 2,364+ manifests | ⭐⭐⭐⭐⭐ (comprehensive) |
| **CI/CD** | Multi-job pipeline | ⭐⭐⭐⭐⭐ (robust) |
| **ProviderConfig Testing** | Auth scenarios | ⭐⭐⭐⭐ (advanced setup) |
| **Documentation** | Inline; some gaps | ⭐⭐⭐ (good Makefile docs) |
| **Controller Testing** | Not visible | ⭐⭐ (opportunity) |

**Key Strengths**:
- Sophisticated Uptest-based e2e framework
- Dynamic monolith testing via subpackage selection
- Two-level Crossplane setup for auth testing
- Breaking change detection and API versioning tests
- Excellent example coverage across AWS services

**Areas for Enhancement**:
- Unit test coverage expansion
- Controller-level integration tests
- Documented datasource patterns
- Test utility library
- Cross-account scenario documentation

---

## Appendix: File Paths & References

### Key Test Files

```
Unit Tests (8 files):
- config/cluster/autoscaling/config_test.go
- config/cluster/common/common_test.go
- config/cluster/elasticache/config_test.go
- config/cluster/rds/utils/engine_version_test.go
- config/namespaced/common/common_test.go
- config/namespaced/rds/utils/engine_version_test.go
- internal/clients/cache_test.go
- internal/clients/partitions_test.go

Test Setup & Configuration:
- cluster/test/setup.sh                           # Main e2e setup
- e2e/providerconfig-aws-e2e-test/test/setup.sh   # ProviderConfig e2e setup
- e2e/providerconfig-aws-e2e-test/Makefile        # ProviderConfig test targets
- e2e/providerconfig-aws-e2e-test/README.md       # Detailed instructions

CI/CD Workflows:
- .github/workflows/ci.yml                        # Main pipeline
- .github/workflows/uptest-trigger.yaml           # E2E trigger
- .github/workflows/publish-provider-packages.yaml
- .github/workflows/tag.yaml
- .github/workflows/stale.yml

Scripts:
- scripts/check-examples.py                       # Example validation
- scripts/family-test.py                          # Family provider testing
- scripts/version_diff.py                         # Schema version diffing
- scripts/tag.sh                                  # Build tagging

Example Directories:
- examples/                                       # 2,364 curated examples
- examples-generated/                             # 2,011 generated examples

Build Configuration:
- Makefile                                        # Main targets
- build/makelib/golang.mk                         # Go build rules
- build/makelib/k8s_tools.mk                      # K8s tool setup
- build/makelib/local.xpkg.mk                     # Local XPKG deployment
- build/makelib/controlplane.mk                   # Kind cluster management
```

### Command Quick Reference

```bash
# Local Development
make test                              # Unit tests
make lint                             # Linting
make local-deploy                     # Local provider deployment
make check-diff                       # Validate generated code

# E2E Testing
UPTEST_CLOUD_CREDENTIALS="..." \
UPTEST_EXAMPLE_LIST="..." \
make uptest                           # Simple E2E

UPTEST_CLOUD_CREDENTIALS="..." \
UPTEST_EXAMPLE_LIST="..." \
make e2e                              # Full E2E with auto provider detection

# ProviderConfig Testing
make providerconfig-e2e               # Build, publish, and test auth scenarios

# CI/CD Local Simulation
make vendor vendor.check
make lint RUN_BUILDTAGGER=true
make -j2 test
make check-diff
./scripts/check-examples.py package/crds examples
```

---

**Report Generated**: 2024 | **Provider**: upbound/provider-aws v2 | **Status**: Comprehensive & Production-Ready
