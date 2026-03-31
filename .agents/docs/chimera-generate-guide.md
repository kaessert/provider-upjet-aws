# Chimera Generate Workflow — Developer & Agent Guide

> **What this is**: Reference guide for the two code-generation paths available
> during the TF-to-native migration ("chimera" phase).  Misunderstanding which
> command to run is the single largest source of accidental mass-diffs (800+
> cosmetic file changes) during development.

---

## Table of Contents

- [Overview — Two Generate Paths](#overview--two-generate-paths)
- [When to Use Which](#when-to-use-which)
- [Chimera Invariants](#chimera-invariants)
- [Common Pitfalls](#common-pitfalls)
- [Troubleshooting](#troubleshooting)

---

## Overview — Two Generate Paths

The repository now has **two distinct code-generation commands** that serve
different purposes.  They must never be confused.

### Path 1: `make generate` — Full Upjet Pipeline

```bash
make generate.init    # download TF provider schema (~15 MB) + docs
make generate         # run full Upjet code-generation pipeline
```

**What it does:**

1. Invokes the Upjet generator (`cmd/generator/main.go`) which reads
   `config/schema.json` and regenerates **all** `zz_*` files across
   `apis/`, `internal/controller/`, `cmd/provider/`, and `examples-generated/`.
2. Runs `controller-gen` to regenerate `zz_generated.deepcopy.go` across
   **all** packages under `apis/`.
3. Runs `angryjet` to regenerate `zz_generated.managed.go` and
   `zz_generated.managedlist.go` across **all** packages under `apis/`.
4. Runs the upjet resolver to regenerate `zz_generated.resolvers.go` for
   all TF-bridged type packages.

**Characteristics:**

| Property | Value |
|----------|-------|
| Duration | **15–30 minutes** (schema download + full code gen) |
| Files touched | **Thousands** — every `zz_*` file across `apis/`, `internal/controller/`, `cmd/provider/` |
| Required tool env | Pinned tool versions (see Makefile); mismatched versions produce cosmetic diffs |
| Safe on CI | Yes — `make check-diff` uses this path |

---

### Path 2: `make generate.native` — Native Types Only

```bash
make generate.native
```

**What it does:**

1. Runs `controller-gen` (`object:` generator) on:
   - `apis/cluster/v1beta1/...` — ProviderConfig types (hand-written)
   - `apis/namespaced/v1beta1/...` — Namespaced ProviderConfig types
   - Every `apis/**/native/` subpackage (dynamically discovered via `find`)
2. Runs `angryjet generate-methodsets` on every `apis/**/native/` subpackage.

**Characteristics:**

| Property | Value |
|----------|-------|
| Duration | **Seconds to ~1 minute** |
| Files touched | Only `zz_generated.deepcopy.go` / `zz_generated.managed.go` inside `native/` subpackages + two ProviderConfig deepcopy files |
| Required tool env | No Terraform schema required |
| Safe for iteration | Yes — does NOT touch existing TF-generated `zz_*` files outside `native/` |

The exact Makefile implementation:

```makefile
generate.native:
	go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen \
		object:headerFile=hack/boilerplate.go.txt \
		paths=./apis/cluster/v1beta1/... \
		paths=./apis/namespaced/v1beta1/... \
		$(shell find apis -path '*/native' -type d | sed 's|^|paths=./|' | tr '\n' ' ')
	go run -tags generate github.com/crossplane/crossplane-tools/cmd/angryjet \
		generate-methodsets \
		--header-file=hack/boilerplate.go.txt \
		$(shell find apis -path '*/native' -type d | sed 's|^|./|' | tr '\n' ' ')
```

---

## When to Use Which

Use the **minimum sufficient command** — running `make generate` when only
native types changed produces hundreds of cosmetic diffs from tool-version
variance.

| Change being made | Command |
|-------------------|---------|
| Adding fields to a native type (`*_raw_types.go`) | `make generate.native` |
| Adding a new native type file in a `native/` package | `make generate.native` |
| Modifying hand-written types in `apis/cluster/v1beta1/types.go` (ProviderConfig) | `make generate.native` |
| Modifying `config/externalname.go` | `make generate` |
| Modifying `config/cluster/<service>/config.go` | `make generate` |
| Modifying `hack/main.go.tmpl` | `make generate` |
| Adding a new service to `config/cluster/provider.go` | `make generate` |
| Upgrading the Terraform provider version | `make generate.init && make generate` |
| Before opening a PR | `make check-diff` (runs full generate internally) |

### The Golden Rule

> If the change is **only** in a `native/` subpackage or in
> `apis/cluster/v1beta1/` / `apis/namespaced/v1beta1/` hand-written types:
> **use `make generate.native`**.
>
> For everything else: **use `make generate`**.

---

## Chimera Invariants

These invariants **must always be true** for the chimera (TF+native parallel)
setup to work correctly.  They are enforced by guardrail tests in
`generate/chimera_test.go` (`go test ./generate/...`).

### Invariant 1 — Native types must not implement `resource.Terraformed`

Native types live in `native/` subpackages and are compiled directly alongside
TF-bridged types.  If a native type accidentally implements the upjet
`resource.Terraformed` interface, the upjet resolver will panic at startup.

The Terraformed interface is identified by these method names:
- `GetTerraformResourceType`
- `GetTerraformSchemaVersion`
- `GetConnectionDetailsMapping`
- `GetObservation` / `SetObservation`
- `GetParameters` / `SetParameters`
- `GetInitParameters` / `GetMergedParameters`
- `LateInitialize`

**Rule**: Native types must not declare any of these methods.

Guardrail test: `TestNativeTypesNotTerraformed`

### Invariant 2 — `native/` packages must be excluded from the upjet resolver

The upjet resolver (`go run github.com/crossplane/upjet/v2/cmd/resolver`) runs
with `-p ../apis/cluster/...` which traverses into `native/` subpackages.
Native packages must be excluded because the resolver only handles types that
implement `resource.Terraformed`.

This exclusion is implemented in `generate/generate.go` via
`FilterNativePackages`, which strips any path containing `/native/` from the
package list before passing it to the resolver.

Guardrail test: `TestFilterNativePackagesAlsoExcludesNamespaced`

### Invariant 3 — `NativeSetupHook_*` variables must exist for every service

The `hack/main.go.tmpl` template generates `cmd/provider/<service>/zz_main.go`
for each service provider.  The template contains:

```go
if h := controller.NativeSetupHook_{{ .Group }}; h != nil {
    if err := h(mgr, o); err != nil { ... }
}
```

For this to compile, a variable `NativeSetupHook_<service>` must be declared in:
- `internal/controller/cluster/native_hooks.go`
- `internal/controller/namespaced/native_hooks.go`

Every subdirectory of `cmd/provider/` must have a corresponding hook variable.

**Rule**: When a new service is added to `cmd/provider/`, add the hook variable
to **both** `native_hooks.go` files before running `make generate`.

Guardrail test: `TestNativeSetupHookVarsMatchServices`

### Invariant 4 — `zz_generated.deepcopy.go` must be regenerated after type changes

When hand-written types in a `native/` package or in `apis/cluster/v1beta1/`
are modified (fields added/removed/renamed), the corresponding
`zz_generated.deepcopy.go` file becomes stale.  A stale deepcopy file causes
compile errors or silent data loss (fields not deep-copied).

**Rule**: Always run `make generate.native` after modifying any hand-written
type definition.

Guardrail test: `TestNativeGenerationIdempotent` (verifies the generation is
deterministic; stale files would manifest as diffs on CI).

---

## Common Pitfalls

### Pitfall 1 — Running `make generate` with different tool versions

The `zz_*` files are sensitive to the exact version of every tool used during
generation (Terraform, upjet, controller-gen, angryjet, etc.).  If your local
tool versions differ from what CI used, `make generate` will produce hundreds of
cosmetic diffs (import alias changes, blank line insertions, comment rewording).

**Symptom**: `make check-diff` fails on CI with "N files changed" even though
you only touched one config file.

**Fix**:
1. Revert all `zz_*` changes: `git checkout -- apis/ internal/controller/ cmd/provider/ examples-generated/`
2. Ensure your environment matches pinned versions (see `Makefile` variables like `TERRAFORM_VERSION`, `UPTEST_VERSION`, etc.)
3. Re-run `make generate` in the CI environment or use `make check-diff` directly

Do **not** commit cosmetic diffs to `zz_*` files.

### Pitfall 2 — Committing `zz_*` files with only formatting changes

Even if `make generate` runs cleanly locally, committing files whose only
changes are formatting (import alias names, blank lines) will cause `make
check-diff` to fail on CI because CI regenerates everything from scratch.

**Fix**: Only stage and commit `zz_*` files when they contain **substantive**
changes driven by your config edits.  If in doubt, run `make check-diff`
locally first.

### Pitfall 3 — Forgetting to regenerate deepcopy after adding fields

After adding a field to a hand-written type (native type or ProviderConfig),
the deepcopy function for that type becomes stale.  The project compiles but
the new field is silently dropped during object cloning.

**Fix**: Run `make generate.native` immediately after any type field change.

### Pitfall 4 — The upjet resolver silently failing on native types

The upjet resolver uses `-s` (`ignorePackageLoadErrors=true`), so if a native
package fails to compile during the resolver's load phase it won't hard-fail
`make generate`.  This can hide bugs where native types have compile errors that
are only surfaced by `go build`.

**Fix**: Always run `go build ./...` (or `go test ./...`) after `make generate`
to catch compile errors in native packages.

### Pitfall 5 — Adding `+crossplane:generate:reference` to native types

If a native type has `+crossplane:generate:reference:type=...` annotations,
`angryjet` generates a `zz_generated.resolvers.go` in the `native/` package.
The upjet resolver will then attempt to transform this file and may fail because
it expects `resource.Terraformed` types.

**Fix**: Until the resolver exclusion (phase-0-15) is fully implemented, do not
add cross-resource reference annotations to native types.  Use manual reference
resolution instead.

### Pitfall 6 — Empty `native/` directory gets deleted by `make generate`

The `generate/generate.go` script runs `find ../apis -type d -empty -delete`
which removes empty directories.  A `native/` package that contains only a
`doc.go` is **not** empty and is preserved.  But a truly empty `native/`
directory (e.g., after a failed partial setup) will be deleted.

**Fix**: Always create `doc.go` first when initialising a new `native/` package:

```go
// Package native contains hand-written native (non-Terraform) type definitions
// for the <service> service.
package native
```

---

## Troubleshooting

### "`make check-diff` fails with 800+ files changed"

**Cause**: Tool version mismatch.  You ran `make generate` locally with a
different version of one or more tools than CI expects.

**Diagnosis**:
```bash
git diff --stat HEAD | grep "zz_" | wc -l
# If this is in the hundreds, it's a tool-version problem
```

**Fix**:
1. Revert all generated file changes:
   ```bash
   git checkout -- apis/ internal/controller/ cmd/provider/ examples-generated/
   ```
2. Check which tool version differs.  Common culprits:
   - `controller-gen` (version in `go.mod` under `sigs.k8s.io/controller-tools`)
   - `angryjet` (version in `go.mod` under `github.com/crossplane/crossplane-tools`)
   - Terraform or AWS provider binary version
3. Run in the CI runner environment, or use Docker to match tool versions exactly.

### "Resolver panics on native type"

**Symptom**: Provider binary panics at startup with a message involving
`resource.Terraformed` or a type assertion failure.

**Cause**: A native type accidentally implements the `resource.Terraformed`
interface (Invariant 1 violation), OR the resolver exclusion is not working
correctly (Invariant 2 violation).

**Diagnosis**:
```bash
# Check if native types have Terraform-specific methods
go test ./generate/... -run TestNativeTypesNotTerraformed -v

# Verify resolver exclusion is active
go test ./generate/... -run TestFilterNativePackagesAlsoExcludesNamespaced -v
```

**Fix for "native type implements Terraformed"**:
- Search for `GetTerraformResourceType` in your native type files
- Remove any TF-specific method or `SetTerraformResourceType` call
- Run `make generate.native` to regenerate

**Fix for "resolver not excluding native packages"**:
- Check `generate/generate.go` for `FilterNativePackages` usage in the resolver
  invocation
- Ensure the `-p` flag passed to the resolver does not include `native/` paths

### "`NativeSetupHook_<service>` undefined"

**Symptom**: Compile error in `cmd/provider/<service>/zz_main.go`:
```
undefined: controller.NativeSetupHook_<service>
```

**Cause**: A new service was added to `cmd/provider/` but the hook variable was
not added to `native_hooks.go` before running `make generate`.

**Diagnosis**:
```bash
go test ./generate/... -run TestNativeSetupHookVarsMatchServices -v
```

**Fix**:
1. Add the missing variable to **both** hook files:
   ```go
   // NativeSetupHook_<service> is set by the native <service> controllers package (if any).
   var NativeSetupHook_<service> func(mgr ctrl.Manager, o xpcontroller.Options) error //nolint:gochecknoglobals
   ```
   Files to edit:
   - `internal/controller/cluster/native_hooks.go`
   - `internal/controller/namespaced/native_hooks.go`
2. Run `make generate` to regenerate `zz_main.go` files.

### "deepcopy function is missing a field"

**Symptom**: Fields added to a native type are not copied during reconciliation;
or a compile error in `zz_generated.deepcopy.go` because a type reference is
outdated.

**Cause**: `make generate.native` was not run after modifying the type
definition.

**Fix**:
```bash
make generate.native
# Then verify the deepcopy file was updated:
git diff apis/**/native/zz_generated.deepcopy.go
```

### "make generate.native runs but produces no output files"

**Symptom**: After running `make generate.native`, no `zz_generated.deepcopy.go`
is created in the `native/` package.

**Cause**: The native type is missing the required deepcopy marker annotation.

**Fix**: Add the deepcopy marker to the type:
```go
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
```
Or, for Crossplane managed resources, ensure the type embeds
`xpv1.ResourceStatus` and `xpv1.ResourceSpec` (these carry the annotations
needed by angryjet).

---

## Quick Reference Card

```
Situation                                  Command
─────────────────────────────────────────  ──────────────────────────
Changed a native type field                make generate.native
Added a new native type file               make generate.native
Changed ProviderConfig types               make generate.native
Changed config/externalname.go             make generate
Changed config/cluster/<svc>/config.go     make generate
Changed hack/main.go.tmpl                  make generate
Added service to cmd/provider/ (new dir)   update native_hooks.go FIRST, then make generate
Before opening a PR                        make check-diff
CI is failing with cosmetic zz_ diffs      git checkout -- apis/ internal/controller/ cmd/provider/
Debugging invariant violations             go test ./generate/... -v
```

---

## Related Documentation

| Document | Location | What it covers |
|----------|----------|----------------|
| Build System Reference | [`.agents/docs/build-system.md`](.agents/docs/build-system.md) | All Makefile targets, CI pipeline, tool versions |
| Build System Isolation | [`.agents/docs/build-system-isolation.md`](.agents/docs/build-system-isolation.md) | Proof that `make generate` does NOT clobber `native/` files |
| Native Controller Guide | [`.agents/docs/native-controller-guide.md`](.agents/docs/native-controller-guide.md) | How to implement native controllers (types, CRUD, setup) |
| Native Controller Pattern Spec | [`.agents/specs/native-controller-pattern.md`](.agents/specs/native-controller-pattern.md) | Dual-scope interface pattern, file layout, templates |
| Migration Design Spec | [`.agents/specs/terraform-removal-migration.md`](.agents/specs/terraform-removal-migration.md) | Full migration plan, phase ordering, batch strategy |
