# Proposal: Planner Role for Gas Town (v2)

**Author**: polecat valkyrie
**Task**: hq-c7uo, hq-cyvm
**Date**: 2026-01-14
**Version**: 2.0
**Status**: Draft
**Foundation**: [docs/research/planning-spec-discipline.md](../research/planning-spec-discipline.md)

## Executive Summary

This proposal introduces the **Planner** role - a specialized agent that transforms vague feature requests into well-shaped specifications through a structured review process before work is assigned to polecats.

**v2 Changes**:
- Added `gt planner attach` command for interactive planning sessions
- Built-in review agents (PM, Developer, Security, Ralph Wiggum)
- Revised workflow: Planning → Proposal → Review → Approval → Spec → Handoff
- Clear authority boundaries (Planner creates planning beads only, messages Mayor when ready)

---

## Problem Statement

### Current Pain Points

1. **Ambiguous Work Assignment**: Polecats receive vague descriptions without clear scope
2. **Missing Perspectives**: Technical, business, and security concerns discovered late
3. **Scope Creep**: Without explicit boundaries, work expands unpredictably
4. **Rework Cycles**: Ambiguity leads to incorrect implementations

### The Gap in Current Workflow

```
Current:  Human → [vague idea] → Polecat → [hope for the best]

Proposed: Human → [idea] → Planner → [multi-perspective review] → Mayor → [shaped spec] → Polecat
```

---

## The Planner Role

### Role Definition

| Attribute | Value |
|-----------|-------|
| **Role Type** | Infrastructure (like Witness, Refinery) |
| **Scope** | Per-rig |
| **Lifecycle** | Persistent, Mayor-managed |
| **Identity** | `<rig>/planner` |
| **Location** | `<rig>/planner/` (no worktree needed) |

### Authority Boundaries

The Planner has **limited authority** by design:

| Can Do | Cannot Do |
|--------|-----------|
| Create `planning` type beads | Create `task`, `bug`, `feature` beads |
| Create `proposal` type beads | Create molecules |
| Create `spec` type beads | Sling work to polecats |
| Send mail to Mayor | Sling work directly |
| Request human input | Approve its own specs |
| Invoke review agents | Bypass Mayor for handoff |

**Key Constraint**: When a spec is ready, Planner **messages Mayor** for approval and handoff. Planner cannot directly dispatch work.

---

## Revised Workflow

### The Six-Phase Pipeline

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ PLANNING │──▶│ PROPOSAL │──▶│  REVIEW  │──▶│ APPROVAL │──▶│   SPEC   │──▶│ HANDOFF  │
└──────────┘   └──────────┘   └──────────┘   └──────────┘   └──────────┘   └──────────┘
     │              │              │              │              │              │
     ▼              ▼              ▼              ▼              ▼              ▼
 raw-idea.md   proposal.md    reviews/      human sign-off   SPEC.md      Mayor slings
               + questions    (4 agents)                    + tasks.md    to polecats
```

### Phase Details

#### Phase 1: Planning
- Human submits raw idea via `gt planner attach` or mail
- Planner creates `planning` bead
- Planner asks clarifying questions
- Human answers via interactive session or mail

#### Phase 2: Proposal
- Planner synthesizes answers into `proposal.md`
- Includes: goal, user stories, scope boundaries, technical approach
- Creates `proposal` bead linked to planning bead

#### Phase 3: Review
- Planner invokes **four built-in review agents**
- Each agent reviews from their perspective
- Reviews stored in `reviews/` directory
- Planner synthesizes feedback

#### Phase 4: Approval
- Planner presents proposal + reviews to human
- Human approves, requests changes, or rejects
- May require multiple iterations

#### Phase 5: Spec
- Upon approval, Planner generates full `SPEC.md`
- Generates `tasks.md` with task breakdown
- Creates `spec` bead with all artifacts

#### Phase 6: Handoff
- Planner **mails Mayor** that spec is ready
- Mayor reviews and approves for execution
- Mayor creates molecule and slings to polecats
- Planner's job is done

---

## Built-in Review Agents

### Overview

The Planner includes four specialized review personas that evaluate proposals from different perspectives:

```
                    ┌─────────────────┐
                    │    Proposal     │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
    ┌─────────┐        ┌─────────┐        ┌─────────┐
    │   PM    │        │Developer│        │Security │
    │ Review  │        │ Review  │        │ Review  │
    └─────────┘        └─────────┘        └─────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  Ralph Wiggum   │
                    │    Review       │
                    └─────────────────┘
```

### Agent Definitions

#### 1. PM Review Agent

**Persona**: Product Manager focused on business value and user needs

**Evaluates**:
- Business value and ROI
- User story completeness
- Market fit and timing
- Priority relative to other work
- Success metrics definition

**Key Questions**:
- "Who specifically benefits from this?"
- "How will we measure success?"
- "What's the cost of NOT doing this?"
- "Does this align with current priorities?"

**Output**: `reviews/pm-review.md`

```markdown
# PM Review: [Feature Name]

## Business Value Assessment
[High/Medium/Low] - [Reasoning]

## User Impact
- Primary users affected: [list]
- Expected behavior change: [description]

## Priority Recommendation
[P0/P1/P2/P3] - [Justification]

## Concerns
- [Concern 1]
- [Concern 2]

## Approval: [APPROVE/NEEDS_CHANGES/REJECT]
```

---

#### 2. Developer Review Agent

**Persona**: Senior engineer focused on technical feasibility and architecture

**Evaluates**:
- Technical feasibility
- Architecture impact
- Complexity estimation (S/M/L/XL)
- Existing code patterns to leverage
- Technical debt implications
- Testing requirements

**Key Questions**:
- "How does this fit with existing architecture?"
- "What's the blast radius if this fails?"
- "Are there simpler alternatives?"
- "What technical debt does this create or resolve?"

**Output**: `reviews/developer-review.md`

```markdown
# Developer Review: [Feature Name]

## Technical Feasibility
[Feasible/Challenging/Risky] - [Reasoning]

## Complexity Estimate
[S/M/L/XL] - [Breakdown]

## Architecture Impact
- Files affected: [estimate]
- New dependencies: [list]
- Breaking changes: [yes/no + details]

## Implementation Approach
[Recommended approach]

## Existing Patterns
- Leverage: [file/pattern references]
- Avoid: [anti-patterns identified]

## Testing Requirements
- Unit tests: [scope]
- Integration tests: [scope]
- E2E tests: [scope]

## Concerns
- [Technical concern 1]
- [Technical concern 2]

## Approval: [APPROVE/NEEDS_CHANGES/REJECT]
```

---

#### 3. Security Review Agent

**Persona**: Security engineer focused on vulnerabilities and data protection

**Evaluates**:
- Authentication/authorization implications
- Data handling and privacy
- Input validation requirements
- Potential attack vectors
- Compliance considerations
- Secrets management

**Key Questions**:
- "What data does this expose or collect?"
- "How could this be abused?"
- "What's the worst case if compromised?"
- "Does this need security review before deploy?"

**Output**: `reviews/security-review.md`

```markdown
# Security Review: [Feature Name]

## Risk Level
[Critical/High/Medium/Low] - [Reasoning]

## Data Handling
- PII involved: [yes/no + types]
- Data storage: [where/how]
- Data transmission: [encrypted/plain]

## Authentication/Authorization
- Auth changes required: [yes/no]
- New permissions: [list]
- Privilege escalation risk: [assessment]

## Attack Surface
- New endpoints: [list]
- Input vectors: [list]
- Potential vulnerabilities: [list]

## Required Mitigations
- [ ] [Mitigation 1]
- [ ] [Mitigation 2]

## Compliance Notes
[GDPR/SOC2/etc. implications]

## Approval: [APPROVE/NEEDS_CHANGES/BLOCK]
```

---

#### 4. Ralph Wiggum Review Agent

**Persona**: The naive questioner who asks "dumb" questions that reveal hidden assumptions

> "I'm helping!" - Ralph Wiggum

**Purpose**: Surface implicit assumptions by asking obvious questions that experts skip over. Named after the Simpsons character whose naive observations often accidentally reveal truth.

**Evaluates**:
- Unstated assumptions
- Jargon that hides complexity
- "Obviously true" statements that aren't
- Edge cases everyone forgot
- User confusion points

**Key Questions**:
- "What does [jargon term] actually mean?"
- "Why can't we just [naive alternative]?"
- "What happens if [unlikely but possible scenario]?"
- "My cat's breath smells like cat food" (occasionally)

**Output**: `reviews/ralph-review.md`

```markdown
# Ralph Wiggum Review: [Feature Name]

## Confused About
- [Thing that seems obvious but isn't explained]
- [Jargon term that needs definition]
- [Assumption that might be wrong]

## Dumb Questions
1. "Why do we need [feature] at all?"
2. "What if the user does [unexpected thing]?"
3. "Couldn't we just [simpler alternative]?"

## Things That Seem Weird
- [Observation that might reveal a flaw]
- [Inconsistency noticed]

## Hidden Assumptions Found
- Assumes: [assumption 1]
- Assumes: [assumption 2]

## My Suggestion
[Often surprisingly insightful or hilariously off-base]

## Approval: [APPROVE/CONFUSED/NEEDS_CRAYONS]
```

---

## Commands

### Primary Command: `gt planner attach`

Start an interactive planning session:

```bash
# Attach to Planner for interactive session
gt planner attach

# Attach with a specific idea to shape
gt planner attach --idea "Add user authentication"

# Attach to continue existing planning session
gt planner attach --resume gt-plan-abc
```

**In the Planner session**:
```
🎯 Planner Session Active

Commands:
  /idea <description>    Submit a new idea
  /answer <response>     Answer pending question
  /status                Show current planning status
  /reviews               Trigger review agents
  /approve               Mark proposal as approved (human only)
  /spec                  Generate final spec
  /handoff               Send to Mayor for execution
  /quit                  Exit session

Current: Awaiting new idea or /status to check existing work
```

### Supporting Commands

```bash
# Check planning status
gt planner status [planning-bead-id]

# List active planning sessions
gt planner list

# View reviews for a proposal
gt planner reviews gt-plan-abc

# Cancel planning session
gt planner cancel gt-plan-abc
```

### Integration with Existing Commands

```bash
# Mayor can request planning
gt mail send planner/ --subject "New feature request" --message "Add OAuth support"

# Planner notifies Mayor when ready
# (Automatic: Planner mails Mayor at handoff)

# Mayor reviews and slings
gt sling gt-spec-abc gastown
```

---

## Bead Types for Planning

### New Bead Types

| Type | Purpose | Created By | Example ID |
|------|---------|------------|------------|
| `planning` | Active planning session | Planner | `gt-plan-abc` |
| `proposal` | Draft proposal under review | Planner | `gt-prop-abc` |
| `spec` | Approved specification | Planner | `gt-spec-abc` |

### Bead Lifecycle

```
         Planner creates           Planner creates           Planner creates
               │                         │                         │
               ▼                         ▼                         ▼
┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│    planning      │───▶│    proposal      │───▶│      spec        │
│   (gt-plan-*)    │    │   (gt-prop-*)    │    │   (gt-spec-*)    │
└──────────────────┘    └──────────────────┘    └──────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
    raw-idea.md            proposal.md              SPEC.md
    questions.md           reviews/*               tasks.md
```

### Bead Structure

```yaml
# Planning bead
id: gt-plan-abc
type: planning
title: "Add user authentication"
status: questioning  # questioning | reviewing | approved | handed_off
created_by: planner
artifacts:
  raw_idea: "planning/raw-idea.md"
  questions: "planning/questions.md"
  answers: "planning/answers.md"

# Proposal bead
id: gt-prop-abc
type: proposal
title: "User Authentication System"
status: reviewing  # draft | reviewing | approved | rejected
parent: gt-plan-abc
created_by: planner
artifacts:
  proposal: "proposal/proposal.md"
  reviews:
    pm: "proposal/reviews/pm-review.md"
    developer: "proposal/reviews/developer-review.md"
    security: "proposal/reviews/security-review.md"
    ralph: "proposal/reviews/ralph-review.md"
review_status:
  pm: approved
  developer: needs_changes
  security: approved
  ralph: confused

# Spec bead
id: gt-spec-abc
type: spec
title: "User Authentication System"
status: ready  # draft | ready | handed_off
parent: gt-prop-abc
created_by: planner
artifacts:
  spec: "spec/SPEC.md"
  tasks: "spec/tasks.md"
scope:
  in_scope: ["JWT auth", "refresh tokens", "logout"]
  out_of_scope: ["OAuth", "SSO", "MFA"]
complexity: "L"
task_count: 24
```

---

## Directory Structure

```
<rig>/.planning/
└── gt-plan-abc/
    ├── planning/
    │   ├── raw-idea.md
    │   ├── questions.md
    │   └── answers.md
    ├── proposal/
    │   ├── proposal.md
    │   └── reviews/
    │       ├── pm-review.md
    │       ├── developer-review.md
    │       ├── security-review.md
    │       └── ralph-review.md
    └── spec/
        ├── SPEC.md
        └── tasks.md
```

---

## Interaction with Existing Roles

| Role | Interaction |
|------|-------------|
| **Mayor** | Receives handoff mail from Planner; approves and slings specs |
| **Human/Crew** | Submits ideas; answers questions; approves proposals |
| **Witness** | Does not interact (Planner is not a polecat) |
| **Refinery** | Does not interact (Planner produces specs, not code) |
| **Polecats** | Receive shaped work from Mayor after Planner handoff |

### Handoff Protocol

```
1. Planner completes spec
2. Planner creates spec bead (gt-spec-abc)
3. Planner sends mail to Mayor:

   To: mayor/
   Subject: Spec Ready: User Authentication System (gt-spec-abc)

   Spec gt-spec-abc is ready for execution.

   Summary: JWT-based auth with refresh tokens
   Complexity: L (24 tasks)
   Reviews: PM ✓, Developer ✓, Security ✓, Ralph 🖍️

   Artifacts:
   - SPEC.md: gastown/.planning/gt-plan-abc/spec/SPEC.md
   - tasks.md: gastown/.planning/gt-plan-abc/spec/tasks.md

   Recommended: Create convoy and sling to 2-3 polecats

4. Mayor reviews and slings
5. Polecats execute
```

---

## Configuration

### Rig-Level Settings

```json
{
  "planner": {
    "enabled": true,
    "require_planning": false,
    "review_agents": ["pm", "developer", "security", "ralph"],
    "auto_reviews": true,
    "approval_required": true,
    "handoff_to": "mayor/"
  }
}
```

### Review Agent Configuration

```json
{
  "review_agents": {
    "pm": {
      "enabled": true,
      "blocking": false
    },
    "developer": {
      "enabled": true,
      "blocking": true
    },
    "security": {
      "enabled": true,
      "blocking": true
    },
    "ralph": {
      "enabled": true,
      "blocking": false
    }
  }
}
```

---

## Implementation Phases

### Phase 1: Core Infrastructure
- [ ] Add `planning`, `proposal`, `spec` bead types
- [ ] Create Planner agent scaffolding
- [ ] Implement `gt planner attach` command
- [ ] Add `.planning/` directory management
- [ ] Implement authority restrictions

### Phase 2: Planning Flow
- [ ] Implement idea intake
- [ ] Implement Q&A system
- [ ] Create proposal generation
- [ ] Add mail integration for handoff

### Phase 3: Review Agents
- [ ] Implement PM review agent
- [ ] Implement Developer review agent
- [ ] Implement Security review agent
- [ ] Implement Ralph Wiggum review agent
- [ ] Create review synthesis

### Phase 4: Approval & Handoff
- [ ] Implement approval workflow
- [ ] Add human approval commands
- [ ] Implement spec generation
- [ ] Create Mayor handoff protocol

### Phase 5: Polish
- [ ] Add Planner to `gt prime` output
- [ ] Dashboard integration
- [ ] Metrics collection
- [ ] Documentation

---

## Success Metrics

| Metric | Baseline | Target |
|--------|----------|--------|
| Polecat rework rate | ~30% | <10% |
| Security issues in prod | Variable | 0 critical |
| Scope creep incidents | Common | Rare |
| Time from idea to spec | N/A | <24h typical |
| Review coverage | 0% | 100% |

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Planning becomes bottleneck | Parallel planning sessions; fast-track for small changes |
| Review agents disagree | Human approval required; synthesis weights |
| Over-specification | Appetite-based planning; cut scope not quality |
| Ralph derails process | Ralph is non-blocking; provides signal not veto |

---

## Conclusion

The Planner role with built-in review agents creates a structured pipeline from idea to execution:

1. **Multi-perspective review** catches issues early
2. **Clear authority boundaries** prevent overreach
3. **Mayor handoff** maintains coordination hierarchy
4. **Ralph Wiggum** surfaces hidden assumptions

The Planner enhances Gas Town's existing roles:
- Mayor remains the coordinator
- Planner becomes the shaper
- Polecats remain the executors
- Reviews become systematic

**Recommendation**: Approve for phased implementation starting with core infrastructure and `gt planner attach`.

---

*Proposal v2 authored by polecat valkyrie, 2026-01-14*
