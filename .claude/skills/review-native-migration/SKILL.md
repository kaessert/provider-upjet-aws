---
description: "Review a completed native migration phase — commits, tickets, ant runs, code quality, and process gaps."
context: fork
agent: general-purpose
argument-hint: "[service-name] e.g. 'sfn', 'sqs'"
---

# Review Native Migration Skill

Perform a comprehensive review of a completed (or in-progress) native migration phase
for one AWS service. This produces an actionable report covering: what happened, what
went well, what didn't, code quality issues, spec/skill gaps, and recommended fixes.

Use `$ARGUMENTS` as the service name (e.g., `sfn`, `sqs`).

---

## Step 1: Gather Context (parallel)

Run all of these in parallel to collect the full picture:

### 1a. Git History

```bash
# All commits related to this service's native migration
git log --all --format="%h %ai %s" --grep="$ARGUMENTS" | head -40
git log --all --format="%h %ai %s" --grep="native.*$ARGUMENTS\|$ARGUMENTS.*native\|$ARGUMENTS.*RAW\|RAW.*$ARGUMENTS" | head -20

# Check for reverts or resets
git log --all --oneline --grep="revert\|reset\|Revert\|Reset" | grep -i "$ARGUMENTS" | head -10

# Working tree status for service directories
git status --short -- "apis/*/$ARGUMENTS/" "internal/controller/*/$ARGUMENTS/" "internal/controller/$ARGUMENTS/"
git ls-files --others --exclude-standard | grep -i "$ARGUMENTS"
```

Count commits by type (feat/fix/chore/docs), identify the timeline, and note any
reset-and-reimplement cycles.

### 1b. Pheromone Tickets

List all tickets with `plan:native-$ARGUMENTS` label. For each ticket, get full details.
Report: status, how many attempts, any failures, orphan recoveries.

Check all statuses: ToDo, InProgress, Done, Failed.

### 1c. Ant Runs

Check the executor ant for recent runs related to this service.
Look at the most recent run's logs for errors, warnings, or workarounds the agent applied.

### 1d. Code Review

Read ALL files in these locations:
- `internal/controller/$ARGUMENTS/*/crud.go` — shared CRUD logic
- `internal/controller/$ARGUMENTS/*/crud_test.go` — tests
- `internal/controller/cluster/$ARGUMENTS/*/controller.go` — cluster wrappers
- `internal/controller/namespaced/$ARGUMENTS/*/controller.go` — namespaced wrappers
- `apis/cluster/$ARGUMENTS/*/native/*_raw_types.go` — RAW types
- `examples/$ARGUMENTS/cluster/*/*raw.yaml` — example manifests

Check specifically:
1. **Compilation**: `go build ./apis/cluster/$ARGUMENTS/... && go build ./internal/controller/$ARGUMENTS/...`
2. **Tests pass**: `go test ./internal/controller/$ARGUMENTS/...`
3. **No untracked files**: `git ls-files --others --exclude-standard | grep $ARGUMENTS`
4. **No build tags** on controller files (they break the build)
5. **Late initialization** implemented (LateInitialize*Ptr calls, ResourceLateInitialized return)
6. **Unavailable() condition** set for non-ACTIVE states
7. **All mutable fields** compared in isUpToDate (no silent drift gaps)
8. **SetTestConditionIfAnnotated** called in Observe (NOT uptest.upbound.io/conditions annotation)
9. **No upjet imports**: `grep -r "crossplane/upjet" internal/controller/$ARGUMENTS/`
10. **CRD YAMLs exist**: `ls package/crds/$ARGUMENTS.aws.upbound.io_*raws.yaml`
11. **Intra-service references** point to RAW types (not TF types)
12. **No extra fields** in RAW types that don't exist in TF types (YAML compatibility)
13. **Example manifests** are exact copies of TF examples (only `kind` changed)
14. **Test account ID `609897127049`** used consistently for all ARNs embedded in opaque JSON strings (policies, state machine definitions, redrive policies)

### 1e. Spec & Skill Alignment

Read the current spec and skill, check if this service's migration revealed any gaps:
- `.agents/specs/native-controller-pattern.md` — Section 17 (Pitfalls) and 18 (Checklist)
- `.claude/skills/plan-native-migration/SKILL.md` — scaffold and implement ticket templates

---

## Step 2: Compile Report

Organize findings into this structure:

### Timeline
Chronological summary of what happened — commits, ticket state changes, key events.
Note the total wall-clock time and any reset/reimplement cycles.

### What Went Well ✅
List things that worked correctly from the start. Give credit for good patterns,
clean code, proper error handling, etc.

### What Didn't Go Well ❌
For each issue found:
- **What happened**: describe the problem
- **Root cause**: why it happened
- **Was it fixed?**: yes/no, and how
- **Is it preventable?**: could a spec/skill change prevent this for future services?

### Code Review Summary

| Category | Count | Key Items |
|----------|-------|-----------|
| 🔴 Critical | N | Compilation failures, missing functionality, spec violations |
| 🟡 Warning | N | Missing best practices, potential issues |
| 💡 Suggestion | N | Improvements, simplifications |
| ✅ Positive | N | Good patterns, well-designed code |

### Process Gaps
Issues that the spec/skill should have prevented but didn't.
For each gap, propose a specific fix (exact section + wording).

### Leaked AWS Resources
Check for any AWS resources left behind from failed e2e runs:
```bash
# Adapt per service — check the regions used in example manifests
aws sqs list-queues --region us-west-1 2>/dev/null | head -10
aws sqs list-queues --region us-east-1 2>/dev/null | head -10
# aws stepfunctions list-state-machines --region us-west-1 ...
# aws s3 ls ...
```

---

## Step 3: Actionable Recommendations

Categorize recommendations by urgency:

1. **Fix now** — blocking issues, compilation failures, leaked resources
2. **Fix before next service** — spec/skill gaps that will cause the same problem again
3. **Nice to have** — improvements that aren't urgent

For each recommendation, be specific:
- What file to change
- What section to update
- Exact wording or code change

---

## Output Format

Present the full report in conversation. Do NOT create tickets or make changes —
this is a read-only review. The user will decide what to act on.

End with a summary table:

| Metric | Value |
|--------|-------|
| Total commits | N |
| Tickets completed | N/N |
| Failed tickets | N |
| Unit tests | N passing |
| Code compiles | ✅/❌ |
| Working tree clean | ✅/❌ |
| Leaked AWS resources | ✅ none / ❌ N found |
| Spec/skill gaps found | N |
| Critical issues | N |
