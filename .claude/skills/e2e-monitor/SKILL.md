---
description: "Monitor a running Kubernetes e2e test with continuous analysis of managed resources, controller logs, and pod state. Detects failures early."
context: fork
agent: general-purpose
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
argument-hint: "[service] e.g. 'sfn', 's3', 'ec2'"
---

# E2E Monitor Skill

Continuously monitor a running Crossplane provider e2e test in a Kind cluster.
Detect failures early through active analysis rather than waiting for chainsaw timeouts.

## Arguments

`$ARGUMENTS` — the AWS service name (e.g. `sfn`, `s3`). Used to identify the
relevant provider pod, managed resources, and controller logs.

If `$ARGUMENTS` is empty, detect the service from running processes:
```bash
ps aux | grep 'make e2e\|uptest' | grep -oP 'SUBPACKAGES="[^"]*"' | head -1
```

---

## Step 1: Establish Baseline

Run all of these in parallel:

```bash
# 1a. Identify the e2e process
ps aux | grep -E 'make e2e|uptest|chainsaw' | grep -v grep

# 1b. Get cluster state
kubectl get providers
kubectl get pods -n upbound-system -o custom-columns='NAME:.metadata.name,STARTED:.status.startTime,READY:.status.containerStatuses[0].ready,RESTARTS:.status.containerStatuses[0].restartCount,IMAGE:.spec.containers[0].image'
kubectl get managed -A

# 1c. Recent commits (to compare against pod age)
git log --format='%H %ai %s' -5

# 1d. What example is being tested
cat /tmp/uptest-e2e/case/test-input.yaml 2>/dev/null || echo "uptest case dir not yet created"
```

Record:
- **E2E process PID** and which phase it's in (build / deploy / uptest)
- **Provider pod start times** vs **latest commit time**
- **Which managed resources exist** and their SYNCED/READY status

---

## Step 2: Stale Binary Detection (Critical — check first)

This is the #1 failure mode. The provider pod may be running code that predates
the controller or fix being tested.

```bash
# Get the service provider pod start time
kubectl get pods -n upbound-system \
  -l pkg.crossplane.io/revision=provider-aws-SERVICE-000000000000 \
  -o jsonpath='{.items[0].status.startTime}'

# Get the latest relevant binary build time
ls -la _output/bin/linux_amd64/SERVICE 2>/dev/null

# Get the latest commit time
git log -1 --format='%ai'

# Check if the pod was restarted during this e2e run
kubectl get pods -n upbound-system \
  -l pkg.crossplane.io/revision=provider-aws-SERVICE-000000000000 \
  -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'
```

**Diagnosis**: If the pod started BEFORE the latest relevant commit AND restart
count is 0, the pod is running stale code. This means:
- Controllers added after the pod started won't exist
- Bug fixes committed after the pod started won't be applied
- The test will timeout waiting for reconciliation that will never happen

**Report** as: `🔴 STALE BINARY — pod started {time}, latest commit at {time}`

---

## Step 3: Monitoring Loop

Poll every 30 seconds. On each iteration, run these checks:

### 3a. Process Health
```bash
# Is the e2e process still running?
pgrep -f 'uptest|chainsaw|make e2e' > /dev/null && echo "E2E RUNNING" || echo "E2E FINISHED"

# Which chainsaw step is active?
ls -lt /tmp/uptest-e2e/case/*.yaml 2>/dev/null
ps aux | grep chainsaw | grep -oP 'test-file \S+' | head -1
```

### 3b. Managed Resource Status
```bash
kubectl get managed -A 2>&1
```

For each managed resource, classify:
| Status | Meaning |
|--------|---------|
| SYNCED=True, READY=True | ✅ Healthy |
| SYNCED=True, READY=False | ⏳ Provisioning or AWS-side delay |
| SYNCED=False | ❌ Controller error — check events |
| No SYNCED/READY columns | ❌ Not being reconciled at all |
| Missing from output | Resource not yet created or was deleted |

### 3c. Resource Detail (for non-Ready resources)
```bash
kubectl describe <kind>.<group>/<name> 2>&1 | tail -30
```

Look for:
- **Events section**: Any warnings or errors
- **Conditions**: Synced=False reason, Ready=False reason
- **Status fields**: Empty status = no reconciliation happening

### 3d. Controller Logs
```bash
kubectl logs -n upbound-system \
  -l pkg.crossplane.io/revision=provider-aws-SERVICE-000000000000 \
  --tail=30 2>&1 | grep -v '^$'
```

Look for:
- `error` / `ERROR` — controller errors
- `cannot` — failed operations
- `requeue` — transient errors being retried
- `not found` — missing dependencies
- `AccessDenied` / `UnauthorizedAccess` — AWS permission issues
- `ValidationException` — bad parameters sent to AWS
- `ThrottlingException` — rate limiting
- Absence of ANY reconciliation logs = controller not watching this resource

### 3e. API Server Health
```bash
kubectl get events -A --sort-by='.lastTimestamp' 2>&1 | tail -10
```

Watch for kube-apiserver unhealthy probes (common in Kind under memory pressure).

---

## Step 4: Early Failure Detection

After each monitoring iteration, evaluate these conditions:

### Immediate Failures (report right away)
| Condition | Detection | Diagnosis |
|-----------|-----------|-----------|
| Stale binary | Pod start < latest commit, restarts=0 | Controller code not in running binary |
| No reconciliation after 60s | Resource has empty .status, no events | Controller not registered or not watching |
| Pod CrashLoopBackOff | Pod status shows CrashLoopBackOff | Binary crash — check logs for panic |
| AWS permission denied | `AccessDenied` in controller logs | IAM role/credentials issue |
| Provider not healthy | `kubectl get providers` shows unhealthy | Provider package install issue |

### Delayed Failures (report after 2 iterations with no progress)
| Condition | Detection | Diagnosis |
|-----------|-----------|-----------|
| Stuck SYNCED=False | SYNCED=False for >60s | Check describe for error reason |
| Stuck READY=False | READY=False for >3min | AWS resource creation slow or failed |
| Reference not resolved | Ref field populated but target field empty | Reference resolver not working |
| No external-name | SYNCED=True but no external-name annotation | Create succeeded but ID not captured |

---

## Step 5: State Tracking

After every 2 monitoring iterations, output a state block:

```
## E2E Monitor State (iteration N, T+Xm)
Phase: [build | deploy | apply | assert | import | delete | finished]
Process: [RUNNING | FINISHED exit=N]

Resources:
- Kind/Name: SYNCED=X READY=X EXTERNAL-NAME=X (age Xs)

Issues:
- 🔴 [CRITICAL issue description]
- 🟡 [WARNING issue description]
- ✅ No issues detected

Controller logs (last notable):
- [timestamp] [summary of important log line]
```

---

## Step 6: Completion

### On Success (all resources Ready, e2e process exits 0)
```
## E2E Result: ✅ PASS

Resources:
- <kind>/<name>: SYNCED=True READY=True EXTERNAL-NAME=<value>

Timeline:
- T+0m: Resources applied
- T+Xm: <kind> reached SYNCED=True
- T+Ym: <kind> reached READY=True
- T+Zm: Delete phase completed
- T+Wm: E2E process exited 0

Evidence:
<last 30 lines of e2e output>
```

### On Failure (e2e process exits non-zero or issue detected)
```
## E2E Result: ❌ FAIL

Root Cause: <one-line summary>
Category: [Stale Binary | No Reconciliation | AWS Error | Controller Bug | Infrastructure | Timeout]

Detail:
<structured analysis of what went wrong and why>

Evidence:
- <controller log lines>
- <resource describe output>
- <process exit info>

Suggested Fix:
<what would need to change to fix this>
```

---

## Monitoring Cadence

| Phase | Poll interval | Focus |
|-------|--------------|-------|
| Build/deploy (no managed resources yet) | 15s | Process health, pod deployment |
| Apply (resources just created) | 10s | Initial reconciliation, stale binary check |
| Assert (waiting for Ready) | 30s | Resource progression, controller logs |
| Delete (cleanup) | 15s | Clean deletion, finalizer issues |

Adapt intervals based on activity — poll faster when resources are actively changing,
slower when waiting for AWS operations.

---

## Common Failure Patterns (reference)

### 1. Stale Pod (most common)
- **Symptom**: Resource has zero status, zero events, controller logs show no reconciliation
- **Cause**: Provider pod started before controller code was committed; `make e2e` reloaded
  image into Kind but pod was never restarted (same dummy sha256:000... hash)
- **Fix**: `kubectl rollout restart deployment -n upbound-system provider-aws-SERVICE-000000000000`

### 2. Missing Reference Resolver
- **Symptom**: Ref fields populated but resolved fields (e.g., roleArn) remain empty
- **Cause**: `zz_generated.resolvers.go` not generated or not in the running binary
- **Fix**: Run `make generate.native` to generate resolver, rebuild and redeploy

### 3. ModernManaged vs LegacyManaged Mismatch
- **Symptom**: Controller logs show type assertion errors or nil pointer on ProviderConfigRef
- **Cause**: Type embeds wrong spec (xpv1.ResourceSpec vs xpv2.ManagedResourceSpec)
- **Fix**: Update type to embed correct spec, regenerate

### 4. External Name Misconfiguration
- **Symptom**: SYNCED=True but Create sends wrong parameters, or Observe can't find resource
- **Cause**: External name strategy doesn't match what config/externalname.go expects
- **Fix**: Align external name handling in native controller with TF resource's pattern

### 5. AWS API Errors
- **Symptom**: SYNCED=False with error in events/conditions
- **Cause**: Invalid parameters, missing required fields, wrong region
- **Fix**: Check AWS API docs, compare with TF provider implementation
