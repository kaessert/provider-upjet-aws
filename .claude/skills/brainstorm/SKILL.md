---
name: brainstorm
description: Brainstorm a design through free-form dialogue. Searches for existing specs to update before creating new ones. Use AskUserQuestion to explore ideas, clarify requirements, and refine the concept before writing a spec to .agents/specs/. Includes an independent review for quality.
allowed-tools:
  - Read
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Write
  - Edit
---

# Brainstorm Skill

Collaborate with the user through dialogue to explore and refine a design idea, then capture it as a spec and have it independently reviewed.

## Instructions

### Step 1: Understand the Initial Idea

If `$ARGUMENTS` contains a description, use it as the starting point. Otherwise, ask:

```
What would you like to brainstorm? Give me a rough idea or problem you're thinking about.
```

### Step 2: Explore Through Dialogue

<dialogue_approach>
Use AskUserQuestion to have a free-form conversation. The goal is to uncover:

- **The core problem**: What pain point or need does this address?
- **The vision**: What does success look like?
- **The scope**: What's in and what's out?
- **The constraints**: Technical, time, resource limitations?
- **The users**: Who benefits and how do they interact?
</dialogue_approach>

<dialogue_guidelines>
1. Ask 1-4 questions at a time (matching AskUserQuestion limits), grouping related topics
2. Build on previous answers - reference what the user said to show active listening
3. Offer concrete options when helpful, but allow open input via "Other"
4. Challenge assumptions gently: "Have you considered...?"
5. Summarize understanding periodically to validate alignment
6. Dig deeper on vague answers: "Can you give an example?"
</dialogue_guidelines>

<proactivity_guidance>
Balance asking questions with suggesting ideas. When you have relevant insights from codebase research or domain knowledge:
- Propose concrete options rather than only asking open-ended questions
- Suggest tradeoffs between approaches you've identified
- Surface patterns from existing code that could inform the design

Don't just extract requirements - actively contribute to shaping the design while respecting user decisions.
</proactivity_guidance>

**Example question patterns**:

```
// Exploring the problem
"What's the current pain point this would solve?"
"Who runs into this problem most often?"

// Clarifying scope
"Should this handle X, or is that out of scope?"
"What's the minimum that would be useful?"

// Technical direction
"Any preferences on approach: A, B, or C?"
"What existing patterns should this follow?"

// Edge cases
"What happens when X fails?"
"How should this behave for new vs existing users?"
```

### Step 3: Research the Codebase

<codebase_research>
ALWAYS research the codebase when the design involves existing code. This grounds the brainstorm in reality and prevents proposing solutions that conflict with current architecture.

Use Read, Glob, Grep, or Task (Explore agent) to:
- Find related existing implementations to follow or extend
- Identify patterns the new design should align with
- Discover constraints from current architecture
- Find integration points and dependencies
- Check for existing specs that might overlap or conflict

Share relevant findings with the user and incorporate into the brainstorm. If you find code patterns that inform the design, explain them.
</codebase_research>

### Step 4: Synthesize and Confirm

Before writing the spec, summarize the design:

```
Here's what we've brainstormed:

**Problem**: [summary]
**Solution**: [summary]
**Key decisions**: [list]
**Open questions**: [if any]

Does this capture what you're envisioning?
```

<summary_verbosity>
Keep synthesis summaries concise - capture the essence in a few sentences per section. The full detail goes in the spec, not the summary. Users should be able to confirm understanding quickly without re-reading everything discussed.
</summary_verbosity>

Use AskUserQuestion to confirm or refine.

### Step 5: Verify No Spec Conflicts

<spec_conflict_check>
Before writing a new spec, use Task (Explore agent) to check if the proposed design contradicts existing specs in `.agents/specs/`.

If contradictions exist:
1. Surface the conflict to the user with specific quotes from conflicting specs
2. Use AskUserQuestion to ask how to resolve: update existing spec, modify new design, or document the intentional difference
3. Update affected specs based on the decision before proceeding
</spec_conflict_check>

### Step 6: Search for Existing Spec to Update

<existing_spec_search>
Before creating a new spec, always search `.agents/specs/` for an existing spec that covers the same topic or a closely related area. Use Glob and Grep to find candidates by filename and content.

If a matching spec is found:
1. Read the existing spec and assess whether the brainstormed design fits as an update, extension, or revision
2. Present the existing spec to the user with AskUserQuestion:
   ```
   I found an existing spec that covers this area: `.agents/specs/{name}.md`
   - Update the existing spec with the new design
   - Create a new separate spec instead
   ```
3. If updating: note the target spec path and continue to Step 8. **Do not modify the existing spec yet** — write the proposed revision as a draft, run the review first (Step 10-11), and only apply changes to the existing spec after the review is addressed.
4. If creating new: continue to Step 7

This prevents spec sprawl and keeps related design decisions consolidated.
</existing_spec_search>

### Step 7: Determine Spec Name

Ask for the spec filename if not obvious:

```
What should I name this spec? (will be saved to .agents/specs/{name}.md)
```

Use lowercase with hyphens (e.g., `user-notifications`, `api-rate-limiting`).

### Step 8: Write the Spec

<spec_content_guidance>
Create a spec document that captures the brainstormed design. The spec should record WHY decisions were made, not just WHAT was decided. Future readers need context.

Adapt the format based on spec type:
</spec_content_guidance>

**For Feature Specs**:
```markdown
# {Feature Name}

## Problem Statement

What problem does this solve? Who has this problem?

## Proposed Solution

High-level description of the approach.

## User Stories

- As a [user], I want [goal] so that [benefit]

## Design Details

### Behavior

How it works from the user's perspective.

### Technical Approach

Implementation strategy, key components.

### Edge Cases

- What happens when...

## Open Questions

- Decisions still to be made

## Out of Scope

What this explicitly does NOT include.
```

**For Architecture Specs**:
```markdown
# {System/Component Name}

## Overview

What is this and why does it exist?

## Goals

- Primary objectives

## Non-Goals

- What this is NOT trying to do

## Design

### Components

Key parts and their responsibilities.

### Data Flow

How data moves through the system.

### Interfaces

APIs, contracts, boundaries.

## Alternatives Considered

Other approaches and why they weren't chosen.

## Dependencies

What this relies on.

## Risks

Potential issues and mitigations.
```

**For API Specs**:
```markdown
# {API Name}

## Overview

Purpose and scope of this API.

## Endpoints

### `METHOD /path`

**Description**: What this does

**Request**:
```json
{
  "field": "type - description"
}
```

**Response**:
```json
{
  "field": "type - description"
}
```

**Errors**:
- `400`: When...
- `404`: When...

## Authentication

How auth works.

## Rate Limiting

Constraints on usage.

## Examples

Common usage patterns.
```

### Step 9: Save the Spec

If creating a new spec, save to `.agents/specs/{name}.md` using the Write tool.

If updating an existing spec (from Step 6), save the proposed revision to `.agents/specs/{name}.draft.md` as a draft. The existing spec remains untouched until the review passes.

### Step 10: Run Independent Review

<review_rationale>
Independent review catches blind spots that emerge during collaborative design. When you and the user build a design together, you share the same assumptions and context. A fresh perspective finds gaps, contradictions, and ambiguities that neither of you noticed.
</review_rationale>

Launch a review agent using the Task tool with subagent_type="general-purpose":

**Prompt for the review agent**:

> You are a critical reviewer. Read the spec at `.agents/specs/{name}.md` and analyze it for:
>
> 1. **Errors**: Factual mistakes, incorrect assumptions, technical impossibilities
> 2. **Inconsistencies**: Contradictions between sections, conflicting requirements
> 3. **Logical flaws**: Circular reasoning, unsupported conclusions, gaps in logic
> 4. **Weak spots**: Vague requirements, undefined edge cases, missing error handling
> 5. **Ambiguities**: Statements that could be interpreted multiple ways
> 6. **Missing pieces**: Important aspects not addressed
>
> Be thorough and critical. Don't be nice - find real problems.
>
> Format your review as:
>
> ## Spec Review: {name}
>
> ### Critical Issues
> (Must be fixed before implementation)
>
> ### Concerns
> (Should be addressed or explicitly acknowledged)
>
> ### Suggestions
> (Nice to have improvements)
>
> ### Summary
> Overall assessment and recommended next steps.
>
> Return ONLY the review content, no preamble.

### Step 11: Present Review and Iterate

After receiving the review:

1. Share the review findings with the user
2. Use AskUserQuestion to ask:
   ```
   The review found some issues. How would you like to proceed?
   - Address critical issues now
   - Note them as open questions in the spec
   - Dismiss specific findings (explain which)
   ```
3. If the user wants to address issues, use Edit to update the spec (or draft)
4. Optionally re-run the review after significant changes
5. **If updating an existing spec**: once the review is addressed, replace the existing spec with the finalized draft content using Write, then delete the `.draft.md` file

### Step 12: Report Completion

Tell the user:
- Where the spec was saved
- Brief summary of what's captured
- Summary of review findings and resolutions
- Suggest next steps (implementation plan via `/plan`, further review, etc.)

## Long Brainstorm Sessions

<context_management>
For complex brainstorms that span many exchanges:
- Periodically summarize progress in the conversation to checkpoint understanding
- If approaching context limits, write a draft spec early and iterate on it rather than losing explored ideas
- Use the spec file itself as state - write early, edit often
</context_management>

## Important Rules

<core_rules>
1. **Dialogue is the core** - Don't rush to write. Explore thoroughly through conversation.
2. **Listen actively** - Reference previous answers in follow-up questions to show understanding.
3. **Stay flexible** - Adapt questions based on what you learn.
4. **Research is mandatory** - Always ground discussions in existing code when the design touches the codebase.
5. **Check for conflicts** - Verify new specs don't contradict existing ones before writing.
6. **Confirm before writing** - Validate understanding before creating the spec.
7. **Capture decisions** - The spec should record WHY, not just WHAT.
8. **Note open questions** - It's OK to have unresolved items documented.
9. **Review is mandatory** - Always run the review agent after writing.
10. **Iterate on feedback** - Give the user agency over addressing review findings.
11. **DO NOT IMPLEMENT** - This skill creates specs only, not code.
</core_rules>

## Example Dialogue Flow

```
User: /brainstorm notification system

Agent: [AskUserQuestion]
"What kind of notifications are you thinking about?"
- In-app notifications (toast messages, badges)
- Email notifications
- Push notifications (mobile/desktop)
- All of the above

User: In-app notifications

Agent: [AskUserQuestion]
"What events should trigger notifications?"
- [text input encouraged via "Other"]

User: When an ant completes a task, when there's an error, when someone shares something with you

Agent: Got it - task completion, errors, and sharing events.
[AskUserQuestion]
"How should notifications be displayed?"
- Toast that auto-dismisses
- Persistent until clicked
- Notification center (bell icon with list)

User: Toast for ephemeral stuff, but also a notification center for history

Agent: Makes sense - ephemeral toasts plus a persistent history.
[Researches existing UI patterns in codebase]
I found you're using the `ratatui` TUI framework. There's an existing toast system in `src/ui/toast.rs`.
[AskUserQuestion]
"Should we extend the existing toast system or build something new?"
...

[Continues until design is clear]
[Checks for conflicting specs in .agents/specs/]
[Writes spec to .agents/specs/notification-system.md]
[Runs review agent]

Agent: The review found a few issues:

**Critical**: The spec doesn't define how notifications are persisted - what happens on restart?
**Concern**: No mention of notification limits - could the history grow unbounded?
**Suggestion**: Consider adding notification categories for filtering.

[AskUserQuestion]
"How would you like to address these?"
- Fix critical issues now
- Add as open questions
- Discuss further

User: Fix the critical one, add the concern as an open question

[Updates spec]
[Reports completion]
```
