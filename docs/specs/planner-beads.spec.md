# Spec: Planner Bead Types & Lifecycle

**Task**: hq-xmev
**Author**: Mayor's Aid
**Date**: 2026-01-14
**Status**: Draft

## Overview

This spec defines the new bead types for the Planner workflow and their state machine transitions.

---

## 1. New Bead Types

### shaping

Active shaping session tracking Q&A and proposal development.

```yaml
type: shaping
title: "Add user authentication"
status: planning  # See lifecycle below
owner: mayor
assignee: <rig>/planner
priority: P2

# Shaping-specific fields
stage: planning | proposal | review | approved | spec | handoff | cancelled
qa_rounds: 2
blocking_questions: 1
reviewers:
  - pm
  - developer
  - security  # optional
  - ralph     # optional
artifacts:
  raw_idea: ".specs/gt-shape-abc/raw-idea.md"
  requirements: ".specs/gt-shape-abc/requirements.md"
  proposal: ".specs/gt-shape-abc/proposal.md"
  spec: null  # populated when generated
  tasks: null
review_verdicts:
  pm: APPROVE | REVISE | REJECT | null
  developer: APPROVE | REVISE | REJECT | null
  security: APPROVE | REVISE | REJECT | null
  ralph: null  # Ralph doesn't give formal verdicts
```

### proposal

Standalone proposal artifact (optional - can be embedded in shaping bead).

```yaml
type: proposal
title: "User Authentication System"
status: draft | review | approved | rejected
parent: gt-shape-abc  # Link to shaping bead
owner: <rig>/planner

# Proposal-specific fields
scope:
  in_scope:
    - "JWT-based authentication"
    - "Token refresh mechanism"
    - "Logout functionality"
  out_of_scope:
    - "OAuth integration"
    - "SSO"
    - "MFA"
complexity: S | M | L | XL
estimated_tasks: 12
```

### spec

Approved specification ready for implementation.

```yaml
type: spec
title: "User Authentication System Specification"
status: generated | handed_off | implemented
parent: gt-shape-abc  # Link to shaping bead
owner: mayor
assignee: null  # Mayor assigns to polecats

# Spec-specific fields
artifacts:
  spec: ".specs/gt-shape-abc/SPEC.md"
  tasks: ".specs/gt-shape-abc/tasks.md"
task_count: 12
complexity_breakdown:
  S: 4
  M: 5
  L: 2
  XL: 1
implementation_beads: []  # Populated by Mayor
```

---

## 2. Shaping Bead Lifecycle

### State Machine

```
                    ┌──────────────────────────────────────────┐
                    │                                          │
                    ▼                                          │
┌─────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐   │
│ pending │───▶│ planning │───▶│ proposal │───▶│  review  │───┤ (revise)
└─────────┘    └──────────┘    └──────────┘    └──────────┘   │
                    │               │               │          │
                    │               │               │          │
                    ▼               ▼               ▼          │
               ┌─────────────────────────────────────────┐    │
               │              cancelled                   │    │
               └─────────────────────────────────────────┘    │
                                                              │
                                                    ┌─────────┘
                                                    │
                                                    ▼
                                              ┌──────────┐
                                              │ approved │
                                              └──────────┘
                                                    │
                                                    ▼
                                              ┌──────────┐
                                              │   spec   │
                                              └──────────┘
                                                    │
                                                    ▼
                                              ┌──────────┐
                                              │ handoff  │
                                              └──────────┘
```

### Stage Descriptions

| Stage | Actor | Actions | Next Trigger |
|-------|-------|---------|--------------|
| **pending** | System | Bead created, awaiting Planner | Planner picks up work |
| **planning** | Planner + Human | Q&A rounds, codebase exploration | All questions answered |
| **proposal** | Planner | Generate proposal.md with scope | Proposal complete |
| **review** | Review Agents | PM, Dev, optional reviewers | All reviews returned |
| **approved** | Human | Human approves via `gt shape approve` | Approval command |
| **spec** | Planner | Generate SPEC.md and tasks.md | Spec complete |
| **handoff** | Planner → Mayor | Mail Mayor, await bead creation | Mayor acknowledges |
| **cancelled** | Any | Can cancel from any stage | Cancel command |

### Transitions

```python
TRANSITIONS = {
    "pending": ["planning", "cancelled"],
    "planning": ["proposal", "cancelled"],
    "proposal": ["review", "cancelled"],
    "review": ["proposal", "approved", "cancelled"],  # Can return to proposal if REVISE
    "approved": ["spec", "cancelled"],
    "spec": ["handoff", "cancelled"],
    "handoff": ["closed"],  # Terminal state
    "cancelled": []  # Terminal state
}
```

### Status vs Stage

- **status**: Standard bead status (open, in_progress, closed)
- **stage**: Shaping-specific workflow position

Mapping:
```python
STAGE_TO_STATUS = {
    "pending": "open",
    "planning": "in_progress",
    "proposal": "in_progress",
    "review": "in_progress",
    "approved": "in_progress",
    "spec": "in_progress",
    "handoff": "in_progress",
    "cancelled": "closed"
}
```

---

## 3. Bead Commands Integration

### Creating Shaping Beads

```bash
# Via gt shape command (preferred)
gt shape "Add feature X" --rig gastown

# Direct bd command (internal use)
bd new "Add feature X" --type=shaping --assignee=gastown/planner
```

### Updating Stage

```bash
# Planner updates stage
bd update gt-shape-abc --stage=proposal

# Custom field updates via bd
bd update gt-shape-abc --field qa_rounds=3
bd update gt-shape-abc --field "review_verdicts.pm=APPROVE"
```

### Querying by Stage

```bash
# Find all beads in review stage
bd list --type=shaping --stage=review

# Find beads awaiting approval
bd list --type=shaping --stage=review --field "review_verdicts.pm!=null"
```

---

## 4. Artifact Management

### Directory Structure

```
<rig>/.specs/
├── <shape-id>/
│   ├── raw-idea.md
│   ├── requirements.md
│   ├── proposal.md
│   ├── reviews/
│   │   ├── pm-review.md
│   │   ├── dev-review.md
│   │   ├── security-review.md
│   │   └── ralph-review.md
│   ├── SPEC.md
│   └── tasks.md
├── <shape-id-2>/
│   └── ...
└── templates/
    ├── raw-idea.template.md
    ├── proposal.template.md
    ├── spec.template.md
    └── tasks.template.md
```

### Artifact Sync

Artifacts are stored in filesystem, paths stored in bead:

```yaml
artifacts:
  raw_idea: ".specs/gt-shape-abc/raw-idea.md"
  requirements: ".specs/gt-shape-abc/requirements.md"
  # ...
```

When bead is read, artifacts can be loaded:
```bash
bd show gt-shape-abc --include-artifacts
```

---

## 5. Templates

### raw-idea.template.md

```markdown
# Raw Idea: {title}

**Created**: {timestamp}
**Requester**: {requester}
**Rig**: {rig}

## Description

{description}

## Initial Context

{any additional context provided}

## Questions (for Planner)

<!-- Planner will add questions here -->
```

### proposal.template.md

```markdown
# Proposal: {title}

**Shape ID**: {shape_id}
**Date**: {timestamp}
**Status**: Draft

## Summary

{one paragraph summary}

## Problem Statement

{what problem does this solve}

## Proposed Solution

{high-level solution description}

## In Scope

- {item 1}
- {item 2}
- {item 3}

## Out of Scope

- {explicitly excluded item 1}
- {explicitly excluded item 2}

## Complexity Assessment

**Overall**: {S|M|L|XL}

| Component | Complexity | Notes |
|-----------|------------|-------|
| {component} | {S|M|L|XL} | {notes} |

## Dependencies

- {dependency 1}
- {dependency 2}

## Open Questions

- {any remaining questions for human}
```

### spec.template.md

```markdown
# Specification: {title}

**Spec ID**: {spec_id}
**Shape ID**: {shape_id}
**Date**: {timestamp}
**Approved By**: {approver}

## Overview

{detailed description of what will be built}

## Requirements

### Functional Requirements

1. **{FR-1}**: {description}
2. **{FR-2}**: {description}

### Non-Functional Requirements

1. **{NFR-1}**: {description}

## Technical Design

### Architecture

{architecture description}

### Components

| Component | Purpose | Notes |
|-----------|---------|-------|
| {name} | {purpose} | {notes} |

### Data Models

{data model descriptions}

### API Contracts

{API specifications}

## Acceptance Criteria

- [ ] {criterion 1}
- [ ] {criterion 2}

## Testing Strategy

{how to verify this works}

## Rollout Plan

{how to deploy safely}
```

### tasks.template.md

```markdown
# Task Breakdown: {title}

**Spec ID**: {spec_id}
**Total Tasks**: {count}
**Complexity Distribution**: S:{s} M:{m} L:{l} XL:{xl}

## Task Groups

### Group 1: {group_name}

| ID | Task | Complexity | Depends On | Acceptance |
|----|------|------------|------------|------------|
| T1 | {task} | S | - | {criteria} |
| T2 | {task} | M | T1 | {criteria} |

### Group 2: {group_name}

| ID | Task | Complexity | Depends On | Acceptance |
|----|------|------------|------------|------------|
| T3 | {task} | L | T2 | {criteria} |

## Dependency Graph

```
T1 → T2 → T3
         ↘
           T4
```

## Implementation Notes

{any notes for implementing polecats}
```

---

## 6. Integration Points

### With Beads System

- New types: `shaping`, `proposal`, `spec`
- New fields: `stage`, `qa_rounds`, `reviewers`, `artifacts`, `review_verdicts`
- Query support: `--type=shaping`, `--stage=review`

### With Mail System

- Stage transitions trigger notifications
- Approval requests via mail
- Handoff notifications to Mayor

### With gt sling

- `require_shaping` config option
- Spec beads can be slung to polecats
- Shaping beads auto-sling to Planner

### With Convoys

- Shaping beads can be part of convoys
- Convoy waits for spec before proceeding
- Convoy tracks shaping progress

---

## Implementation Notes

### Database Schema Changes

```sql
-- New columns for beads table
ALTER TABLE beads ADD COLUMN stage TEXT;
ALTER TABLE beads ADD COLUMN qa_rounds INTEGER DEFAULT 0;
ALTER TABLE beads ADD COLUMN reviewers TEXT;  -- JSON array
ALTER TABLE beads ADD COLUMN artifacts TEXT;  -- JSON object
ALTER TABLE beads ADD COLUMN review_verdicts TEXT;  -- JSON object

-- Index for stage queries
CREATE INDEX idx_beads_stage ON beads(stage) WHERE type IN ('shaping', 'proposal', 'spec');
```

### JSONL Format

```jsonl
{"id":"gt-shape-abc","type":"shaping","stage":"planning","qa_rounds":0,"reviewers":["pm","developer"],...}
```

---

## Testing Checklist

- [ ] Shaping bead creation with all fields
- [ ] Stage transitions follow state machine
- [ ] Invalid transitions are rejected
- [ ] Artifact paths are created and tracked
- [ ] Query by stage works
- [ ] Templates render correctly
- [ ] Integration with gt shape commands
- [ ] Mail notifications on stage change
