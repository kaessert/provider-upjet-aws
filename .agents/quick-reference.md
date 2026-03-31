# Quick Reference: Tools & Setup

## One-Time Setup

```bash
# 1. Clone with submodules
git clone --recurse-submodules https://github.com/upbound/provider-upjet-aws.git
cd provider-upjet-aws

# 2. Initialize submodules (fallback if not cloned with --recurse-submodules)
git submodule sync
git submodule update --init --recursive

# 3. Verify Go version
go version  # Must be 1.25.8+

# 4. Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# 5. Download Go modules
make vendor vendor.check
```

## Pre-Build Checklist

```bash
# Check system
go version           # Must output 1.25.8+
which curl           # Must exist
which unzip          # Must exist
which git            # Must exist
df -h /              # At least 10GB free

# Verify submodules
test -f build/makelib/common.mk && echo "✓ build submodule present" || echo "✗ build submodule MISSING"

# Verify Go.mod
head -2 go.mod       # Must show module and go version
```

## Essential Commands for Development

### Code Generation
```bash
# Download Terraform schema (first time, ~5min)
make generate.init

# Generate all types and controllers
make generate

# Verify generated files match source (CI gate)
make check-diff
```

### Building
```bash
# Build all services
make build

# Build specific services (faster)
make build SUBPACKAGES="config ec2"

# Build and push to Docker
DOCKERHUB_ORG=myname make build.all publish

# Run provider locally
make run
```

### Testing
```bash
# Unit tests (fast)
make test

# Lint code (requires buildtagger download, ~5min first time)
make lint

# E2E test (requires Kind, Crossplane, AWS creds, ~10min)
make uptest UPTEST_EXAMPLE_LIST="examples/s3/cluster/v1beta1/bucket.yaml"

# Local deployment (sets up Kind + Crossplane)
make local-deploy

# Full E2E (family test)
make family-e2e
```

### Validation (CI gates)
```bash
# Check for breaking CRD changes
make crddiff

# Check for native schema version changes
make schema-version-diff

# Verify no generated files are stale
make check-diff
```

## Environment Variables

### For Building
```bash
export SUBPACKAGES="config,ec2,s3"        # Which services to build
export DOCKERHUB_ORG=myusername           # For publishing to Docker Hub
export XPKG_REG_ORGS=index.docker.io/myusername  # Package registry
```

### For E2E Testing
```bash
# AWS credentials (required for actual AWS calls)
export UPTEST_CLOUD_CREDENTIALS='[default]
aws_access_key_id=AKIA...
aws_secret_access_key=...'

# Example to test
export UPTEST_EXAMPLE_LIST="examples/s3/cluster/v1beta1/bucket.yaml,examples/ec2/cluster/v1beta1/instance.yaml"

# Optional: dynamic value injection
export UPTEST_DATASOURCE_PATH=/path/to/datasource.json
```

### Build Optimization
```bash
export RUN_BUILDTAGGER=true           # Enable build tagging (default)
export SKIP_LINTER_ANALYSIS=false     # Skip cache build (CI only)
export GOGC=50                        # Reduce memory (slower but uses less RAM)
export GO_TEST_PARALLEL=4             # Limit test parallelism
```

## Tool Versions at a Glance

| Tool | Version | Installed How |
|------|---------|---|
| Go | **1.25.8+** | Manual (golang.org) |
| Terraform | **1.5.5** | Auto (`make generate.init`) |
| Terraform AWS | **6.34.0** | Auto (`make generate.init`) |
| golangci-lint | **2.11.4** | Auto (`make lint`) |
| buildtagger | **v0.12.0-rc.0.28.gdc5d6f3** | Auto (`make lint`) |
| Kind | **v0.30.0** | Auto (`make local-deploy`) |
| Uptest | **v2.2.0** | Auto (`make uptest`) |
| Kustomize | **v5.3.0** | Auto (`make kustomize-crds`) |
| YQ | **v4.40.5** | Auto (`make kustomize-crds`) |
| Crossplane CLI | **v2.2.0** | Auto (`make local-deploy`) |
| crddiff | **v0.12.1** | Auto (go run) |

## Where Tools Are Stored

```
~/.cache/golangci-lint/        # Linter cache
~/.cache/go-build/             # Go build cache
.work/tools/                   # Downloaded binaries (Terraform, Kind, etc.)
.work/terraform/               # Terraform working directory
build/                         # Build system submodule
```

## Disk Space Usage

| Phase | Size | Time |
|-------|------|------|
| After `git clone` | ~500MB | — |
| After `make vendor` | ~2.5GB | 1-2min |
| After `make generate.init` | ~3GB | 5min (network-dependent) |
| After first `make build` | ~5GB | 15min |
| Full `.work/` tools | ~1GB | (cached) |
| **Total typical setup** | ~10GB | 30-45min (first time) |

## Memory Usage During Build

| Operation | Peak RAM | Notes |
|-----------|----------|-------|
| `make lint` | 4-6 GB | Most memory-intensive |
| `make generate` | 3-5 GB | Code generation |
| `make build` | 2-3 GB | Per-service compilation |
| `make test` | 2-3 GB | Unit tests |
| **Recommended minimum** | 16 GB | Required per CLAUDE.md |

## Recommended Workflow

### Initial Setup (30-45 min)
```bash
cd provider-upjet-aws
make vendor vendor.check      # Download dependencies
make generate.init            # Download Terraform schema
make check-diff              # Verify everything works
```

### Daily Development (per feature)
```bash
# 1. Make changes to config/externalname.go or config/cluster/{service}/config.go

# 2. Regenerate types and controllers
make generate
make check-diff

# 3. Run tests
make lint
make test

# 4. Test locally if needed
make local-deploy
make uptest UPTEST_EXAMPLE_LIST="examples/..."

# 5. Verify no breakage
make crddiff  # If you modified CRDs
```

### Before Committing
```bash
git status                     # Review changes
make check-diff               # Ensure generated files match
make lint                     # No lint errors
make test                     # All tests pass
git diff --stat               # Review scope
```

### Before Pushing
```bash
# These run in CI, but useful to verify locally
make check-diff               # CI gate 1
make lint                     # CI gate 2
make test                     # CI gate 3
make crddiff                  # CI gate 4 (if CRDs changed)

# Check the workflow can complete
make local-deploy             # Optional but recommended
```

## When Something Goes Wrong

### "Makefile not found" or "command not found: make"
```bash
# Install make
# macOS
brew install make

# Ubuntu/Debian
sudo apt-get install make

# RHEL/CentOS
sudo yum install make
```

### "go: command not found"
```bash
# Install Go 1.25.8 from https://golang.org/dl/
# Or use asdf:
asdf plugin add golang
asdf install golang 1.25.8
asdf global golang 1.25.8
```

### "build submodule not found"
```bash
git submodule sync
git submodule update --init --recursive
# Or re-clone:
git clone --recurse-submodules https://github.com/upbound/provider-upjet-aws.git
```

### "Out of memory during make lint"
```bash
# Option 1: Increase swap
# Option 2: Reduce parallelism
make lint GOGC=50 GO_TEST_PARALLEL=1

# Option 3: Lint specific packages
make lint LINT_ARGS="./config/..."
```

### "Terraform provider download fails"
```bash
# Check network
curl -I https://github.com/upbound/terraform-provider-aws/releases

# Clear cached download
rm -rf .work/tools/terraform-provider-aws*

# Retry
make generate.init
```

### "Generated files are stale"
```bash
# Regenerate everything
make generate
make check-diff

# If still failing, verify schema
ls -la config/schema.json
# If missing:
make generate.init
```

## Common Make Variables

Override on command line:
```bash
make build SUBPACKAGES="ec2 rds"        # Build specific services
make lint GOGC=50                       # Reduce GC pressure
make test -j4                           # Parallel test execution
make uptest UPTEST_CLOUD_CREDENTIALS='...'  # E2E with creds
```

## Resource Files

- **Config**: `config/externalname.go`, `config/cluster/{svc}/config.go`
- **Generated types**: `apis/cluster/{svc}/{version}/zz_*.go`
- **Controllers**: `internal/controller/cluster/zz_*.go`
- **Examples**: `examples/{svc}/cluster/v1beta1/*.yaml`
- **CRDs**: `package/crds/*.yaml`
- **Build scripts**: `Makefile`, `build/makelib/`, `scripts/`, `hack/`

## GitHub Actions / CI Reference

PR/push triggers: `lint` → `check-diff` → `unit-tests` → `local-deploy` → `check-examples`

To test locally:
```bash
# Lint phase
make lint

# Check-diff phase
make check-diff

# Unit tests phase
make test

# Local deploy phase (requires Docker)
make local-deploy

# Check examples phase (Python)
python3 scripts/check-examples.py
```

## Get Help

- **Build issues**: Check `.agents/build-dependencies.md` (this dir)
- **Architecture**: Read `CLAUDE.md` and `.agents/docs/architecture.md`
- **Adding resources**: See `.agents/docs/config-patterns.md`
- **Testing**: See `.agents/docs/testing.md`
- **Upjet docs**: https://github.com/crossplane/upjet/tree/main/docs
- **Crossplane**: https://crossplane.io
