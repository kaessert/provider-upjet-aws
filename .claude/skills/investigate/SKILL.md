---
name: investigate
description: Troubleshoot app functionality by autonomously exploring code, runtime state, and tests. Provide root cause analysis for bugs, unexpected behavior, and integration issues.
allowed-tools:
  # Code exploration
  - Read
  - Glob
  - Grep
  - Task
  # Runtime diagnostics
  - Bash
  # Browser diagnostics
  - mcp__claude-in-chrome__read_console_messages
  - mcp__claude-in-chrome__read_network_requests
  - mcp__claude-in-chrome__tabs_context_mcp
  - mcp__claude-in-chrome__read_page
  - mcp__claude-in-chrome__computer
  - mcp__claude-in-chrome__javascript_tool
  # Investigation workflow
  - AskUserQuestion
  - Write
  - Edit
  - Skill
---

# Investigate Skill

Autonomously troubleshoot app functionality issues by exploring code, runtime state, and running diagnostics to identify root causes.

<critical_constraint>
**DIAGNOSIS ONLY**

This skill identifies root causes; it does not implement fixes. The Edit tool may only be used to save investigation reports to `.agents/bugs/`.

Why: Separating diagnosis from fixing ensures thorough root cause analysis. Jumping to fixes often addresses symptoms rather than causes, and misses opportunities to identify systemic patterns. A complete diagnosis enables better fixes through `/plan`.
</critical_constraint>

<autonomous_execution>
Act immediately without asking permission. The user invoked /investigate because they want autonomous troubleshooting. Run health checks, search code, query databases, and examine logs without confirmation.

Only use AskUserQuestion when:
- The problem description is too vague to form any hypothesis
- Multiple equally-likely root causes require user input to prioritize
- The investigation is complete and you need to confirm next steps
</autonomous_execution>

<parallel_tool_calls>
Execute independent diagnostics simultaneously. In a single response, call multiple tools:

```
# Execute all in one response:
Bash: curl -s http://localhost:8080/api/health
Bash: sqlite3 ~/.local/share/hive/hive.db "SELECT * FROM ants;"
Bash: pgrep -a hived
```

Parallelize: health checks, database queries, file searches, browser diagnostics.
Sequence: only when one result determines the next query.
</parallel_tool_calls>

<state_tracking>
**After your 5th tool call**, output an Investigation State block. Update it every 3-4 calls thereafter.

```
## Investigation State (updated after tool call #X)
Hypotheses:
- [REFUTED] Database corruption - schema intact, data valid
- [CONFIRMED] Field name mismatch - types.ts vs handlers.rs
- [UNTESTED] Similar mismatches in other endpoints

Evidence collected:
- Health: OK | Network request: {"type": "brainstorm"} | Error: "missing field conversation_type"

Next: Search for similar field name mismatches
```

This prevents context loss in long investigations and provides an audit trail.
</state_tracking>

## Instructions

### Step 1: Understand the Problem

If `$ARGUMENTS` contains a description, use it. Otherwise ask:
```
What issue are you experiencing? Describe the symptom.
```

Extract: **What** (affected component), **When** (trigger conditions), **Expected** vs **Actual** behavior.

Classify: **Backend** (Rust/hived) or **Frontend** (SvelteKit/browser).

### Step 2: Form Hypotheses

**Backend Components**:
| Issue | Location |
|-------|----------|
| TUI | `crates/hive/src/tui/` |
| API | `crates/hived/src/http/`, `crates/hived/src/services/` |
| Database | `crates/hived/src/state/` |
| Ant operations | `crates/hived/src/providers/` |
| gRPC | `crates/hive-proto/` |
| Mind/agents | `crates/mind/`, `crates/hived/src/mind/` |

**Frontend Components**:
| Issue | Location |
|-------|----------|
| Pages | `hive-frontend/src/routes/` |
| Components | `hive-frontend/src/lib/components/` |
| API client | `hive-frontend/src/lib/api/` |
| State | `hive-frontend/src/lib/stores/` |

### Step 3: Run Health Checks

Execute in parallel:

**Backend**:
```bash
curl -s http://localhost:8080/api/health
curl -s http://localhost:8080/api/ants | jq .
pgrep -a hived
```

**Frontend** (via Chrome plugin, in parallel):
- `read_console_messages` pattern="error|Error|fail"
- `read_network_requests` urlPattern="/api/"
- `computer` action=screenshot

<frontend_api_bugs>
**For HTTP errors (4xx/5xx) from frontend API calls**: Always capture the actual network request via Chrome plugin BEFORE reproducing with curl. The browser's actual payload may differ from what you reconstruct manually.

```
# REQUIRED for frontend API bugs:
1. tabs_context_mcp (get tab ID)
2. read_network_requests urlPattern="/api/<endpoint>"  # See actual request payload
3. read_console_messages pattern="error|<status-code>"  # See error details
```

Common frontend-backend integration bugs:
- **Field name mismatch**: Frontend sends `type`, backend expects `conversation_type`
- **Enum format mismatch**: Frontend sends string, backend expects number (or vice versa)
- **Missing required fields**: Frontend type has optional field, backend requires it
- **Case mismatch**: Frontend sends `camelCase`, backend expects `snake_case`
</frontend_api_bugs>

<early_exit>
If health check returns "connection refused": Stop. Report "Daemon not running. Start with `cargo run -p hived`."
If frontend returns connection error: Stop. Report "Frontend not running. Start with `cd hive-frontend && npm run dev`."
</early_exit>

### Step 4: Explore Code

<tool_selection>
Use **Task with Explore agent** when you don't know the specific file, need to understand component connections, or the search spans multiple directories.

Use **direct Glob/Grep/Read** when you have an exact file path, precise function name, or specific stack trace to follow.
</tool_selection>

Strategy:
1. Find entry point (UI trigger, CLI command, API endpoint)
2. Trace execution path to affected area
3. Identify dependencies
4. Check error handling (look for swallowed errors)

### Step 5: Examine State

**Backend**:
```bash
sqlite3 ~/.local/share/hive/hive.db ".schema"
sqlite3 ~/.local/share/hive/hive.db "SELECT * FROM ants;"
```

**Frontend**: Use `read_page` for DOM, `javascript_tool` for runtime state.

### Step 6: Run Tests

```bash
# Backend
cargo test -p hived -- <keyword> --nocapture

# Frontend
cd hive-frontend && npm run test
```

### Step 7: Enable Debug Logging (if needed)

```bash
RUST_LOG=debug cargo run -p hived
RUST_LOG=hived::services=trace cargo run -p hived
```

### Step 8: Verify Hypotheses

<hypothesis_format>
**You MUST use this exact format for each hypothesis:**

```
## Hypothesis 1: [TESTING] Brief description
Belief: "I believe [X causes Y] because [Z]"
Confidence: low/medium/high
Test: [Specific command or check to run]
```

After executing:
```
## Hypothesis 1: [CONFIRMED/REFUTED] Brief description
Result: [What the test showed]
Evidence: [Specific output, line number, or data]
```
</hypothesis_format>

Even when the cause seems obvious, formally state and test at least one hypothesis. This prevents confirmation bias and creates an audit trail.

<termination>
**Success**: Root cause identified with high-confidence evidence (logs, failing test, code path confirmed).

**Inconclusive**: After testing 5 distinct hypotheses without finding root cause, stop and report what was eliminated.

**Blocked**: If investigation requires unavailable information (credentials, external service access), stop and explain what's needed.

A hypothesis attempt means: forming a specific theory, designing a test, and executing it. Reading code without testing doesn't count.
</termination>

### Step 9: Analyze Spec Origin

<mandatory_spec_search>
**Always search for specs.** Even "obvious" bugs may have spec implications. Only skip the classification if the search returns zero results AND the bug is clearly a typo/syntax error.

```bash
# REQUIRED - run these commands:
ls .agents/specs/ | grep -i "<feature-keyword>"
grep -r "<feature-keyword>" .agents/specs/
```

If no spec found, explicitly state: "**Spec**: None found (searched for '<keywords>')"
</mandatory_spec_search>

**Classification**:
| Type | Meaning | Action |
|------|---------|--------|
| **Code Bug** | Spec clear, implementation wrong | Fix code per spec |
| **Spec Bug** | Spec explicitly requires buggy behavior | Fix spec first |
| **Spec Ambiguity** | Spec unclear, implementation misinterpreted | Clarify spec first |
| **Spec Gap** | Spec doesn't cover this scenario | Extend spec first |
| **No Spec** | Feature implemented without spec | Create spec first |

### Step 10: Analyze Prevention (Code Bugs only)

For Code Bugs, launch an Explore agent:

```
Analyze missing validations for this bug:
- Root cause: [description]
- Location: [file:line]

Find:
1. Input validation gaps - earlier checks that could catch bad state
2. Type system gaps - stricter types preventing this class of bug
3. Assertion opportunities - debug_assert! or invariant checks
4. Error handling gaps - similar patterns of swallowed errors
5. Test coverage gaps - cases that would catch this

Report: specific locations, recommended guards, one-off vs systemic.
```

### Step 11: Write Report

<report_format>
## Investigation: [Problem]

### Summary
One-sentence root cause.

### Root Cause
**Location**: `path/to/file.rs:line` - description
**Mechanism**: How the bug manifests

### Spec Analysis
**Classification**: [Code Bug | Spec Bug | Spec Ambiguity | Spec Gap | No Spec]
**Spec**: `.agents/specs/<name>.md` or "None"
**Finding**: What's wrong

### Prevention (Code Bugs)
**Validation gaps**: What checks could have caught this
**Recommended guards**: Specific assertions to add
**Pattern**: One-off | Systemic

### Evidence
- Concrete proof: logs, test output, code paths

### Fix Approach
What needs to change (without implementing).

**Spec changes** (required for Spec Bug, Spec Ambiguity, Spec Gap, No Spec):
List the specific spec files that need updating/creating and what changes are needed.
Example: "Update `.agents/specs/conversation-agent-rust.md` to define cancellation UX behavior (partial response preservation, user feedback, session continuity)."

**Code changes**:
List the code changes needed to fix the bug.

Run `/plan fix-{issue-name}` to create pheromone tickets for the fix — executor ant picks them up (include spec changes when classification is not Code Bug).
</report_format>

### Step 12: Complete Investigation

1. Present the report
2. Ask: "Save to `.agents/bugs/{issue-name}.md`?"
3. If the spec changes are **clear-cut** (e.g., adding a missing field, documenting existing behavior, extending a section with well-understood requirements): end with "Run `/plan fix-{issue-name}` to create pheromone tickets for the fix — executor ant picks them up."
4. If the spec changes require **design exploration** (e.g., multiple valid approaches, ambiguous requirements, new behavior that needs user input): offer `/brainstorm` first — "Run `/brainstorm` to design the spec changes, then `/plan fix-{issue-name}` to create pheromone tickets — executor ant picks them up."

**Important**: When classification is Spec Bug, Spec Ambiguity, Spec Gap, or No Spec, the Fix Approach section MUST list the specific spec files to update/create and what changes are needed. Use your judgment on whether the spec changes are clear enough for `/plan` alone or need `/brainstorm` to explore the design first.

## Quick Reference

| Check | Backend | Frontend |
|-------|---------|----------|
| Running | `pgrep -a hived` | `curl http://10.0.0.6:5173/` |
| Health | `curl localhost:8080/api/health` | `read_console_messages` |
| State | `sqlite3 ~/.local/share/hive/hive.db` | `javascript_tool` |
| Logs | `RUST_LOG=debug` | `read_console_messages` |

| Symptom | Likely Cause | Check |
|---------|--------------|-------|
| Connection refused | Service not running | pgrep, curl |
| Entity not found | Wrong ID or deleted | Database |
| UI freeze | Blocking call | Event loop, async code |
| 500 error | Backend panic/error | hived logs |
| Page won't load | Build error | npm output, console |
| 422 from frontend | Field name/type mismatch | Chrome plugin network, compare types.ts vs handlers.rs |
| 400 from frontend | Missing required field | Chrome plugin network, check request payload |

## Pre-Completion Checklist

Before presenting the final report, verify:

- [ ] **Hypothesis tested**: At least one hypothesis formally stated and tested (Step 8)
- [ ] **Spec searched**: Ran `grep -r "<keyword>" .agents/specs/` (Step 9)
- [ ] **Prevention analyzed**: For Code Bugs, launched Explore agent for similar patterns (Step 10)
- [ ] **Frontend observed**: For frontend bugs, used Chrome plugin to capture actual request/error (Step 3)
- [ ] **Spec changes listed**: For non-Code-Bug classifications, Fix Approach lists specific spec files to update/create and what changes are needed (Step 11)
- [ ] **Handoff included**: Report ends with `/plan fix-{issue-name}` instruction (creates pheromone tickets; executor picks them up) (Step 12)

If any box is unchecked, go back and complete that step before finalizing.

## Rules

1. **Act immediately** - Run diagnostics without asking; results over conversation
2. **Verify before concluding** - Test hypotheses; first guess is often wrong
3. **Show evidence** - Include logs, test output, code paths
4. **Track state** - Long investigations lose context; update state explicitly
5. **Know when to stop** - 5 tested hypotheses without root cause = inconclusive
6. **Complete the workflow** - Diagnosis includes spec analysis and handoff

## Examples

### Code Bug

```
User: /investigate database connection panics on startup

Agent: [Runs health checks, finds panic in logs]
Agent: [Reads db.rs:45, finds unwrap on env var]
Agent: [Checks spec - says "validate config before connecting"]
Agent: [Launches Explore agent - finds 3 similar unwrap patterns]

## Investigation: Database Connection Panic

### Summary
Missing DATABASE_URL env var causes unwrap panic.

### Root Cause
**Location**: `crates/hived/src/state/db.rs:45`
**Mechanism**: `env::var("DATABASE_URL").unwrap()` panics when unset

### Spec Analysis
**Classification**: Code Bug
**Spec**: `.agents/specs/daemon-startup.md`
**Finding**: Spec requires config validation; code skips it

### Prevention
**Validation gaps**: No startup config validation
**Recommended guards**: Config validator checking all required env vars
**Pattern**: Systemic - 3 other unwrap calls on env vars

### Evidence
- Panic backtrace at line 45
- DATABASE_URL missing from environment

### Fix Approach
Replace unwrap with error handling. Add config validation module.
Run `/plan fix-db-panic` to create a pheromone ticket — executor picks it up.

---
Save to `.agents/bugs/db-panic.md`? [Yes]
```

### Spec Ambiguity

```
User: /investigate ant creation fails silently

Agent: [Health OK, searches code, finds error swallowed at line 142]
Agent: [Checks spec - "appropriate error responses" undefined]

## Investigation: Silent Ant Creation Failure

### Summary
Provider quota errors caught and swallowed.

### Root Cause
**Location**: `crates/hived/src/services/ant_service.rs:142`
**Mechanism**: Provider errors logged at debug, but returns Ok()

### Spec Analysis
**Classification**: Spec Ambiguity
**Spec**: `.agents/specs/rest-ant-endpoints.md` section 4.2
**Finding**: "Appropriate error responses" doesn't define provider errors

### Evidence
- Debug log: "Provider error: quota exceeded"
- Database: no ant record created
- API response: 200 OK

### Fix Approach
**Spec changes**: Update `.agents/specs/rest-ant-endpoints.md` section 4.2 — the phrase "appropriate error responses" needs to be replaced with concrete definitions: which HTTP status codes map to which provider errors, what the error response body looks like, and whether clients should retry.

**Code changes**: Propagate provider errors instead of swallowing them. Map provider error types to the HTTP status codes defined in the updated spec.

The spec changes here involve design decisions (error taxonomy, retry semantics) that need exploration.
Run `/brainstorm` to design the error response spec, then `/plan fix-silent-ant-creation` to create pheromone tickets — executor picks them up.

---
Save to `.agents/bugs/silent-ant-creation.md`? [Yes]
```

### Frontend-Backend Integration Bug

```
User: /investigate creating brainstorm conversation gives HTTP 422

Agent: [Gets tab context, captures network request via Chrome plugin]
Agent: [Sees request payload: {"workspace_id": "...", "type": "brainstorm"}]
Agent: [Sees response: 422 "missing field `conversation_type`"]

## Hypothesis 1: [TESTING] Field name mismatch
Belief: "Frontend sends 'type' but backend expects 'conversation_type'"
Confidence: high (error message is explicit)
Test: Compare frontend types.ts with backend handlers.rs

Agent: [Reads types.ts:665 - defines 'type: ConversationType']
Agent: [Reads handlers.rs:2345 - expects 'conversation_type: String']

## Hypothesis 1: [CONFIRMED] Field name mismatch
Result: Frontend and backend use different field names
Evidence: types.ts:665 vs handlers.rs:2345

Agent: [Searches specs: grep -r "conversation" .agents/specs/]
Agent: [Launches Explore agent for similar mismatches]

## Investigation: HTTP 422 Creating Conversation

### Summary
Field name mismatch: frontend sends `type`, backend expects `conversation_type`.

### Root Cause
**Location**: `hive-frontend/src/lib/api/types.ts:665` and `crates/hived/src/http/handlers.rs:2345`
**Mechanism**: Frontend CreateConversationSessionRequest uses `type`, backend deserializer expects `conversation_type`

### Spec Analysis
**Classification**: Code Bug
**Spec**: None found (searched for 'conversation', 'session')
**Finding**: No API contract spec exists; field names inconsistent

### Prevention
**Validation gaps**: No integration tests verifying frontend payloads match backend expectations
**Recommended guards**: API contract tests or OpenAPI spec generation
**Pattern**: Potentially systemic - check other request types for similar mismatches

### Evidence
- Chrome plugin network capture: payload contains `type`
- Backend error: "missing field `conversation_type`"
- curl with `conversation_type` succeeds (HTTP 201)

### Fix Approach
Add `#[serde(alias = "type")]` to backend, or rename frontend field.
Run `/plan fix-conversation-422` to create a pheromone ticket — executor picks it up.

---
Save to `.agents/bugs/conversation-422.md`? [Yes]
Run `/plan fix-conversation-422` to create a pheromone ticket — executor picks it up.
```
