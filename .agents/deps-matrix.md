# Dependency Matrix

## Quick Lookup Table

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                    TOOL & DEPENDENCY QUICK REFERENCE                         ║
╚══════════════════════════════════════════════════════════════════════════════╝

┌─ REQUIRED (Must Install Manually) ─────────────────────────────────────────┐
│                                                                              │
│  ✓ Go 1.25.8+                    From: golang.org/dl or asdf               │
│    └─ golang.org/x/tools/cmd/goimports   (installed via CI, not critical)  │
│                                                                              │
│  ✓ git, curl, unzip              Standard Unix tools (pre-installed)        │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌─ AUTO-DOWNLOADED (Via Make Targets) ──────────────────────────────────────┐
│                                                                              │
│  ▶ Code Generation                                                           │
│    ├─ Terraform CLI 1.5.5            → make generate.init                   │
│    └─ Terraform AWS Provider 6.34.0  → make generate.init                   │
│                                                                              │
│  ▶ Linting & Analysis                                                        │
│    ├─ golangci-lint 2.11.4           → make lint                            │
│    └─ buildtagger v0.12.0-rc...      → make lint (or scripts/tag.sh)        │
│                                                                              │
│  ▶ Testing & Local Development                                               │
│    ├─ Kind v0.30.0                   → make local-deploy                    │
│    ├─ Uptest v2.2.0                  → make uptest                          │
│    ├─ Kustomize v5.3.0               → make kustomize-crds                  │
│    ├─ YQ v4.40.5                     → make kustomize-crds                  │
│    └─ Crossplane CLI v2.2.0          → make local-deploy                    │
│                                                                              │
│  ▶ Validation                                                                 │
│    └─ crddiff v0.12.1                → make crddiff (go run)                │
│                                                                              │
│  ▶ Go Modules (Dependencies)                                                 │
│    └─ Downloaded via: make vendor vendor.check                              │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌─ BUILD SUBMODULE (Git Submodule) ─────────────────────────────────────────┐
│                                                                              │
│  build/                          (git submodule)                            │
│  ├─ makelib/common.mk            (common utilities)                         │
│  ├─ makelib/golang.mk            (Go build system)                          │
│  ├─ makelib/k8s_tools.mk         (Kind, Kustomize, Uptest)                 │
│  ├─ makelib/imagelight.mk        (container images)                         │
│  ├─ makelib/xpkg.mk              (Crossplane packages)                      │
│  └─ [more...]                                                               │
│                                                                              │
│  Initialize via: git submodule update --init --recursive                    │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Dependency Download Timeline

```
┌─ First-Time Setup (45 minutes) ─────────────────────────────────────────┐
│                                                                           │
│  $ git clone --recurse-submodules <url>                   ▶ 2 min       │
│    └─ Clones repo + build/ submodule (~500 MB)                          │
│                                                                           │
│  $ make vendor vendor.check                               ▶ 2-3 min     │
│    └─ Downloads Go dependencies to vendor/ (~2.5 GB)                    │
│                                                                           │
│  $ make generate.init                                     ▶ 5 min       │
│    ├─ Downloads Terraform CLI (1.5.5)                                   │
│    ├─ Downloads Terraform AWS Provider (6.34.0)                         │
│    └─ Generates config/schema.json (~500 MB)                            │
│                                                                           │
│  $ make check-diff                                        ▶ 15 min      │
│    ├─ Runs cmd/generator (uses schema.json)                             │
│    ├─ Generates apis/cluster/**/*.go                                    │
│    ├─ Generates internal/controller/**/*.go                             │
│    ├─ Generates package/crds/*.yaml                                     │
│    └─ Compares with git (verifies nothing changed)                      │
│                                                                           │
│  $ make lint (first time)                                 ▶ 10-15 min   │
│    ├─ Downloads buildtagger                                             │
│    ├─ Downloads golangci-lint                                           │
│    ├─ Builds analysis cache                                             │
│    └─ Runs linting                                                       │
│                                                                           │
│  ═════════════════════════════════════════════════════════════════════ │
│  Total First Build: ~45 minutes, ~6 GB downloads                         │
│  ═════════════════════════════════════════════════════════════════════ │
│                                                                           │
│  Subsequent builds:                                       ▶ 5-10 min     │
│  └─ Most tools cached in .work/ and vendor/                             │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

## CI Pipeline Dependency Flow

```
┌─ GitHub Actions (CI) ──────────────────────────────────────────────────┐
│                                                                         │
│  detect-noop                                                            │
│   └─ Check if changes are non-trivial                                  │
│       ▼                                                                 │
│  report-breaking-changes                                               │
│   └─ requires: crddiff v0.12.1                                         │
│       ▼                                                                 │
│  lint                                                                   │
│   ├─ requires: Go 1.25.8, golangci-lint 2.11.4, buildtagger            │
│   ├─ downloads: build submodule (if not present)                       │
│   └─ env: GOGC=50 (memory optimization)                                │
│       ▼                                                                 │
│  check-diff                                                             │
│   ├─ requires: Go 1.25.8, goimports, git                               │
│   ├─ runs: make vendor, make generate, make check-diff                 │
│   └─ downloads: Terraform (if schema.json missing)                     │
│       ▼                                                                 │
│  unit-tests                                                             │
│   ├─ requires: Go 1.25.8, git                                          │
│   └─ runs: make test -j2 (parallel with 2 workers)                     │
│       ▼                                                                 │
│  local-deploy                                                           │
│   ├─ requires: Go 1.25.8, Docker, Kind, Crossplane CLI                 │
│   └─ runs: make local-deploy (builds and deploys to Kind)              │
│       ▼                                                                 │
│  check-examples                                                         │
│   ├─ requires: Python 3, yq, kubectl                                   │
│   └─ runs: scripts/check-examples.py                                   │
│                                                                         │
│  All jobs run on: Ubuntu-Jumbo-Runner (16 CPU, 64 GB RAM)              │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Storage Breakdown

```
┌─ Disk Space Usage ──────────────────────────────────────────────────┐
│                                                                      │
│  repository root (repo)              ~500 MB  (source code)         │
│  ├─ apis/                            ~200 MB  (generated types)      │
│  ├─ internal/controller/             ~300 MB  (generated controllers)│
│  ├─ package/crds/                    ~50 MB   (generated CRDs)      │
│  └─ [other sources]                  ~50 MB                         │
│                                                                      │
│  vendor/                             ~2.5 GB  (go modules)          │
│                                                                      │
│  .work/                              ~2.5 GB  (tools & artifacts)   │
│  ├─ tools/                           ~1.0 GB  (Terraform, Kind, etc)│
│  ├─ terraform/                       ~0.5 GB  (working directory)   │
│  ├─ build/                           ~0.5 GB  (output files)        │
│  └─ helm/                            ~0.5 GB  (Helm cache)          │
│                                                                      │
│  ~/.cache/golangci-lint/             ~0.5 GB  (linter cache)        │
│  ~/.cache/go-build/                  ~0.5 GB  (Go build cache)      │
│  ~/.cache/go-mod/                    ~0.5 GB  (Go module cache)     │
│                                                                      │
│  ═════════════════════════════════════════════════════════════════ │
│  TOTAL:                              ~10 GB   (typical setup)       │
│  ═════════════════════════════════════════════════════════════════ │
│                                                                      │
│  CI artifacts (.work/cache/):        ~1-2 GB  (Docker buildx)       │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

## Version Sources

```
┌─ Where Versions Are Defined ────────────────────────────────────────┐
│                                                                      │
│  Go 1.25.8                    go.mod (line 7)                       │
│                               github.com/upbound/provider-aws/v2    │
│                                                                      │
│  Terraform 1.5.5              Makefile (line 13)                    │
│                               export TERRAFORM_VERSION              │
│                                                                      │
│  Terraform AWS 6.34.0         Makefile (line 14)                    │
│                               export TERRAFORM_PROVIDER_VERSION     │
│                                                                      │
│  golangci-lint 2.11.4         Makefile (line 55)                    │
│                               GOLANGCILINT_VERSION                  │
│                                                                      │
│  buildtagger v0.12.0-rc...    Makefile (line 62)                    │
│                               BUILDTAGGER_VERSION                   │
│                                                                      │
│  Kind v0.30.0                 Makefile (line 82)                    │
│                               KIND_VERSION                          │
│                                                                      │
│  Uptest v2.2.0                Makefile (line 83)                    │
│                               UPTEST_VERSION                        │
│                                                                      │
│  Kustomize v5.3.0             Makefile (line 84)                    │
│                               KUSTOMIZE_VERSION                     │
│                                                                      │
│  YQ v4.40.5                   Makefile (line 85)                    │
│                               YQ_VERSION                            │
│                                                                      │
│  Crossplane v2.2.0            Makefile (line 86)                    │
│                               CROSSPLANE_VERSION                    │
│                                                                      │
│  Crossplane CLI v2.2.0        Makefile (line 87)                    │
│                               CROSSPLANE_CLI_VERSION                │
│                                                                      │
│  crddiff v0.12.1              Makefile (line 88)                    │
│                               CRDDIFF_VERSION                       │
│                                                                      │
│  Build Submodule              .gitmodules                           │
│                               build/ (git submodule)                │
│                                                                      │
│  Go Module Versions           go.mod, go.sum                        │
│                               All dependencies defined here         │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

## Installation Decision Tree

```
                         Want to build provider-upjet-aws?
                                    |
                  ┌─────────────────┼─────────────────┐
                  |                 |                 |
         Install Go 1.25.8?     Have git?      Have curl/unzip?
         (REQUIRED)             (REQUIRED)      (REQUIRED)
                  |                 |                 |
           golang.org/dl      pre-installed     pre-installed
                  |                 |                 |
                  └─────────────────┼─────────────────┘
                                    |
                  ╔═════════════════╩════════════════╗
                  ║   You're Ready to Build!         ║
                  ║   Everything else auto-downloads ║
                  ╚═════════════════╦════════════════╝
                                    |
                  ┌─────────────────┼─────────────────┐
                  |                 |                 |
         git clone --recurse    make vendor       make generate.init
         submodules <url>       vendor.check          ▼
                  |                 |           Terraform, schema.json
                  ▼                 ▼
         Repository + build/   Go modules
                  |                 |
                  └─────────────────┼─────────────────┘
                                    |
                           ╔════════╩═════════╗
                           ║  Ready to code!  ║
                           ║  make check-diff ║
                           ║  make lint       ║
                           ║  make test       ║
                           ╚══════════════════╝
```

---

## Summary Statistics

| Metric | Value | Notes |
|--------|-------|-------|
| **Required manual installs** | 1 | Go 1.25.8+ |
| **Auto-downloaded tools** | 10 | All via Make |
| **Total tool versions tracked** | 10 | In Makefile |
| **Build submodule files** | 6+ | In build/makelib/ |
| **First-time setup time** | 30-45 min | Network dependent |
| **Subsequent build time** | 5-10 min | Most cached |
| **Total disk usage** | ~10 GB | Typical setup |
| **Memory required** | 16 GB min | 32 GB recommended |
| **Platforms supported** | 2 | linux_amd64, linux_arm64 |
| **CI job parallelism** | 6 | Run in parallel on PR |

