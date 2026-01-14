# Research: Strategic Planning and Spec Writing Discipline

**Source**: User's codesetup repo (https://github.com/irelandpaul/codesetup)
**Analyzed From**: `/home/paul-ireland/screencoach/` implementation
**Task**: hq-cthl
**Date**: 2026-01-13

## Executive Summary

The codesetup repo implements a disciplined approach to AI-driven development through structured spec writing. The core philosophy: **"Shape before you build"** - investing upfront in clear requirements prevents expensive rework and AI hallucination.

The system uses three commands (`/aos:shape-spec`, `/aos:write-spec`, `/aos:create-tasks`) to transform raw ideas into executable task lists through a rigorous refinement process.

---

## The AOS Spec Workflow

### Overview

```
Raw Idea → Shape Spec → Write Spec → Create Tasks → Execute
    |           |            |             |
    v           v            v             v
raw-idea.md  requirements.md  SPEC.md    tasks.md
```

### Command Flow

| Command | Input | Output | Purpose |
|---------|-------|--------|---------|
| `/aos:shape-spec` | Feature idea | raw-idea.md + requirements.md | Requirements discussion with Q&A |
| `/aos:write-spec` | requirements.md | SPEC.md | Full technical specification |
| `/aos:create-tasks` | SPEC.md | tasks.md | Executable task breakdown |

---

## Document Formats

### 1. raw-idea.md

Captures the initial feature concept in structured format.

```markdown
# Raw Idea: [Feature Name]

**Date**: YYYY-MM-DD
**Feature Name**: [Name]
**Project**: [Project Name]

## Description
[1-2 sentence summary]

## Context
[Why this feature, what problem it solves]

## Next Steps
This raw idea will be refined into detailed requirements...
```

**Purpose**: Establish a clear starting point before refinement begins.

---

### 2. requirements.md

The critical "shaping" document where ambiguity is eliminated through Q&A.

```markdown
# Spec Requirements: [Feature Name]

## Initial Description
[Original description from raw-idea.md]

## Requirements Discussion

### First Round Questions
**Q1:** [Specific question about ambiguous requirement]
**Answer:** [Human's answer]

**Q2:** [Another question]
**Answer:** [Answer with reasoning]

### Existing Code to Reference
[Files that contain similar patterns to leverage]

### Follow-up Questions
[Second round of questions based on initial answers]

## Visual Assets
### Files Provided:
[List of mockups, diagrams, screenshots]

### Visual Insights:
[Key observations from visual assets]

## Requirements Summary

### Functional Requirements
[Bulleted list of what the system must do]

### Reusability Opportunities
[Existing code/patterns to leverage]

### Scope Boundaries
**In Scope:**
- [Feature A]
- [Feature B]

**Out of Scope:**
- [Deferred Feature X]
- [Future Enhancement Y]

### Technical Considerations
[Architecture decisions, technology choices]
```

**Key Insight**: The Q&A format forces explicit decisions on ambiguous requirements. This prevents the AI from making assumptions that may be wrong.

---

### 3. SPEC.md

The comprehensive technical specification with all implementation details.

```markdown
# Specification: [Feature Name]

## Goal
[1-2 sentence goal statement]

## User Stories
- As a [role], I want [feature] so that [benefit]
- As the system, I want [capability] so that [outcome]

## Specific Requirements
[Detailed technical requirements organized by subsystem]

**[Subsystem Name]**
- Requirement 1 with specific details
- Requirement 2 with implementation guidance
- [Continue for all requirements]

## Visual Design
[Reference to mockups, wireframes, or design assets]

## Existing Code to Leverage
[Specific files and what to extract from them]

## Out of Scope
[Explicit list of what this spec does NOT include]

---

## Technical Architecture

### System Diagram
[ASCII or Mermaid diagram showing components]

### Component Breakdown
[Detailed component specifications]

### Data Models
[Schema definitions with field specifications]

### API Endpoints
[Endpoint specifications with request/response formats]
```

**Key Pattern**: The "Out of Scope" section is critical - it explicitly bounds the work and prevents scope creep during implementation.

---

### 4. tasks.md

Executable task breakdown with dependencies and acceptance criteria.

```markdown
# Task Breakdown: [Feature Name]

## Overview
**Total Estimated Tasks:** [N] subtasks across [M] task groups
**Estimated Total Effort:** [Size] ([Time estimate])

## Execution Order Summary
[ASCII diagram showing phase dependencies]

---

## Task Groups

### Task Group 1: [Group Name]
**Dependencies:** [None | Task Group X]
**Complexity:** [S|M|L|XL]
**Specialist:** [Role needed]

- [ ] 1.0 [Parent task description]
  - [ ] 1.1 [Subtask with implementation details]
    - Specific guidance line 1
    - Specific guidance line 2
  - [ ] 1.2 [Next subtask]
  - [ ] 1.3 [Test task - tests written FIRST]

**Acceptance Criteria:**
- [Measurable outcome 1]
- [Measurable outcome 2]

---

### Task Group 2: [Group Name]
**Dependencies:** Task Group 1
...
```

**Key Patterns**:
1. **Tests First**: Test tasks appear early in each group (TDD approach)
2. **Dependencies Explicit**: Clear execution order prevents parallelization mistakes
3. **Acceptance Criteria**: Measurable outcomes for each group
4. **Complexity Ratings**: S/M/L/XL for estimation

---

## The "Shape Up" Philosophy

The spec workflow implements Ryan Singer's "Shape Up" methodology adapted for AI development:

### 1. Appetite, Not Estimates
- Define how much time you're willing to invest (the "appetite")
- Shape the work to fit the appetite
- Cut scope rather than extend timeline

### 2. Fixed Time, Variable Scope
- Tasks.md shows total effort estimates
- Out of Scope lists features cut to fit appetite
- Prevents endless feature creep

### 3. Shaping Before Building
- Requirements.md Q&A process is the "shaping"
- Removes ambiguity before implementation
- AI cannot make good decisions on ambiguous requirements

### 4. Bets, Not Backlogs
- Each spec is a discrete "bet" on a feature
- Complete implementation or nothing
- No half-finished features accumulating

---

## AI Development Implications

The CLAUDE.md in codesetup emphasizes:

> "This project is 100% AI-developed. All code is written by Claude or other AI assistants."

### Why Specs Matter More for AI

| Challenge | Spec Solution |
|-----------|---------------|
| AI hallucination | Q&A eliminates ambiguity |
| Context limits | Structured docs fit context windows |
| Inconsistency | SPEC.md is single source of truth |
| Scope creep | Out of Scope explicitly bounds work |
| Rework cycles | Shape before build prevents redo |

### Architecture Principles for AI

From CLAUDE.md:
- **Type safety**: Types catch AI mistakes at compile time
- **Single language per context**: Less context switching = fewer AI errors
- **Established patterns**: More training data = better AI output
- **Simple over clever**: AI generates clearer straightforward code
- **Explicit over implicit**: AI can follow explicit patterns

---

## Quick Commands Reference

| Command | Purpose |
|---------|---------|
| `/cook [task]` | Build feature (execute with context) |
| `/cook:auto:fast [task]` | Automated build (minimal interaction) |
| `/plan [task]` | Plan feature (research phase) |
| `/fix [issue]` | Fix bug |
| `/scout [query]` | Find files |
| `/ask [question]` | Technical question |
| `/aos:shape-spec` | Start spec process with Q&A |
| `/aos:write-spec` | Write full specification |
| `/aos:create-tasks` | Generate task breakdown |
| `/git:cm` | Commit |
| `/git:cp` | Commit and push |

---

## Document Organization

```
project/
├── docs/               # Product documentation
├── specs/              # Implementation specs
│   └── YYYY-MM-DD-feature-name/
│       └── planning/
│           ├── raw-idea.md
│           ├── requirements.md
│           ├── SPEC.md
│           └── tasks.md
├── investigations/     # Internal codebase analysis
├── research/           # External research
└── plans/              # Active planning work
```

---

## Recommendations for Gas Town

### Potential Adoption

1. **Spec-First Convoys**: Require SPEC.md before convoy creation
2. **Shape Beads**: Add `shaping` bead type for requirements Q&A
3. **Task Molecules**: Generate molecules from tasks.md automatically
4. **Out of Scope Enforcement**: Prevent polecat scope creep via spec bounds

### Integration Points

| Codesetup Concept | Gas Town Equivalent |
|-------------------|---------------------|
| raw-idea.md | Issue/bead creation |
| requirements.md | Requirements discussion (could be bead thread) |
| SPEC.md | Could be pinned reference bead |
| tasks.md | Molecule steps |
| /aos:create-tasks | `gt mol pour` from spec |

### Philosophical Alignment

Both systems share:
- **Explicit over implicit**: Gas Town's propulsion principle requires clear work
- **Bounded scope**: Out of Scope aligns with discrete work units
- **Accountability**: Specs create paper trail like capability ledger
- **Test-first**: Tasks.md puts tests early like quality-focused polecats

---

## Conclusion

The codesetup spec discipline provides a rigorous framework for AI-driven development:

1. **Shape before build** - Q&A eliminates ambiguity
2. **Explicit scope** - Out of Scope prevents creep
3. **Tests first** - Quality built into task structure
4. **Phased execution** - Dependencies prevent chaos
5. **Single source of truth** - SPEC.md is the contract

For Gas Town, the key insight is that **structured pre-work (shaping) dramatically improves autonomous execution**. A polecat with a well-shaped SPEC.md will execute far more reliably than one given a vague description.

---

*Research completed by polecat valkyrie, 2026-01-13*
