# Assessment & Review Audit: Gap Analysis

**Audit ID**: hq-brbb
**Date**: 2026-01-13
**Auditor**: gastown/polecats/morsov

---

## Executive Summary

Audited assessment documents and code reviews across all rigs to identify follow-up actions that lack corresponding beads. Found **10 missing beads** from infrastructure review recommendations and **3 missing beads** from desktop assessment recommendations.

---

## 1. Documents Found

### Assessment Documents

| Document | Location | Status |
|----------|----------|--------|
| DESKTOP-BUILD-DEPLOY-TEST-ASSESSMENT.md | screencoach/docs/ | 4 recommended actions |

### Review Documents

| Document | Location | Purpose |
|----------|----------|---------|
| infrastructure-review.md | gastown/docs/reviews/ | Code cleanup, 7 priority items |
| rlm-gastown-philosophy-review.md | gastown/docs/reviews/ | Philosophy alignment |
| rlm-vs-gastown-review.md | gastown/docs/reviews/ | Architecture comparison |

---

## 2. Related Beads Found

| Bead ID | Title | Status | Source Document |
|---------|-------|--------|-----------------|
| hq-972 | Desktop App Assessment: Build/Deploy/Test Pipeline | CLOSED | Assessment creation task |
| hq-qndw | RLM + Librarian Knowledge Graph Architecture | OPEN (epic) | Architecture proposal |
| hq-qndw.1 | Review: RLM architecture vs Gas Town philosophy | HOOKED | Philosophy review task |
| hq-qndw.2 | Review: Is RLM architecture over-engineering? | HOOKED | Engineering review task |
| hq-c69x | Implement Playwright E2E tests for desktop UI | HOOKED | Desktop assessment action #1 |
| hq-d8dh | Run Playwright login test | CLOSED | Follow-up testing |

---

## 3. Gap Analysis

### From: DESKTOP-BUILD-DEPLOY-TEST-ASSESSMENT.md

**Recommended Next Steps** (from Section 5 "Conclusion"):

| Action Item | Existing Bead | Status |
|-------------|---------------|--------|
| 1. Implement Playwright automation for UI testing | hq-c69x | COVERED |
| 2. Add headless test mode to UI for CI | None | **MISSING** |
| 3. Complete extension store publishing for stable IDs | None | **MISSING** |
| 4. Enhance test data seeding for fully automated setup | None | **MISSING** |

### From: infrastructure-review.md

**Recommended Priority Items** (from Section "Recommended Priority"):

| Priority | Action Item | Existing Bead | Status |
|----------|-------------|---------------|--------|
| 1 | Delete keepalive package (entire package unused) | None | **MISSING** |
| 2 | Fix claude/RoleTypeFor() (incorrect behavior) | None | **MISSING** |
| 3 | Fix config/GetAccount() (pointer to stack bug) | None | **MISSING** |
| 4 | Fix polecat/pending.go (non-atomic writes) | None | **MISSING** |
| 5 | Delete 21 unused constants (maintenance burden) | None | **MISSING** |
| 6 | Consolidate atomic write pattern (DRY) | None | **MISSING** |
| 7 | Add checkpoint tests (crash recovery critical) | None | **MISSING** |

### From: rlm-gastown-philosophy-review.md

**Recommendations** (Section 4):

| Recommendation | Type | Bead Needed? |
|----------------|------|--------------|
| Allow RLM-style internal processing for large inputs | Adopt | No (guidance, not action) |
| Do not create Librarian agent role | Skip | No (rejection, not action) |
| Use labels for knowledge classification | Guidance | No (design choice) |
| Preserve core philosophy | Guidance | No (non-actionable) |

### From: rlm-vs-gastown-review.md

This is a comparative analysis document with synthesis recommendations but no specific actionable follow-ups requiring beads.

---

## 4. Summary of Missing Beads

### High Priority (Bugs/Fixes)

| Proposed Title | Type | Priority | Source |
|----------------|------|----------|--------|
| Fix claude/RoleTypeFor() missing deacon and crew mapping | bug | P1 | infrastructure-review.md |
| Fix config/GetAccount() pointer to loop variable bug | bug | P1 | infrastructure-review.md |
| Fix polecat/pending.go non-atomic writes | bug | P1 | infrastructure-review.md |

### Medium Priority (Cleanup/Technical Debt)

| Proposed Title | Type | Priority | Source |
|----------------|------|----------|--------|
| Delete internal/keepalive package (100% unused) | task | P2 | infrastructure-review.md |
| Remove 21 unused constants from constants.go | task | P2 | infrastructure-review.md |
| Consolidate atomic write pattern across packages | task | P2 | infrastructure-review.md |
| Add unit tests for checkpoint package (crash recovery) | task | P2 | infrastructure-review.md |

### ScreenCoach Improvements

| Proposed Title | Type | Priority | Source |
|----------------|------|----------|--------|
| Add headless test mode to ScreenCoach.UI for CI | task | P2 | Desktop assessment |
| Publish Edge extension to store for stable ID | task | P2 | Desktop assessment |
| Enhance test data seeding for fully automated setup | task | P2 | Desktop assessment |

---

## 5. Recommendations

1. **Create beads for the 3 P1 bugs** identified in infrastructure review - these affect correctness
2. **Create epic for infrastructure cleanup** to track the 4 cleanup tasks as children
3. **Create beads for 3 ScreenCoach improvements** under screencoach rig
4. **Consider periodic audit** - assessments and reviews should trigger bead creation workflow

---

## 6. Proposed Beads to Create

### Immediate (P1 Bugs)

```
bd create --title="Fix claude/RoleTypeFor() missing deacon and crew mapping" --type=bug --priority=1
bd create --title="Fix config/GetAccount() pointer to loop variable bug" --type=bug --priority=1
bd create --title="Fix polecat/pending.go non-atomic writes" --type=bug --priority=1
```

### Technical Debt Epic

```
bd create --title="(EPIC) Infrastructure Cleanup from Code Review" --type=epic --priority=2
# Children:
bd create --title="Delete internal/keepalive package" --type=task --priority=2 --parent=<epic-id>
bd create --title="Remove unused constants from constants.go" --type=task --priority=2 --parent=<epic-id>
bd create --title="Consolidate atomic write pattern" --type=task --priority=2 --parent=<epic-id>
bd create --title="Add checkpoint package unit tests" --type=task --priority=2 --parent=<epic-id>
```

### ScreenCoach Improvements (in screencoach rig)

```
bd create --title="Add headless test mode to UI for CI" --type=task --priority=2
bd create --title="Publish Edge extension to store for stable ID" --type=task --priority=2
bd create --title="Enhance test data seeding automation" --type=task --priority=2
```

---

*Audit completed by gastown/polecats/morsov, 2026-01-13*
