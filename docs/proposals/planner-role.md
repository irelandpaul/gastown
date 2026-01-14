# Proposal: Planner Role for Gas Town

**Author**: polecat valkyrie
**Task**: hq-c7uo
**Date**: 2026-01-13
**Status**: Draft
**Foundation**: [docs/research/planning-spec-discipline.md](../research/planning-spec-discipline.md)

## Executive Summary

This proposal introduces the **Planner** role - a specialized agent that transforms vague feature requests into well-shaped specifications before work is assigned to polecats. The Planner implements the "shape before build" discipline, dramatically improving autonomous execution quality.

**Core insight**: Polecats execute reliably when given clear specifications. They struggle with ambiguous requirements. The Planner eliminates ambiguity upstream.

---

## Problem Statement

### Current Pain Points

1. **Ambiguous Work Assignment**: Polecats receive vague descriptions like "add user authentication" without clear scope boundaries
2. **Scope Creep**: Without explicit "Out of Scope" definitions, polecats over-engineer or miss requirements
3. **Rework Cycles**: Ambiguity leads to incorrect implementations that require costly rework
4. **Hallucination Risk**: AI agents make assumptions when requirements are unclear, often incorrectly

### The Gap in Current Workflow

```
Current:  Human → [vague idea] → Polecat → [hope for the best]

Proposed: Human → [vague idea] → Planner → [shaped spec] → Polecat → [reliable execution]
```

---

## Proposed Solution

### The Planner Role

| Attribute | Value |
|-----------|-------|
| **Role Type** | Infrastructure (like Witness, Refinery) |
| **Scope** | Per-rig |
| **Lifecycle** | Persistent, Mayor-managed |
| **Identity** | `<rig>/planner` |
| **Location** | `<rig>/planner/` (no worktree needed) |

### Core Responsibilities

1. **Shape Work**: Transform raw ideas into structured specifications
2. **Q&A Facilitation**: Ask clarifying questions to eliminate ambiguity
3. **Scope Definition**: Explicitly define In Scope and Out of Scope
4. **Task Generation**: Break specifications into executable molecules
5. **Handoff**: Prepare work for polecat assignment

---

## Workflow Integration

### The Shaping Pipeline

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Human     │     │   Planner   │     │   Mayor     │     │  Polecats   │
│  (idea)     │────▶│  (shape)    │────▶│  (assign)   │────▶│  (execute)  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
      │                    │                   │                   │
      │                    │                   │                   │
      ▼                    ▼                   ▼                   ▼
  raw-idea.md         SPEC.md            molecule           completion
                    + tasks.md           + sling
```

### Interaction with Existing Roles

| Role | Interaction with Planner |
|------|--------------------------|
| **Mayor** | Dispatches shaping requests to Planner; receives shaped specs |
| **Human/Crew** | Answers Planner's clarifying questions |
| **Witness** | Does not interact (Planner is not a polecat) |
| **Refinery** | Does not interact (Planner produces specs, not code) |
| **Polecats** | Receive shaped work from Mayor after Planner completes |

---

## Bead Types for Planning

### New Bead Types

| Type | Purpose | Example ID |
|------|---------|------------|
| `shaping` | Active shaping session with Q&A | `gt-shape-abc` |
| `spec` | Completed specification | `gt-spec-abc` |

### Shaping Bead Lifecycle

```
pending → shaping → shaped → assigned
   │         │         │         │
   │         │         │         └─ Molecule created, work slung
   │         │         └─ SPEC.md complete, ready for assignment
   │         └─ Planner asking questions, human answering
   └─ Raw idea submitted, awaiting Planner
```

### Bead Structure

```yaml
# Shaping bead (gt-shape-abc)
type: shaping
title: "Add user authentication"
status: shaping
assignee: gastown/planner
artifacts:
  raw_idea: "gt-shape-abc/raw-idea.md"
  requirements: "gt-shape-abc/requirements.md"
  spec: null  # Populated when shaped
  tasks: null
qa_rounds: 2
blocking_questions: 1

# Spec bead (gt-spec-abc)
type: spec
title: "User Authentication System"
status: shaped
parent: gt-shape-abc
artifacts:
  spec: "gt-spec-abc/SPEC.md"
  tasks: "gt-spec-abc/tasks.md"
scope:
  in_scope: ["JWT auth", "refresh tokens", "logout"]
  out_of_scope: ["OAuth", "SSO", "MFA"]
estimated_effort: "L"
task_count: 24
```

---

## Commands

### New Commands

```bash
# Request shaping for a feature
gt shape "Add user authentication" --rig gastown

# Check shaping status
gt shape status gt-shape-abc

# Answer Planner's questions (interactive)
gt shape answer gt-shape-abc

# View shaped spec
gt shape show gt-spec-abc

# Approve spec and generate molecule
gt shape approve gt-spec-abc --convoy "Auth Sprint"

# Cancel shaping
gt shape cancel gt-shape-abc
```

### Integration with Existing Commands

```bash
# Mayor can request shaping before convoy creation
gt convoy create "Auth Sprint" --shape "Add user authentication"

# Sling can require shaped spec
gt sling gt-spec-abc gastown  # Works
gt sling gt-raw-idea gastown  # Error: "Work must be shaped first"
```

---

## Document Artifacts

### Directory Structure

```
<rig>/.specs/
└── gt-shape-abc/
    ├── raw-idea.md       # Initial concept
    ├── requirements.md   # Q&A and shaped requirements
    ├── SPEC.md          # Full specification
    └── tasks.md         # Task breakdown
```

### Document Formats

Planner produces documents following the codesetup discipline:

1. **raw-idea.md**: Captures initial feature concept
2. **requirements.md**: Q&A rounds eliminating ambiguity
3. **SPEC.md**: Complete technical specification with:
   - Goal statement
   - User stories
   - Specific requirements
   - Out of Scope (critical)
   - Technical architecture
4. **tasks.md**: Executable breakdown with:
   - Task groups with dependencies
   - Complexity ratings (S/M/L/XL)
   - Acceptance criteria
   - Tests-first ordering

---

## Q&A Protocol

### How Planner Asks Questions

1. Planner reads raw idea and existing codebase
2. Identifies ambiguities and decision points
3. Posts questions to shaping bead
4. Human/Crew receives notification
5. Human answers via `gt shape answer` or mail
6. Planner incorporates answers, may ask follow-ups
7. When no blocking questions remain, generates SPEC.md

### Question Categories

| Category | Example |
|----------|---------|
| **Scope** | "Should this include OAuth integration or just username/password?" |
| **Behavior** | "What should happen when a token expires mid-session?" |
| **Edge Cases** | "How should the system handle concurrent logins?" |
| **Existing Patterns** | "I found auth code in `lib/auth.go`. Should I follow this pattern?" |
| **Trade-offs** | "JWT is simpler but sessions offer revocation. Which is preferred?" |

### Escalation

If human doesn't respond within configurable timeout:
1. Planner nudges via mail
2. After second timeout, escalates to Mayor
3. Mayor can assign different human or defer shaping

---

## Molecule Generation

### From tasks.md to Molecule

Planner generates molecules from tasks.md:

```toml
# Auto-generated from gt-spec-abc/tasks.md
description = "User Authentication System"
formula = "shaped-spec"
version = 1
source_spec = "gt-spec-abc"

[[steps]]
id = "setup"
title = "Project Infrastructure"
description = "Initialize auth module structure"
complexity = "S"

[[steps]]
id = "models"
title = "Data Models"
description = "Create User, Session, Token models"
needs = ["setup"]
complexity = "M"

[[steps]]
id = "tests-models"
title = "Model Tests"
description = "Write tests for data models"
needs = ["models"]
complexity = "S"

# ... continues for all task groups
```

### Automatic Dependency Resolution

- Task groups become molecule steps
- Dependencies from tasks.md become `needs` in molecule
- Complexity ratings preserved for estimation

---

## Configuration

### Rig-Level Settings

```json
{
  "planner": {
    "enabled": true,
    "require_shaping": false,  // If true, all work must be shaped
    "qa_timeout": "24h",       // Time to wait for human answers
    "max_qa_rounds": 5,        // Maximum Q&A iterations
    "auto_approve_trivial": true  // Auto-approve S-complexity specs
  }
}
```

### Per-Convoy Settings

```bash
# Convoy that requires shaped work
gt convoy create "Q1 Features" --require-shaped

# Convoy that accepts unshaped work
gt convoy create "Bug Fixes" --allow-unshaped
```

---

## Implementation Phases

### Phase 1: Core Infrastructure
- [ ] Add `shaping` and `spec` bead types
- [ ] Create Planner agent scaffolding
- [ ] Implement `gt shape` command family
- [ ] Add `.specs/` directory management

### Phase 2: Q&A System
- [ ] Implement question posting to beads
- [ ] Add `gt shape answer` interactive mode
- [ ] Create mail integration for notifications
- [ ] Add timeout and escalation logic

### Phase 3: Document Generation
- [ ] Implement raw-idea.md template
- [ ] Implement requirements.md with Q&A formatting
- [ ] Implement SPEC.md generation
- [ ] Implement tasks.md generation

### Phase 4: Molecule Integration
- [ ] Create molecule generator from tasks.md
- [ ] Add `gt shape approve` workflow
- [ ] Integrate with `gt sling` for shaped work
- [ ] Add `--require-shaped` convoy option

### Phase 5: Polish
- [ ] Add Planner role to `gt prime` output
- [ ] Dashboard integration for shaping status
- [ ] Metrics: shaping time, rework reduction
- [ ] Documentation and examples

---

## Success Metrics

| Metric | Baseline | Target |
|--------|----------|--------|
| Polecat rework rate | ~30% | <10% |
| Scope creep incidents | Common | Rare |
| Time to first working code | Variable | Predictable |
| Human clarification interrupts | Frequent | Front-loaded |

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Shaping becomes bottleneck | Auto-approve trivial work; parallel shaping |
| Over-specification | Appetite-based shaping; cut scope, not quality |
| Human doesn't answer questions | Timeout escalation; sensible defaults |
| Planner asks too many questions | Limit Q&A rounds; batch related questions |

---

## Alternatives Considered

### 1. Inline Shaping by Mayor
**Rejected**: Mayor already coordinates; adding shaping overloads the role.

### 2. Shaping by Polecats
**Rejected**: Polecats are execution-focused; shaping requires different skills (questioning vs. building).

### 3. Human-Only Shaping
**Rejected**: AI can identify ambiguities humans miss; hybrid approach better.

### 4. No Shaping (Status Quo)
**Rejected**: Current rework rates and scope creep justify investment.

---

## Conclusion

The Planner role addresses a fundamental gap in Gas Town's workflow: the translation of vague ideas into actionable specifications. By implementing the "shape before build" discipline:

1. **Polecats receive clear work** - No more guessing at requirements
2. **Scope is explicit** - Out of Scope prevents over-engineering
3. **Questions are front-loaded** - Interruptions happen before coding, not during
4. **Quality improves** - Shaped specs lead to better outcomes

The Planner complements existing roles without replacing them:
- Mayor dispatches and coordinates
- Planner shapes and specifies
- Polecats execute shaped work
- Refinery merges quality code

**Recommendation**: Approve for Phase 1 implementation to validate the core concept before full buildout.

---

*Proposal authored by polecat valkyrie, 2026-01-13*
