# Proposal: AI User Testing for ScreenCoach

**Task**: hq-t58n (revised hq-skab)
**Author**: Mayor's Aid
**Date**: 2026-01-14
**Status**: Draft (Revised with PM/QA/User feedback)

## Executive Summary

This proposal introduces **AI User Testing** - a system where Claude acts as real users to test ScreenCoach applications, identify UX friction points, and suggest improvements. Testing agents embody user personas and navigate the apps as a real user would, while producing recorded artifacts for human review.

**Core insight**: AI can simulate diverse user types at scale, catching UX issues that scripted tests miss. By recording everything (video, traces, screenshots), humans can review the exact moment of confusion.

**Revision notes**: Updated based on PM, QA, and user feedback to narrow Phase 1 scope, add reliability features, and incorporate unified user story format.

---

## Target Applications

| Application | Type | Primary Tool | Phase |
|-------------|------|--------------|-------|
| Parent Portal | Web App | Playwright MCP | **Phase 1** |
| Desktop Client | Electron | MCPControl | Phase 2+ |
| Browser Extension | Chrome Extension | Playwright on Windows | Phase 2+ |

**Note**: Phase 1 focuses exclusively on Parent Portal web testing. Desktop/Extension testing deferred per PM review feedback.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Linux Host (Gas Town)                        │
│                                                                     │
│  ┌──────────────┐    ┌─────────────────┐    ┌───────────────────┐  │
│  │ gt tester    │───▶│ Preflight       │───▶│ Test Agent        │  │
│  │ (command)    │    │ Checks          │    │ (Haiku/Sonnet)    │  │
│  └──────────────┘    └─────────────────┘    └────────┬──────────┘  │
│                                                       │             │
│                                                       ▼             │
│  ┌──────────────┐    ┌─────────────────┐    ┌───────────────────┐  │
│  │ Retry Logic  │◀───│ Infrastructure  │◀───│ Playwright MCP    │  │
│  │ (exponential │    │ Error Handler   │    │ (web testing)     │  │
│  │  backoff)    │    └─────────────────┘    └───────────────────┘  │
│  └──────────────┘                                                   │
│                                                                     │
│  ┌──────────────┐    ┌─────────────────┐    ┌───────────────────┐  │
│  │ Artifacts    │───▶│ Observations    │───▶│ Stability         │  │
│  │ (video,trace)│    │ (severity,conf) │    │ Metrics           │  │
│  └──────────────┘    └─────────────────┘    └───────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Key Decisions

### 1. Agent Model: Fresh Haiku Per Scenario

**Pattern**: Polecat model - spawn, execute, complete

| Factor | Benefit |
|--------|---------|
| Token efficiency | Clean context each run |
| State isolation | No bleed-through between tests |
| Parallelization | Run multiple scenarios simultaneously |
| Persona fidelity | Agent fully embodies one user |
| Failure isolation | One crash doesn't affect others |

**Model Selection:**
- **Default: Haiku** - fast, cheap, sufficient for navigation
- **Override: Sonnet** - for complex UX reasoning (error recovery scenarios)
- **Optional: Gemini** - for very long sessions

### 2. Recording Requirements

Every test run MUST produce:

| Artifact | Format | Purpose |
|----------|--------|---------|
| Video | `.webm` | Full session replay |
| Trace | `.zip` | Interactive Playwright trace |
| Screenshots | `.png` | Key moments, confusion points (require human review) |
| Observations | `.json` | Structured findings with severity/confidence |
| Summary | `.md` | Human-readable report |

### 3. Reliability Requirements (NEW)

| Feature | Implementation |
|---------|----------------|
| Preflight checks | Environment validation before tests |
| Retry logic | Exponential backoff for infrastructure failures |
| Flake detection | Auto-quarantine tests with >10% flake rate |
| Test data isolation | Unique emails, cleanup strategies |
| Observation confidence | Agent self-assessment for human validation |

### 4. Tooling by Target

| Target | Tool | Reason | Phase |
|--------|------|--------|-------|
| Web apps | Playwright MCP | Headless, scriptable, CI-friendly | Phase 1 |
| Electron | MCPControl | Windows UI automation required | Phase 2+ |
| Extension | Playwright on Windows | Needs local browser access | Phase 2+ |

---

## Test Scenario Format

Scenarios now support two formats: unified **user story** (recommended) or detailed **persona**.

### User Story Format (Recommended)

```yaml
# scenarios/user-stories/sarah-registers.yaml
scenario: sarah_registers
version: 1
description: "First-time parent registration"
tags: [registration, critical-path, p0]

user_story:
  name: new_parent_registration
  persona: Sarah, first-time parent, not tech-savvy, 2 kids ages 8 and 11
  goal: Register for ScreenCoach and set up profiles for both children
  context: |
    Found ScreenCoach through school newsletter.
    Wants to limit gaming during homework hours.

target:
  app: parent-portal
  environment: staging

steps:
  - Navigate to homepage
  - Find and click registration
  - Complete signup form
  - Add children
  - View dashboard

success_criteria:
  - Account created successfully
  - Child profiles added
  - Dashboard visible

# Reliability settings
wait_strategies:
  network_idle: true
  animation_complete: true
  min_load_time: 1500

retry:
  max_attempts: 3
  on_errors: [browser_crash, timeout, network_error]
  backoff: exponential

# Observation settings
observations:
  require_severity: true
  require_confidence: true
  auto_triage:
    P0_P1: create_bead
    P2_P3: log_only

# Test data isolation
test_data:
  email_pattern: "sarah+{run_id}@screencoach.test"
  cleanup_strategy:
    on_success: delete_account
    on_failure: mark_for_review

recording:
  video: true
  trace: true
  screenshots:
    on_failure: true
    on_confusion: true
    require_review: true

timeout: 600
model: haiku
```

---

## Agent Context (Minimal)

Each spawned agent receives only:

1. **Testing CLAUDE.md** (~50 lines) - How to be a user tester
2. **User story/Persona** - Who they are, their goal
3. **App context** - What ScreenCoach is, key concepts
4. **Success criteria** - What defines completion
5. **Playwright MCP tools** - Browser control

**NO accumulated state, NO prior test results.**

---

## Artifact Storage

```
screencoach/test-results/
└── 2026-01-14/
    └── sarah_registers/
        ├── run-001/
        │   ├── video.webm
        │   ├── trace.zip
        │   ├── screenshots/
        │   │   ├── 01-landing-page.png
        │   │   ├── 02-confusion-signup-button.png
        │   │   └── 03-success-dashboard.png
        │   ├── observations.json
        │   └── summary.md
        └── run-002/
            └── ...
```

---

## Observation Format (Updated)

Observations now include severity and confidence:

```json
{
  "scenario": "sarah_registers",
  "persona": "Sarah",
  "completed": true,
  "duration_seconds": 222,
  "observations": [
    {
      "type": "confusion",
      "severity": "P2",
      "confidence": "high",
      "timestamp": "00:23",
      "location": "homepage",
      "description": "Signup button not visible without scrolling",
      "screenshot": "confusion-signup-hidden.png",
      "validated": null,
      "false_positive": null
    }
  ],
  "success_criteria_met": ["Account created", "Child added"],
  "success_criteria_failed": [],
  "overall_experience": "Mostly smooth but signup button hard to find",
  "retry_count": 0,
  "infrastructure_errors": []
}
```

### Severity Levels

| Level | Description | Auto Action |
|-------|-------------|-------------|
| `P0` | Blocking - user cannot complete goal | Create bead |
| `P1` | Significant friction - user likely to abandon | Create bead |
| `P2` | Minor friction - noticeable but not blocking | Log only |
| `P3` | Nitpick - improvement opportunity | Log only |

---

## Integration with Gas Town

### Beads Integration

| Event | Bead Created |
|-------|--------------|
| Test batch started | Convoy bead tracking all scenarios |
| Scenario complete | Result bead with observations |
| P0/P1 issue found | Bug bead automatically (auto_triage) |
| Test failure | Bug bead with trace link |

### Workflow (Updated)

```
1. Human/Mayor creates test batch
   gt tester batch scenarios/parent-portal/*.yaml

2. Preflight checks run (once per batch)
   ✓ Environment reachable
   ✓ API quota sufficient
   ✓ Disk space OK

3. For each scenario (quarantined skipped):
   - Spawn Haiku agent
   - Agent runs test as persona
   - On infrastructure error, retry with backoff
   - Agent produces observations with severity/confidence
   - Result bead created (P0/P1 only)

4. Batch completes
   - Stability metrics updated
   - Flaky tests flagged/quarantined
   - Summary generated

5. Human reviews (gt tester review):
   - Validate observations
   - Mark false positives
   - Watch videos of P0/P1 issues

6. Calibration (periodically):
   - Compare AI observations to real user feedback
   - Adjust persona if FP rate > 20%

7. UX fixes become work beads
   - Slung to polecats
   - Implement improvements
   - Re-run scenarios to verify (--compare-to)
```

---

## Commands Overview

| Command | Purpose |
|---------|---------|
| `gt tester run <scenario>` | Run single test with preflight and retry |
| `gt tester batch <pattern>` | Run multiple scenarios (skips quarantined) |
| `gt tester preflight` | Check environment readiness |
| `gt tester flaky` | View/manage flaky tests |
| `gt tester metrics` | View stability and accuracy metrics |
| `gt tester review` | Validate observations (interactive/batch) |
| `gt tester results` | View test results |
| `gt tester clean` | Cleanup artifacts and orphaned data |

See `tester-commands.spec.md` for full details.

---

## Implementation Phases (Revised)

### Phase 1: Web App Testing (MVP) - Narrowed Scope

**Target**: Parent Portal only, single persona (Sarah) to start

**Must Have (MVP):**
- [ ] Testing agent CLAUDE.md
- [ ] Scenario YAML parser (user story format)
- [ ] `gt tester run` with preflight and retry
- [ ] Playwright MCP integration
- [ ] Wait strategies (network_idle, animation_complete)
- [ ] Test data isolation (unique emails)
- [ ] Artifact recording (video, trace, screenshots)
- [ ] Observations with severity/confidence
- [ ] Basic beads integration (P0/P1 auto-create)

**Should Have (Phase 1 Complete):**
- [ ] `gt tester batch` for multiple scenarios
- [ ] 3 personas (Sarah, Mike, Emma)
- [ ] Flake detection and quarantine
- [ ] Basic regression comparison (`--compare-to`)
- [ ] `gt tester review` for validation
- [ ] Error recovery scenario
- [ ] Convoy integration

**Calibration Step (Before Phase 1 sign-off):**
- [ ] Run same scenarios with 3 real users (think-aloud)
- [ ] Compare AI observations to human observations
- [ ] Calculate correlation score
- [ ] Tune persona if correlation < 70%

### Phase 2: Persona Library & Expansion
- [ ] Additional personas (Rose, teacher, etc.)
- [ ] Desktop client testing (MCPControl)
- [ ] Browser extension testing
- [ ] Observation validation tracking
- [ ] Headed verification runs

### Phase 3: Continuous Testing
- [ ] Scheduled test runs
- [ ] Full regression detection
- [ ] Automatic issue creation
- [ ] Dashboard integration
- [ ] Persona calibration automation

---

## Success Metrics (Specific)

| Metric | Baseline | Phase 1 Target | Measurement |
|--------|----------|----------------|-------------|
| UX issues found | Manual (few) | 5+ unique issues per bi-weekly run | Count distinct validated observations |
| Time to regression detection | Days | <4 hours | Time from deploy to first observation |
| Critical flow coverage | 0% | 3 flows (registration, onboarding, dashboard) | Scenario count |
| Human review time | N/A | <5 min per scenario | Measured via `gt tester review` |
| False positive rate | N/A | <20% | Validated vs total observations |
| Test reliability | N/A | >85% | Passes / total runs |

---

## Reliability Requirements (NEW)

### Test Data Isolation

```yaml
test_data:
  email_pattern: "test+{scenario}+{run_id}@screencoach.test"
  cleanup_strategy:
    on_success: delete_account
    on_failure: mark_for_review
    on_crash: cleanup_job
  isolation:
    unique_suffix: true
```

### Retry Logic

```yaml
retry:
  max_attempts: 3
  on_errors: [browser_crash, timeout, network_error]
  not_on: [test_failure, blocked]
  backoff: exponential
  backoff_base: 1000
```

### Stability Thresholds

| Threshold | Default | Action |
|-----------|---------|--------|
| Flake rate | >10% | Quarantine test |
| False positive rate | >20% | Flag persona for recalibration |
| Observation confidence | <medium | Require human review |

---

## Risks and Mitigations (Updated)

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Token costs | Low | Medium | Haiku default, chunk flows |
| Flaky tests | High | High | Preflight, retry, quarantine |
| False positives | Medium | High | Confidence levels, human validation, calibration |
| Persona drift | Medium | Medium | Periodic calibration with real users |
| Test data pollution | High | Medium | Unique emails, cleanup jobs |
| MCPControl complexity | High | Low | Defer to Phase 2 |

---

## Related Specs

| Spec | Purpose |
|------|---------|
| `tester-claude-md.spec.md` | Agent prompt for testing role |
| `tester-scenario-format.spec.md` | Full YAML schema (updated) |
| `tester-commands.spec.md` | gt tester command details (updated) |
| `tester-example-scenarios.spec.md` | Example scenarios |

---

## Conclusion

AI User Testing transforms QA from scripted checks to realistic user simulation. By:

1. **Embodying diverse users** - Personas capture real user behavior
2. **Recording everything** - Video + trace for human review
3. **Scaling horizontally** - Parallel scenario execution
4. **Ensuring reliability** - Preflight, retry, flake detection
5. **Validating accuracy** - Human review, calibration with real users
6. **Integrating with Gas Town** - Results become actionable beads

**Recommendation**: Approve Phase 1 (Web App Testing) for ScreenCoach Parent Portal with narrowed scope:
- Start with Sarah persona only
- Include calibration step before expanding
- Add reliability features from Day 1

---

*Proposal authored by Mayor's Aid, 2026-01-14 (hq-t58n)*
*Revised with PM/QA/User feedback, 2026-01-14 (hq-skab)*
