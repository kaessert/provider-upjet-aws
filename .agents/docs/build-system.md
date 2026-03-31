# Build System Reference — provider-upjet-aws

> Comprehensive guide to the Makefile-based build system, CI/CD pipelines, code
> generation, linting, and testing for `provider-aws`.

---

## Table of Contents

- [Overview](#overview)
- [Key Variables](#key-variables)
- [Make Targets](#make-targets)
  - [Code Generation](#code-generation)
  - [Building](#building)
  - [Testing](#testing)
  - [Validation](#validation)
  - [Local Development](#local-development)
  - [Publishing](#publishing)
  - [Utilities](#utilities)
- [Buildtagger System](#buildtagger-system)
- [Code Generation Flow](#code-generation-flow)
- [Tool Versions](#tool-versions)
- [CI/CD Workflows](#cicd-workflows)
- [Linting Configuration](#linting-configuration)
- [Gotchas & Tips](#gotchas--tips)

---

## Overview

The build system is a **440-line root Makefile** that delegates to reusable
includes from the `build/` git submodule (sourced from
[crossplane/build](https://github.com/crossplane/build)).

### Submodule Includes

The `build/makelib/` directory provides modular `.mk` files, loaded with
`-include` (silent on first run before submodule init):

| Include File          | Purpose                                       |
|-----------------------|-----------------------------------------------|
| `common.mk`          | Shared variables, `help`, `check-diff`, etc.  |
| `output.mk`          | Output directory setup (`_output/`)           |
| `golang.mk`          | Go build, test, lint, vendor targets          |
| `k8s_tools.mk`       | Install KIND, kubectl, kustomize, uptest, yq  |
| `imagelight.mk`      | Lightweight OCI image building                |
| `xpkg.mk`            | Crossplane package (xpkg) build & publish     |
| `local.xpkg.mk`      | Local xpkg deploy to KIND cluster             |
| `controlplane.mk`    | KIND cluster lifecycle (`controlplane.up/down`)|

On first `make` invocation, the `fallthrough` target initialises the submodule:

```bash
make          # runs: git submodule sync && git submodule update --init --recursive
```

---

## Key Variables

Defined at the top of the Makefile:

```makefile
PROVIDER_NAME            := aws
PROJECT_NAME             := provider-aws
PROJECT_REPO             := github.com/upbound/provider-aws/v2

# Terraform
TERRAFORM_VERSION        := 1.5.5
TERRAFORM_PROVIDER_VERSION := 6.34.0
TERRAFORM_PROVIDER_RELEASE := v6.34.0-upjet.1          # Upbound-patched release
TERRAFORM_PROVIDER_SOURCE := hashicorp/aws
TERRAFORM_PROVIDER_REPO  ?= https://github.com/hashicorp/terraform-provider-aws
TERRAFORM_DOCS_PATH      ?= website/docs/r

# Build targets
PLATFORMS                ?= linux_amd64 linux_arm64
BATCH_PLATFORMS          ?= linux_amd64,linux_arm64

# Subpackages — controls which provider binaries to build
SUBPACKAGES              ?= monolith                    # default: single monolith binary
# Override examples:
#   SUBPACKAGES="ec2 rds"     → build only ec2 and rds
#   SUBPACKAGES="*"           → build all service packages (auto-discovered from cmd/provider/)

# Go
GO_REQUIRED_VERSION      ?= <read from go.mod>
GOPRIVATE                 = github.com/upbound/*
GO_TEST_PARALLEL         := $(NPROCS / 2)               # half CPU count

# Registry
REGISTRY_ORGS            ?= xpkg.upbound.io/upbound
XPKG_REG_ORGS           ?= xpkg.upbound.io/upbound

# Buildtagger
RUN_BUILDTAGGER          ?= true
BUILDTAGGER_VERSION      ?= v0.12.0-rc.0.28.gdc5d6f3
```

---

## Make Targets

### Code Generation

| Command | Description |
|---------|-------------|
| `make generate.init` | Download Terraform provider binary + extract `config/schema.json` + clone TF provider docs |
| `make generate` | Run full Upjet code generation pipeline (types, controllers, examples, deepcopy) |
| `make check-diff` | Regenerate and verify no uncommitted diff — **CI gate** |

```bash
# Full regeneration workflow
make generate.init    # fetch schema (~15 MB JSON) + docs
make generate         # run Upjet pipeline → generates zz_* files
make check-diff       # verify everything is committed (CI uses this)
```

The `generate.init` target chains two sub-targets:
1. **`$(TERRAFORM_PROVIDER_SCHEMA)`** — Installs Terraform + the AWS provider
   binary, runs `terraform providers schema -json` to produce `config/schema.json`.
2. **`pull-docs`** — Shallow-clones the Terraform AWS provider repo to extract
   resource documentation from `website/docs/r/`.

### Building

| Command | Description |
|---------|-------------|
| `make build` | Compile provider binaries for configured `SUBPACKAGES` |
| `make build.all` | Build all subpackages |
| `SUBPACKAGES="ec2 rds" make build` | Build specific service packages |
| `make build-provider.ec2,rds` | Build specific providers (comma-separated) |

```bash
# Build the monolith (default)
make build

# Build specific service providers
SUBPACKAGES="ec2 s3 rds" make build

# Build all individual service providers
SUBPACKAGES="*" make build

# Build via the provider-specific target (comma-separated)
make build-provider.ec2,rds,config
```

The `build.init` target runs before any build and does two things:
1. Ensures the Crossplane CLI (`$(CROSSPLANE_CLI)`) is available.
2. Runs `kustomize-crds` — copies CRDs to `_output/package/` and applies
   Kustomize transforms.

### Testing

| Command | Description |
|---------|-------------|
| `make test` | Run Go unit tests (parallelism: `NPROCS / 2`) |
| `make uptest` | E2E tests via Uptest framework against real AWS resources |
| `make e2e` | Alias for `family-e2e` — full family provider E2E |
| `make family-e2e` | Parse examples → build required providers → deploy → run uptest |
| `make providerconfig-e2e` | E2E tests for auth methods against EKS |
| `make cobertura` | Generate Cobertura coverage XML (filters out `zz_` generated files) |

```bash
# Unit tests
make -j2 test

# E2E: requires AWS credentials and example list
export UPTEST_CLOUD_CREDENTIALS="DEFAULT='$(cat ~/.aws/credentials)'"
export UPTEST_EXAMPLE_LIST="examples/ec2/v1beta1/instance.yaml"
make e2e

# Provider config E2E (builds ec2, rds, kafka, config; publishes; tests on EKS)
make providerconfig-e2e

# Coverage report
make test
make cobertura    # → _output/tests/coverage/cobertura-coverage.xml
```

**Uptest environment variables:**
- `UPTEST_EXAMPLE_LIST` — Comma-separated list of example manifests to test.
- `UPTEST_CLOUD_CREDENTIALS` — AWS credentials. Supports multiple named
  credential sets (`DEFAULT`, `PEER`).
- `UPTEST_DATASOURCE_PATH` — Optional data source for dynamic value injection.

### Validation

| Command | Description |
|---------|-------------|
| `make check-diff` | Regenerate code and fail if working tree is dirty — **CI gate** |
| `make lint` | Run golangci-lint with buildtagger (see [Buildtagger](#buildtagger-system)) |
| `make crddiff` | Detect breaking CRD OpenAPI v3 schema changes against base branch |
| `make schema-version-diff` | Detect Terraform native state schema version changes |

```bash
# Lint with build tags (standard CI flow)
RUN_BUILDTAGGER=true make lint

# Check for breaking API changes (requires GITHUB_BASE_REF and MODIFIED_CRD_LIST)
GITHUB_BASE_REF=main MODIFIED_CRD_LIST="package/crds/ec2.aws.upbound.io_instances.yaml" make crddiff

# Check schema version drift
GITHUB_BASE_REF=main make schema-version-diff
```

### Local Development

| Command | Description |
|---------|-------------|
| `make run` | Build and run provider locally out-of-cluster |
| `make local-deploy` | Build config provider + deploy to local KIND cluster |
| `make local-deploy.ec2,rds` | Deploy specific service packages to local cluster |
| `make controlplane.up` | Create a local KIND cluster with Crossplane installed |
| `make controlplane.down` | Tear down the local KIND cluster |

```bash
# Quick local run (out-of-cluster, no KIND needed)
make run
# → runs: UPBOUND_CONTEXT="local" ./monolith --debug --certs-dir=""

# Full local deployment to KIND
make local-deploy                    # builds + deploys config provider
make local-deploy.ec2,rds            # add ec2 and rds providers

# Deploy everything
SUBPACKAGES="*" make build
make local-deploy.ec2,s3,rds,lambda
```

The `local-deploy` sequence:
1. `controlplane.up` — Creates KIND cluster, installs Crossplane (v2.2.0) in
   `upbound-system` namespace.
2. Patches the RBAC manager to use `index.docker.io` as registry.
3. For each API package: builds xpkg, loads into cluster, waits for `Healthy`.

### Publishing

| Command | Description |
|---------|-------------|
| `make publish` | Build OCI images and push packages to registry |
| `make build.all publish` | Build all subpackages and publish |

```bash
# Publish all subpackages
SUBPACKAGES="*" make build.all publish

# Publish specific services
SUBPACKAGES="ec2 rds config" make build.all publish
```

Default registry: `xpkg.upbound.io/upbound`

### Utilities

| Command | Description |
|---------|-------------|
| `make submodules` | Sync and update git submodules |
| `make vendor` | Download Go module dependencies |
| `make vendor.check` | Verify vendor directory matches go.sum |
| `make kustomize-crds` | Aggregate CRDs via Kustomize into `_output/package/` |
| `make go.cachedir` | Print Go build cache location |
| `make go.mod.cachedir` | Print Go module cache location |
| `make go.lint.analysiskey` | Print cache key for golangci-lint analysis |
| `make print-subpackages` | Print the resolved list of subpackages |
| `make help` | Show all available targets |

---

## Buildtagger System

The provider codebase is massive (~2,300+ controllers). Compiling everything at
once during linting causes OOM. The **buildtagger** tool solves this:

```makefile
RUN_BUILDTAGGER ?= true
BUILDTAGGER_VERSION ?= v0.12.0-rc.0.28.gdc5d6f3
```

### How It Works

1. **`lint.init → build-lint-cache`**: Before linting, `scripts/tag.sh`
   downloads buildtagger and adds Go build constraints (e.g.,
   `//go:build provider_ec2`) to source files. This limits compilation scope.
2. **Analysis cache warm-up**: A preliminary lint pass runs with the small
   `account` API group and `staticcheck` only, to populate the golangci-lint
   analysis cache without OOM risk.
3. **Full lint**: `make lint` runs golangci-lint with `--build-tags all`.
4. **`lint.done → delete-build-tags`**: After linting, build constraints are
   removed and deepcopy tags are restored.

```bash
# The full lint lifecycle (automatic):
make lint
#  1. lint.init → build-lint-cache (tag files, warm cache)
#  2. go.lint (actual lint run)
#  3. lint.done → delete-build-tags (clean up tags)
```

> **Warning:** Never manually commit build constraint tags added by buildtagger.
> They are transient and cleaned up by `delete-build-tags`.

---

## Code Generation Flow

### Pipeline Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌────────────────────┐
│ make generate.init │ → │  make generate    │ → │  make check-diff   │
│                   │   │                    │   │  (CI verification)  │
│ 1. Install TF    │   │ 1. Load TF schema │   │                    │
│ 2. Install AWS   │   │ 2. Load configs   │   │  git diff --exit   │
│    provider       │   │ 3. Run pipeline   │   │                    │
│ 3. Extract schema│   │ 4. Post-process   │   │                    │
│ 4. Pull docs     │   │                    │   │                    │
└─────────────────┘     └──────────────────┘     └────────────────────┘
```

### Step-by-Step

1. **`make generate.init`**
   - Downloads Terraform `1.5.5` from Upbound's fork.
   - Downloads the Upbound-patched AWS provider (`v6.34.0-upjet.1`).
   - Runs `terraform init` + `terraform providers schema -json` →
     `config/schema.json` (~15 MB, all 6,000+ AWS resources).
   - Clones TF provider docs from `website/docs/r/`.

2. **`cmd/generator/main.go`** — Upjet pipeline entry point:
   - Parses arguments: `repo-root`, `--skipped-resources-csv`,
     `--generated-resource-list`.
   - Initialises both **framework** and **SDK** Terraform providers via
     `xpprovider.GetProvider()`.
   - Loads **cluster-scoped** config (`config.GetProvider()`) and
     **namespace-scoped** config (`config.GetProviderNamespaced()`).
   - Calls `pipeline.Run(pc, pns, absRootDir)` — the Upjet code generation
     engine.
   - Post-processes: converts singleton lists to embedded objects in
     `examples-generated/`.
   - Dumps `config/generated.lst` (JSON array of generated resources) and
     optionally a skipped resources CSV.

3. **Generated Artifacts** (all prefixed `zz_`):
   - `apis/**/zz_types.go` — Go types for each CRD.
   - `apis/**/zz_groupversion_info.go` — GroupVersion registration.
   - `internal/controller/**/zz_controller.go` — Reconciliation controllers.
   - `apis/**/zz_generated.deepcopy.go` — DeepCopy implementations.
   - `apis/**/zz_generated.conversion*.go` — Conversion functions.
   - `apis/**/zz_generated.managed*.go` — Managed resource interfaces.
   - `examples-generated/` — Example manifests.
   - `config/generated.lst` — Full list of generated resource names.

4. **Configuration** — `config/` directory:
   - `config/provider.go` — Root config, registers all resource groups.
   - `config/<service>/config.go` — Per-service resource customizations
     (references, overrides, sensitive fields, late-initialization).
   - `config/schema.json` — Raw Terraform provider schema.

---

## Tool Versions

All tool versions are pinned in the Makefile:

| Tool              | Variable                | Version                          |
|-------------------|-------------------------|----------------------------------|
| Terraform         | `TERRAFORM_VERSION`     | `1.5.5`                         |
| AWS TF Provider   | `TERRAFORM_PROVIDER_VERSION` | `6.34.0`                   |
| AWS TF Release    | `TERRAFORM_PROVIDER_RELEASE` | `v6.34.0-upjet.1`          |
| KIND              | `KIND_VERSION`          | `v0.30.0`                       |
| Uptest            | `UPTEST_VERSION`        | `v2.2.0`                        |
| Kustomize         | `KUSTOMIZE_VERSION`     | `v5.3.0`                        |
| yq                | `YQ_VERSION`            | `v4.40.5`                       |
| Crossplane        | `CROSSPLANE_VERSION`    | `2.2.0`                         |
| Crossplane CLI    | `CROSSPLANE_CLI_VERSION`| `v2.2.0`                        |
| CRD Diff          | `CRDDIFF_VERSION`       | `v0.12.1`                       |
| golangci-lint     | `GOLANGCILINT_VERSION`  | `2.11.4`                        |
| Buildtagger       | `BUILDTAGGER_VERSION`   | `v0.12.0-rc.0.28.gdc5d6f3`     |
| Go                | `GO_REQUIRED_VERSION`   | Read dynamically from `go.mod`  |

---

## CI/CD Workflows

All workflows live in `.github/workflows/`.

### 1. `ci.yml` — Main CI Pipeline

**Triggers:** Push to `main` / `release-*`, pull requests, manual dispatch.

**Job graph:**

```
detect-noop
  ├── report-breaking-changes   (CRD diff + schema version diff)
  ├── lint                      (golangci-lint with buildtagger)
  ├── check-diff                (regenerate + verify clean tree)
  ├── unit-tests                (make -j2 test)
  ├── local-deploy              (build + deploy config to KIND)
  └── check-examples            (validate example manifests against CRDs)
```

| Job | Runner |
|-----|--------|
| `detect-noop` | `ubuntu-latest` |
| `report-breaking-changes` | `ubuntu-latest` |
| `lint` | Oracle 16-CPU / Ubuntu Jumbo (16 CPU, 64 GB) |
| `check-diff` | Oracle 16-CPU / Ubuntu Jumbo |
| `unit-tests` | Oracle 16-CPU / Ubuntu Jumbo |
| `local-deploy` | `ubuntu-latest` |
| `check-examples` | `ubuntu-latest` |

**Key details:**
- `detect-noop` skips CI for doc-only changes (`**.md`, `**.png`, `**.jpg`).
- `lint` caches the golangci-lint analysis directory between runs (keyed by
  `go.sum` hash, rotated weekly).
- `check-diff` uploads a `skipped_resources.csv` artifact.
- `unit-tests` runs with `-j2` parallelism.

### 2. `publish-provider-packages.yaml` — Manual Publish

**Trigger:** `workflow_dispatch` with inputs for subpackages, size, concurrency,
version, go-version.

Calls a reusable workflow from `crossplane-contrib/provider-workflows` to build
OCI images and push to `xpkg.upbound.io/upbound`.

**Runner:** Oracle 16-CPU / 64 GB.

### 3. `tag.yaml` — Manual Tag Creation

**Trigger:** `workflow_dispatch` with `version` and `message` inputs.

Creates a git tag using `negz/create-tag`. Runner: `ubuntu-latest`.

### 4. `uptest-trigger.yaml` — E2E on PR Comment

**Trigger:** `issue_comment` (created) with `/test-examples` keyword.

Validates commenter permissions (admin/write), extracts example list from PR,
then triggers Uptest E2E.

### 5. `stale.yml` — Stale Issue/PR Cleanup

**Trigger:** Daily cron (`04:15 UTC`) + manual dispatch.

Uses `actions/stale@v10`:
- **90 days** idle → labelled stale.
- **14 more days** (104 total) → auto-closed.

Runner: `ubuntu-24.04`.

---

## Linting Configuration

Config file: `.golangci.yml` (golangci-lint v2 format).

### Global Settings

```yaml
version: "2"
run:
  timeout: 90m
  concurrency: 1
```

### Enabled Linters

| Linter       | Purpose                                         |
|--------------|--------------------------------------------------|
| `errcheck`   | Unchecked error returns                          |
| `govet`      | Go vet checks (shadow disabled)                  |
| `gocyclo`    | Cyclomatic complexity (min: 10)                  |
| `gocritic`   | Opinionated style checks (performance tag)       |
| `goconst`    | Repeated string constants (min-len: 3, min-occurrences: 5) |
| `prealloc`   | Slice preallocation suggestions (simple + range) |
| `revive`     | Extensible linter (confidence: 0.8)              |
| `staticcheck`| Merged gosimple + stylecheck + staticcheck       |
| `unconvert`  | Unnecessary type conversions                     |
| `unused`     | Unused code detection                            |
| `misspell`   | Spelling errors in comments/strings              |
| `nakedret`   | Naked returns in functions > 30 lines            |

Preset: `bugs`

### Enabled Formatters

| Formatter    | Config                                           |
|--------------|--------------------------------------------------|
| `gofmt`      | With `-s` simplification                         |
| `goimports`  | Local prefix: `github.com/upbound/provider-aws/v2` |

### Exclusion Rules

| Scope | Rule |
|-------|------|
| Generated files (`zz_.*go$`) | Excluded from all path-based linters |
| Generated files (`zz_.+\.go$`) | `misspell` disabled (group names trigger false positives) |
| Test files (`_test(ing)?\.go`) | `gocyclo`, `errcheck`, `dupl`, `gosec`, `scopelint`, `unparam` relaxed |
| Test files (`_test\.go`) | `gocritic` `unnamedResult`/`exitAfterDefer` suppressed |
| All files | `hugeParam`/`rangeValCopy` gocritic warnings suppressed |
| All files | `SA3000` staticcheck (TestMain os.Exit) suppressed |
| All files | `G101`/`G104` gosec (false-positive secrets/errors) suppressed |

### Issues Settings

```yaml
issues:
  new: false                   # check all code, not just changes
  max-issues-per-linter: 0     # no limit
  max-same-issues: 0           # no limit
```

---

## Gotchas & Tips

### Memory Requirements
- **16 GB+ RAM recommended.** Code generation and linting load thousands of
  packages simultaneously.
- CI uses 16-CPU / 64-GB runners for lint, check-diff, and unit-tests.
- The buildtagger system exists specifically to prevent OOM during linting.

### `check-diff` Is a CI Gate
Always regenerate after changing anything in `config/`:
```bash
make generate.init
make generate
make check-diff       # must exit 0 or CI fails
```

### Buildtagger Tags Are Transient
- `scripts/tag.sh` adds `//go:build` constraints before lint.
- `delete-build-tags` removes them after.
- **Never commit files with buildtagger-added constraints.** If you see
  `//go:build provider_xxx` in generated files after a failed lint, run:
  ```bash
  EXTRA_BUILDTAGGER_ARGS="--delete" RESTORE_DEEPCOPY_TAGS="true" ./scripts/tag.sh
  ```

### `go.mod` / `go.sum` May Appear Empty
They can be managed externally. Always run `make vendor vendor.check` before
building.

### Subpackages Wildcard
`SUBPACKAGES="*"` auto-discovers all service directories under `cmd/provider/`
(excluding `monolith`):
```bash
# See what packages are available
make print-subpackages
SUBPACKAGES="*" make print-subpackages
```

### First-Time Setup
```bash
git clone --recursive <repo-url>   # or:
make submodules                     # init build/ submodule
make vendor                         # download Go dependencies
make generate.init                  # fetch TF schema + docs
make generate                       # generate code
make build                          # compile
```

### Useful Debug Patterns
```bash
# Check what Go cache dir is in use (build submodule overrides XDG_CACHE_HOME)
make go.cachedir

# Get the current lint analysis cache key
make go.lint.analysiskey

# Run provider locally with debug logging
make run
# → UPBOUND_CONTEXT="local" ./_output/bin/linux_amd64/monolith --debug --certs-dir=""
```
