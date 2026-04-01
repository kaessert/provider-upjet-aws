---
description: "Create pheromone tickets to migrate one AWS service from Terraform to native SDK."
context: fork
agent: general-purpose
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - CreateTicket
argument-hint: "[service-name] e.g. 'sfn', 's3', 'dynamodb'"
---

# Plan Native Migration Skill

Create all pheromone tickets required to migrate one AWS service from Terraform-bridged
controllers to native AWS SDK v2 controllers. Follow each step in order — the output is a
complete dependency-chained ticket graph that the executor ant can work through autonomously.

This skill generates approximately 7 + (4 × N) tickets per service, where N is the number
of resources. For a 2-resource service like `sfn`, that is 15 tickets total.

---

## Spec References (Read These First)

Before creating any tickets, read both specs in full:

1. **`.agents/specs/native-controller-pattern.md`** — Implementation guide: file layout, controller
   setup, dual-scope interface pattern, CRUD patterns, external name handling, async operations.

2. **`.agents/specs/terraform-removal-migration.md`** — Full migration design: dual scope
   architecture, Phase 0 prerequisites, per-service migration cycle (Steps 0–6), batch ordering,
   Per-Ticket Description Template, and the Plan Skill Design section (~line 658).

---

## Step 1: Parse and Validate the Service Name

Extract the service name from `$ARGUMENTS`. Trim whitespace, lowercase.

```bash
SERVICE="$ARGUMENTS"   # e.g., "sfn"
```

If `$ARGUMENTS` is empty or whitespace only, stop immediately and print:
```
ERROR: No service name provided.
Usage: /plan-native-migration <service-name>
Examples: sfn, s3, dynamodb, kms, iam, eks
```

Verify the service exists:
```bash
ls config/cluster/$SERVICE/config.go
```

If not found, also try `config/namespaced/$SERVICE/config.go`. If neither exists, stop and print:
```
ERROR: Service '<service>' not found.
Expected: config/cluster/<service>/config.go
```

---

## Step 2: Discover Resources

Read `config/cluster/<service>/config.go` and extract every `AddResourceConfigurator("aws_...")` call.
Each call represents one TF resource to migrate.

```bash
grep -o 'AddResourceConfigurator("[^"]*"' config/cluster/$SERVICE/config.go \
  | grep -o '"aws_[^"]*"' \
  | tr -d '"'
```

This produces a list of TF resource names, one per line, e.g.:
```
aws_sfn_state_machine
aws_sfn_activity
```

Store this list. If empty, stop and print:
```
ERROR: No AddResourceConfigurator calls found in config/cluster/<service>/config.go
This service may already be migrated or may not use standard configuration.
```

---

## Step 3: Derive Name Variants for Each Resource

For each TF resource name (e.g., `aws_sfn_state_machine`), compute the following name variants
that will be used throughout all tickets and file paths.

Given `TF_NAME = "aws_sfn_state_machine"` and `SERVICE = "sfn"`:

**resource_bare** — Strip `aws_<service>_` prefix, keep underscores:
```
aws_sfn_state_machine  →  state_machine
aws_sfn_activity       →  activity
```

**resource_slug** — Same as resource_bare but replace `_` with `-` (for ticket IDs):
```
state_machine  →  state-machine
activity       →  activity
```

**resource_file** — Lowercase, no separators (matches the `zz_*` file name convention):
```
state_machine  →  statemachine
activity       →  activity
```

**resource_go** — PascalCase (each `_`-separated word capitalized):
```
state_machine  →  StateMachine
activity       →  Activity
```

**kind_raw** — RAW kind name:
```
StateMachine  →  StateMachineRAW
Activity      →  ActivityRAW
```

**controller_dir** — Controller directory name (lowercase, no separators, + "raw"):
```
statemachine  →  statemachineraw
activity      →  activityraw
```

If the TF name does NOT start with `aws_<service>_`, emit a warning but still process it using
the full name minus `aws_` as the resource_bare.

---

## Step 4: Gather Metadata for Each Resource

For each resource, run these checks. Collect results for use in ticket descriptions.
Many checks use `grep` with context lines — read enough lines to capture the full configurator block.

### 4a: External Name Strategy

```bash
grep '"<TF_NAME>"' config/externalname.go
```

Capture the full line. Classify the strategy:

| Pattern in line | Strategy label |
|-----------------|---------------|
| `config.IdentifierFromProvider` | `IdentifierFromProvider` — AWS assigns an opaque ID; set external name from Create response |
| `config.NameAsIdentifier` | `NameAsIdentifier` — the `name` parameter IS the external name |
| `config.ParameterAsIdentifier("<field>")` | `ParameterAsIdentifier(<field>)` — a specific param field is the external name |
| `config.TemplatedStringAsIdentifier(...)` | `TemplatedStringAsIdentifier(template)` — composite ID from template |
| Any other Go identifier/function call | `Custom` — complex logic; flag for careful review |

If the resource is NOT in `externalname.go` (line not found), record strategy as `Unknown — not in externalname.go; search config/externalname.go manually`.

### 4b: UseAsync

```bash
grep -A60 '"<TF_NAME>"' config/cluster/$SERVICE/config.go | grep -m1 "UseAsync"
```

Record `true` or `false` (default false if not present).

**If UseAsync=true**: The implement ticket gets priority `High` and the description must call out the async annotation pattern from `internal/native/async.go`.

### 4c: MoveToStatus Fields

```bash
grep -A60 '"<TF_NAME>"' config/cluster/$SERVICE/config.go | grep -A3 "MoveToStatus"
```

Extract the field names listed (comma-separated string arguments after the first `r.TerraformResource` arg).

Example: `config.MoveToStatus(r.TerraformResource, "arn", "endpoint", "status")` → fields: `arn`, `endpoint`, `status`.

These fields go in the Observation struct ONLY, not in Parameters.

### 4d: TerraformConfigurationInjector

```bash
grep -A60 '"<TF_NAME>"' config/cluster/$SERVICE/config.go | grep -c "TerraformConfigurationInjector"
```

Record yes/no. If yes: business logic resource — flag in summary warnings and reference
`.agents/specs/tf-business-logic-catalog.md` in the implement ticket.

### 4e: TerraformCustomDiff

```bash
grep -A60 '"<TF_NAME>"' config/cluster/$SERVICE/config.go | grep -c "TerraformCustomDiff"
```

Record yes/no. If yes: drift-suppression logic resource — flag and reference
`.agents/specs/tf-business-logic-catalog.md` in the implement ticket.

### 4f: AdditionalConnectionDetailsFn

```bash
grep -A60 '"<TF_NAME>"' config/cluster/$SERVICE/config.go | grep -c "AdditionalConnectionDetailsFn"
```

Record yes/no. If yes, also read the surrounding code block to extract the connection detail
keys being published (e.g., `id`, `arn`, `endpoint`, `password`, `username`).

### 4g: LateInitializer

```bash
grep -A60 '"<TF_NAME>"' config/cluster/$SERVICE/config.go | grep -A5 "LateInitializer"
```

Extract `IgnoredFields` list if present. Record as comma-separated list or "none".

### 4h: Reference Annotations

Locate the TF types file. First, determine which versions exist:

```bash
ls apis/cluster/$SERVICE/
# Output: v1beta1  v1beta2  (or just v1beta1)
```

For each version, check if the resource has a types file:
```bash
ls apis/cluster/$SERVICE/*/zz_<resource_file>_types.go 2>/dev/null
```

Read the file(s) and extract lines containing `+crossplane:generate:reference`:
```bash
grep "+crossplane:generate:reference" apis/cluster/$SERVICE/<version>/zz_<resource_file>_types.go
```

Record which fields have references. Also note any `resource.TerraformID()` extractor in
the annotations — these MUST be rewritten to `resource.ExtractResourceID()` in RAW types
(TerraformID() requires `resource.Terraformed` which RAW types do not implement).

### 4i: Multiple API Versions and Storage Version

```bash
ls apis/cluster/$SERVICE/
```

If both `v1beta1` and `v1beta2` exist, check if this specific resource has a v1beta2 type:
```bash
ls apis/cluster/$SERVICE/v1beta2/zz_<resource_file>_types.go 2>/dev/null
```

**Storage version** (the version controllers reconcile):
- If `apis/cluster/$SERVICE/v1beta2/zz_<resource_file>_types.go` exists → storage version is `v1beta2`
- Otherwise → storage version is `v1beta1`

Multi-version resources need conversion webhooks at cutover — flag in summary warnings.

### 4j: Find Example Manifest

```bash
ls examples/$SERVICE/cluster/*/
```

Look for `<resource_file>.yaml` in the highest available version directory (prefer v1beta2 over
v1beta1 when both exist).

```bash
# Check v1beta2 first, fall back to v1beta1
ls examples/$SERVICE/cluster/v1beta2/<resource_file>.yaml 2>/dev/null \
  || ls examples/$SERVICE/cluster/v1beta1/<resource_file>.yaml 2>/dev/null
```

Record the full path (e.g., `examples/sfn/cluster/v1beta2/statemachine.yaml`).

If no example manifest exists, record as `MISSING` — the baseline ticket must note this and
the executor must locate or create one before running e2e.

### 4k: IAM Policy Field (Infer)

If the TF resource name contains `policy` (e.g., `aws_s3_bucket_policy`, `aws_iam_role_policy`,
`aws_kms_key_policy`), flag this resource as requiring semantic IAM policy comparison via
`internal/native/policy.go`. Record yes/no.

### 4l: Schema Field Classification for isUpToDate

Extract field flags from `config/schema.json` to determine how each field should be treated
in the `isUpToDate` comparison. This prevents update loops caused by AWS returning default
values for fields the user didn't set.

```bash
python3 -c "
import json
with open('config/schema.json') as f:
    schema = json.load(f)
ps = schema.get('provider_schemas', {})
for provider_key, provider in ps.items():
    rs = provider.get('resource_schemas', {})
    if '<TF_NAME>' in rs:
        sm = rs['<TF_NAME>']
        block = sm.get('block', {})
        # Top-level attributes
        for name, attr in sorted(block.get('attributes', {}).items()):
            flags = [f for f in ['computed','optional','required'] if attr.get(f)]
            print(f'attr  {name}: {\" \".join(flags)}')
        # Block types (nested objects)
        for name, bt in sorted(block.get('block_types', {}).items()):
            min_i, max_i = bt.get('min_items', 0), bt.get('max_items', 0)
            print(f'block {name}: min={min_i} max={max_i}')
            for attr_name, attr in sorted(bt.get('block', {}).get('attributes', {}).items()):
                flags = [f for f in ['computed','optional','required'] if attr.get(f)]
                print(f'  {attr_name}: {\" \".join(flags)}')
"
```

Classify each field into one of these categories and record the results:

| Schema flags | isUpToDate treatment | Example |
|--------------|----------------------|---------|
| `required` | Always compare — no nil guard needed | `definition`, `role_arn` |
| `optional` (no `computed`) | Nil-guard: `if spec.Field != nil && *spec.Field != observed` | `type`, `publish` |
| `optional` + `computed` | Nil-guard OR late-init from AWS | `name`, `region` |
| `computed` only | Never compare — observation-only field | `arn`, `status`, `creation_date` |
| Block with `min=0` | If spec sub-struct is nil, accept AWS defaults: `if spec.X == nil { return true }` | `logging_configuration`, `encryption_configuration` |
| Block with `min=1` | Always compare — required block | (rare) |

**Critical**: Optional blocks with `min=0` are the #1 cause of update loops in native controllers.
AWS always returns default values for these blocks (e.g., `logging_configuration` with `level=OFF`,
`encryption_configuration` with `type=AWS_OWNED_KEY`), but the user's spec has nil. If the
`isUpToDate` comparison treats nil-spec + non-nil-observed as "not up to date", the controller
will call Update on every reconcile, and the uptest Test condition will never be set.

Record the classified fields as a table in the implement ticket description under
"### Schema Field Classification".

---

## Step 5: Create Baseline Tickets (1 per resource)

For each resource, create a baseline E2E ticket. This is an absolute hard gate.

Use `CreateTicket` with the following parameters. Substitute actual values — no placeholders
in the final ticket.

**Ticket structure:**

```
id:      native-<SERVICE>-baseline-<resource_slug>
title:   Baseline E2E: <SERVICE>/<resource_slug> TF controller passes
plan:    native-<SERVICE>
phase:   baseline
labels:  ["stage:executor"]
priority: High
```

**Description** (fill in all `<...>` with real values):

```
## Baseline E2E: <TF_NAME> → TF controller

Run the existing Terraform-backed e2e test to establish a verified baseline BEFORE any
migration work on this resource. This is a HARD GATE.

**If this ticket fails, ALL downstream tickets for this resource are blocked:**
- native-<SERVICE>-<resource_slug> (implement)
- native-<SERVICE>-<resource_slug>-e2e (E2E RAW)
Other resources in the service are unaffected.

### Spec Reference
Read `.agents/specs/terraform-removal-migration.md` section "Step 0: Baseline E2E"
for complete rules and failure handling.

### Example Manifest
<example_manifest_path>
  (or: MISSING — locate or create before running)

### E2E Command
```bash
export UPTEST_CLOUD_CREDENTIALS="DEFAULT='[default]
aws_access_key_id = ${AWS_ACCESS_KEY_ID}
aws_secret_access_key = ${AWS_SECRET_ACCESS_KEY}'"

export UPTEST_EXAMPLE_LIST="<example_manifest_path>"
make e2e SUBPACKAGES="config <SERVICE>"
```

### Failure Rules — ZERO TOLERANCE
| Failure                        | Action                                      |
|-------------------------------|---------------------------------------------|
| E2E test fails for any reason  | Mark FAILED — do NOT proceed                |
| AWS permission denied          | Mark FAILED — do NOT attempt IAM workarounds|
| AWS quota/limit exceeded       | Mark FAILED — do NOT request increases      |
| Example manifest is broken     | Mark FAILED — do NOT fix the manifest       |
| Timeout / infrastructure issue | Mark FAILED — do NOT retry                  |

When this ticket fails, record the exact error output in the ticket (update description)
and mark the ticket Failed.

### Result Capture
Paste the final `uptest` output here (last 50 lines) before marking Done.
```

**Acceptance criteria** (as array):
```
["TF resource <resource_go> reaches Ready condition",
 "TF resource <resource_go> deletes cleanly with no errors",
 "No AWS permission errors in output",
 "No AWS quota or limit errors in output",
 "E2E output captured in ticket"]
```

---

## Step 6: Create Scaffold Ticket (1 per service)

Create ONE scaffold ticket covering ALL resources in the service, both cluster and namespaced
scopes. This ticket does NOT implement CRUD — it creates types and empty stubs that compile.

The scaffold ticket has NO depends_on so it can start immediately in parallel with baselines.

**Ticket structure:**

```
id:      native-<SERVICE>-scaffold
title:   Scaffold RAW types and controller stubs for <SERVICE> (both scopes, all resources)
plan:    native-<SERVICE>
phase:   scaffold
labels:  ["stage:executor"]
priority: High
```

**Description** (list ALL resources with their specific files):

```
## Scaffold: RAW Types and Controller Stubs for <SERVICE>

Create RAW type definitions and empty controller stubs for ALL resources in <SERVICE>.
Both cluster AND namespaced scopes. No CRUD implementation — stubs must compile, nothing more.

### Spec References
- `.agents/specs/native-controller-pattern.md` — Sections 1 (File Layout), 2 (Controller
  Setup Pattern), 2.1 (Setup function signature), 2.3 (Dual Scope Interface Pattern)
- `.agents/specs/terraform-removal-migration.md` — Section "Step 1: Scaffold"

### Resources in This Service (<N> total)
<for each resource, one line: "- <TF_NAME> → Kind: <resource_go>RAW  storage: <version>  multi-version: yes/no">

### Critical Field Placement Rules
1. Fields listed in MoveToStatus → Observation struct ONLY (do NOT put in Parameters)
2. All user-writable fields → Parameters struct
3. Read-only AWS output fields (IDs, ARNs, timestamps) → Observation struct
4. `spec.forProvider.region` → ALWAYS in Parameters — REQUIRED for credential resolution
5. Copy `+crossplane:generate:reference` annotations from TF types verbatim, EXCEPT:
   - Replace `resource.TerraformID()` extractor → `resource.ExtractResourceID()`
   - Reason: TerraformID() requires resource.Terraformed which RAW types do not implement
6. Use `json:"..."` tags ONLY. NO `tf:"..."` tags anywhere in RAW types.
7. Kind must be `<resource_go>RAW` (json:"kind" value in the CRD will be set by Kubernetes)

### Phase 0 Prerequisites
Before the scaffold compiles, these internal packages must exist (Phase 0 infrastructure):
- `internal/native/` — connector, errors, tags, lateinit, externalname, async helpers
- `internal/controller/cluster/native_hook.go` — NativeSetupHook_<SERVICE> var
- `internal/controller/namespaced/native_hook.go` — NativeSetupHook_<SERVICE> var
If they don't exist yet, create stubs that compile (empty package with the right exports).

### Files to Create

Repeat this block for EACH resource in the service:

---
**Resource: <TF_NAME> → <resource_go>RAW**
MoveToStatus fields: <list or "none">
References: <list of fields with +crossplane:generate:reference or "none">
TerraformID() rewrite needed: <yes/no>

```
apis/cluster/<SERVICE>/<version>/native/<resource_file>_raw_types.go
```
  - Package: `package native`
  - Types: `<resource_go>RAWParameters`, `<resource_go>RAWInitParameters`,
           `<resource_go>RAWObservation`, `<resource_go>RAWSpec`, `<resource_go>RAWStatus`,
           `<resource_go>RAW`, `<resource_go>RAWList`
  - Embed: `v1.ResourceSpec` in Spec, `v1.ConditionedStatus` in Status (crossplane-runtime)
  - `<resource_go>RAW` must have markers: `+kubebuilder:object:root=true`,
    `+kubebuilder:subresource:status`, `+kubebuilder:resource:scope=Cluster`
  - Implement the `<resource_go>CR` interface declared in crud.go (compiler-verified)
  - Include `spec.forProvider.region *string \`json:"region"\`` in Parameters

```
apis/namespaced/<SERVICE>/<version>/native/<resource_file>_raw_types.go
```
  - Mirror of cluster type; resource scope = Namespaced
  - `+kubebuilder:resource:scope=Namespaced`
  - Same field layout, same interfaces

```
internal/controller/<SERVICE>/<resource_file>/crud.go
```
  - Package: `package <resource_file>`
  - Declare interface:
    ```go
    type <resource_go>CR interface {
        resource.Managed
        GetForProvider() *v1beta1native.<resource_go>RAWParameters
        GetInitProvider() *v1beta1native.<resource_go>RAWInitParameters
        GetAtProvider() v1beta1native.<resource_go>RAWObservation
        SetAtProvider(v1beta1native.<resource_go>RAWObservation)
    }
    ```
    (import cluster native package for the types; namespaced types are identical)
  - Stub Observe/Create/Update/Delete returning `errors.New("not implemented")`
  - Must compile without upjet imports

```
internal/controller/cluster/<SERVICE>/<resource_file>raw/controller.go
```
  - Package: `package <controller_dir>`
  - Setup() function: registers the controller using managed.WithTypedExternalConnector
  - Wire cluster native type (`apis/cluster/<SERVICE>/<version>/native.<resource_go>RAW`)
    to the shared ExternalClient from crud.go
  - Register: `func init() { clustercontroller.NativeSetupHook_<SERVICE> = append(..., Setup) }`

```
internal/controller/namespaced/<SERVICE>/<resource_file>raw/controller.go
```
  - Package: `package <controller_dir>`
  - Mirror of cluster controller but wires namespaced native type
  - Register: `func init() { namespacedcontroller.NativeSetupHook_<SERVICE> = append(..., Setup) }`

```
examples/<SERVICE>/cluster/<version>/<resource_file>raw.yaml
```
  - Copy of the TF example manifest; change `kind: <resource_go>` → `kind: <resource_go>RAW`
  - Change `apiVersion` group to the native sub-package group if different

```
examples/<SERVICE>/namespaced/<version>/<resource_file>raw.yaml
```
  - Namespaced scope variant

---

<repeat the above block for each resource in the service>

### Acceptance Criteria (All Must Pass Before Any Implement Ticket)
```

**Acceptance criteria** (as array):
```
["go build ./apis/cluster/<SERVICE>/... passes with no errors",
 "go build ./apis/namespaced/<SERVICE>/... passes with no errors",
 "go build ./internal/controller/<SERVICE>/... passes with no errors",
 "go vet ./apis/cluster/<SERVICE>/... passes",
 "go vet ./internal/controller/<SERVICE>/... passes",
 "All RAW types implement the <Resource>CR interface (verified by compiler)",
 "No upjet imports in any new file (grep -r 'crossplane/upjet' apis/cluster/<SERVICE>/*/native/ internal/controller/<SERVICE>/ — must return empty)",
 "Example RAW yaml manifests created for each resource"]
```

---

## Step 7: Create Implement Tickets (1 per resource)

For each resource, create one ticket covering full CRUD implementation plus both scope wrappers.

**Ticket structure:**

```
id:         native-<SERVICE>-<resource_slug>
title:      Implement native <resource_go>RAW controller (<SERVICE>)
plan:       native-<SERVICE>
phase:      implement
labels:     ["stage:executor"]
priority:   High   (if UseAsync=true OR has ConfigurationInjector OR has CustomDiff)
            Normal (all other resources)
depends_on: ["native-<SERVICE>-scaffold", "native-<SERVICE>-baseline-<resource_slug>"]
```

**Description** — fill in ALL metadata gathered in Step 4:

```
## Resource: <TF_NAME> → <resource_go>RAW

Implement the native AWS SDK v2 controller. Follow strict TDD: write failing tests FIRST,
then implement, then verify tests pass. Do NOT commit until all acceptance criteria pass.

### Spec References (Read ALL before writing any code)
1. `.agents/specs/native-controller-pattern.md` — COMPLETE GUIDE. Every section is relevant.
2. `.agents/specs/terraform-removal-migration.md` — "Step 2: Implement CRUD", "Dual Scope Architecture"
3. TF reference implementation:
   `vendor/github.com/upbound/terraform-provider-aws/internal/service/<SERVICE>/`
   (search for files matching <resource_file>*.go)
4. AWS SDK v2 client API:
   `vendor/github.com/aws/aws-sdk-go-v2/service/<SERVICE>/api_op_*.go`

### Files to Modify (all created as stubs by scaffold ticket)
- `internal/controller/<SERVICE>/<resource_file>/crud.go`
  Shared CRUD logic (interface-based). This is the main implementation file.
- `internal/controller/cluster/<SERVICE>/<resource_file>raw/controller.go`
  Cluster scope Setup() + thin wrapper. Wire concrete cluster type to shared CRUD.
- `internal/controller/namespaced/<SERVICE>/<resource_file>raw/controller.go`
  Namespaced scope Setup() + thin wrapper. Wire concrete namespaced type to shared CRUD.

### AWS SDK v2 Operations
These are inferred from the resource name. VERIFY against the SDK at
`vendor/github.com/aws/aws-sdk-go-v2/service/<SERVICE>/` before implementing:
- Create:  `<SERVICE>.Create<resource_go>Input` via `svc.Create<resource_go>(ctx, &input)`
- Observe: `<SERVICE>.Describe<resource_go>Input` via `svc.Describe<resource_go>(ctx, &input)`
           (may be `Get<resource_go>` — check SDK; some use List* + filter)
- Update:  `<SERVICE>.Update<resource_go>Input` via `svc.Update<resource_go>(ctx, &input)`
           (may be `Modify<resource_go>` — check SDK; some resources are immutable)
- Delete:  `<SERVICE>.Delete<resource_go>Input` via `svc.Delete<resource_go>(ctx, &input)`

For NotFound detection in Observe(): check TF code for the error code strings it uses
(typically `smithy.GenericAPIError.ErrorCode()` == "ResourceNotFoundException" or similar).

### External Name
Strategy: <external_name_strategy>
Full config entry: `<exact_line_from_externalname.go>`

Implementation guide:
- In Create(): after calling AWS, set: `meta.SetExternalName(cr, <id from response>)`
  - IdentifierFromProvider: ID is the main resource identifier in response (ARN, ID field)
  - NameAsIdentifier: ID is the name you passed in; `meta.GetExternalName(cr)` gives it
  - ParameterAsIdentifier(<field>): ID comes from `cr.Spec.ForProvider.<Field>`
  - TemplatedStringAsIdentifier: reconstruct the composite ID from the template fields
- In Observe/Update/Delete(): `meta.GetExternalName(cr)` retrieves the stored external name
- Use `internal/native/externalname.go` helpers if available

### Special Handling

**Async operations**: <yes | no>
<If yes: >
  UseAsync=true — this resource has long-running operations (creation/deletion may take
  10–15 minutes). Use the annotation-based async pattern from `internal/native/async.go`:
  - In Create()/Update()/Delete(): after calling AWS, set an operation annotation on cr
  - In Observe(): check for the annotation; if present, poll AWS for completion status
  - If still running: return ResourceExists=true, ResourceUpToDate=true (triggers re-poll)
  - If complete: clear annotation, return actual observed state
  - If failed: clear annotation, return the error
  The managed reconciler's PollInterval handles re-scheduling automatically.

**Business logic (TerraformConfigurationInjector)**: <yes | no>
<If yes: >
  This resource has a TerraformConfigurationInjector. Check
  `.agents/specs/tf-business-logic-catalog.md` for the entry for `<TF_NAME>`. The injector
  adds/modifies fields before TF applies them — port this logic to:
  - Observe(): when comparing spec vs AWS state, apply the same normalization
  - Create(): apply defaults before calling AWS (e.g., force_destroy=false for S3 Bucket)

**Custom diff (TerraformCustomDiff)**: <yes | no>
<If yes: >
  This resource has a TerraformCustomDiff. Check `.agents/specs/tf-business-logic-catalog.md`
  for the entry for `<TF_NAME>`. The custom diff suppresses certain field diffs to avoid
  reconcile loops — port this logic to Observe() comparison (return ResourceUpToDate=true
  for those fields even if AWS state differs).

**Connection details**: <yes | no>
<If yes: >
  This resource publishes connection details. Return these from Create() via
  `managed.ExternalCreation{ConnectionDetails: connection.Details{...}}`:
  Keys to publish: <list the keys found in AdditionalConnectionDetailsFn, e.g.:
    - "id" → cr.Status.AtProvider.ID (or from Create response)
    - "arn" → cr.Status.AtProvider.ARN
    - "endpoint" → cr.Status.AtProvider.Endpoint
    - "password" → (from Create response only — not stored in spec)>

**Late init ignored fields**: <list | "none">
<If any: >
  Skip these fields during late initialization (do NOT copy from AWS response to spec):
  <list fields>
  Use `internal/native/lateinit.go` with the ignore list configured.

**MoveToStatus fields**: <list | "none">
<If any: >
  These fields are in Observation struct only. They are read-only and populated from AWS
  responses. Do NOT attempt to set them via AWS API calls, and do NOT include them in
  update comparisons. Populate them in Observe() from the AWS describe response.

**Cross-resource references**: <list | "none">

### Schema Field Classification

The following table is extracted from `config/schema.json` for `<TF_NAME>`.
It determines how each field must be handled in the `isUpToDate` comparison.
**Getting this wrong causes infinite update loops.**

| Field | Schema Flags | isUpToDate Treatment |
|-------|-------------|----------------------|
<for each field from step 4l, one row>

Key rules:
- `required` fields: always compare, no nil guard
- `optional` fields: nil-guard (`if spec.X != nil && *spec.X != observed`)
- `computed`-only fields: NEVER compare (observation-only, e.g., `arn`, `status`)
- **Optional blocks with `min=0`**: if spec sub-struct is nil, return true (accept AWS defaults).
  AWS always returns default values for these blocks even when the user didn't set them.
  Treating nil-spec + non-nil-observed as "not up to date" causes an infinite Update loop.
<If any: >
  Fields with +crossplane:generate:reference annotations. Reference resolution code is
  generated by crossplane-tools and lives in the native/ package's zz_resolve_references.go.
  In your CRUD code: the resolved values are already present in spec ForProvider by the time
  CRUD methods are called — just read them normally (e.g., cr.Spec.ForProvider.RoleARN).
  Do NOT re-resolve them in CRUD code.

**Multi-version CRD**: <yes | no>
<If yes: >
  This resource has both v1beta1 and v1beta2 (storage version: <version>). During the
  parallel phase, implement against the storage version only. At cutover, the cutover
  ticket handles registering all versions and implementing conversion.

**IAM policy field**: <yes | no>
<If yes: >
  This resource has a policy field containing JSON. Use `internal/native/policy.go`
  (wraps `awspolicyequivalence.PoliciesAreEquivalent`) for comparison in Observe().
  Never use `reflect.DeepEqual` or string comparison for IAM policy JSON — AWS normalizes
  key ordering and whitespace, causing infinite reconcile loops.

**Tags**: yes (all resources with taggable AWS resources)
  Use `internal/native/tags.go` for tag diff computation. Read ProviderConfig default_tags
  and merge with spec.forProvider.tags before comparing against AWS resource tags.
  Populate `status.atProvider.tagsAll` with the merged tag set.

### TF Reference Implementation
Read `vendor/github.com/upbound/terraform-provider-aws/internal/service/<SERVICE>/` to understand:
1. What AWS API calls TF makes in resourceCreate/resourceRead/resourceUpdate/resourceDelete
2. What error codes signal "not found" (use these in NotFound detection in Observe)
3. How TF handles eventual consistency (retry loops, waiter functions)
4. Field normalization or transformation before API calls
5. What fields TF computes vs what it passes through from config

### Implementation Checklist (TDD Order)
1. Write unit tests in `internal/controller/<SERVICE>/<resource_file>/crud_test.go`
   with a mocked AWS client. Tests MUST FAIL before implementation.
2. Implement Observe() — call Describe/Get, map to Observation, detect NotFound
3. Run `go test ./internal/controller/<SERVICE>/<resource_file>/...` — Observe tests pass
4. Implement Create() — call Create API, set external name, publish connection details
5. Implement Update() — call Update/Modify API (respect business logic)
6. Implement Delete() — call Delete API, handle NotFound gracefully (idempotent)
7. Wire thin wrappers in both scope controller.go files
8. Run full test suite: `go test ./internal/controller/<SERVICE>/...`
9. Verify no upjet imports: `grep -r "crossplane/upjet" internal/controller/<SERVICE>/`
10. Verify late initialization for AWS-defaulted fields (e.g., type, logging level)
11. Verify Unavailable() condition for non-ACTIVE states
12. Verify ALL mutable fields compared in isUpToDate (cross-check with schema.json — no gaps)
13. Verify build tags present: `head -1 internal/controller/<SERVICE>/<resource_file>/crud.go`
14. Verify working tree is clean: `git status --short -- 'apis/*/<SERVICE>/' 'internal/controller/*/<SERVICE>/'`
```

**Acceptance criteria** (as array):
```
["go build ./internal/controller/<SERVICE>/<resource_file>/... passes with zero errors",
 "go build ./internal/controller/cluster/<SERVICE>/<resource_file>raw/... passes with zero errors",
 "go build ./internal/controller/namespaced/<SERVICE>/<resource_file>raw/... passes with zero errors",
 "go test ./internal/controller/<SERVICE>/<resource_file>/... passes (unit tests with mocked AWS client)",
 "No upjet imports: grep -r 'crossplane/upjet' internal/controller/<SERVICE>/<resource_file>/ returns empty",
 "External name correctly set on Create via meta.SetExternalName",
 "Observe returns ResourceExists=false for missing resources (NotFound from AWS)",
 "Observe returns ResourceUpToDate=false when spec differs from AWS state",
 "Observe returns ResourceUpToDate=true when spec matches AWS state",
 "Delete is idempotent (returns nil when resource is already gone)",
 "All MoveToStatus fields populated in status.atProvider from Observe",
 "Connection details published in ExternalCreation (if applicable for this resource)",
 "Late initialization implemented for all AWS-defaulted fields (e.g., type, logging defaults) — Observe returns ResourceLateInitialized: true",
 "Unavailable() condition set in Observe for non-ACTIVE states (DELETING, PENDING, etc.)",
 "ALL mutable fields in every sub-struct compared in isUpToDate — no silent drift gaps (e.g., check every field in encryption, logging, tracing config blocks)",
 "No misleading idempotency comments (verify AWS Create API behavior before commenting)",
 "Build tag //go:build <SERVICE> || all present on all controller files",
 "Async operations tracked via CR annotations (if UseAsync=true for this resource)"]
```

---

## Step 8: Create E2E RAW Tickets (1 per resource)

For each resource, create an E2E ticket for the RAW controller. Same hard-gate rules as baseline.

**Ticket structure:**

```
id:         native-<SERVICE>-<resource_slug>-e2e
title:      E2E test <resource_go>RAW controller (<SERVICE>)
plan:       native-<SERVICE>
phase:      e2e
labels:     ["stage:executor"]
priority:   Normal
depends_on: ["native-<SERVICE>-<resource_slug>"]
```

**Description:**

```
## E2E Test: <resource_go>RAW (<SERVICE>)

Run end-to-end test of the native RAW controller using the RAW example manifest.
Same hard failure rules as baseline — any failure means mark FAILED, no workarounds.

### Spec Reference
`.agents/specs/terraform-removal-migration.md` — Section "Step 3: E2E Test RAW"

### RAW Example Manifest
examples/<SERVICE>/cluster/<version>/<resource_file>raw.yaml
(Created by scaffold ticket. Change `kind: <resource_go>` → `kind: <resource_go>RAW`.)

### E2E Command
```bash
export UPTEST_CLOUD_CREDENTIALS="DEFAULT='[default]
aws_access_key_id = ${AWS_ACCESS_KEY_ID}
aws_secret_access_key = ${AWS_SECRET_ACCESS_KEY}'"

export UPTEST_EXAMPLE_LIST="examples/<SERVICE>/cluster/<version>/<resource_file>raw.yaml"
make e2e SUBPACKAGES="config <SERVICE>"
```

### Failure Rules — ZERO TOLERANCE (same as baseline)
- Permission denied → Mark FAILED
- Quota/limit exceeded → Mark FAILED
- Infrastructure issue → Mark FAILED
- Controller error → Mark FAILED
No workarounds. Record exact error and mark Failed.

### What to Verify
1. `<resource_go>RAW` resource transitions to Ready condition
2. No TF-related errors in controller logs (must be pure native SDK path)
3. `<resource_go>RAW` deletes cleanly (DeletedSuccessfully condition)
4. External name is set correctly on the CR after creation
5. Full compilation check passes BEFORE running e2e:
   ```bash
   go build ./apis/cluster/<SERVICE>/... && go build ./apis/namespaced/<SERVICE>/... && \
   go build ./internal/controller/<SERVICE>/... && \
   go build ./internal/controller/cluster/<SERVICE>/... && \
   go build ./internal/controller/namespaced/<SERVICE>/... && \
   go test ./internal/controller/<SERVICE>/...
   ```
6. Working tree is clean — no untracked files in service directories:
   ```bash
   test -z "$(git ls-files --others --exclude-standard -- 'apis/*/<SERVICE>/' 'internal/controller/*/<SERVICE>/')" || \
     echo "FAIL: untracked files found"
   ```

### Result Capture
Paste the final uptest output (last 50 lines) before marking Done.
```

**Acceptance criteria** (as array):
```
["<resource_go>RAW resource reaches Ready condition during e2e",
 "<resource_go>RAW resource deletes cleanly with no errors",
 "Controller logs show no upjet or terraform errors",
 "External name correctly populated on the CR after creation",
 "No AWS permission errors",
 "Full compilation check passes before e2e (go build + go test on all service packages)",
 "No untracked files in service directories (git ls-files --others)",
 "E2E output captured in ticket"]
```

---

## Step 9: Create TF Regression Ticket (1 per service)

Create ONE regression ticket that depends on ALL e2e RAW tickets for the service.

**Ticket structure:**

```
id:         native-<SERVICE>-tf-regression
title:      TF regression test for <SERVICE> — re-run TF e2e after native work
plan:       native-<SERVICE>
phase:      regression
labels:     ["stage:executor"]
priority:   High
depends_on: ["native-<SERVICE>-<resource_slug_1>-e2e",
             "native-<SERVICE>-<resource_slug_2>-e2e",
             ... one entry per resource ...]
```

**Description:**

```
## TF Regression Test: <SERVICE>

Re-run ALL original TF e2e tests for <SERVICE> AFTER the native implementation work.
This confirms native changes did not break existing TF controllers.

### Why This Gate Exists
Native controller work touches shared infrastructure: options struct, connector, build hooks,
credential cache. A regression here means native changes broke the TF path — which MUST NOT
happen during the parallel phase. This gate must pass before verification and cutover.

### Spec Reference
`.agents/specs/terraform-removal-migration.md` — Section "Step 4: TF Regression Test"

### Resources to Test (run one e2e per resource)

<for each resource:>
**<resource_go>** (<TF_NAME>)
Example: <example_manifest_path>

```bash
export UPTEST_EXAMPLE_LIST="<example_manifest_path>"
make e2e SUBPACKAGES="config <SERVICE>"
```

### Failure Rules
Same hard gate as baseline: any failure → Mark FAILED. No workarounds. Record exact error.

Note: Run all resources even if one fails — collect all results before marking the ticket.

### Acceptance Criteria
```

**Acceptance criteria** (as array — one per resource plus the overall):
```
["<resource_go_1> TF controller still passes e2e (kind=<resource_go_1>, not RAW)",
 "<resource_go_2> TF controller still passes e2e (kind=<resource_go_2>, not RAW)",
 ... one per resource ...
 "All TF resources in <SERVICE> reach Ready condition with original TF kinds",
 "No regressions introduced by native controller infrastructure changes",
 "E2E output for each resource captured in ticket"]
```

---

## Step 10: Create Agent Verification Ticket (1 per service)

**Ticket structure:**

```
id:         native-<SERVICE>-verify
title:      Agent verification: parity test for <SERVICE>
plan:       native-<SERVICE>
phase:      verify
labels:     ["stage:executor"]
priority:   High
depends_on: ["native-<SERVICE>-tf-regression"]
```

**Description:**

```
## Agent Verification: <SERVICE>

Deploy TF and RAW resources side-by-side and verify behavioral parity. This is the final
gate before the service is considered migration-ready. Use the nohup+poll pattern — this
may run 20–60+ minutes.

### Spec References
`.agents/specs/terraform-removal-migration.md` — Sections "Step 5: Agent Verification"
and "Agent Verification Detail" (parity report format, verification tiering)

### Prerequisites
- AWS CLI must be available. If not installed, install it first:
  ```bash
  cd /tmp && curl -sL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" \
    && unzip -qo awscliv2.zip && sudo ./aws/install --update
  aws --version   # must succeed
  aws sts get-caller-identity  # must return the test account
  ```

### Mutable Fields to Test

The following fields MUST be patched on both TF and RAW resources during verification.
These are all non-computed, user-settable fields from the Schema Field Classification
(see the implement ticket for the full table from `config/schema.json`).

<for each resource, list ALL mutable fields from the schema classification:>
**<resource_go>**:
| Field | Schema | Patch Test |
|-------|--------|------------|
<for each `required` field:>
| `<field>` | required | Patch to new value, verify AWS state matches on both TF and RAW |
<for each `optional` (non-computed) field:>
| `<field>` | optional | Set value, verify AWS state matches; then remove, verify AWS defaults restored |
<for each block with min=0:>
| `<block_name>` | block min=0 | Set explicit values, verify; then remove, verify AWS defaults |
<for tags:>
| `tags` | optional | Add tag, verify; update tag value, verify; remove tag, verify |

Do NOT skip fields. Every mutable field must be tested for parity.

### Verification Steps (for each resource)

For each resource `<resource_go>`:

**Phase 1: Create + AWS-side verification**
1. Apply TF example manifest → wait for Ready (timeout 10m)
2. Apply RAW example manifest → wait for Ready (timeout 10m)
3. **AWS CLI: describe both resources and compare state** (MANDATORY)
   ```bash
   # Example for SFN:
   aws stepfunctions describe-state-machine --state-machine-arn <TF_ARN> --region <REGION> > /tmp/tf-state.json
   aws stepfunctions describe-state-machine --state-machine-arn <RAW_ARN> --region <REGION> > /tmp/raw-state.json
   # Compare relevant fields (exclude timestamps, ARNs, names):
   diff <(jq '{definition,roleArn,type,loggingConfiguration,tracingConfiguration,encryptionConfiguration}' /tmp/tf-state.json) \
        <(jq '{definition,roleArn,type,loggingConfiguration,tracingConfiguration,encryptionConfiguration}' /tmp/raw-state.json)
   ```
   Adapt the AWS CLI command and jq filter for the specific service/resource.

**Phase 2: Update each mutable field**
4. For EACH mutable field in the table above:
   a. Patch the field on the TF resource → wait for Synced=True
   b. Patch the same field on the RAW resource → wait for Synced=True
   c. **AWS CLI: describe both and compare** — the field must have the same value
   d. Record result: ✓ or ✗ with details

**Phase 3: Tags lifecycle**
5. Add a new tag to both → verify via AWS CLI
6. Update the tag value on both → verify via AWS CLI
7. Remove the tag from both → verify via AWS CLI

**Phase 4: Delete + cleanup verification**
8. Delete both resources → wait for Gone (timeout 10m)
9. **AWS CLI: confirm both resources are actually deleted** (MANDATORY)
   ```bash
   aws stepfunctions describe-state-machine --state-machine-arn <TF_ARN> --region <REGION> 2>&1 | grep -i "does not exist"
   aws stepfunctions describe-state-machine --state-machine-arn <RAW_ARN> --region <REGION> 2>&1 | grep -i "does not exist"
   ```

### nohup+Poll Pattern (REQUIRED — long-running verification)
```bash
cat > /tmp/run-verification-<SERVICE>.sh << 'SCRIPT'
#!/bin/bash
set -euo pipefail
LOG=/tmp/verify-<SERVICE>.log

echo "=== Agent Verification: <SERVICE> ===" | tee -a $LOG
echo "Date: $(date -u +%Y-%m-%dT%H:%M:%SZ)" | tee -a $LOG

# Verify AWS CLI is available
aws --version || { echo "FAIL: AWS CLI not installed" | tee -a $LOG; exit 1; }
aws sts get-caller-identity || { echo "FAIL: AWS credentials not working" | tee -a $LOG; exit 1; }

# ... verification steps for each resource ...

echo "=== Verification Complete ===" | tee -a $LOG
SCRIPT
chmod +x /tmp/run-verification-<SERVICE>.sh

nohup /tmp/run-verification-<SERVICE>.sh >> /tmp/verify-<SERVICE>.log 2>&1 &
VPID=$!
echo "Verification PID: $VPID"

# Poll every 60s until complete
while kill -0 $VPID 2>/dev/null; do
    echo "--- Still running ($(date -u +%H:%M:%S)) ---"
    tail -5 /tmp/verify-<SERVICE>.log
    sleep 60
done
echo "=== Verification finished ==="
cat /tmp/verify-<SERVICE>.log
```

### Parity Report Format (attach to ticket before marking Done)
```
=== Agent Verification Report: <SERVICE> ===
Service: <SERVICE>
Resources: <N>
Date: <ISO timestamp>
AWS CLI: <version>
Account: <account ID from sts get-caller-identity>

<For each resource:>
Resource: <resource_go>
  Status: PASS/FAIL
  Lifecycle: create ✓/✗ | observe ✓/✗ | update ✓/✗ | delete ✓/✗

  AWS-side create comparison:
    <diff output or "IDENTICAL">

  Mutable fields tested: N/M
    ✓/✗ definition: TF AWS={...} RAW AWS={...}
    ✓/✗ role_arn: TF AWS={...} RAW AWS={...}
    ✓/✗ <field>: TF AWS={...} RAW AWS={...}
    ✓/✗ tags (add/update/remove): TF AWS={...} RAW AWS={...}

  AWS-side delete confirmation:
    TF: <not found / error>
    RAW: <not found / error>

Overall: N/N PASS → MIGRATION READY
  (or: N/M PASS — <list failures> — NOT READY)
```
```

**Acceptance criteria** (as array):
```
["AWS CLI installed and working (aws sts get-caller-identity succeeds)",
 "Parity report generated and attached to ticket",
 "ALL resources in <SERVICE> show PASS in parity report",
 "Lifecycle verified for each resource: create ✓ observe ✓ update ✓ delete ✓",
 "ALL mutable fields patched and compared via AWS CLI (not just tags)",
 "AWS-side state compared after create (describe both resources via CLI)",
 "AWS-side deletion confirmed via CLI (both resources gone from AWS)",
 "Report concludes: MIGRATION READY",
 "No resource left in error state in the test cluster after verification"]
```

---

## ~~Step 11: Cutover~~ — NOT created per service

Cutover (renaming RAW → original, removing TF scaffolding) is a **global operation** done once
after ALL services have been migrated to native controllers. It is NOT part of the per-service
migration plan. Cutover tickets will be created separately when the full migration is complete.

---

## Step 12: Print Summary Report

After ALL tickets have been created successfully, print the following report to the
conversation (this is NOT a ticket — just output to the user):

```
## Migration Plan Created: <SERVICE>

Plan label: plan:native-<SERVICE>
Executor ant will pick up stage:executor tickets automatically.

### Resources Discovered (<N> total)
<For each resource:>
  <TF_NAME>
    → Kind: <resource_go>RAW
    → Storage version: <version>
    → External name: <strategy>
    → Multi-version: <yes/no>
    → Async: <yes/no>
    → Business logic: <yes/no>
    → Connection details: <yes/no>
    → Example: <path or MISSING>

### Tickets Created (<total_count> total)
  Phase baseline:   <N> tickets  (native-<SERVICE>-baseline-*)
  Phase scaffold:   1 ticket     (native-<SERVICE>-scaffold)
  Phase implement:  <N> tickets  (native-<SERVICE>-<resource_slug>)
  Phase e2e:        <N> tickets  (native-<SERVICE>-*-e2e)
  Phase regression: 1 ticket     (native-<SERVICE>-tf-regression)
  Phase verify:     1 ticket     (native-<SERVICE>-verify)
  ─────────────────────────────────────────────────────────
  TOTAL:            <4N+3> tickets

### Dependency Chain
  baseline-<resource> ──┐
                         ├──→ <resource> (implement) ──→ <resource>-e2e ──┐
  scaffold ─────────────┘                                                   ├──→ tf-regression ──→ verify
  (repeat for each resource) ────────────────────────────────────────────────┘

Note: Cutover is a global operation done after ALL services are migrated. Not created here.

### ⚠️  Warnings
<Only print sections that apply:>

🔴 ASYNC resources (UseAsync=true) — require async annotation pattern:
   <list TF names>

🔴 BUSINESS LOGIC resources (TerraformConfigurationInjector or TerraformCustomDiff):
   Must check .agents/specs/tf-business-logic-catalog.md for each:
   <list TF names>

🟡 MULTI-VERSION resources — conversion webhooks needed at cutover:
   <list TF names with versions>

🟡 MISSING EXAMPLE MANIFESTS — baseline executor must locate before running e2e:
   <list TF names>

🟡 TerraformID() EXTRACTOR — must rewrite to ExtractResourceID() in RAW types:
   <list TF names with affected fields>

🟡 CONNECTION DETAILS — publish in ExternalCreation:
   <list TF names with their keys>

🟡 IAM POLICY FIELDS — use internal/native/policy.go for comparison:
   <list TF names>

### Next Steps
1. Executor ant will automatically claim baseline and scaffold tickets (no depends_on).
2. Implement tickets unblock once both scaffold AND their baseline are Done.
3. If a baseline ticket Fails, mark it Failed — downstream tickets for that resource
   are automatically blocked (depends_on not satisfied).
4. The tf-regression gate depends on ALL e2e tickets. If some e2e tickets Failed,
   tf-regression is still claimable (Failed tickets count as "resolved" for depends_on).
   See spec for exact semantics.

Phase 0 infrastructure tickets (native connector, async system, etc.) must be Done
before scaffold and implement tickets can succeed. Verify Phase 0 status separately.
```

---

## Error Handling

Throughout execution:

- If `CreateTicket` fails for any ticket, report the failure with the ticket ID and error,
  then continue creating remaining tickets. List all failures at the end.

- If metadata collection (Step 4) fails for a resource (e.g., grep error), use "Unknown —
  manual inspection required" for that field and continue. Flag the resource in warnings.

- If a resource name does not match `aws_<SERVICE>_*` pattern, emit a warning:
  ```
  ⚠️  Unexpected resource name: <TF_NAME> (expected aws_<SERVICE>_* prefix)
     Using full name after aws_ as resource identifier.
  ```
  Then continue processing.

- Never stop mid-execution due to a single resource failure. Complete the full ticket graph.
