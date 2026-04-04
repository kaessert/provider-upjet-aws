---
description: "Review a completed native migration service, create fix tickets if issues found, or create a planner ticket if clean."
context: fork
agent: general-purpose
argument-hint: "[service-name] e.g. 'elasticache', 'sqs'"
---

# Review and Fix Skill

Autonomous review loop: review a completed service migration, create fix tickets for issues,
or signal readiness for the next service.

> **⚠️ THE PARITY RULE — NON-NEGOTIABLE**: Native controllers MUST be 100% drop-in replacements
> for their Terraform counterparts. When the `RAW` prefix is removed, every existing user manifest
> MUST work identically — same fields, same types, same behavior, same defaults, same errors.
> Zero behavioral differences are acceptable. A single missed field is a **review failure**.

Use `$ARGUMENTS` as the service name.

---

## Step 1: Run the Review

Execute the review-native-migration skill logic against `$ARGUMENTS`. Specifically:

### 1a. Code Quality

```bash
# Compile
go build ./apis/cluster/$ARGUMENTS/... 2>&1
go build ./apis/namespaced/$ARGUMENTS/... 2>&1
go build ./internal/controller/$ARGUMENTS/... 2>&1

# Tests
go test ./internal/controller/$ARGUMENTS/... 2>&1

# No upjet imports
grep -r "crossplane/upjet" internal/controller/$ARGUMENTS/ apis/cluster/$ARGUMENTS/*/native/ 2>/dev/null

# CRDs exist
ls package/crds/$ARGUMENTS.aws.upbound.io_*raws.yaml 2>/dev/null
ls package/crds/$ARGUMENTS.aws.m.upbound.io_*raws.yaml 2>/dev/null

# Namespaced extractors don't reference cluster config
grep -r "config/cluster/common" apis/namespaced/$ARGUMENTS/*/native/ 2>/dev/null
```

### 1b. Code Review

Read ALL files:
- `internal/controller/$ARGUMENTS/*/crud.go`
- `internal/controller/$ARGUMENTS/*/crud_test.go`
- `apis/cluster/$ARGUMENTS/*/native/*_raw_types.go`
- `apis/namespaced/$ARGUMENTS/*/native/*_raw_types.go`

Check against `.agents/specs/native-controller-pattern.md` §17 (Pitfalls) and §18 (Checklist):
1. Late initialization implemented for AWS-defaulted fields
2. Unavailable() condition set for non-ACTIVE states
3. All mutable fields compared in isUpToDate
4. SetTestConditionIfAnnotated called in Observe
5. No build tags on controller files
6. security_group_names never compared or updated (if applicable)
7. JSON policy comparisons use semantic equality
8. For every field where isUpToDate returns false, Update handles that value
9. All references from config.go AND KnownReferencers are annotated
10. Connection details match the catalog
11. **SetForProvider** called after `GetForProvider()` + late-init in `Observe()` — example:
    ```go
    spec := cr.GetForProvider()
    if lateInitialize(spec, awsResp) {
        cr.SetForProvider(*spec)  // ⚠️ mandatory write-back
        return managed.ExternalObservation{..., ResourceLateInitialized: true}, nil
    }
    ```
    🔴 **Critical**: Namespaced types return a *copy* from `GetForProvider()`. Without
    `SetForProvider()`, late-initialized fields are silently discarded, causing an infinite
    reconcile loop where `ResourceLateInitialized: true` is returned every cycle and the
    `Ready` condition never becomes `True`.

### 1c. Parity Check

For each resource, compare the RAW type fields against the TF type fields:
- Count fields in Parameters, InitParameters, Observation
- Verify they match (same json tags, same Go types)
- Check for missing or extra fields
- Verify `SetForProvider()` exists on **both** cluster and namespaced RAW types:
  ```bash
  grep -n "SetForProvider" apis/cluster/$ARGUMENTS/*/native/*_raw_types.go
  grep -n "SetForProvider" apis/namespaced/$ARGUMENTS/*/native/*_raw_types.go
  ```
  If missing from either scope type → **Critical** issue (infinite late-init loop).

### 1d. Git Status

```bash
git status --short -- "apis/*/$ARGUMENTS/" "internal/controller/*/$ARGUMENTS/" "internal/controller/$ARGUMENTS/"
git ls-files --others --exclude-standard | grep -i "$ARGUMENTS"
```

---

## Step 2: Classify Findings

Categorize every finding:

| Severity | Definition |
|----------|-----------|
| 🔴 Critical | Breaks compilation, infinite reconciliation, data loss, missing functionality |
| 🟠 Medium | Behavioral difference from TF, missing error handling, wrong references |
| 🟡 Warning | Code quality, missing tests, documentation gaps |
| 🔵 Convention | Typos, naming, style |

> **Known Critical Pattern — Missing SetForProvider**: If `SetForProvider()` is absent on
> either scope type, or `Observe()` calls `GetForProvider()` to late-init but never calls
> `SetForProvider(*spec)`, classify as **🔴 Critical** — infinite reconcile loop.

---

## Step 3: Decision Point

### If ANY Critical or Medium issues found:

Create fix tickets and a follow-up review ticket.

**For each Critical/Medium issue**, create one executor ticket:

```
CreateTicket:
  id: fix-$ARGUMENTS-<issue-slug>
  title: "Fix: <concise issue description> ($ARGUMENTS)"
  description: |
    ## Issue
    <what's wrong>

    ## Files to Modify
    <exact file paths>

    ## Fix
    <specific code change>

    ## Spec Reference
    <relevant spec section>
  acceptance_criteria: [<specific, testable criteria>]
  labels: ["stage:executor"]
  priority: Critical (for 🔴) or High (for 🟠)
  test_plan: "go test ./internal/controller/$ARGUMENTS/..."
```

**If any fixes touch code (not just docs)**, also create an E2E retest ticket:

```
CreateTicket:
  id: e2e-$ARGUMENTS-post-review-<N>
  title: "E2E retest $ARGUMENTS RAW resources after review fixes (round <N>)"
  description: "Re-run all RAW examples after review fixes. <list example paths>"
  acceptance_criteria: ["All RAW examples reach Ready condition", "No leaked AWS resources"]
  labels: ["stage:executor"]
  priority: High
  depends_on: [<all fix ticket IDs>]
```

**Always create a follow-up review ticket** (the recursive loop):

```
CreateTicket:
  id: review-$ARGUMENTS-<N+1>
  title: "Review native $ARGUMENTS migration (round <N+1>)"
  description: |
    Re-review after fixes from round <N>.
    Previous review found <count> issues: <brief list>.
    Fix tickets: <list IDs>.
    
    Run the review-and-fix skill for $ARGUMENTS.
  acceptance_criteria: ["Review completed", "All critical/medium issues addressed or new fix tickets created"]
  labels: ["stage:reviewer"]
  priority: High
  depends_on: ["e2e-$ARGUMENTS-post-review-<N>"] or [<fix ticket IDs if no e2e needed>]
```

### If NO Critical or Medium issues found (clean review):

Create a planner ticket to trigger the next service migration:

```
CreateTicket:
  id: plan-next-after-$ARGUMENTS
  title: "Plan next native migration service (after $ARGUMENTS)"
  description: |
    $ARGUMENTS migration is complete and reviewed clean.
    
    Analyze remaining services and pick the next best candidate based on:
    1. Migration tier ordering (Tier 1 → Tier 4)
    2. Complexity progression (build on skills learned)
    3. Resource count (manageable batch sizes)
    
    Use the plan-next-service skill to select and plan.
  acceptance_criteria: ["Next service selected", "All tickets created for next service"]
  labels: ["stage:planner"]
  priority: Normal
```

Also update the spec/skill if any Warning-level gaps were found (create doc tickets).

---

## Step 4: Report

Output a summary:

```
## Review Complete: $ARGUMENTS (Round <N>)

| Metric | Value |
|--------|-------|
| Critical issues | <N> |
| Medium issues | <N> |
| Warnings | <N> |
| Fix tickets created | <N> |
| E2E retest | <yes/no> |
| Outcome | <FIX_LOOP / CLEAN → PLAN_NEXT> |

### Issues Found
<table of issues with severity>

### Tickets Created
<list of ticket IDs and titles>

### Next Step
<"Executor will work through fix tickets, then this review repeats" OR "Planner will select next service">
```
