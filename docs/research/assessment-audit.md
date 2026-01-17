# Assessment & Review Audit: Gap Analysis

**Audit ID**: hq-brbb
**Date**: 2026-01-16
**Auditor**: gastown/polecats/morsov

---

## Executive Summary

Audited 10 assessment and review documents across screencoach, gastown, and beads rigs. Found:
- **14 verified missing beads** across screencoach and gastown
- **8 items covered** by existing beads
- **7 items** that are guidance/documentation (no bead needed)
- **9 critical security items** were marked as COMPLETED in the code review (Phase 1 done)

---

## 1. Documents Audited

### Assessment Documents

| Document | Location | Status |
|----------|----------|--------|
| DESKTOP-BUILD-DEPLOY-TEST-ASSESSMENT.md | screencoach/docs/ | 4 recommended actions |

### Review Documents

| Document | Location | Purpose | Beads Created? |
|----------|----------|---------|----------------|
| rlm-gastown-philosophy-review.md | gastown/docs/reviews/ | Philosophy alignment | Yes (hq-qndw.1) |
| rlm-vs-gastown-review.md | gastown/docs/reviews/ | Architecture comparison | Yes (hq-qndw.2) |
| ai-testing-qa-review.md | gastown/docs/reviews/ | QA concerns for AI testing | Partial (hq-zlx2 epic) |
| ai-testing-pm-review.md | gastown/docs/reviews/ | PM scope recommendations | Partial (hq-zlx2 epic) |
| 2024-12-28-project-review-checkpoint.md | screencoach/investigations/ | Early project status | Outdated |
| 2026-01-02-systematic-code-review.md | screencoach/investigations/ | Security code review | Phase 1 DONE |
| 2026-01-03-coordination-pm-review.md | screencoach/investigations/ | Coordination service gaps | No beads found |
| pr-752-chaos-testing-review.md | beads/docs/ | PR review guidance | N/A (PR review) |
| ai-coach-technical-review.md | screencoach/specs/ | AI Coach tech gaps | No beads found |

---

## 2. Related Beads Found

| Bead ID | Title | Status | Source Document |
|---------|-------|--------|-----------------|
| hq-c69x | Implement Playwright E2E tests for desktop UI | HOOKED | Desktop assessment |
| hq-qndw.1 | Review: RLM architecture vs Gas Town philosophy | HOOKED | Philosophy review task |
| hq-qndw.2 | Review: Is RLM architecture over-engineering? | HOOKED | Engineering review task |
| hq-zlx2 | AI User Testing Implementation | OPEN (epic) | AI Testing reviews |
| hq-zlx2.6 | Implement artifact recording | HOOKED | AI Testing QA review |
| hq-a9yd | AI Exploratory UI Testing - ScreenCoach | OPEN (epic) | AI Testing |
| hq-wx0a | Knowledge Capture Infrastructure | OPEN (epic) | RLM discussions |

---

## 3. Gap Analysis by Document

### From: DESKTOP-BUILD-DEPLOY-TEST-ASSESSMENT.md

**Section 5: "Conclusion - Recommended Next Steps"**

| # | Action Item | Existing Bead | Status |
|---|-------------|---------------|--------|
| 1 | Implement Playwright automation for UI testing | hq-c69x | **COVERED** |
| 2 | Add headless test mode to UI for CI | None | **MISSING** |
| 3 | Complete extension store publishing for stable IDs | None | **MISSING** |
| 4 | Enhance test data seeding for fully automated setup | None | **MISSING** |

---

### From: 2026-01-02-systematic-code-review.md

**Phase 1: Critical Security (COMPLETED 2026-01-09)**

| # | Action Item | Status |
|---|-------------|--------|
| 1 | Fix JWT fallback secret | DONE |
| 2 | Fix remote debugging exposure | DONE |
| 3 | Fix command injection in alerts.service.ts | DONE |
| 4 | Fix command injection in MCP server | DONE |
| 5 | Add authorization to screenshot endpoints | DONE |
| 6 | Implement CSRF protection in admin portal | DONE |
| 7 | Fix migration duplicate enum creation | DONE |
| 8 | Change default role from SUPER_ADMIN to PARENT | DONE |
| 9 | Add missing FK constraints | DONE |

**Phase 2: High Priority (Pending)**

| # | Action Item | Existing Bead | Status |
|---|-------------|---------------|--------|
| 9 | Update vulnerable npm dependencies | None | **MISSING** |
| 10 | Add error handling to device auth guard | None | **MISSING** |
| 11 | Fix 64-bit compatibility in LockdownWindow | None | **MISSING** |
| 12 | Implement proper error boundaries in frontend | None | **MISSING** |
| 13 | Fix accessibility issues | None | **MISSING** |
| 14 | Remove hardcoded IPs from config files | None | **MISSING** |
| 15 | Add route-level permissions to admin | None | **MISSING** |

---

### From: 2026-01-03-coordination-pm-review.md

**P0: Must Fix (Blocking Issues)**

| # | Action Item | Existing Bead | Status |
|---|-------------|---------------|--------|
| 1 | Document inbox API | None | **MISSING** |
| 2 | Add message length validation | None | **MISSING** |
| 3 | Lock race condition fix | None | **MISSING** |

**P1: Should Fix Soon**

| # | Action Item | Existing Bead | Status |
|---|-------------|---------------|--------|
| 4 | File conflict detection | None | **MISSING** |
| 5 | Session history view | None | **MISSING** |
| 6 | Heartbeat watchdog | None | **MISSING** |
| 7 | Offline queue | None | **MISSING** |
| 8 | Add notification sound | None | **MISSING** |
| 9 | Authentication for dashboard | None | **MISSING** |
| 10 | Windows troubleshooting docs | None | **MISSING** |

---

### From: ai-testing-qa-review.md & ai-testing-pm-review.md

**QA Blocking Concerns (PM says "must fix before Phase 1")**

| # | Action Item | Existing Bead | Status |
|---|-------------|---------------|--------|
| 1 | Test data isolation strategy (unique emails, cleanup) | None | **MISSING** |
| 2 | Basic retry logic for infrastructure failures | None | **MISSING** |
| 3 | Preflight environment checks | None | **MISSING** |

Note: Some QA concerns may be addressed within the hq-zlx2 epic implementation. The epic exists but specific QA hardening tasks aren't broken out as individual beads.

---

### From: ai-coach-technical-review.md

**Must Have Before Implementation**

| # | Action Item | Existing Bead | Status |
|---|-------------|---------------|--------|
| 1 | Implement BedrockProvider | None | **MISSING** |
| 2 | Define AI Self-Improvement Workflow | None | **MISSING** |
| 3 | Add A/B Test Management (lifecycle, significance) | None | **MISSING** |
| 4 | Add Prompt Injection Protection | None | **MISSING** |

---

### From: rlm-gastown-philosophy-review.md & rlm-vs-gastown-review.md

These reviews provide guidance and recommendations rather than specific action items requiring beads. The verdict was "partial adoption with modifications" - treat RLM as internal technique, not orchestration model. No specific implementation beads needed; these inform design decisions.

---

### From: pr-752-chaos-testing-review.md

This is a PR review document providing decision guidance ("Merge with modifications"). Not an assessment requiring follow-up beads.

---

## 4. Summary of Missing Beads

### ScreenCoach - High Priority

| # | Proposed Title | Type | Priority | Source |
|---|----------------|------|----------|--------|
| 1 | Add headless test mode to ScreenCoach.UI for CI | task | P2 | Desktop assessment |
| 2 | Publish Edge extension to store for stable ID | task | P2 | Desktop assessment |
| 3 | Enhance test data seeding automation | task | P2 | Desktop assessment |

### ScreenCoach - Code Review Phase 2

| # | Proposed Title | Type | Priority | Source |
|---|----------------|------|----------|--------|
| 4 | Update vulnerable npm dependencies | task | P1 | Systematic code review |
| 5 | Add error handling to device auth guard | bug | P2 | Systematic code review |
| 6 | Fix 64-bit compatibility in LockdownWindow | bug | P2 | Systematic code review |
| 7 | Implement error boundaries in parent portal frontend | task | P2 | Systematic code review |
| 8 | Fix accessibility issues (keyboard nav, aria) | task | P2 | Systematic code review |
| 9 | Remove hardcoded IPs from config files | task | P2 | Systematic code review |
| 10 | Add route-level permissions to admin portal | task | P2 | Systematic code review |

### ScreenCoach - AI Coach

| # | Proposed Title | Type | Priority | Source |
|---|----------------|------|----------|--------|
| 11 | Implement BedrockProvider for AI Coach | task | P1 | AI Coach tech review |
| 12 | Define AI self-improvement workflow for prompts | task | P1 | AI Coach tech review |
| 13 | Implement A/B test management (lifecycle, stats) | task | P2 | AI Coach tech review |
| 14 | Add prompt injection protection | bug | P1 | AI Coach tech review |

### ScreenCoach - Coordination Service

| # | Proposed Title | Type | Priority | Source |
|---|----------------|------|----------|--------|
| 15 | Document inbox API in coordination-setup.md | task | P1 | Coordination PM review |
| 16 | Add message length validation (10k chars) | bug | P1 | Coordination PM review |
| 17 | Fix lock race condition with SQLite transaction | bug | P1 | Coordination PM review |
| 18 | Implement file conflict detection | task | P1 | Coordination PM review |
| 19 | Add session history view to dashboard | task | P2 | Coordination PM review |
| 20 | Implement heartbeat process watchdog | task | P2 | Coordination PM review |
| 21 | Add offline message queue | task | P2 | Coordination PM review |
| 22 | Add notification sound option | task | P3 | Coordination PM review |
| 23 | Add basic authentication to dashboard | task | P1 | Coordination PM review |
| 24 | Add Windows troubleshooting section to docs | task | P2 | Coordination PM review |

### Gastown - AI Testing QA Hardening

| # | Proposed Title | Type | Priority | Source |
|---|----------------|------|----------|--------|
| 25 | AI Testing: Implement test data isolation strategy | task | P1 | AI Testing QA review |
| 26 | AI Testing: Add retry logic for infrastructure failures | task | P1 | AI Testing QA review |
| 27 | AI Testing: Implement preflight environment checks | task | P1 | AI Testing QA review |

---

## 5. Recommendations

1. **Create beads for ScreenCoach Code Review Phase 2** - 7 items pending from security review
2. **Create beads for Coordination Service P0/P1** - Core stability and documentation gaps
3. **Create beads for AI Coach implementation gaps** - Provider and self-improvement workflow
4. **Add QA hardening tasks to AI Testing epic** - Test reliability concerns
5. **Consider periodic audit** - Assessments and reviews should trigger bead creation workflow

---

## 6. Proposed Beads to Create

### Immediate (P1 - This Sprint)

```bash
# ScreenCoach - Security Review Phase 2
bd create --title="Update vulnerable npm dependencies (glob, qs, nodemailer)" --type=task --priority=1

# ScreenCoach - AI Coach
bd create --title="Implement BedrockProvider for AI Coach" --type=task --priority=1
bd create --title="Add prompt injection protection to AI Coach" --type=bug --priority=1

# ScreenCoach - Coordination Service
bd create --title="Document inbox API in coordination-setup.md" --type=task --priority=1
bd create --title="Add message length validation to coordination service" --type=bug --priority=1
bd create --title="Fix lock race condition with SQLite transaction" --type=bug --priority=1
bd create --title="Add basic authentication to coordination dashboard" --type=task --priority=1
bd create --title="Implement file conflict detection in coordination service" --type=task --priority=1

# Gastown - AI Testing
bd create --title="AI Testing: Implement test data isolation strategy" --type=task --priority=1
bd create --title="AI Testing: Add retry logic for infrastructure failures" --type=task --priority=1
bd create --title="AI Testing: Implement preflight environment checks" --type=task --priority=1
```

### Soon (P2 - Next Sprint)

```bash
# ScreenCoach - Desktop
bd create --title="Add headless test mode to ScreenCoach.UI for CI" --type=task --priority=2
bd create --title="Publish Edge extension to store for stable ID" --type=task --priority=2
bd create --title="Enhance test data seeding automation" --type=task --priority=2

# ScreenCoach - Code Quality
bd create --title="Add error handling to device auth guard" --type=bug --priority=2
bd create --title="Fix 64-bit compatibility in LockdownWindow" --type=bug --priority=2
bd create --title="Implement error boundaries in parent portal" --type=task --priority=2
bd create --title="Fix accessibility issues in parent portal" --type=task --priority=2
bd create --title="Remove hardcoded IPs from config files" --type=task --priority=2
bd create --title="Add route-level permissions to admin portal" --type=task --priority=2

# ScreenCoach - AI Coach
bd create --title="Implement A/B test management for AI prompts" --type=task --priority=2
bd create --title="Define AI self-improvement workflow for prompts" --type=task --priority=2

# ScreenCoach - Coordination
bd create --title="Add session history view to coordination dashboard" --type=task --priority=2
bd create --title="Implement heartbeat process watchdog" --type=task --priority=2
bd create --title="Add offline message queue to coordination" --type=task --priority=2
bd create --title="Add Windows troubleshooting to coordination docs" --type=task --priority=2
```

---

## 7. Items Already Covered

| Document | Recommended Action | Existing Bead |
|----------|-------------------|---------------|
| Desktop Assessment | Implement Playwright E2E tests | hq-c69x |
| AI Testing Reviews | AI User Testing Implementation | hq-zlx2 (epic) |
| AI Testing Reviews | Artifact recording | hq-zlx2.6 |
| RLM Reviews | Review RLM philosophy | hq-qndw.1 |
| RLM Reviews | Review RLM engineering | hq-qndw.2 |
| Systematic Code Review | Phase 1 Critical Security | All 9 items COMPLETED |
| AI Testing | AI Exploratory Testing | hq-a9yd (epic) |
| RLM/Knowledge | Knowledge Capture | hq-wx0a (epic) |

---

## 8. Items Not Requiring Beads

| Document | Item | Reason |
|----------|------|--------|
| rlm-gastown-philosophy-review.md | Allow RLM-style internal processing | Guidance, not action |
| rlm-gastown-philosophy-review.md | Do not create Librarian agent role | Rejection decision |
| rlm-gastown-philosophy-review.md | Use labels for knowledge classification | Design guidance |
| rlm-gastown-philosophy-review.md | Preserve core philosophy | Non-actionable principle |
| rlm-vs-gastown-review.md | Synthesis recommendations | Comparative analysis |
| pr-752-chaos-testing-review.md | PR decision framework | Review guidance, not action |
| 2024-12-28-project-review-checkpoint.md | All items | Outdated (2024) - superseded by later reviews |

---

*Audit completed by gastown/polecats/morsov, 2026-01-16*
