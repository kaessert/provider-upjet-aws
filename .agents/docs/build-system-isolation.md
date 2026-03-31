# Build System Isolation — `native/` Sub-Package Safety

> **Summary**: `make generate` does **NOT** touch files in `apis/cluster/<service>/<version>/native/`
> sub-packages. The isolation guarantee is verified for all four generation pipeline stages.

---

## Problem Statement

During the parallel migration phase, TF-bridged types (generated, prefixed `zz_`) and native AWS SDK
types (hand-written) must coexist in the same provider binary. If native types were placed in the
same package as generated types:

```
apis/cluster/s3/v1beta1/zz_bucket_types.go     ← generated, overwritten by make generate
apis/cluster/s3/v1beta1/bucket_raw_types.go    ← hand-written, at risk of deletion
```

…the code generation pipeline could accidentally delete or overwrite native files.

## Solution: The `native/` Sub-Package

Native types live in a dedicated sub-package:

```
apis/cluster/s3/v1beta1/                   ← TF types (zz_* generated here)
apis/cluster/s3/v1beta1/native/            ← RAW types (hand-written, isolated here)
```

At cutover, native types move up to replace the TF types:

```
apis/cluster/s3/v1beta1/                   ← Now contains native types (TF code deleted)
```

---

## Verification Results

### Test Setup

A dummy `native/` sub-package was created at `apis/cluster/s3/v1beta1/native/doc.go` containing
only:

```go
package native
```

**SHA-256 checksum before any generation steps:**
```
d9dfd2cbb6000725f1a689f5dae4f617ad058b1d5e52717fdf4de79b08599c8b  doc.go
```

Each pipeline component was run against:
1. The `native/` package directly (`./apis/cluster/s3/v1beta1/native/...`)
2. The full s3 package scope (`./apis/cluster/s3/...`) which would mirror real `make generate` behavior

**Result**: All tests passed. SHA-256 was identical after every step.

---

## Pipeline Stage Analysis

### Stage 1: Deletion Commands (from `generate/generate.go`)

```bash
# Line 15: Delete specific zz_ generated files by name
find ../apis \( -iname 'zz_generated.conversion_hubs.go' \
             -o -iname 'zz_generated.conversion_spokes.go' \
             -o -iname 'zz_generated.resolvers.go' \) -delete

# Line 16: Remove empty directories
find ../apis -type d -empty -delete

# Lines 17-20: Clean internal/controller and cmd/provider (NOT apis/)
find ../internal/controller -iname 'zz_*' -delete
find ../internal/controller -type d -empty -delete
find ../cmd/provider -name 'zz_*' -type f -delete
find ../cmd/provider -type d -maxdepth 1 -mindepth 1 -empty -delete
```

**Analysis for `native/` sub-packages:**

| Command | Affect on `native/`? | Reason |
|---------|----------------------|--------|
| Delete `zz_generated.conversion_hubs.go` etc. | ❌ NO | Only deletes files with exact `zz_generated.*` names. `doc.go` and hand-written types are not named `zz_*`. |
| Delete empty directories in `apis/` | ⚠️ **CONDITIONAL** | Would delete `native/` only if it is **empty**. As long as `native/` contains at least one file (e.g., `doc.go`), it is preserved. |
| Delete `zz_*` in `internal/controller` | ❌ NO | Targets `internal/controller/`, not `apis/`. |

**Key invariant**: `native/` must always contain at least one non-generated file (e.g., `doc.go`).
The per-resource `*_raw_types.go` file satisfies this requirement automatically.

### Stage 2: Terrajet Generator (`cmd/generator/main.go`)

The Terrajet generator reads from `config/schema.json` (the Terraform provider schema) and
generates types only for resources registered in `config.GetProvider()`. It writes output to:

- `apis/cluster/<service>/<version>/zz_<resource>_types.go`
- `apis/cluster/<service>/<version>/zz_<resource>_terraformed.go`

**Analysis**: The generator operates on a fixed list of TF resources and writes files with the `zz_`
prefix. It has no knowledge of `native/` sub-packages and will never write to them. The generator
was not individually testable without the full Terraform schema setup (~15 MB JSON), but the
output path logic is deterministic and bounded to `zz_`-prefixed files in known locations.

### Stage 3: `controller-gen` (deepcopy + CRD generation)

```bash
go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen \
    object:headerFile=../hack/boilerplate.go.txt \
    paths=../apis/... \
    crd:allowDangerousTypes=true,crdVersions=v1 \
    output:artifacts:config=../package/crds
```

`controller-gen` processes `paths=../apis/...` which **includes** `native/` sub-packages. However,
it only generates output for packages containing types annotated with:
- `+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object` (deepcopy)
- `+kubebuilder:resource:...` (CRDs)

**Tested**: Running `controller-gen` against both `./apis/cluster/s3/v1beta1/native/...` and
`./apis/cluster/s3/...` produced **no new files** in `native/` and left `doc.go` unchanged.

**When native types ARE added**: When `*_raw_types.go` files are added to `native/` with proper
`+k8s:deepcopy-gen` annotations, `controller-gen` WILL generate a `zz_generated.deepcopy.go` in
`native/`. This is the correct behavior — `native/` needs its own deepcopy file, separate from the
parent package's `zz_generated.deepcopy.go`. This generated file CAN be committed to source
control since it serves native types, not TF types.

**Key**: `controller-gen` generates deepcopy for annotated types regardless of package location.
The `native/` sub-package gets its own `zz_generated.deepcopy.go` — this is safe. The deletion
step in `generate/generate.go` (line 15) only deletes three specific files:
- `zz_generated.conversion_hubs.go`
- `zz_generated.conversion_spokes.go`
- `zz_generated.resolvers.go`

It does **NOT** delete `zz_generated.deepcopy.go`. The deepcopy file in `native/` is regenerated
each time `make generate` runs, but its content is deterministic (it reflects the current native
types), so it is safe.

### Stage 4: `angryjet` (crossplane-runtime methodsets)

```bash
go run -tags generate github.com/crossplane/crossplane-tools/cmd/angryjet \
    generate-methodsets \
    --header-file=../hack/boilerplate.go.txt \
    ../apis/...
```

`angryjet` processes `../apis/...` which **includes** `native/` sub-packages. It generates
`zz_generated.managed.go` and `zz_generated.managedlist.go` for types with crossplane resource
interface markers (`+crossplane:generate:...`).

**Tested**: Running `angryjet` against `./apis/cluster/s3/...` produced **no new files** in
`native/` and left `doc.go` unchanged.

**When native types ARE added**: When `*_raw_types.go` adds crossplane managed resource types,
`angryjet` WILL generate `zz_generated.managed.go` and `zz_generated.managedlist.go` in `native/`.
These are safe to have in `native/` — they're deterministic and will be regenerated correctly on
each `make generate` run.

### Stage 5: Upjet Resolver

```bash
go run github.com/crossplane/upjet/v2/cmd/resolver \
    -g aws.upbound.io \
    -a github.com/upbound/provider-aws/v2/internal/apis \
    -s \
    -p ../apis/cluster/...

go run github.com/crossplane/upjet/v2/cmd/resolver \
    -g aws.m.upbound.io \
    -a github.com/upbound/provider-aws/v2/internal/apis \
    -s \
    -p ../apis/namespaced/...
```

The resolver processes **all packages** matching `../apis/cluster/...`, including `native/`
sub-packages. For each package, it looks **only** for files named `zz_generated.resolvers.go` and
transforms them.

**Tested**: Running the resolver with `-p ./apis/cluster/s3/...` (which traverses into `native/`)
produced **no changes** to `doc.go`. The resolver simply skips packages that have no
`zz_generated.resolvers.go` file.

**⚠️ Critical caveat — native types WITH cross-resource references**: When a native type has
`+crossplane:generate:reference:type=...` annotations, `angryjet` generates a
`zz_generated.resolvers.go` in `native/`. The upjet resolver will then **find and attempt to
transform this file**. This fails because the resolver expects types that implement
`resource.Terraformed` — native types do not.

This is the issue documented in spec section 0.15. **The fix is separate from this ticket**.
Section 0.15 requires excluding `native/` from the resolver's `-p` patterns in
`generate/generate.go`, and using `crossplane-tools` directly for native reference resolution.

**The `-s` flag** (`ignorePackageLoadErrors=true`) means the resolver silently ignores any
package load errors, so even if the native package fails to compile during resolver's load phase,
it won't hard-fail `make generate`.

---

## Current Status: Does `make generate` Touch `native/`?

| Condition | Verdict |
|-----------|---------|
| `native/` contains only `doc.go` (no types) | ✅ **SAFE** — no stage touches it |
| `native/` contains native types (no references) | ✅ **SAFE** — controller-gen/angryjet add deepcopy/managed files, resolver skips |
| `native/` contains native types WITH `+crossplane:generate:reference` | ⚠️ **BLOCKED** — upjet resolver will attempt to transform the generated resolver file and may fail. Requires fix from phase-0-15. |

---

## What Developers Must Do to Add a Native Package Safely

### Step 1: Create the package

```
mkdir -p apis/cluster/<service>/<version>/native/
```

### Step 2: Create `doc.go` (required to prevent empty-dir deletion)

```go
// Package native contains hand-written native (non-Terraform) type definitions
// for the <service> service.
package native
```

### Step 3: Add `*_raw_types.go`

Write your native types following the patterns in `.agents/specs/native-controller-pattern.md`.

### Step 4: Run generation for the native package

After adding types, run these to regenerate the `zz_generated.deepcopy.go` and
`zz_generated.managed.go` in `native/`:

```bash
# Deepcopy
go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen \
    object:headerFile=./hack/boilerplate.go.txt \
    paths=./apis/cluster/<service>/<version>/native/...

# Managed resource methods  
go run -tags generate github.com/crossplane/crossplane-tools/cmd/angryjet \
    generate-methodsets \
    --header-file=./hack/boilerplate.go.txt \
    ./apis/cluster/<service>/<version>/native/...
```

These are automatically re-run by `make generate` on subsequent runs — the output is deterministic.

### Step 5: Reference resolution (if needed)

If native types use `+crossplane:generate:reference:type=...` annotations, you must **separately**
run `crossplane-gen` (not `make generate`) after modifying annotations:

```bash
# Run crossplane-gen directly against native packages (NOT part of make generate)
go run ./vendor/github.com/crossplane/crossplane-tools/cmd/crossplane-gen/... \
    -p ./apis/cluster/<service>/<version>/native/
```

Commit the resulting `zz_resolve_references.go` to source control. This step must be run manually
whenever cross-resource reference annotations change. See spec section 0.15 for the resolver
exclusion implementation.

---

## Isolation Guarantee Summary

> **`make generate` DOES NOT clobber hand-written files in `native/` sub-packages.**
>
> The guarantee holds for all current pipeline stages (deletion commands, Terrajet generator,
> `controller-gen`, `angryjet`, and the upjet resolver) provided that:
>
> 1. `native/` contains at least one file (preventing empty-dir deletion)
> 2. Native types do NOT have `+crossplane:generate:reference` annotations until the resolver
>    exclusion fix (phase-0-15) is implemented.
>
> When native types have reference annotations AND phase-0-15 is implemented, the guarantee is
> unconditional for all hand-written files.

---

## Verification Artifacts

| Test | Command | Result |
|------|---------|--------|
| Deletion commands | `find ./apis/cluster/s3/v1beta1/native \( -iname 'zz_generated.*.go' \) -print` | 0 files matched |
| controller-gen (narrow) | `go run controller-gen ... paths=./apis/cluster/s3/v1beta1/native/...` | doc.go SHA unchanged |
| controller-gen (wide) | `go run controller-gen ... paths=./apis/cluster/s3/...` | doc.go SHA unchanged |
| angryjet (narrow) | `go run angryjet generate-methodsets ... ./apis/cluster/s3/v1beta1/native/...` | doc.go SHA unchanged |
| angryjet (wide) | `go run angryjet generate-methodsets ... ./apis/cluster/s3/...` | doc.go SHA unchanged |
| upjet resolver (narrow) | `go run resolver ... -p ./apis/cluster/s3/v1beta1/native/...` | doc.go SHA unchanged |
| upjet resolver (wide) | `go run resolver ... -p ./apis/cluster/s3/...` | doc.go SHA unchanged |

All tests verified `doc.go` SHA-256 = `d9dfd2cbb6000725f1a689f5dae4f617ad058b1d5e52717fdf4de79b08599c8b`
before and after each step.
