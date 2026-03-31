# CRD Versioning Strategy for Native Migration

> **Spec reference**: Section 0.8 of `terraform-removal-migration.md`

## Overview

Several resources in this provider serve multiple CRD API versions (e.g. v1beta1 + v1beta2, or
v1beta1 + v1beta2 + v1beta3) with conversion webhooks. This document describes:

1. How versioning works during the **parallel phase** (TF and RAW types coexist)
2. What must happen **at cutover** (when the native type replaces the TF type)
3. Which resources are affected (multi-version resource table)
4. A known blocker: the **conversion-webhook type-assertion problem**

---

## Parallel Phase: Single Version Only

**During the parallel phase, RAW native types use a single API version.** There is no need to
implement CRD conversion webhooks for RAW types.

Rationale:
- RAW types live in a sub-package `apis/cluster/<service>/v1beta1/native/` (see spec 0.9).
- No user-facing objects are stored under RAW kinds during the parallel phase.
- No version upgrade path is needed until cutover.

**Implication**: The scaffold and implement phases create RAW types with a single version.
Do NOT add `+kubebuilder:storageversion` to multiple versions, and do NOT implement
`ConvertTo`/`ConvertFrom` methods on RAW types.

---

## At Cutover: Serve All Historical Versions

When a native type replaces its TF counterpart (cutover phase), it **MUST** serve every API
version the TF type ever served. This is required so existing objects stored in etcd under
any previous version can still be retrieved by the API server.

Requirements at cutover for each multi-version resource:

1. **Register all versions**: The native type's `+kubebuilder:resource` annotations must
   include all versions served by the TF type.
2. **Match storage version**: The native type's storage version (`+kubebuilder:storageversion`)
   must match the TF type's storage version (see table below).
3. **Implement conversion**: Either implement `ConvertTo`/`ConvertFrom` conversion webhooks,
   or perform an in-place migration of all stored objects to the new storage version before
   removing old versions.

If in-place object migration is chosen, a migration job must iterate over all objects and
re-save them at the new version before dropping the old version from the served set.

---

## Multi-Version Resource Table

### Resources With Explicit Version Config in `config/cluster/*/config.go`

These resources have `PreviousVersions`, `SetCRDStorageVersion`, or `ControllerReconcileVersion`
explicitly set in their per-service config files (the source searched by the test plan).

| Kind | Service | Terraform Resource | All Versions Served | CRD Storage Version | Config Source |
|------|---------|-------------------|--------------------|--------------------|---------------|
| `BucketLifecycleConfiguration` | `s3` | `aws_s3_bucket_lifecycle_configuration` | v1beta1, v1beta2 | **v1beta1** | `config/cluster/s3/config.go` — explicit `Version="v1beta2"`, `PreviousVersions=["v1beta1"]`, `SetCRDStorageVersion("v1beta1")`, `ControllerReconcileVersion="v1beta1"` |
| `Cluster` | `redshift` | `aws_redshift_cluster` | v1beta1, v1beta2 | **v1beta1** | `config/cluster/redshift/config.go` — explicit `Version="v1beta2"`, `SetCRDStorageVersion("v1beta1")`, `ControllerReconcileVersion="v1beta1"` |

### Resources With Explicit `r.Version = "v1beta2"` in `config/cluster/*/config.go`

These resources also have explicit version management but the storage version is derived from
the `bumpVersionsWithEmbeddedLists` auto-bump mechanism (see below).

| Kind | Service | Terraform Resource | All Versions Served | CRD Storage Version | Notes |
|------|---------|-------------------|--------------------|--------------------|-------|
| `ReplicationGroup` | `elasticache` | `aws_elasticache_replication_group` | v1beta1, v1beta2 | **v1beta2** | `config/cluster/elasticache/config.go` — custom field-level converters, NOT processed by auto-bump |
| `Cluster` | `kafka` (MSK) | `aws_msk_cluster` | v1beta1, v1beta2, v1beta3 | **v1beta2** | `config/cluster/kafka/config.go` — set to v1beta2, then auto-bumped to v1beta3; storage stays v1beta2 |
| `AutoscalingGroup` | `autoscaling` | `aws_autoscaling_group` | v1beta1, v1beta2, v1beta3 | **v1beta2** | `config/cluster/autoscaling/config.go` — set to v1beta2, then auto-bumped to v1beta3; storage stays v1beta2 |
| `RoutingProfile` | `connect` | `aws_connect_routing_profile` | v1beta1, v1beta2 | **v1beta2** | `config/cluster/connect/config.go` — exception in auto-bump (field renaming already bumped it); storage v1beta2 |

### Resources With Multiple Versions Via Auto-Bump (Broad Scope)

The function `bumpVersionsWithEmbeddedLists` in `config/registry_cluster.go` automatically
bumps API versions for **325 resources** listed in `config/old-singleton-list-apis.txt`.
These resources all have:

- Two (or three) API versions served
- CRD storage version set to the pre-bump version (e.g. v1beta1)
- `ControllerReconcileVersion` set to the pre-bump version

**For cutover tickets of these 325 resources, the same versioning requirements apply**: the
native type must serve all versions the TF type served, with the same storage version.

To determine the exact versions for a specific auto-bumped resource, check:

```bash
grep -A3 "name: v1beta" package/crds/<group>_<resource>.yaml | grep -E "name:|storage:"
```

---

## Conversion Webhook Type-Assertion Problem

### The Issue

The generated file `zz_generated.conversion_spokes.go` in every spoke-version package contains
conversion methods that cast the hub type to `resource.Terraformed`:

```go
// Example from apis/cluster/s3/v1beta1/zz_generated.conversion_spokes.go
func (tr *BucketLifecycleConfiguration) ConvertTo(dstRaw conversion.Hub) error {
    if err := ujconversion.RoundTrip(dstRaw.(resource.Terraformed), tr); err != nil {
        // ...
    }
}
```

The `dstRaw.(resource.Terraformed)` type assertion **WILL PANIC** if the hub type does not
implement the `resource.Terraformed` interface.

### Why RAW Types Cannot Be Hubs

Native (RAW) types do not implement `resource.Terraformed` — they are plain Crossplane managed
resources with no Terraform state. Therefore:

- **You cannot make a native type the hub** in a CRD that also serves TF spoke versions.
- At cutover for multi-version resources, the `zz_generated.conversion_spokes.go` file in
  the old spoke packages **must be replaced** with a hand-written conversion implementation
  that does NOT rely on `resource.Terraformed`.

### Cutover Action Required

For every multi-version resource listed in the tables above, the cutover ticket MUST include:

1. **Rewrite conversion methods** in old spoke-version packages (v1beta1, v1beta2 where they
   are spokes) to NOT cast to `resource.Terraformed`. Instead, implement field-by-field
   conversion between the old spoke struct and the new native hub struct.

2. The existing `zz_generated.conversion_spokes.go` files are auto-generated and immutable.
   The cutover work must either:
   - Replace the file with a hand-written non-`zz_` prefixed file, OR
   - Delete the old spoke-version package entirely (only viable if in-place object migration
     is performed first)

3. Reference ticket **phase-0-18** for the detailed conversion-webhook strategy.

---

## Summary: Cutover Checklist for Multi-Version Resources

For each resource in the tables above, the cutover ticket must verify:

- [ ] Native type registers all versions the TF type served
- [ ] Native type's `+kubebuilder:storageversion` matches the TF storage version (see table)
- [ ] `ConvertTo`/`ConvertFrom` in old spoke packages do NOT use `.(resource.Terraformed)` cast
- [ ] Conversion logic is functionally equivalent to TF conversion (field mapping preserved)
- [ ] E2E test passes after cutover (existing objects can be read/updated under all versions)

---

## How to Determine Versions for a Given Resource

```bash
# 1. Check explicit per-service config
grep -n "Version\|PreviousVersions\|SetCRDStorageVersion\|ControllerReconcileVersion" \
    config/cluster/<service>/config.go

# 2. Check if resource is in old singleton list APIs (auto-bump applies)
grep "<terraform_resource_name>" config/old-singleton-list-apis.txt

# 3. Verify actual CRD storage version
grep -A3 "name: v1beta" package/crds/<group>_<resource>.yaml | grep -E "name:|storage:"

# 4. Confirm hub version from generated code
grep -r "func.*Hub()" apis/cluster/<service>/
```
