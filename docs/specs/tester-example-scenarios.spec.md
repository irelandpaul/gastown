# Spec: Example Test Scenarios

**Task**: hq-t58n
**Author**: Mayor's Aid
**Date**: 2026-01-14
**Status**: Draft

## Overview

This spec provides example test scenarios for ScreenCoach Parent Portal, demonstrating the scenario format and persona variety.

---

## Persona Library

### Sarah (Low Tech Parent)

```yaml
# personas/sarah-low-tech.yaml
name: Sarah
role: parent
context: |
  First-time user, not tech-savvy, has 2 kids (ages 8 and 12).
  Found ScreenCoach through school recommendation.
  Primary goal: limit screen time during homework hours.
  Gets confused by technical jargon.
  Will read instructions carefully if they're clear.
tech_comfort: low
patience: medium
device: desktop
```

### Mike (Tech Savvy Parent)

```yaml
# personas/mike-tech-savvy.yaml
name: Mike
role: parent
context: |
  Software developer, comfortable with technology.
  Has 3 kids (ages 6, 9, 14) with different needs.
  Expects efficient, powerful tools.
  Will find workarounds if UI is slow.
  Notices small UX issues others miss.
tech_comfort: high
patience: low
device: desktop
```

### Grandma Rose (Very Low Tech)

```yaml
# personas/grandma-rose.yaml
name: Rose
role: parent
context: |
  Grandmother setting up ScreenCoach for grandchild visits.
  Very uncomfortable with technology.
  Needs explicit, simple instructions.
  Will call for help if confused.
  Using shared family computer.
tech_comfort: low
patience: high
device: desktop
```

### Emma (Mobile-First Parent)

```yaml
# personas/emma-mobile.yaml
name: Emma
role: parent
context: |
  Busy working mom, does everything on phone.
  Medium tech comfort, uses many apps.
  Impatient with slow-loading pages.
  Often distracted, may abandon if too many steps.
tech_comfort: medium
patience: low
device: mobile
```

---

## Critical Path Scenarios

### 1. New Parent Registration

```yaml
# scenarios/parent-portal/registration/new-parent.yaml
scenario: register_new_parent
version: 1
description: "First-time parent registration - critical path"
tags: [registration, critical-path, p0]

persona:
  name: Sarah
  role: parent
  context: |
    First-time user, not tech-savvy, has 2 kids (ages 8 and 12).
    Found ScreenCoach through school recommendation.
    Primary goal: limit screen time during homework hours.
  tech_comfort: low
  patience: medium

target:
  app: parent-portal
  environment: staging

goal: |
  Register for ScreenCoach and complete the signup process.
  Navigate as Sarah would - carefully reading everything,
  possibly making mistakes a non-technical user would make.

steps:
  - Navigate to homepage
  - Find signup/registration option
  - Complete registration form
  - Verify email if required
  - Complete any onboarding steps

success_criteria:
  - Account created successfully
  - Logged into dashboard
  - No unrecoverable errors

evaluate:
  - Is the signup button easy to find?
  - Are form field labels clear?
  - Are error messages helpful?
  - Is email verification smooth?
  - Would Sarah complete this or give up?

recording:
  video: true
  trace: true
  screenshots:
    on_failure: true
    on_confusion: true
    on_demand: true

timeout: 600
model: haiku
```

### 2. Add First Child

```yaml
# scenarios/parent-portal/onboarding/add-first-child.yaml
scenario: add_first_child
version: 1
description: "Add first child after registration"
tags: [onboarding, critical-path, p0]

persona:
  name: Sarah
  role: parent
  context: |
    Just registered, now needs to add her first child.
    Child is 8 years old, uses tablet for games.
    Wants to limit to 2 hours per day.
  tech_comfort: low
  patience: medium

target:
  app: parent-portal
  environment: staging

goal: |
  Add first child profile with basic settings.
  Set up reasonable screen time limits.
  Understand what the child profile means.

steps:
  - Start from dashboard (post-registration)
  - Find "Add Child" option
  - Enter child information
  - Set basic time limits
  - Understand next steps (install on child device)

success_criteria:
  - Child profile created
  - Time limits set
  - Clear next steps shown

evaluate:
  - Is "Add Child" button obvious?
  - Is the form asking for too much info?
  - Are time limit controls intuitive?
  - Does Sarah understand what to do next?

depends_on:
  - register_new_parent

timeout: 300
model: haiku
```

### 3. View Child Activity

```yaml
# scenarios/parent-portal/dashboard/view-activity.yaml
scenario: view_child_activity
version: 1
description: "View child's screen time activity"
tags: [dashboard, monitoring, p1]

persona:
  name: Sarah
  role: parent
  context: |
    Has been using ScreenCoach for a week.
    Child has some activity logged.
    Wants to see what apps child is using.
  tech_comfort: low
  patience: medium

target:
  app: parent-portal
  environment: staging

goal: |
  Check child's screen time activity for today.
  Understand which apps are being used.
  See if time limits are working.

steps:
  - Log into parent portal
  - Navigate to child's profile
  - View today's activity
  - Understand the data shown

success_criteria:
  - Activity data displayed
  - Can see app breakdown
  - Understands time remaining

evaluate:
  - Is the activity data clear?
  - Are charts/graphs easy to understand?
  - Can Sarah tell if limits are working?
  - Is navigation intuitive?

test_data:
  seed_account: test-parent-with-activity@example.com
  cleanup: false

timeout: 300
model: haiku
```

---

## Edge Case Scenarios

### 4. Registration with Errors

```yaml
# scenarios/parent-portal/registration/registration-errors.yaml
scenario: registration_with_errors
version: 1
description: "Test error handling during registration"
tags: [registration, error-handling, p2]

persona:
  name: Rose
  role: parent
  context: |
    Grandmother, very uncomfortable with computers.
    Will make mistakes - typos, wrong format.
    Needs very clear error messages.
  tech_comfort: low
  patience: high

target:
  app: parent-portal
  environment: staging

goal: |
  Attempt registration but make common mistakes:
  - Mistype email address
  - Use weak password
  - Skip required fields
  Observe how the system helps Rose recover.

steps:
  - Navigate to signup
  - Enter email with typo (missing @)
  - Try to submit
  - Fix email, enter weak password
  - Try to submit
  - Eventually complete successfully

success_criteria:
  - Error messages are shown
  - Rose can understand and fix errors
  - Eventually succeeds

evaluate:
  - Are errors shown near the field?
  - Is the language simple (no tech jargon)?
  - Can Rose recover without help?
  - Does the form remember correct entries?

timeout: 600
model: sonnet  # Complex UX analysis needed
```

### 5. Mobile Registration

```yaml
# scenarios/parent-portal/registration/mobile-registration.yaml
scenario: mobile_registration
version: 1
description: "Test registration on mobile device"
tags: [registration, mobile, p1]

persona:
  name: Emma
  role: parent
  context: |
    Doing everything on phone while commuting.
    Quick glances, may be interrupted.
    Expects mobile-friendly experience.
  tech_comfort: medium
  patience: low
  device: mobile

target:
  app: parent-portal
  environment: staging

goal: |
  Complete registration on mobile phone.
  Note any mobile-specific issues:
  - Touch targets too small
  - Keyboard covering form
  - Layout issues

steps:
  - Open on mobile browser
  - Navigate signup flow
  - Complete registration
  - Note mobile UX issues

success_criteria:
  - Registration completes on mobile
  - All elements touchable
  - No layout issues blocking progress

evaluate:
  - Are touch targets large enough?
  - Does keyboard obscure important elements?
  - Is text readable without zooming?
  - Is the experience comparable to desktop?

timeout: 600
model: haiku
```

---

## Power User Scenarios

### 6. Tech Savvy Setup

```yaml
# scenarios/parent-portal/onboarding/tech-savvy-setup.yaml
scenario: tech_savvy_complete_setup
version: 1
description: "Power user completing full setup"
tags: [onboarding, power-user, p2]

persona:
  name: Mike
  role: parent
  context: |
    Software developer, wants full control.
    Has 3 kids with different needs.
    Looking for advanced features.
  tech_comfort: high
  patience: low

target:
  app: parent-portal
  environment: staging

goal: |
  Set up ScreenCoach for all 3 children with custom settings:
  - Different schedules per child
  - App-specific rules
  - Notification preferences
  Find any "missing" features Mike would expect.

steps:
  - Register or login
  - Add 3 children with different ages
  - Set up unique schedules for each
  - Configure advanced settings
  - Note missing features

success_criteria:
  - 3 children added
  - Different schedules working
  - Advanced settings accessible

evaluate:
  - Can Mike find advanced options?
  - Is bulk setup efficient?
  - Are there shortcuts for power users?
  - What features does Mike expect but not find?

timeout: 900
model: sonnet  # Complex multi-child setup
```

---

## Batch Configuration

```yaml
# scenarios/parent-portal/batch-critical-path.yaml
batch:
  name: Critical Path Tests
  description: "Run all P0 critical path scenarios"
  scenarios:
    - registration/new-parent.yaml
    - onboarding/add-first-child.yaml
    - dashboard/view-activity.yaml
  parallel: 2
  stop_on_fail: true
  convoy: critical-path-tests
```

---

## Running Examples

```bash
# Run single scenario
gt tester run scenarios/parent-portal/registration/new-parent.yaml

# Run with visible browser
gt tester run scenarios/parent-portal/registration/new-parent.yaml --headed

# Run critical path batch
gt tester batch scenarios/parent-portal/*.yaml --filter critical-path

# Run all parent portal scenarios in parallel
gt tester batch "scenarios/parent-portal/**/*.yaml" --parallel 3

# Run with sonnet for better UX analysis
gt tester run scenarios/parent-portal/registration/registration-errors.yaml --model sonnet
```

---

## Expected Observations

Based on common UX issues, these scenarios should catch:

| Scenario | Likely Observations |
|----------|---------------------|
| new-parent | Signup button visibility, form length |
| add-first-child | Age picker UX, time limit controls |
| view-activity | Chart clarity, data presentation |
| registration-errors | Error message quality, recovery flow |
| mobile-registration | Touch targets, keyboard issues |
| tech-savvy-setup | Missing power user features |

---

## Scenario Coverage Matrix

| Flow | Low Tech | High Tech | Mobile |
|------|----------|-----------|--------|
| Registration | Sarah ✓ | Mike | Emma ✓ |
| Add Child | Sarah ✓ | Mike ✓ | |
| View Activity | Sarah ✓ | | |
| Settings | | Mike | |
| Error Recovery | Rose ✓ | | |

Gaps to fill:
- High tech registration
- Mobile onboarding
- Settings for non-power users
