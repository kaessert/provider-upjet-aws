---
name: plan
description: Create an implementation plan as pheromone tickets (stage:executor). Each task becomes one ticket on the board, ready for the executor ant. Does NOT execute — only creates tickets.
allowed-tools:
  - Read
  - Glob
  - Grep
  - Agent
  - AskUserQuestion
  - Bash
  - Write
  - CreateTicket
---

# Plan Creation Skill

Create a detailed implementation plan and publish it as pheromone tickets. The executor ant will automatically pick up and work through each ticket.

<scope_boundaries>
This skill ONLY creates plans. It does NOT:
- Execute the plan
- Start implementing code
- Check off any tasks
</scope_boundaries>

## Instructions

### Step 1: Determine Plan Name

Parse `$ARGUMENTS` for the plan name:
- **First word/argument present**: Use it as the plan name
- **No arguments**: Use AskUserQuestion to ask for a name

Valid name format: lowercase with hyphens (e.g., `phase-0`, `native-sfn`, `add-credential-cache`).

### Step 2: Gather Requirements

Understand the task:
- What is the goal?
- What constraints exist?
- What is the scope?

<when_to_ask>
Use AskUserQuestion when:
- The task could be interpreted in multiple distinct ways
- A key constraint (technology, pattern, scope) is genuinely ambiguous
- The user's intent has meaningful alternatives

Do NOT ask when:
- You can infer the intent from context
- Standard patterns exist in the codebase
- The ambiguity is minor and can be resolved by following conventions
</when_to_ask>

### Step 3: Research the Codebase

<research_approach>
Use Read, Glob, Grep, and Agent (with explore subagent) to understand:
- Relevant existing code and patterns
- Files that will need creation or modification
- Dependencies and constraints
- Existing specs in `.agents/specs/` that define requirements

Start with `.agents/specs/INDEX.md` — it lists all specs and catalogs. Read the relevant spec
sections that map to the plan being created.

Thoroughness: Spend adequate time exploring before writing the plan. A well-researched plan prevents rework.
</research_approach>

### Step 4: Design the Tasks

<verification_scope_guidance>
Based on what the plan modifies, include relevant verification:

| Modified Area | Verification Type |
|--------------|-------------------|
| `internal/native/` | `go test ./internal/native/...` |
| `internal/controller/` | `go test ./internal/controller/...` |
| `apis/cluster/` or `apis/namespaced/` | `go build ./apis/...` |
| Config changes | `make check-diff` |
| Build hooks / templates | `make generate` + verify output |
| Any Go code | `golangci-lint run ./<changed-packages>/...` |
| E2E test resources | `make e2e SUBPACKAGES="config <service>"` |
</verification_scope_guidance>

Design tasks using these guidelines:

<task_format_rules>
- ONE ticket per logical unit (30min - 2hrs of effort)
- Related small items are details within one ticket's description, not separate tickets
- [BLOCKING] = Must be done first, unblocks other tasks
- [P0] = Critical path, high priority
- No marker = Normal priority
- Order tasks by critical path (blocking/P0 first)
- Every implementation ticket needs acceptance criteria
</task_format_rules>

### Step 5: Create Pheromone Tickets

Create one ticket per task using `CreateTicket`. Use readable IDs like `<plan-name>-01`, `<plan-name>-02`:

```
CreateTicket:
  id: <plan-name>-01
  title: "<Task title>"
  description: "<Description with file paths, implementation details, and spec references>"
  acceptance_criteria: ["criteria 1", "criteria 2"]
  labels: ["stage:executor", "plan:<plan-name>"]
  plan: <plan-name>
  priority: High  (for BLOCKING/P0) or Normal
```

**For tasks with dependencies**: Use `depends_on` to enforce ordering:

```
CreateTicket:
  id: <plan-name>-02
  title: "<Dependent task>"
  depends_on: ["<plan-name>-01"]
```

**Ticket content guidelines**:
- Title: concise, action-oriented (e.g., "Create native connector framework at internal/native/connector.go")
- Description: task details + file paths + spec references
- Acceptance criteria: testable conditions that define "done"
- Labels: always include `stage:executor` and `plan:<plan-name>`
- Reference specs: every ticket description should point to the relevant section of `.agents/specs/`

### Step 6: Report Completion

Tell the user:
- Number of tickets created on the pheromone board
- Key [BLOCKING] / [P0] tickets highlighted
- Dependency chain visualization
- The executor ant will automatically pick them up

## Important Rules

<critical_constraints>
1. **DO NOT EXECUTE** — Only create tickets
2. **DO NOT IMPLEMENT** — Planning only
3. **ALL TICKETS START AS ToDo** — Never create InProgress tickets
</critical_constraints>

<task_format_rules>
4. **One ticket per logical unit** — 30min to 2hr effort each
5. **Order by critical path** — [BLOCKING] and [P0] tasks first, use depends_on
6. **Include acceptance criteria** — As a list on every implementation ticket
7. **Verification tickets** — Just describe the command to run
</task_format_rules>

<quality_guidelines>
8. **Be thorough** — Research codebase and specs before writing
9. **Be specific** — Include file paths, function names, implementation details in descriptions
10. **Be realistic** — Break complex tasks into manageable units
11. **Follow conventions** — Match existing code patterns
12. **Reference specs** — Every ticket should point to relevant spec sections
13. **Never edit zz_ files** — They are auto-generated
</quality_guidelines>

## Project Context

This is a Go project (provider-upjet-aws) — a Crossplane provider managing AWS resources.
Key specs live in `.agents/specs/`. The executor ant follows TDD (test first, then implement).
Build/test commands: `go test`, `go build`, `golangci-lint run`, `make generate`, `make check-diff`.
See `CLAUDE.md` at the repo root for full project context.

## Example Ticket Creation

```
CreateTicket:
  id: phase-0-01
  title: "[BLOCKING] Create native controller options struct"
  description: |
    Create internal/native/options.go with a native Options struct that wraps
    crossplane-runtime's controller.Options. No upjet imports.

    Spec: .agents/specs/terraform-removal-migration.md section 0.1

    Files:
    - internal/native/options.go
  acceptance_criteria:
    - "Options struct compiles without upjet imports"
    - "go test ./internal/native/... passes"
    - "Can be used in a Setup() function signature"
  labels: ["stage:executor", "plan:phase-0"]
  plan: phase-0
  priority: High

CreateTicket:
  id: phase-0-02
  title: "Create native connector framework"
  description: |
    Create internal/native/connector.go with ProviderConfig → aws.Config resolution.
    Must call clients.GetAWSConfigWithTracking.

    Spec: .agents/specs/terraform-removal-migration.md section 0.2

    Files:
    - internal/native/connector.go
  acceptance_criteria:
    - "Connector resolves ProviderConfig to aws.Config"
    - "Integration test passes with mock ProviderConfig"
    - "No upjet imports"
  labels: ["stage:executor", "plan:phase-0"]
  plan: phase-0
  depends_on: ["phase-0-01"]
  priority: High
```
