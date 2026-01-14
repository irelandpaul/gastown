# QA Review: AI User Testing Proposal

**Reviewer**: QA Engineering Perspective
**Date**: 2026-01-14
**Task**: hq-xixk
**Status**: Review

## Executive Summary

The AI User Testing proposal introduces an innovative approach to UX testing using AI personas. From a QA engineering perspective, the proposal is **well-architected** with strong artifact collection and isolation patterns. However, there are gaps in **test reliability**, **environment management**, and **regression detection** that need addressing before implementation.

**Verdict**: **APPROVE with QA concerns addressed** - Solid foundation, needs reliability hardening and clearer test hygiene practices.

---

## Strengths

### 1. Excellent Test Isolation Model

The "fresh Haiku per scenario" pattern is sound QA practice:

| Pattern | QA Benefit |
|---------|------------|
| Clean context per test | No state bleed-through |
| Parallel execution | Faster feedback loops |
| Failure isolation | One flaky test doesn't cascade |
| Reproducible runs | Same inputs, comparable outputs |

This mirrors best practices from modern test frameworks (Jest workers, pytest-xdist).

### 2. Comprehensive Artifact Collection

The mandatory recording requirements create excellent debugging capability:

```
video.webm    → Visual reproduction without re-running
trace.zip     → Step-through debugging in Playwright UI
screenshots/  → Quick triage without watching full video
observations.json → Structured data for automation
summary.md    → Human-readable report
```

**QA endorsement**: This artifact set is sufficient for root cause analysis without requiring reproduction.

### 3. Clear Success Criteria Pattern

The success_criteria field in scenarios provides:
- Binary pass/fail determination
- Specific, testable conditions
- Automated verification potential

This prevents the "observation overload" problem where everything is noted but nothing is actionable.

### 4. Scenario Versioning

The `version` field in scenario YAML enables:
- Breaking change detection
- Backward compatibility checks
- Migration tooling

---

## QA Concerns

### 1. Test Flakiness Risk: Timing Dependencies

**Critical concern**: Web UI testing is notoriously flaky. The proposal mentions Playwright's auto-wait but doesn't address:

| Timing Issue | Risk Level | Current Coverage |
|--------------|------------|------------------|
| Network latency variations | High | Not addressed |
| Animation/transition waits | Medium | Assumed by Playwright |
| Dynamic content loading | High | Not addressed |
| Third-party service delays | High | Not addressed |

**Example failure modes**:
- Signup API is slow → Agent times out → False failure
- Animation on button → Agent clicks too early → Element not interactable
- CDN image loading → Page appears ready but isn't

**Recommendation**: Add explicit wait strategies to scenario format:

```yaml
wait_strategies:
  network_idle: true           # Wait for no pending requests
  animation_complete: true     # Wait for CSS transitions
  custom_selectors:
    - "#app-loaded"            # App-specific ready indicators
  min_load_time: 2000          # Minimum wait (ms) after navigation
```

### 2. Test Data Management Gap

**Critical gap**: The proposal mentions `test_data.seed_account` but lacks comprehensive test data strategy:

| Concern | Impact |
|---------|--------|
| Account creation pollution | Staging DB fills with test accounts |
| Email verification | Where do verification emails go? |
| Unique constraint collisions | Same email used across parallel runs |
| Data cleanup failures | Orphaned test data accumulates |

**Questions needing answers**:
1. How are unique test emails generated? (`sarah+{uuid}@test.screencoach.com`?)
2. Is there a test email inbox service (Mailinator, test SMTP)?
3. Who cleans up failed test data when `cleanup: true` but test crashes?
4. How do parallel tests avoid account collision?

**Recommendation**: Add test data spec:

```yaml
# In scenario format
test_data:
  email_pattern: "test+{scenario}+{run_id}@screencoach.test"
  email_inbox: mailhog  # or: skip_verification
  cleanup_strategy:
    on_success: delete_account
    on_failure: mark_for_review
    on_crash: cleanup_job  # Background cleanup task
  isolation:
    unique_suffix: true  # Append UUID to all created data
```

### 3. Environment Parity Concerns

The proposal targets "staging" but doesn't address:

| Gap | Risk |
|-----|------|
| Staging/prod feature flags | Tests pass on staging, fail on prod |
| Database state differences | Staging has different user distributions |
| Third-party sandbox vs prod | OAuth may behave differently |
| Performance differences | Timeouts tuned for staging may fail on slower prod |

**Recommendation**: Add environment validation:

```bash
# Before test run
gt tester env-check staging
  ✓ Parent Portal reachable
  ✓ API health check passed
  ✓ Database connection OK
  ✗ OAuth sandbox mode differs from prod config
    Warning: Social login tests may not match prod behavior
```

### 4. Observation Reliability

**Core QA concern**: How do we know AI observations are accurate?

| Observation Type | Reliability Risk |
|------------------|------------------|
| "Confusion point" | Subjective - AI may misinterpret |
| "Friction" | Depends on persona calibration |
| "Error" | More reliable - concrete failures |
| "Success" | Most reliable - criteria-based |

**The calibration problem**:
- PM review mentions calibration with real users (R3)
- But no ongoing calibration mechanism
- Persona accuracy may drift over time

**Recommendation**: Add observation confidence and validation:

```json
{
  "observations": [
    {
      "type": "confusion",
      "confidence": "high",  // Agent self-assessment
      "validated": null,     // Human validation status
      "false_positive": null // Set by human review
    }
  ]
}
```

Track false positive rate over time. If FP rate > 20%, flag persona for recalibration.

### 5. Missing Regression Detection

The proposal mentions "comparison with previous runs" in Phase 4, but this is essential for QA value:

**Without regression detection**:
- Every observation is "new"
- No way to verify fixes worked
- No baseline for improvement metrics

**Minimum viable regression**:

```bash
gt tester run scenario.yaml --compare-to run-001

Output:
  Comparing to: 2026-01-13/register_new_parent/run-001

  Fixed since baseline:
    ✓ Signup button visibility (was: confusion, now: OK)

  New issues:
    ! Password field: new error message unclear

  Recurring issues:
    • Email re-entry still required (3 consecutive runs)

  Regression score: +1 (improved)
```

**Recommendation**: Elevate basic regression to Phase 1, not Phase 4.

### 6. Error Handling Gaps

The spec shows exit codes but not recovery scenarios:

| Scenario | Current Handling | Recommended |
|----------|------------------|-------------|
| Browser crash mid-test | Unknown | Save partial artifacts, mark incomplete |
| Agent timeout | Exit code 2 | Capture last state, screenshot, force-kill |
| Playwright connection lost | Unknown | Retry with fresh browser |
| VM unreachable (MCPControl) | Unknown | Fail fast with clear message |
| Out of API quota | Unknown | Queue for retry, don't mark as failure |

**Recommendation**: Add error taxonomy:

```yaml
# In observations.json
{
  "status": "error",
  "error_type": "infrastructure",  // vs "test_failure" vs "blocked"
  "error_category": "browser_crash",
  "recoverable": true,
  "retry_count": 0,
  "partial_artifacts": ["video_partial.webm", "last_screenshot.png"]
}
```

### 7. Headed vs Headless Parity

The proposal defaults to `headed: false` (headless) but:

- Some bugs only appear in headed mode (focus issues, popups)
- Some bugs only appear in headless mode (GPU rendering)
- Viewport sizes may differ

**Recommendation**: Periodic headed verification:

```yaml
recording:
  headed: false
  headed_verification: weekly  # Run headed once per week
```

---

## Test Coverage Analysis

### Coverage Strengths

| Flow | Coverage | Notes |
|------|----------|-------|
| Registration | Strong | Multiple personas, clear criteria |
| Onboarding | Good | Child add flow covered |
| Dashboard | Partial | View covered, complex actions not |

### Coverage Gaps

| Flow | Missing | Risk |
|------|---------|------|
| Password reset | Not in scenarios | Common UX friction point |
| Account deletion | Not mentioned | GDPR/compliance flow |
| Multi-child management | Single child only | Scale UX |
| Mobile responsive | Mentioned, not specified | Device field exists but no scenarios |
| Error recovery | Not tested | What if user makes mistake? |

**Recommendation**: Phase 1 should include error recovery scenario:

```yaml
scenario: registration_with_errors
persona: sarah
goal: |
  Register for ScreenCoach, but make typical mistakes:
  - Enter invalid email first
  - Use weak password
  - Skip required field
  Observe error message clarity and recovery UX.
```

---

## Reliability Recommendations

### R1: Add Test Stability Metrics

Track and report:

```yaml
stability_metrics:
  flake_rate: 5%        # Tests that fail then pass on retry
  false_positive_rate: 12%  # Observations marked invalid by humans
  mean_duration: 3m 42s
  duration_variance: 45s
```

Fail tests that exceed flake threshold.

### R2: Implement Test Quarantine

Automatically quarantine flaky tests:

```bash
gt tester run scenario.yaml
  ✗ Failed (attempt 1/3)
  ✗ Failed (attempt 2/3)
  ✓ Passed (attempt 3/3)

  ⚠ Test marked as FLAKY (2/3 failures)
  Quarantined until reviewed. See: gt tester flaky
```

### R3: Add Pre-flight Checks

Before running tests:

```bash
gt tester preflight
  ✓ Playwright installed and working
  ✓ MCP server connected
  ✓ Staging environment reachable
  ✓ Test email service responding
  ✓ Sufficient API quota (1000 tokens remaining)
  ✓ Disk space for artifacts (5GB free)
```

### R4: Implement Retry Logic

In scenario format:

```yaml
retry:
  max_attempts: 3
  on_errors: [browser_crash, timeout, network_error]
  not_on: [test_failure, blocked]  # Don't retry actual failures
  backoff: exponential  # 1s, 2s, 4s
```

### R5: Add Health Monitoring

For ongoing test reliability:

```yaml
# Weekly report
Test Health Report (2026-01-14)

  Scenarios: 8 total
    Healthy: 6 (75%)
    Flaky: 1 (12.5%)
    Failing: 1 (12.5%)

  Observation Accuracy:
    Confirmed useful: 45
    False positives: 8
    Accuracy rate: 85%

  Infrastructure:
    Browser crashes: 2
    Timeouts: 5
    API errors: 0
```

---

## Phase 1 Scope (QA Perspective)

### Must Have (MVP)

- [ ] `gt tester run` with retry logic
- [ ] Preflight environment checks
- [ ] Test data isolation (unique emails)
- [ ] Partial artifact collection on failure
- [ ] Basic pass/fail with exit codes
- [ ] Observation output with confidence levels

### Should Have (Phase 1 Complete)

- [ ] Flake detection and quarantine
- [ ] Basic regression comparison (vs last run)
- [ ] Test stability metrics
- [ ] Error recovery scenario
- [ ] Cleanup job for orphaned test data

### Defer to Phase 2

- [ ] Full regression analysis
- [ ] Observation validation tracking
- [ ] Headed verification runs
- [ ] Complex multi-persona scenarios
- [ ] Performance baseline comparison

---

## Risk Assessment (QA View)

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Flaky tests erode trust | High | High | Quarantine, retry, stability metrics |
| False positive overload | Medium | High | Confidence levels, human validation |
| Test data pollution | High | Medium | Isolation, cleanup jobs |
| Environment differences | Medium | Medium | Preflight checks, parity validation |
| Artifact storage bloat | Low | Medium | Cleanup policy, retention limits |
| Agent timeout too aggressive | Medium | Low | Adaptive timeouts based on flow |

---

## Verdict

**APPROVE with QA concerns addressed**

The proposal has a solid architectural foundation. The artifact collection and isolation patterns are excellent. However, for this to be a reliable QA tool rather than a novelty, the following must be addressed:

**Blocking concerns (must fix before Phase 1)**:
1. Test data isolation strategy (unique emails, cleanup)
2. Basic retry logic for infrastructure failures
3. Preflight environment checks

**High priority concerns (should fix in Phase 1)**:
4. Observation confidence levels
5. Flake detection mechanism
6. Basic regression comparison

**Estimated reliability**: With above fixes, expect 85%+ test reliability (vs ~60% without).

**QA endorsement**: Once reliability concerns are addressed, this becomes a valuable addition to the testing portfolio. AI persona testing fills a gap between scripted E2E tests (no UX insight) and manual testing (expensive, slow).

---

*Review by QA Engineering perspective, 2026-01-14 (hq-xixk)*
