---
description: "Select the next AWS service to migrate and create all migration tickets using plan-native-migration."
context: fork
agent: general-purpose
argument-hint: "[completed-service] e.g. 'after elasticache'"
---

# Plan Next Service Skill

Autonomously select the next AWS service for native migration and create all tickets.

---

## Step 1: Inventory Completed Services

Check pheromone for completed migration plans:

```bash
# Find all completed verify tickets (these mark a service as done)
pheromone list --status Done --include-label stage:reviewer 2>/dev/null || true
pheromone list --status Done --include-label stage:planner 2>/dev/null || true
```

Also check git for evidence of completed migrations:

```bash
# Services with native/ subdirectories = scaffolded or completed
ls -d apis/cluster/*/v1beta*/native/ 2>/dev/null | sed 's|apis/cluster/||;s|/v1beta.*||' | sort -u
```

Build a list of already-migrated services.

## Step 2: Analyze Candidates

Read `.agents/specs/terraform-removal-migration.md` for the tiering strategy.

For each candidate service NOT yet migrated, gather:

```bash
SERVICE="<candidate>"

# Count resources
grep "aws_${SERVICE}_" config/externalname.go | wc -l

# Check for multi-version
ls apis/cluster/$SERVICE/v1beta2/ 2>/dev/null && echo "HAS_V1BETA2" || echo "V1BETA1_ONLY"

# Check for async
grep -A5 "aws_${SERVICE}_" config/cluster/$SERVICE/config.go 2>/dev/null | grep "UseAsync"

# Check for business logic
grep -c "aws_${SERVICE}_" .agents/specs/tf-business-logic-catalog.md 2>/dev/null || echo "0"

# Check for examples
ls examples/$SERVICE/cluster/v1beta*/*.yaml 2>/dev/null | wc -l
```

## Step 3: Select Next Service

Apply these criteria in order:
1. **Tier progression**: Complete lower tiers before higher (Tier 1 → 4)
2. **Skill building**: Choose services that introduce one new complexity at a time
3. **Resource count**: Prefer 3-10 resources (manageable batch)
4. **Example availability**: Skip services with no examples (can't e2e test)

If there's a spec at `.agents/specs/native-tier3-pattern.md` or similar that covers the
selected service, note it. If not, the plan-native-migration skill will handle discovery.

## Step 4: Create the Migration Plan

Use the plan-native-migration skill (invoked via UseSkill) to create all tickets for
the selected service:

```
UseSkill:
  name: plan-native-migration
  arguments: "<selected-service>"
```

This creates scaffold, implement, e2e, regression, verify, and review tickets automatically.

## Step 5: Report

```
## Next Service Selected: <SERVICE>

### Rationale
- Tier: <N>
- Resources: <count>
- New patterns: <what this service introduces>
- Complexity: <low/medium/high>

### Tickets Created
<summary from plan-native-migration output>

### Pipeline Status
Services completed: <list>
Services remaining: <approximate count>
Estimated tier progress: <current tier>/<total tiers>
```
