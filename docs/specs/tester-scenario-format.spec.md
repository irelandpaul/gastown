# Spec: Test Scenario YAML Format

**Task**: hq-t58n (revised hq-skab)
**Author**: Mayor's Aid
**Date**: 2026-01-14
**Status**: Draft (Revised with PM/QA/User feedback)

## Overview

This spec defines the YAML format for AI User Testing scenarios. Each scenario describes a user story (persona + goal), evaluation criteria, and reliability settings.

**Revision notes**: Updated based on PM, QA, and user feedback to add unified user story format, wait strategies, retry logic, test data isolation, and observation confidence levels.

---

## Full Schema

```yaml
# Required fields marked with *

# === Metadata ===
scenario: string*           # Unique identifier (snake_case)
version: integer            # Schema version (default: 1)
description: string         # Human-readable description
tags: [string]              # For filtering/organization

# === User Story (Unified Format) ===
# Combines persona and goal into cohesive narrative
user_story:
  name: string*             # Story identifier (e.g., "new_parent_registration")
  persona: string*          # One-line persona description
  goal: string*             # What the user is trying to accomplish
  context: string           # Additional background/situation

# === Alternative: Detailed Persona (for complex scenarios) ===
persona:
  name: string*             # User's name (for think-aloud)
  role: string*             # parent | child | admin | teacher
  context: string*          # Background, goals, situation
  tech_comfort: string      # low | medium | high (default: medium)
  patience: string          # low | medium | high (default: medium)
  device: string            # desktop | mobile | tablet (default: desktop)

goal: string                # What user is trying to accomplish (if not in user_story)

# === Target Application ===
target:
  app: string*              # parent-portal | desktop-client | extension
  environment: string       # staging | production (default: staging)
  url: string               # Override base URL (optional)

# === Guided Steps (Optional) ===
steps:                      # Guidance, agent can deviate
  - string                  # Step description

# === Success Criteria ===
success_criteria:           # What defines successful completion
  - string*                 # At least one required

# === Evaluation Focus ===
evaluate:                   # What to pay attention to
  - string

# === Observation Settings (NEW) ===
observations:
  require_severity: boolean # Require severity on each observation (default: true)
  require_confidence: boolean # Require confidence level (default: true)
  auto_triage:
    P0_P1: string           # Action for blocking issues (default: create_bead)
    P2_P3: string           # Action for minor issues (default: log_only)

# === Recording Settings ===
recording:
  video: boolean            # Record video (default: true)
  trace: boolean            # Playwright trace (default: true)
  screenshots:
    on_failure: boolean     # Auto screenshot on failure (default: true)
    on_confusion: boolean   # Agent screenshots confusion (default: true)
    on_demand: boolean      # Agent can screenshot anytime (default: true)
    require_review: boolean # Screenshots need human validation (default: true)
  headed: boolean           # Show browser (default: false)
  headed_verification: string # Run headed periodically (default: null, options: daily|weekly|monthly)

# === Wait Strategies (NEW - QA) ===
wait_strategies:
  network_idle: boolean     # Wait for no pending requests (default: true)
  animation_complete: boolean # Wait for CSS transitions (default: true)
  min_load_time: integer    # Minimum wait after navigation in ms (default: 1000)
  custom_selectors:         # App-specific ready indicators
    - string                # e.g., "#app-loaded", "[data-ready='true']"

# === Retry Configuration (NEW - QA) ===
retry:
  max_attempts: integer     # Max retry attempts (default: 3)
  on_errors:                # Error types to retry
    - string                # browser_crash | timeout | network_error
  not_on:                   # Error types to NOT retry
    - string                # test_failure | blocked
  backoff: string           # none | linear | exponential (default: exponential)
  backoff_base: integer     # Base delay in ms (default: 1000)

# === Execution Settings ===
timeout: integer            # Max seconds (default: 600)
model: string               # haiku | sonnet | gemini (default: haiku)

# === Test Data (EXPANDED - QA) ===
test_data:
  # Account creation
  email_pattern: string     # Pattern for unique emails (default: "test+{scenario}+{run_id}@screencoach.test")
  email_inbox: string       # Email service: mailhog | skip_verification | real (default: skip_verification)

  # Seeding
  seed_account: string      # Pre-created test account
  seed_data: object         # Pre-populated data

  # Cleanup
  cleanup_strategy:
    on_success: string      # delete_account | keep | mark_for_review (default: delete_account)
    on_failure: string      # keep | mark_for_review (default: mark_for_review)
    on_crash: string        # cleanup_job | keep (default: cleanup_job)

  # Isolation
  isolation:
    unique_suffix: boolean  # Append UUID to all created data (default: true)

# === Dependencies ===
depends_on:                 # Run after these scenarios
  - string                  # scenario names

# === Desktop/Extension Specific ===
windows:                    # For desktop client / extension tests (Phase 2+)
  use_mcpcontrol: boolean   # Use MCPControl for Windows (default: false)
  app_path: string          # Path to Electron app
  extension_id: string      # Chrome extension ID
```

---

## User Story Format (Recommended)

The unified user story format combines persona and goal into a more cohesive narrative. This is the **preferred format** for new scenarios.

```yaml
# Unified user story - simpler, more cohesive
user_story:
  name: new_parent_registration
  persona: Sarah, first-time parent, not tech-savvy, 2 kids ages 8 and 11
  goal: Register for ScreenCoach and set up profiles for both children
  context: Found ScreenCoach through school newsletter, wants to limit homework-time gaming
```

vs. the detailed persona format (still supported for complex scenarios):

```yaml
# Detailed persona - for complex UX analysis
persona:
  name: Sarah
  role: parent
  context: |
    First-time user, not tech-savvy, has 2 kids (ages 8 and 12).
    Found ScreenCoach through school recommendation.
    Primary goal: limit screen time during homework hours.
  tech_comfort: low
  patience: medium
  device: desktop

goal: |
  Register for ScreenCoach and set up first child profile.
```

**Validation rule**: Scenario must have EITHER `user_story` OR (`persona` + `goal`), not both.

---

## Field Details

### user_story.persona (One-liner Format)

Compact persona description. Include:
- Name
- Key characteristic (e.g., "tech-savvy", "first-time user")
- Relevant context (e.g., "2 kids ages 8 and 11")

Examples:
- `"Sarah, first-time parent, not tech-savvy, 2 kids ages 8 and 11"`
- `"Mike, software developer, power user, 3 kids with different needs"`
- `"Rose, grandmother, very uncomfortable with technology, setting up for grandchild visits"`
- `"Emma, busy working mom, mobile-first, impatient with slow-loading pages"`

### persona.tech_comfort

| Level | Behavior |
|-------|----------|
| `low` | Confused by jargon, needs clear labels, hesitates often |
| `medium` | Comfortable with common patterns, occasional confusion |
| `high` | Quick to navigate, expects efficiency, notices subtle issues |

### persona.patience

| Level | Behavior |
|-------|----------|
| `low` | Gives up quickly, frustrated by friction |
| `medium` | Tries a few times, documents frustration |
| `high` | Persistent, tries multiple approaches |

### persona.device

| Device | Implications |
|--------|--------------|
| `desktop` | Full browser, hover states, keyboard shortcuts |
| `mobile` | Touch, small screen, no hover |
| `tablet` | Touch, medium screen, mixed patterns |

### target.app

| App | Description | Tool | Phase |
|-----|-------------|------|-------|
| `parent-portal` | Web application for parents | Playwright MCP | Phase 1 |
| `desktop-client` | Electron Windows app | MCPControl | Phase 2+ |
| `extension` | Chrome browser extension | Playwright on Windows | Phase 2+ |

### model

| Model | When to Use |
|-------|-------------|
| `haiku` | Default, fast, cheap, good for straightforward flows |
| `sonnet` | Complex UX analysis, nuanced observations |
| `gemini` | Very long sessions (token budget) |

### observations.auto_triage

| Setting | Action |
|---------|--------|
| `create_bead` | Automatically create a bug bead for this observation |
| `log_only` | Log observation but don't create bead |
| `ignore` | Don't track this severity level |

### retry.backoff

| Strategy | Behavior |
|----------|----------|
| `none` | No delay between retries |
| `linear` | backoff_base, 2*base, 3*base, ... |
| `exponential` | backoff_base, 2*base, 4*base, 8*base, ... |

### retry.on_errors (Error Types)

| Error Type | Description | Should Retry? |
|------------|-------------|---------------|
| `browser_crash` | Playwright browser process died | Yes |
| `timeout` | Test exceeded timeout limit | Yes |
| `network_error` | Network connectivity issue | Yes |
| `test_failure` | Test criteria not met (actual bug) | No |
| `blocked` | User flow blocked, cannot proceed | No |
| `api_quota` | Out of API tokens | Queue for later |

---

## Observation Output Format

Observations produced by the agent must include severity and confidence:

```json
{
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
  ]
}
```

### Severity Levels

| Level | Description | Auto Action |
|-------|-------------|-------------|
| `P0` | Blocking - user cannot complete goal | create_bead |
| `P1` | Significant friction - user likely to abandon | create_bead |
| `P2` | Minor friction - noticeable but not blocking | log_only |
| `P3` | Nitpick - improvement opportunity | log_only |

### Confidence Levels

| Level | Description |
|-------|-------------|
| `high` | Agent is very sure this is a real UX issue |
| `medium` | Agent thinks this might be an issue |
| `low` | Agent is uncertain, flagging for human review |

### Human Validation Fields

After human review, these fields are updated:

```json
{
  "validated": true,       // Human confirmed observation
  "false_positive": false  // Was this a false alarm?
}
```

Track false positive rate over time. If FP rate > 20%, flag persona for recalibration.

---

## Validation Rules

### Required Fields
- `scenario`: Must be unique, snake_case
- Either `user_story` OR (`persona` + `goal`) required, not both
- If using `user_story`:
  - `user_story.name`: Non-empty string
  - `user_story.persona`: Non-empty string (minimum 10 characters)
  - `user_story.goal`: At least 20 characters
- If using `persona`:
  - `persona.name`: Non-empty string
  - `persona.role`: One of: parent, child, admin, teacher
  - `persona.context`: At least 20 characters
  - `goal`: At least 20 characters
- `target.app`: One of: parent-portal, desktop-client, extension
- `success_criteria`: At least one criterion

### Defaults

```yaml
# Applied if not specified
version: 1
persona:
  tech_comfort: medium
  patience: medium
  device: desktop
target:
  environment: staging
observations:
  require_severity: true
  require_confidence: true
  auto_triage:
    P0_P1: create_bead
    P2_P3: log_only
recording:
  video: true
  trace: true
  screenshots:
    on_failure: true
    on_confusion: true
    on_demand: true
    require_review: true
  headed: false
  headed_verification: null
wait_strategies:
  network_idle: true
  animation_complete: true
  min_load_time: 1000
  custom_selectors: []
retry:
  max_attempts: 3
  on_errors: [browser_crash, timeout, network_error]
  not_on: [test_failure, blocked]
  backoff: exponential
  backoff_base: 1000
timeout: 600
model: haiku
test_data:
  email_pattern: "test+{scenario}+{run_id}@screencoach.test"
  email_inbox: skip_verification
  cleanup_strategy:
    on_success: delete_account
    on_failure: mark_for_review
    on_crash: cleanup_job
  isolation:
    unique_suffix: true
windows:
  use_mcpcontrol: false
```

---

## Directory Structure

```
screencoach/scenarios/
├── parent-portal/
│   ├── registration/
│   │   ├── new-parent.yaml
│   │   ├── registration-errors.yaml    # Error recovery scenario
│   │   └── mobile-registration.yaml
│   ├── onboarding/
│   │   ├── add-first-child.yaml
│   │   └── add-multiple-children.yaml
│   └── dashboard/
│       ├── view-activity.yaml
│       └── modify-settings.yaml
├── user-stories/                        # NEW: Unified user story files
│   ├── sarah-registers.yaml
│   ├── mike-power-setup.yaml
│   └── rose-error-recovery.yaml
├── personas/
│   ├── sarah-low-tech.yaml
│   ├── mike-tech-savvy.yaml
│   ├── rose-very-low-tech.yaml
│   └── emma-mobile.yaml
└── config.yaml
```

---

## Example: User Story Format (Recommended)

```yaml
# scenarios/user-stories/sarah-registers.yaml
scenario: sarah_registers
version: 1
description: "First-time parent registration using unified user story format"
tags: [registration, critical-path, p0]

user_story:
  name: new_parent_registration
  persona: Sarah, first-time parent, not tech-savvy, 2 kids ages 8 and 11
  goal: Register for ScreenCoach and set up profiles for both children
  context: |
    Found ScreenCoach through school newsletter.
    Wants to limit gaming during homework hours.
    Gets confused by technical jargon.

target:
  app: parent-portal
  environment: staging

steps:
  - Navigate to homepage
  - Find and click registration
  - Complete signup form
  - Add first child (age 8)
  - Add second child (age 11)
  - View dashboard with both children

success_criteria:
  - Account created successfully
  - Two child profiles added
  - Dashboard visible with both children listed

evaluate:
  - Is the signup button easy to find?
  - Are error messages helpful?
  - Is adding multiple children intuitive?
  - Any points where Sarah would give up?

observations:
  require_severity: true
  require_confidence: true
  auto_triage:
    P0_P1: create_bead
    P2_P3: log_only

recording:
  video: true
  trace: true
  screenshots:
    on_failure: true
    on_confusion: true
    require_review: true

wait_strategies:
  network_idle: true
  animation_complete: true
  min_load_time: 1500
  custom_selectors:
    - "[data-testid='app-ready']"

retry:
  max_attempts: 3
  on_errors: [browser_crash, timeout, network_error]
  not_on: [test_failure, blocked]
  backoff: exponential

timeout: 600
model: haiku

test_data:
  email_pattern: "sarah+{run_id}@screencoach.test"
  email_inbox: skip_verification
  cleanup_strategy:
    on_success: delete_account
    on_failure: mark_for_review
  isolation:
    unique_suffix: true
```

---

## Example: Error Recovery Scenario (QA Recommended)

```yaml
# scenarios/parent-portal/registration/registration-errors.yaml
scenario: registration_with_errors
version: 1
description: "Test error handling during registration - error recovery UX"
tags: [registration, error-handling, p1]

user_story:
  name: error_recovery_registration
  persona: Rose, grandmother, very uncomfortable with technology
  goal: |
    Register for ScreenCoach, but make typical mistakes:
    - Enter invalid email first (missing @)
    - Use weak password
    - Skip required fields
    Observe error message clarity and recovery UX.
  context: |
    Setting up for grandchild visits.
    Needs explicit, simple instructions.
    Will give up if errors aren't clear.

target:
  app: parent-portal
  environment: staging

steps:
  - Navigate to signup
  - Enter email with typo (missing @)
  - Try to submit
  - Fix email, enter weak password
  - Try to submit
  - Fix password, skip required field
  - Try to submit
  - Eventually complete successfully

success_criteria:
  - Error messages are shown for each mistake
  - Rose can understand and fix each error
  - Form remembers previously correct entries
  - Eventually succeeds without help

evaluate:
  - Are errors shown near the field (not just top of form)?
  - Is the language simple (no tech jargon)?
  - Does the form preserve correct entries?
  - Are password requirements shown BEFORE submission?
  - Would Rose give up or call for help?

observations:
  require_severity: true
  require_confidence: true

timeout: 600
model: sonnet  # Complex UX analysis

test_data:
  email_pattern: "rose+error+{run_id}@screencoach.test"
  email_inbox: skip_verification
  cleanup_strategy:
    on_success: delete_account
    on_failure: mark_for_review
```

---

## Global Config

```yaml
# screencoach/scenarios/config.yaml
defaults:
  environment: staging
  model: haiku
  timeout: 600

environments:
  staging:
    parent_portal_url: https://staging.screencoach.example.com
    api_url: https://api.staging.screencoach.example.com
  production:
    parent_portal_url: https://app.screencoach.example.com
    api_url: https://api.screencoach.example.com

# Test data defaults
test_data_defaults:
  email_domain: screencoach.test
  email_inbox: skip_verification
  cleanup_enabled: true

# Stability thresholds
stability:
  flake_rate_threshold: 0.10      # Quarantine if >10% flaky
  false_positive_threshold: 0.20  # Recalibrate persona if >20% FP

app_context: |
  ScreenCoach is a parental control app for managing children's screen time.
  Key concepts:
  - Parents create accounts and add children
  - Each child has a profile with time limits
  - Parents can set schedules and block apps
  - Children use a separate app/extension

  You're testing the Parent Portal - where parents manage settings.

windows_vm:
  host: 192.168.122.90
  mcpcontrol_port: 3232
```

---

## Validation Command

```bash
# Validate scenario file
gt tester validate scenarios/user-stories/sarah-registers.yaml

# Validate all scenarios
gt tester validate scenarios/**/*.yaml
```

---

## Testing Checklist

- [ ] User story format validates correctly
- [ ] Detailed persona format still works
- [ ] Cannot use both user_story and persona
- [ ] Wait strategies applied during execution
- [ ] Retry logic triggers on appropriate errors
- [ ] Test data isolation generates unique emails
- [ ] Cleanup runs on success/failure/crash
- [ ] Observation severity/confidence required
- [ ] Defaults applied correctly
- [ ] Global config merges properly
