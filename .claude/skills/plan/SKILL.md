---
name: plan
description: Create an implementation plan as pheromone tickets (stage:executor). Each task becomes one ticket on the board, ready for the executor ant. Does NOT execute — only creates tickets.
allowed-tools:
  - Read
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Bash
  - Write
---

# Plan Creation Skill (hive-classic-ralph)

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

Valid name format: lowercase with hyphens (e.g., `add-user-auth`, `refactor-api`).

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
Use Read, Glob, Grep, and Task (with Explore agent) to understand:
- Relevant existing code and patterns
- Files that will need modification
- Dependencies and constraints
- Similar implementations to follow

Thoroughness: Spend adequate time exploring before writing the plan. A well-researched plan prevents rework. Use the Explore agent for open-ended searches.
</research_approach>

**E2E Spec Check**: If the plan touches user-facing features, check `.agents/tests/e2e/INDEX.md` for existing specs.

### Step 4: Design the Tasks

<verification_scope_guidance>
Based on what the plan modifies, include relevant verification:

| Modified Area | Verification Type |
|--------------|-------------------|
| `crates/hive/` (TUI) | tmux pane testing |
| `crates/hive/` (CLI) | CLI command testing |
| `crates/hived/` | Backend unit tests + API testing |
| `hive-frontend/` | Vitest + Playwright E2E |
| User-facing features | Playwright E2E spec |
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

Create one ticket per task using `pheromone create`. Use readable IDs like `<plan-name>-01`, `<plan-name>-02`:

```bash
pheromone create <plan-name>-01 \
  --status ToDo \
  --title "<Task title>" \
  --description "<Description>\n\nAcceptance: <measurable criteria>" \
  --labels "stage:executor,plan:<plan-name>"
```

**For [BLOCKING] / [P0] tasks**: Create these first. The executor ant processes in creation order.

**For tasks with dependencies**: Use `--depends-on <ticket-id>` to enforce ordering:

```bash
pheromone create <plan-name>-02 \
  --status ToDo \
  --title "<Dependent task>" \
  --description "..." \
  --labels "stage:executor,plan:<plan-name>" \
  --depends-on <plan-name>-01
```

**For verification tasks** (run tests, check linting): Create them last:

```bash
pheromone create <plan-name>-verify-01 \
  --status ToDo \
  --title "All tests pass: run cargo test" \
  --description "Run cargo test -- --quiet. Confirm all tests pass with no failures." \
  --labels "stage:executor,plan:<plan-name>"
```

**Ticket content guidelines**:
- Title: concise, action-oriented (e.g., "Add user auth types to models.rs")
- Description: task details + acceptance criteria in one field
- Labels: always include `stage:executor` and `plan:<plan-name>`

### Step 6: Report Completion

Tell the user:
- Number of tickets created on the pheromone board
- Key [BLOCKING] / [P0] tickets highlighted
- The executor ant will automatically pick them up
- If execution gets stuck, run `/investigate <symptom>` then `/plan fix-<issue>` to create a fix ticket

## Important Rules

<critical_constraints>
1. **DO NOT EXECUTE** — Only create tickets
2. **DO NOT IMPLEMENT** — Planning only
3. **ALL TICKETS START AS ToDo** — Never create InProgress tickets
</critical_constraints>

<task_format_rules>
4. **One ticket per logical unit** — 30min to 2hr effort each
5. **Order by critical path** — [BLOCKING] and [P0] tasks first, use depends-on
6. **Include acceptance criteria** — In the description field of every implementation ticket
7. **Verification tickets are simple** — Just describe the command to run
</task_format_rules>

<quality_guidelines>
8. **Be thorough** — Research codebase before writing
9. **Be specific** — Include file paths, function names, implementation details in descriptions
10. **Be realistic** — Break complex tasks into manageable units
11. **Follow conventions** — Match existing code patterns
12. **Executor performs verification** — Include verify tickets at the end
13. **Playwright for E2E** — Use Playwright tests in `hive-frontend/e2e/`, not deprecated Chrome plugin specs
</quality_guidelines>

## Example Ticket Creation

```bash
# BLOCKING: types must exist before middleware can use them
pheromone create add-auth-01 \
  --status ToDo \
  --title "[BLOCKING] Create auth types in src/types/auth.rs" \
  --description "Define User, Session, and JWT payload types.\n- User struct: id, email, password_hash fields\n- Session struct: user_id, token, expires_at\n- JwtPayload struct for token claims\n\nAcceptance: Types compile; used by auth middleware" \
  --labels "stage:executor,plan:add-auth"

pheromone create add-auth-02 \
  --status ToDo \
  --title "Add JWT middleware to protected routes" \
  --description "Validate JWT tokens on protected routes.\n- Extract token from Authorization header or cookie\n- Verify signature and expiration\n- Attach user to request context\n\nAcceptance: Middleware rejects invalid tokens, passes valid ones" \
  --labels "stage:executor,plan:add-auth" \
  --depends-on add-auth-01

pheromone create add-auth-verify-01 \
  --status ToDo \
  --title "Auth tests pass" \
  --description "Run cargo test -p hived -- auth. All auth tests pass." \
  --labels "stage:executor,plan:add-auth"
```
