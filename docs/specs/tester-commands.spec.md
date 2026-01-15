# Spec: gt tester Commands

**Task**: hq-t58n (revised hq-skab)
**Author**: Mayor's Aid
**Date**: 2026-01-14
**Status**: Draft (Revised with PM/QA/User feedback)

## Overview

This spec defines the command interface for AI User Testing, including running scenarios, batch execution, reliability tools, and results management.

**Revision notes**: Updated based on QA feedback to add preflight checks, flake detection, retry logic, and basic regression comparison.

---

## 1. Single Test Execution

### gt tester run

Run a single test scenario.

```bash
gt tester run <scenario.yaml> [flags]
```

**Arguments:**
- `<scenario.yaml>`: Path to scenario file

**Flags:**
- `--model <model>`: Override model (haiku, sonnet, gemini)
- `--headed`: Show browser window (default: headless)
- `--env <env>`: Target environment (staging, production)
- `--output <dir>`: Custom output directory
- `--no-video`: Disable video recording
- `--no-trace`: Disable Playwright trace
- `--timeout <sec>`: Override timeout
- `--verbose`: Show agent output in real-time
- `--retry <n>`: Override retry attempts (default from scenario)
- `--no-retry`: Disable retry logic
- `--compare-to <run>`: Compare results to previous run (NEW)
- `--skip-preflight`: Skip environment preflight checks (not recommended)

**Behavior:**
1. Run preflight checks (unless --skip-preflight)
2. Parse and validate scenario YAML
3. Load user story/persona and app context
4. Spawn test agent (Haiku by default)
5. Agent executes test with Playwright MCP
6. On infrastructure failure, retry with backoff
7. Collect artifacts (video, trace, screenshots)
8. Generate observations.json and summary.md
9. Create result bead (if P0/P1 observations)
10. Compare to previous run (if --compare-to)
11. Print summary

**Output:**
```
Running: sarah_registers
  User Story: new_parent_registration
  Persona: Sarah, first-time parent, not tech-savvy
  App: parent-portal (staging)
  Model: haiku

Preflight checks...
  ✓ Playwright installed
  ✓ Staging environment reachable
  ✓ API health OK
  ✓ Disk space sufficient

Starting browser...
Agent navigating...

[00:23] P2/high confusion - Signup button hard to find
[01:45] P3/medium friction - Email re-entry required
[02:30] ✓ Success: Account created
[03:15] ✓ Success: Child profile added

Test Complete
  Duration: 3m 42s
  Observations: 2 issues (0 P0/P1, 1 P2, 1 P3)
  Success criteria: 3/3 met
  Retries: 0

Artifacts:
  Video: test-results/2026-01-14/sarah_registers/run-001/video.webm
  Trace: test-results/2026-01-14/sarah_registers/run-001/trace.zip
  Summary: test-results/2026-01-14/sarah_registers/run-001/summary.md

Result: PASS (no bead created - no P0/P1 issues)
```

**Output with Retry:**
```
Running: sarah_registers
  ...

Starting browser...
  ✗ Browser crash (attempt 1/3)
  Retrying in 1s...
  ✗ Timeout (attempt 2/3)
  Retrying in 2s...
  ✓ Started (attempt 3/3)

Agent navigating...
  ...
```

**Output with Comparison:**
```
gt tester run scenario.yaml --compare-to run-001

...

Comparison to: 2026-01-13/sarah_registers/run-001

  Fixed since baseline:
    ✓ Signup button visibility (was: P2 confusion, now: OK)

  New issues:
    ! P2/medium: Password field error message unclear

  Recurring issues (3 consecutive runs):
    • P3: Email re-entry still required

  Regression score: +1 (improved)
```

**Exit Codes:**
- 0: Test passed (all criteria met)
- 1: Test failed (criteria not met)
- 2: Test error (crash, timeout after all retries)
- 3: Invalid scenario
- 4: Preflight failed

---

## 2. Preflight Checks (NEW)

### gt tester preflight

Run environment preflight checks before testing.

```bash
gt tester preflight [flags]
```

**Flags:**
- `--env <env>`: Target environment (default: staging)
- `--fix`: Attempt to fix issues automatically
- `--json`: Output as JSON

**Behavior:**
1. Check Playwright installation
2. Check MCP server connection
3. Check target environment reachability
4. Check API health endpoints
5. Check test email service (if configured)
6. Check API quota (tokens remaining)
7. Check disk space for artifacts

**Output:**
```
Preflight Checks (staging)

  ✓ Playwright installed (1.40.0)
  ✓ MCP server connected
  ✓ Parent Portal reachable (https://staging.screencoach.example.com)
  ✓ API health check passed (200ms)
  ✓ Test email service: skip_verification (no check needed)
  ✓ API quota: 8,500 tokens remaining
  ✓ Disk space: 12GB free (>5GB required)

All checks passed. Ready to run tests.
```

**Output with Failures:**
```
Preflight Checks (staging)

  ✓ Playwright installed (1.40.0)
  ✓ MCP server connected
  ✗ Parent Portal unreachable
    Error: Connection refused (https://staging.screencoach.example.com)
    Fix: Check VPN connection or staging server status
  ✓ API health check passed
  ⚠ API quota: 200 tokens remaining (low)
    Warning: May run out during long tests
  ✓ Disk space: 12GB free

1 check failed, 1 warning. Fix issues before running tests.
```

**Exit Codes:**
- 0: All checks passed
- 1: Some checks failed
- 2: Critical failure (cannot continue)

---

## 3. Flake Detection (NEW)

### gt tester flaky

View and manage flaky tests.

```bash
gt tester flaky [flags]
```

**Flags:**
- `--scenario <name>`: Filter by scenario
- `--quarantined`: Show only quarantined tests
- `--unquarantine <scenario>`: Remove from quarantine
- `--threshold <rate>`: Override flake threshold (default: 0.10)
- `--json`: Output as JSON

**Behavior:**
1. Analyze recent test runs
2. Calculate flake rate per scenario
3. Identify quarantined tests
4. Show stability metrics

**Output:**
```
Flaky Tests (last 30 days)

Quarantined (>10% flake rate):
  ⚠ social_login
    Flake rate: 23% (7/30 runs)
    Last failure: 2026-01-14 (timeout)
    Quarantined: 2026-01-12
    Run: gt tester flaky --unquarantine social_login

Borderline (5-10% flake rate):
  ~ mobile_registration
    Flake rate: 8% (2/25 runs)
    Last failure: 2026-01-10 (browser_crash)

Stable (<5% flake rate):
  ✓ sarah_registers: 0% (0/45 runs)
  ✓ add_first_child: 2% (1/50 runs)
  ✓ view_activity: 0% (0/30 runs)

Summary:
  Total scenarios: 8
  Quarantined: 1
  Borderline: 1
  Stable: 6
```

### gt tester flaky --unquarantine

Remove a scenario from quarantine.

```bash
gt tester flaky --unquarantine social_login
```

**Output:**
```
Unquarantining: social_login

  Previous flake rate: 23%
  Quarantine reason: High flake rate
  Quarantine date: 2026-01-12

  ⚠ Warning: This test failed 7 times in last 30 days.
     Consider investigating root cause before unquarantining.

Unquarantined. Will run in next batch.
```

---

## 4. Batch Execution

### gt tester batch

Run multiple scenarios.

```bash
gt tester batch <pattern> [flags]
```

**Arguments:**
- `<pattern>`: Glob pattern for scenario files

**Flags:**
- `--parallel <n>`: Run N scenarios simultaneously (default: 1)
- `--stop-on-fail`: Stop batch on first failure
- `--convoy <name>`: Create convoy bead for tracking
- `--model <model>`: Override model for all scenarios
- `--env <env>`: Target environment
- `--filter <tag>`: Only run scenarios with tag
- `--exclude <tag>`: Skip scenarios with tag
- `--include-quarantined`: Include quarantined tests (default: skip)
- `--compare-to <batch>`: Compare to previous batch run (NEW)
- `--skip-preflight`: Skip preflight (runs once per batch)

**Behavior:**
1. Run preflight checks (once for batch)
2. Find all matching scenario files
3. Filter out quarantined tests (unless --include-quarantined)
4. Create convoy bead (if --convoy)
5. Run scenarios (parallel if specified)
6. Retry infrastructure failures per scenario config
7. Aggregate results
8. Update flake metrics
9. Compare to previous batch (if --compare-to)
10. Print batch summary

**Output:**
```
Batch: scenarios/parent-portal/**/*.yaml
  Found: 8 scenarios (1 quarantined, skipped)
  Running: 7 scenarios
  Parallel: 3
  Convoy: parent-portal-tests

Preflight...
  ✓ All checks passed

Running...
  ✓ sarah_registers (3m 42s) - 2 issues (P2, P3)
  ✓ add_first_child (2m 15s) - 1 issue (P2)
  ✓ registration_errors (4m 10s) - 0 issues
  ✓ view_activity (1m 30s) - 0 issues
  ✓ mobile_registration (3m 05s) - 1 issue (P3)
  ✗ modify_settings (failed) - Blocked at settings page
  ↻ power_user_setup (retry 2/3) - 1 issue (P2)

Batch Complete
  Passed: 6/7
  Failed: 1/7
  Skipped: 1 (quarantined)
  Total time: 12m 45s (parallel)
  Total observations: 5 issues (0 P0, 0 P1, 3 P2, 2 P3)
  Retries: 2 (power_user_setup)

Stability:
  Flake rate this batch: 14% (1/7)
  New quarantine candidates: modify_settings (investigate)

Convoy: gt-convoy-xyz
Results: test-results/2026-01-14/batch-001/
```

**Exit Codes:**
- 0: All tests passed
- 1: Some tests failed
- 2: Batch error

---

## 5. Scenario Management

### gt tester list

List available scenarios.

```bash
gt tester list [flags]
```

**Flags:**
- `--app <app>`: Filter by target app
- `--tag <tag>`: Filter by tag
- `--quarantined`: Show quarantine status
- `--json`: Output as JSON

**Output:**
```
Available Scenarios

parent-portal/
  registration/
    ✓ sarah-registers       Sarah (low tech)     critical-path
    ✓ registration-errors   Rose (very low)      error-handling
    ⚠ social-login          Sarah (low tech)     oauth [QUARANTINED]
    ✓ mobile-registration   Emma (mobile)        mobile
  onboarding/
    ✓ add-first-child       Sarah (low tech)     critical-path
    ✓ power-user-setup      Mike (high tech)     power-user
  dashboard/
    ✓ view-activity         Sarah (low tech)     monitoring
    ✓ modify-settings       Mike (high tech)     settings

Total: 8 scenarios (1 quarantined)
```

### gt tester validate

Validate scenario files.

```bash
gt tester validate <pattern>
```

**Output:**
```
Validating: scenarios/**/*.yaml

  ✓ parent-portal/registration/sarah-registers.yaml
  ✓ parent-portal/registration/registration-errors.yaml
  ✗ parent-portal/registration/social-login.yaml
    Error: Cannot use both user_story and persona (line 8)
  ✓ parent-portal/onboarding/add-first-child.yaml

Validation: 3/4 passed
```

---

## 6. Results Management

### gt tester results

View test results.

```bash
gt tester results [date] [flags]
```

**Arguments:**
- `[date]`: Date to view (default: today)

**Flags:**
- `--scenario <name>`: Filter by scenario
- `--failed`: Show only failed tests
- `--observations`: Include observation details
- `--severity <P0-P3>`: Filter by observation severity
- `--pending-review`: Show observations needing human validation
- `--json`: Output as JSON

**Output:**
```
Test Results: 2026-01-14

sarah_registers/
  run-001 (10:30) ✓ Passed - 2 issues
    • P2/high confusion: Signup button hard to find [pending review]
    • P3/medium friction: Email re-entry required
  run-002 (14:15) ✓ Passed - 1 issue
    • P2/medium friction: Password requirements unclear

add_first_child/
  run-001 (10:45) ✓ Passed - 1 issue
    • P2/high confusion: Age picker not intuitive [pending review]

modify_settings/
  run-001 (11:00) ✗ Failed
    • P0/high blocked: Settings page not loading
    • Blocked at: /settings (timeout after 30s)

Summary: 4 scenarios, 4 runs, 3 passed, 1 failed
Observations: 5 total (1 P0, 0 P1, 3 P2, 1 P3)
Pending review: 2 observations
```

### gt tester review (NEW)

Review and validate observations.

```bash
gt tester review [flags]
```

**Flags:**
- `--scenario <name>`: Filter by scenario
- `--date <date>`: Filter by date
- `--interactive`: Interactive review mode

**Output (non-interactive):**
```
Pending Review: 2 observations

1. sarah_registers run-001 [00:23]
   P2/high confusion: Signup button hard to find
   Screenshot: confusion-signup-hidden.png

   Validate: gt tester review --validate 1
   Mark false positive: gt tester review --false-positive 1

2. add_first_child run-001 [01:15]
   P2/high confusion: Age picker not intuitive
   Screenshot: confusion-age-picker.png

   Validate: gt tester review --validate 2
   Mark false positive: gt tester review --false-positive 2
```

**Interactive mode:**
```bash
gt tester review --interactive

Reviewing: sarah_registers run-001 [00:23]
  P2/high confusion: Signup button hard to find

  Opening screenshot: confusion-signup-hidden.png
  Opening video at timestamp 00:23...

  [v] Validate  [f] False positive  [s] Skip  [q] Quit
  > v

  Validated. Moving to next...
```

### gt tester artifacts

Open test artifacts.

```bash
gt tester artifacts <run-path> [flags]
```

**Flags:**
- `--video`: Open video player
- `--trace`: Open Playwright Trace Viewer
- `--summary`: Show summary markdown
- `--screenshot <name>`: Open specific screenshot

**Examples:**
```bash
# Open trace viewer
gt tester artifacts test-results/2026-01-14/sarah_registers/run-001 --trace

# Open video
gt tester artifacts test-results/2026-01-14/sarah_registers/run-001 --video

# Open specific screenshot
gt tester artifacts test-results/2026-01-14/sarah_registers/run-001 --screenshot confusion-signup-hidden.png
```

---

## 7. Stability Metrics (NEW)

### gt tester metrics

View test stability and accuracy metrics.

```bash
gt tester metrics [flags]
```

**Flags:**
- `--period <days>`: Analysis period (default: 30)
- `--scenario <name>`: Filter by scenario
- `--json`: Output as JSON

**Output:**
```
Test Stability Report (last 30 days)

Overall:
  Total runs: 156
  Pass rate: 94%
  Flake rate: 6%
  Mean duration: 3m 12s

Observations:
  Total logged: 89
  Validated: 67 (75%)
  False positives: 12 (13%)
  Pending review: 10 (11%)
  Accuracy rate: 85%

By Severity:
  P0: 2 (100% validated)
  P1: 8 (88% validated)
  P2: 45 (82% validated)
  P3: 34 (79% validated)

Persona Calibration:
  Sarah: 88% accuracy (OK)
  Mike: 91% accuracy (OK)
  Rose: 72% accuracy (needs recalibration)
  Emma: 85% accuracy (OK)

Infrastructure:
  Browser crashes: 5
  Timeouts: 12
  Network errors: 3
  API quota exhaustion: 0

Recommendations:
  • Rose persona has >20% false positive rate - consider recalibration
  • social_login has high timeout rate - investigate OAuth flow
```

---

## 8. Cleanup Commands

### gt tester clean

Clean up test artifacts.

```bash
gt tester clean [flags]
```

**Flags:**
- `--before <date>`: Clean before date
- `--keep <n>`: Keep last N runs per scenario
- `--dry-run`: Show what would be deleted
- `--orphaned-data`: Clean orphaned test data (DB cleanup)

**Output:**
```
Cleaning test results...
  Keeping: last 5 runs per scenario
  Before: 2026-01-07

Would delete:
  test-results/2026-01-05/ (45 MB)
  test-results/2026-01-06/ (38 MB)

Total: 83 MB

Run without --dry-run to delete.
```

### gt tester clean --orphaned-data

Clean orphaned test accounts from database.

```bash
gt tester clean --orphaned-data --dry-run
```

**Output:**
```
Scanning for orphaned test data...

Found orphaned accounts (created by tests, never cleaned up):
  test+sarah_registers+run-abc123@screencoach.test (2026-01-10)
  test+registration_errors+run-def456@screencoach.test (2026-01-11)
  test+mobile_registration+run-ghi789@screencoach.test (2026-01-12)

Total: 3 accounts

Run without --dry-run to delete.
```

---

## 9. Configuration

### gt tester config

View/edit test configuration.

```bash
gt tester config [key] [value]
```

**Examples:**
```bash
# View all config
gt tester config

# View specific key
gt tester config defaults.model

# Set value
gt tester config defaults.timeout 900

# Set stability thresholds
gt tester config stability.flake_threshold 0.15
gt tester config stability.false_positive_threshold 0.25
```

---

## 10. Integration Commands

### gt tester create-bead

Create a UX issue bead from test observation.

```bash
gt tester create-bead <observation-id> [flags]
```

**Flags:**
- `--priority <P0-P4>`: Override priority (default: from observation severity)
- `--assign <rig>`: Assign to rig

**Output:**
```
Creating bead from observation...

Observation: P2/high confusion - Signup button hard to find
From: sarah_registers run-001 at 00:23
Screenshot: confusion-signup-hidden.png

Created: sc-abc
  Title: UX: Signup button hard to find
  Type: bug
  Priority: P2
  Attachments: screenshot, trace link, video timestamp
```

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GT_TESTER_MODEL` | Default model | haiku |
| `GT_TESTER_ENV` | Default environment | staging |
| `GT_TESTER_TIMEOUT` | Default timeout (sec) | 600 |
| `GT_TESTER_PARALLEL` | Default parallelism | 1 |
| `GT_TESTER_OUTPUT` | Results directory | test-results/ |
| `GT_TESTER_FLAKE_THRESHOLD` | Flake rate threshold | 0.10 |
| `GT_TESTER_FP_THRESHOLD` | False positive threshold | 0.20 |
| `GT_TESTER_RETRY_MAX` | Default max retries | 3 |

---

## Configuration File

```yaml
# ~/.gt/tester.yaml
defaults:
  model: haiku
  environment: staging
  timeout: 600
  parallel: 1

output:
  directory: test-results
  keep_runs: 10

video:
  format: webm
  quality: medium

trace:
  enabled: true
  screenshots_in_trace: true

# NEW: Reliability settings
retry:
  max_attempts: 3
  backoff: exponential
  backoff_base: 1000

stability:
  flake_threshold: 0.10
  false_positive_threshold: 0.20
  quarantine_auto: true

preflight:
  enabled: true
  check_api_quota: true
  min_disk_space_gb: 5
```

---

## Implementation Notes

### Agent Spawning

Tests spawn agents via the Task tool:

```python
Task(
    description=f"Test: {scenario_name}",
    prompt=render_tester_prompt(scenario),
    subagent_type="general-purpose",
    model=scenario.model or "haiku"
)
```

### Playwright Integration

Playwright MCP is available in Claude Code:
- Automatic video recording
- Trace collection
- Screenshot on failure
- Network idle detection
- Animation wait support

### Retry Logic

```python
def run_with_retry(scenario):
    attempts = 0
    max_attempts = scenario.retry.max_attempts

    while attempts < max_attempts:
        try:
            result = run_test(scenario)
            return result
        except InfrastructureError as e:
            if e.type in scenario.retry.not_on:
                raise  # Don't retry actual test failures
            attempts += 1
            if attempts < max_attempts:
                delay = calculate_backoff(attempts, scenario.retry)
                time.sleep(delay)

    raise RetryExhausted(f"Failed after {max_attempts} attempts")
```

### Artifact Collection

After agent completes:
1. Collect video from Playwright
2. Collect trace archive
3. Parse observations from agent output (with severity/confidence)
4. Validate observation format
5. Generate summary markdown
6. Store all in results directory
7. Update stability metrics

---

## Testing Checklist

- [ ] `gt tester run` spawns agent correctly
- [ ] Preflight checks run before test
- [ ] Retry logic triggers on infrastructure errors
- [ ] Retry does NOT trigger on test failures
- [ ] Video and trace recorded
- [ ] Observations include severity/confidence
- [ ] Result bead created for P0/P1 only
- [ ] `--compare-to` shows regression analysis
- [ ] `gt tester batch` parallelizes correctly
- [ ] Quarantined tests skipped in batch
- [ ] Convoy created for batch
- [ ] `gt tester flaky` shows correct metrics
- [ ] `gt tester metrics` shows accuracy data
- [ ] `gt tester review` enables validation
- [ ] Cleanup respects keep count
- [ ] Orphaned data cleanup works
