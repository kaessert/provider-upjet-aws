# Build Dependencies Summary

## Overview

This is a **Crossplane provider** built with Upjet that exposes 1000+ AWS resources. The build system is complex due to code generation from Terraform AWS provider schema, but most tools are auto-downloaded.

---

## Essential Checklist

### What You Must Install Manually

Only **one requirement** cannot be auto-downloaded:

- ✅ **Go 1.25.8 or later** — Download from [golang.org](https://golang.org/dl/) or use asdf

Everything else (Terraform, Kind, Uptest, golangci-lint, buildtagger, Kustomize, YQ, etc.) is downloaded automatically by Make targets.

### What Gets Auto-Downloaded

| Tool | Version | Downloaded On |
|------|---------|---|
| Terraform CLI | 1.5.5 | `make generate.init` |
| Terraform AWS Provider | 6.34.0 | `make generate.init` |
| golangci-lint | 2.11.4 | `make lint` |
| buildtagger | v0.12.0-rc.0.28.gdc5d6f3 | `make lint` |
| Kind | v0.30.0 | `make local-deploy` |
| Uptest | v2.2.0 | `make uptest` |
| Kustomize | v5.3.0 | `make kustomize-crds` |
| YQ | v4.40.5 | `make kustomize-crds` |
| Crossplane CLI | v2.2.0 | `make local-deploy` |
| crddiff | v0.12.1 | `make crddiff` (go run) |

---

## Versions Reference

| Component | Version | Source |
|-----------|---------|--------|
| **Go** | **1.25.8+** (required) | `go.mod` |
| **Terraform** | **1.5.5** | `Makefile` line 13 |
| **Terraform AWS Provider** | **6.34.0** | `Makefile` line 14 |
| **golangci-lint** | **2.11.4** | `Makefile` line 55 |
| **buildtagger** | **v0.12.0-rc.0.28.gdc5d6f3** | `Makefile` line 62 |
| **Kind** | **v0.30.0** | `Makefile` line 82 |
| **Uptest** | **v2.2.0** | `Makefile` line 83 |
| **Kustomize** | **v5.3.0** | `Makefile` line 84 |
| **YQ** | **v4.40.5** | `Makefile` line 85 |
| **Crossplane CLI** | **v2.2.0** | `Makefile` line 87 |
| **crddiff** | **v0.12.1** | `Makefile` line 88 |

---

## System Requirements

### Minimum
- **OS**: Linux (recommended), macOS, or WSL
- **RAM**: 16GB
- **Disk**: 10GB free
- **Go**: 1.25.8+
- **CLI tools**: `git`, `curl`, `unzip`

### Recommended
- **RAM**: 32GB+
- **Disk**: 50GB+
- **CPU**: 8+ cores
- **Docker**: For container builds (optional)
- **Python 3**: For validation scripts

---

## Setup Time

| Step | Duration | What Happens |
|------|----------|---|
| Clone repo | 2 min | ~500 MB download |
| `make vendor` | 2-3 min | Download Go dependencies (~2.5 GB) |
| `make generate.init` | 5 min | Download Terraform schema (~500 MB) |
| First `make check-diff` | 15 min | Generate types, controllers, CRDs |
| First `make lint` | 10-15 min | Download buildtagger, build analysis cache |
| **Total first build** | **30-45 min** | Everything cached afterward |

---

## How It Works

```
Source Code
    ↓
config/externalname.go (manual)
config/cluster/{svc}/config.go (manual)
    ↓
make generate
    ↓
cmd/generator/main.go
(reads Terraform schema from config/schema.json)
    ↓
Generate APIs, Controllers, Examples
    ↓
apis/cluster/**/*types.go (auto)
internal/controller/**/*.go (auto)
package/crds/*.yaml (auto)
examples-generated/**/*.yaml (auto)
```

---

## Command Flow for Common Tasks

### Setting Up for Development
```
make vendor vendor.check
→ downloads Go dependencies

make generate.init
→ downloads Terraform, AWS provider, generates schema

make check-diff
→ runs full code generation, verifies output
```

### Before Committing
```
make lint
→ runs golangci-lint (downloads buildtagger if needed)

make test
→ runs unit tests

make check-diff
→ verifies generated files are current

make crddiff  (if CRDs changed)
→ checks for breaking API changes
```

### For Local Testing
```
make local-deploy
→ builds provider, creates local Kind cluster, deploys Crossplane

make uptest UPTEST_EXAMPLE_LIST="examples/..."
→ runs E2E tests against actual AWS (requires credentials)
```

---

## Where Files Go

| Directory | Purpose |
|-----------|---------|
| `.work/` | Downloaded tools and build artifacts |
| `.work/tools/` | Terraform, Kind, Uptest binaries |
| `.work/terraform/` | Terraform working directory |
| `vendor/` | Go modules cache |
| `config/schema.json` | Terraform AWS provider schema |
| `apis/cluster/` | Generated cluster-scoped K8s types |
| `apis/namespaced/` | Generated namespace-scoped K8s types |
| `internal/controller/` | Generated Kubernetes controllers |
| `package/crds/` | Custom Resource Definition YAML files |
| `examples-generated/` | Auto-generated example manifests |
| `_output/` | Build artifacts (during CI) |

---

## Environment Variables You Might Use

### For Building
```bash
SUBPACKAGES="ec2,rds"                   # Which AWS services to build
DOCKERHUB_ORG=myusername               # Your Docker Hub account
XPKG_REG_ORGS=index.docker.io/username # Package registry
RUN_BUILDTAGGER=true                   # Use build tags (default)
```

### For Testing
```bash
UPTEST_EXAMPLE_LIST="examples/s3/cluster/v1beta1/bucket.yaml"
UPTEST_CLOUD_CREDENTIALS='[default]\naws_access_key_id=...'
UPTEST_DATASOURCE_PATH=/path/to/datasource.json
```

### For CI/Optimization
```bash
GOGC=50                        # Reduce memory during build
GO_TEST_PARALLEL=4             # Limit test parallelism
SKIP_LINTER_ANALYSIS=false     # Skip cache rebuild (CI only)
```

---

## Key Files & Locations

### Configuration (Hand-Written)
- `config/externalname.go` — AWS resource ID mapping (3,676 lines)
- `config/overrides.go` — Global resource overrides
- `config/groups.go` — Terraform→Crossplane group/kind mapping
- `config/cluster/{service}/config.go` — Per-service configuration

### Build System
- `Makefile` — Main orchestrator (440 lines)
- `build/` — Git submodule with reusable makelib
- `scripts/tag.sh` — Buildtagger wrapper
- `scripts/check-examples.py` — Example validation
- `scripts/version_diff.py` — Schema version diffing

### Generated (Never Edit `zz_` Files!)
- `apis/cluster/{svc}/{ver}/zz_*.go` — API types
- `internal/controller/cluster/zz_*.go` — Controllers
- `package/crds/*.yaml` — CRD definitions
- `examples-generated/` — Example manifests

### CI/Documentation
- `.github/workflows/ci.yml` — GitHub Actions pipeline
- `CLAUDE.md` — Project conventions
- `.agents/docs/` — Architecture & implementation guides
- `.agents/specs/` — Feature specifications

---

## Troubleshooting Quick Links

| Problem | Solution |
|---------|----------|
| "command not found: make" | Install: `brew install make` (macOS) or `apt-get install make` (Linux) |
| "go: command not found" | Install Go 1.25.8+ from golang.org or use asdf |
| "build submodule not initialized" | Run: `git submodule update --init --recursive` |
| "Out of memory during lint" | Run: `make lint GOGC=50` or use larger machine |
| "Terraform download fails" | Check network, delete `.work/tools/terraform*`, retry `make generate.init` |
| "Generated files are stale" | Run: `make generate && make check-diff` |
| "crddiff reports breaking changes" | Review `package/crds/` diffs; may be intentional |

---

## Next Steps

1. **Install Go 1.25.8+** — Only manual prerequisite
2. **Clone repo**: `git clone --recurse-submodules <repo-url> && cd provider-upjet-aws`
3. **Setup**: `make vendor vendor.check && make generate.init && make check-diff`
4. **Start coding**: Edit `config/` files as needed
5. **Verify**: `make lint && make test && make check-diff`

---

## Reference Documentation

For deeper dives, see:
- **Architecture**: [`.agents/docs/architecture.md`](.agents/docs/architecture.md)
- **Build System**: [`.agents/docs/build-system.md`](.agents/docs/build-system.md)
- **Config Patterns**: [`.agents/docs/config-patterns.md`](.agents/docs/config-patterns.md)
- **Testing**: [`.agents/docs/testing.md`](.agents/docs/testing.md)
- **Full Details**: [`.agents/build-dependencies.md`](.agents/build-dependencies.md)
- **Quick Ref**: [`.agents/quick-reference.md`](.agents/quick-reference.md)

---

**Last updated**: Based on `go.mod` (1.25.8) and `Makefile` (v2024.12)
