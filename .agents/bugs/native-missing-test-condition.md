## Investigation: Native Controllers Missing Uptest `Test` Condition

### Summary
Native (RAW) controllers never set the `Test` condition that uptest/chainsaw asserts on, causing e2e tests to timeout after 20 minutes even when the resource is fully SYNCED and READY.

### Root Cause
**Location**: `upjet/v2/pkg/resource/conditions.go:118-134` — the `SetUpToDateCondition` function that produces the `Test` condition lives exclusively in the upjet package.

**Mechanism**:

The uptest e2e framework uses a two-part protocol for verifying resource readiness:

1. **Annotation**: Uptest annotates each test resource with `upjet.upbound.io/test=true` (done in chainsaw's `00-apply.yaml` script step)
2. **Condition**: The controller checks for this annotation via `IsTest(mg)`, and when the resource is up-to-date, sets a condition `Type: "Test", Status: True, Reason: "UpToDate"` on the resource
3. **Assert**: Chainsaw asserts `status.conditions[?type=='Test'].status == "True"` — this is controlled by `--default-conditions="Test"` in the Makefile (line 273)

The chain for **Upjet/TF controllers**:
```
uptest annotates resource → TF Observe runs → IsTest() returns true → SetUpToDateCondition() sets Test:True → chainsaw assert passes
```

The chain for **Native controllers**:
```
uptest annotates resource → Native Observe runs → ??? → Test condition never set → chainsaw times out
```

The gap is that `SetUpToDateCondition()` is called from exactly **3 places**, all inside upjet:
- `pkg/controller/external_tfpluginsdk.go:572` — TF Plugin SDK Observe path
- `pkg/controller/external_tfpluginfw.go:616` — TF Plugin Framework Observe path
- `pkg/controller/external.go:355` — Classic TF Workspace Observe path

**crossplane-runtime's managed reconciler has zero awareness of the Test condition.** It only sets `Synced` and `Ready`. The `Test` condition is a pure upjet-layer concept with no native equivalent.

### Spec Analysis
**Classification**: Spec Gap
**Spec**: `.agents/specs/native-controller-pattern.md` line 1359
**Finding**: The spec says "Avoid upjet imports; use crossplane-runtime conditions" in the anti-patterns table, but provides **no guidance** on what native controllers should do to satisfy the uptest `Test` condition. The migration spec (`.agents/specs/terraform-removal-migration.md` line 437) describes uptest as "waits for Ready" which is inaccurate — it actually waits for `Test` by default.

### Evidence
- `UpToDateCondition()` returns `{Type: "Test", Status: True, Reason: "UpToDate"}` — upjet `conditions.go:118-125`
- `SetUpToDateCondition()` gates on `IsTest(mg) && upToDate` — upjet `conditions.go:128-134`
- `IsTest()` checks `mg.GetAnnotations()["upjet.upbound.io/test"] == "true"` — upjet `lateinit.go:444-446`
- Makefile line 273: `--default-conditions="Test"` hardcoded for all uptest runs
- IAM Role (upjet) has 4 conditions: `Synced, LastAsyncOperation, Ready, Test` ✅
- StateMachineRAW (native) has 2 conditions: `Synced, Ready` — no `Test` ❌
- Uptest supports per-resource override via `uptest.upbound.io/conditions` annotation (confirmed in `uptest e2e --help`) — currently unused by any example in the repo

### Fix Approach

There are two viable approaches (not mutually exclusive):

**Option A — Per-resource annotation override (quick fix, no code changes)**:
Add `uptest.upbound.io/conditions: "Ready"` annotation to native RAW example manifests. This tells uptest to assert on `Ready` instead of `Test` for that specific resource. No controller changes needed.

```yaml
# examples/sfn/cluster/v1beta2/statemachineraw.yaml
metadata:
  annotations:
    uptest.upbound.io/conditions: "Ready"
```

**Option B — Native controllers set Test condition (parity fix)**:
Add a small helper in `internal/native/` that replicates `SetUpToDateCondition` without importing upjet:

```go
// internal/native/conditions.go
func SetTestCondition(mg resource.Managed, upToDate bool) {
    if upToDate && mg.GetAnnotations()["upjet.upbound.io/test"] == "true" {
        mg.SetConditions(xpv1.Condition{
            Type:   "Test",
            Status: corev1.ConditionTrue,
            Reason: "UpToDate",
        })
    }
}
```

Call from the native Observe path when `ResourceUpToDate` is true. This achieves full behavioral parity with upjet controllers for e2e testing.

**Spec changes required**:
1. Update `.agents/specs/native-controller-pattern.md` — replace the anti-pattern line "Avoid upjet imports; use crossplane-runtime conditions" with explicit guidance: either use the `uptest.upbound.io/conditions` annotation override on RAW examples, or call the native `SetTestCondition` helper in Observe.
2. Update `.agents/specs/terraform-removal-migration.md` — clarify that uptest waits for `Test` condition (not `Ready`), and document the native workaround in the e2e test step description.

Run `/plan fix-native-test-condition` to create pheromone tickets for the fix — executor ant picks them up.
