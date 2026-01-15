# PM Review: AI User Testing Proposal

**Reviewer**: Product Manager Perspective
**Date**: 2026-01-14
**Task**: hq-ndap
**Status**: Review

## Executive Summary

The AI User Testing proposal introduces a system where Claude embodies user personas to test ScreenCoach applications and identify UX friction. The approach is **innovative and high-value** but needs tighter Phase 1 scope and clearer success metrics.

**Verdict**: **APPROVE with modifications** - Strong foundation, needs scope refinement before implementation.

---

## Strengths

### 1. Solves a Real Problem

Traditional QA catches broken functionality but misses UX issues. This proposal addresses the gap:

| Current State | Proposed State |
|---------------|----------------|
| Manual UX testing (expensive, slow) | Automated persona-based testing |
| Scripted E2E tests (no UX insight) | Natural navigation with observations |
| Few test iterations per release | Many test iterations (parallel) |

**Business value**: Faster release cycles with higher UX confidence.

### 2. Excellent Artifact Strategy

The mandatory recording requirement is excellent PM thinking:

- **Video**: Human reviewers can see exact confusion moments
- **Traces**: Developers can debug issues without reproduction
- **Screenshots**: Documentation for UX issue tickets
- **observations.json**: Structured data for tracking/metrics

This creates **accountability and evidence** - crucial for UX improvements.

### 3. Persona Variety

The example personas cover a good spectrum:

| Persona | Archetype | Value |
|---------|-----------|-------|
| Sarah (low tech) | Majority of real parents | Core use case |
| Mike (high tech) | Power users who notice details | Edge case finder |
| Rose (very low tech) | Accessibility edge case | Error message quality |
| Emma (mobile-first) | Mobile UX validator | Platform coverage |

This diversity will catch issues that single-type testing misses.

### 4. Gas Town Integration

Smart reuse of existing infrastructure:

- Beads for issue tracking
- Convoys for batch coordination
- Polecat model (spawn, execute, complete)
- Haiku default (cost-conscious)

No new infrastructure required - immediate leverage.

---

## Concerns

### 1. Scope Creep Risk: Phase 1 Too Broad

The proposal mentions three targets in Phase 1:

| Target | Tool | Complexity |
|--------|------|------------|
| Parent Portal (Web) | Playwright MCP | Low |
| Desktop Client | MCPControl | High |
| Browser Extension | Playwright on Windows | Medium |

**Recommendation**: Phase 1 should be **Parent Portal only**. Desktop/Extension should wait for Phase 2.

Rationale:
- Windows VM automation adds operational complexity
- MCPControl is new, unproven in this context
- Web testing via Playwright is well-understood
- Proves concept with minimal variables

### 2. Success Metrics Need Specificity

Current metrics are directional but not measurable:

| Proposed Metric | Issue |
|-----------------|-------|
| "10+ per test cycle" | What's a cycle? Per sprint? Per release? |
| "Hours" for regression detection | Baseline unclear |
| "80% of critical flows" | What defines critical? |

**Recommendation**: Define specific measurable targets:
- "Find X new UX issues per bi-weekly regression run"
- "Reduce mean time to identify UX regression from Y days to Z hours"
- "Cover registration, onboarding, and dashboard flows (3 of 5 critical)"

### 3. False Positive Risk

AI acting as users may produce observations that aren't real UX issues:

- Agent misunderstands context human would have
- Agent moves too fast/slow compared to real user
- Agent notices things users wouldn't (or vice versa)

**Recommendation**: Add calibration step:
1. Run scenarios manually with real users first
2. Compare AI observations to human observations
3. Tune persona definitions based on delta

### 4. Persona Authenticity Uncertainty

The personas are reasonable archetypes, but:

- Are they based on actual ScreenCoach user research?
- Do they match the real user distribution?
- Who validated these personas?

**Recommendation**:
- Document persona sources (user interviews, analytics, assumptions)
- Tag personas with confidence level (researched vs. assumed)
- Plan persona refinement based on real user feedback

### 5. Observation Triage

The proposal doesn't address observation volume management:

- What if 50 observations per run?
- Who triages AI observations?
- How to prevent alert fatigue?

**Recommendation**: Add severity/confidence levels:
- `P0`: Blocking (user cannot complete goal)
- `P1`: Significant friction (user likely to abandon)
- `P2`: Minor friction (noticeable but not blocking)
- `P3`: Nitpick (improvement opportunity)

Agent should self-classify. Humans only review P0/P1 initially.

---

## Recommendations

### R1: Narrow Phase 1 Scope

```
Phase 1: Parent Portal Web Testing Only
- 3-5 critical path scenarios
- Single persona (Sarah) to start
- Playwright MCP only
- Validate concept before expanding
```

### R2: Define Concrete Success Criteria

Before implementation, establish:

1. **Baseline**: Manual UX testing findings for same scenarios
2. **Target**: AI finds 80% of baseline issues + bonus discoveries
3. **Timeline**: 2-week validation period with daily runs

### R3: Add Calibration Phase

Before declaring Phase 1 complete:

1. Run same scenarios with 3 real users (think-aloud)
2. Compare observations (human vs AI)
3. Calculate correlation score
4. Tune persona if correlation < 70%

### R4: Implement Severity Classification

Add to scenario format:

```yaml
observation_config:
  require_severity: true
  auto_triage:
    P0_P1: create_bead
    P2_P3: log_only
```

### R5: Persona Provenance

Document in each persona file:

```yaml
provenance:
  source: user_research | analytics | assumed
  confidence: high | medium | low
  last_validated: 2026-01-14
  validation_notes: Based on support ticket analysis
```

---

## Phase 1 Scope Recommendation

### Must Have (MVP)

- [ ] `gt tester run` for single scenario
- [ ] Parent Portal target only
- [ ] Sarah persona only
- [ ] Video + screenshot recording
- [ ] observations.json output
- [ ] Basic bead integration

### Should Have (Phase 1 Complete)

- [ ] `gt tester batch` for multiple scenarios
- [ ] 3 personas (Sarah, Mike, Emma)
- [ ] Playwright trace recording
- [ ] Convoy integration
- [ ] Severity classification

### Won't Have (Phase 2+)

- Desktop client testing (MCPControl)
- Browser extension testing
- Scheduled/automated runs
- Comparison with previous runs
- Dashboard integration

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| AI observations don't match real users | Medium | High | Calibration phase |
| Token costs exceed budget | Low | Medium | Haiku default, chunking |
| Flaky tests from timing | Medium | Low | Playwright auto-wait |
| Alert fatigue from volume | Medium | Medium | Severity filtering |
| MCPControl VM issues | High | Low | Defer to Phase 2 |

---

## Verdict

**APPROVE with modifications**

The proposal addresses a genuine gap in QA coverage. The polecat-model execution and comprehensive recording strategy are well-designed.

**Conditions for approval**:

1. Narrow Phase 1 to Parent Portal only
2. Start with single persona (Sarah) for initial validation
3. Add calibration step before declaring Phase 1 success
4. Define measurable success metrics before implementation begins

**Estimated business value**: High. If calibration shows AI observations correlate with real user friction, this becomes a scalable UX quality gate for every release.

---

*Review by PM perspective, 2026-01-14 (hq-ndap)*
