# Using the Planner

> **Status**: Feature planned - not yet implemented. This documents the intended usage.

## Overview

The **Planner** is a specialized Gas Town agent that transforms vague feature ideas into well-shaped specifications before work is assigned to polecats. It implements the "shape before build" discipline.

**Core insight**: Polecats execute reliably when given clear specifications. They struggle with ambiguous requirements. The Planner eliminates ambiguity upstream.

## When to Use the Planner

Use the Planner when:
- Starting a new feature with unclear scope
- The work involves multiple components or polecats
- You want multi-perspective review (PM, Developer, Security)
- You need explicit In Scope / Out of Scope boundaries

Skip the Planner for:
- Bug fixes with clear reproduction steps
- Single-file changes with obvious implementation
- Refactoring with well-defined scope

## Quick Start

```bash
# Start an interactive planning session
gt planner attach

# Submit an idea for shaping
gt shape "Add user authentication with JWT" --rig gastown

# Check status of a shaping request
gt shape status gt-shape-abc

# Answer Planner's questions
gt shape answer gt-shape-abc

# Approve a proposal
gt shape approve gt-shape-abc
```

---

## The Shaping Workflow

```
PLANNING → PROPOSAL → REVIEW → APPROVAL → SPEC → HANDOFF
```

### Stage 1: Planning

You submit an idea. Planner asks clarifying questions.

```bash
gt shape "Add user authentication" --rig gastown
# Creates: gt-shape-abc
# Planner starts asking questions
```

**Question categories:**
| Category | Example |
|----------|---------|
| Scope | "Should this include OAuth or just username/password?" |
| Behavior | "What should happen when a token expires?" |
| Edge Cases | "How should we handle concurrent login?" |
| Trade-offs | "Option A is simpler but B is more flexible. Preference?" |

### Stage 2: Proposal

After Q&A, Planner generates a proposal with:
- Summary of the feature
- Explicit **In Scope** items
- Explicit **Out of Scope** items
- Complexity estimate (S/M/L/XL)
- Dependencies

```bash
# View the proposal
gt shape show gt-shape-abc
```

### Stage 3: Review

Planner spawns review agents that evaluate from different perspectives:

| Reviewer | Focus | Verdict |
|----------|-------|---------|
| **PM** | Business value, user impact | APPROVE / REVISE / REJECT |
| **Developer** | Technical feasibility, architecture | APPROVE / REVISE / REJECT |
| **Security** | Attack vectors, data handling | APPROVE / REVISE / BLOCK |
| **Ralph Wiggum** | Hidden assumptions, "dumb" questions | APPROVE / CONFUSED / NEEDS_CRAYONS |

```bash
# Request additional reviewers
gt shape review gt-shape-abc --security --ralph
```

### Stage 4: Approval

You review the proposal and all reviews, then approve:

```bash
# Approve and proceed to spec generation
gt shape approve gt-shape-abc --message "Looks good, proceed"
```

### Stage 5: Spec

Planner generates detailed artifacts:
- `SPEC.md` - Full specification with requirements, design, acceptance criteria
- `tasks.md` - Task breakdown with complexity estimates

### Stage 6: Handoff

Planner mails Mayor. Mayor creates implementation beads and slings to polecats.

---

## Commands Reference

### Session Management

```bash
# Start Planner session
gt planner start [--rig <rig>]

# Attach to running Planner
gt planner attach [--rig <rig>]

# Check Planner status
gt planner status [--rig <rig>] [--json]

# Restart Planner session
gt planner restart [--rig <rig>]
```

### Shaping Commands

```bash
# Create shaping request
gt shape "<description>" --rig <rig> [--security] [--ralph] [--expedite]

# Check shaping status
gt shape status <shape-id> [--json]

# Answer Planner's questions
gt shape answer <shape-id>

# View shaping artifacts
gt shape show <shape-id> [--stage <stage>]

# Approve proposal
gt shape approve <shape-id> [--message "<note>"]

# Request additional reviewers
gt shape review <shape-id> [--security] [--ralph]

# Cancel shaping request
gt shape cancel <shape-id> [--reason "<reason>"]
```

---

## Artifacts

Each shaping session creates a directory of artifacts:

```
<rig>/.specs/<shape-id>/
├── raw-idea.md           # Initial concept
├── requirements.md       # Q&A rounds and clarifications
├── proposal.md           # Shaped proposal with scope
├── reviews/
│   ├── pm-review.md      # Product Manager review
│   ├── dev-review.md     # Developer review
│   ├── security-review.md # Security review (if requested)
│   └── ralph-review.md   # Ralph Wiggum review (if requested)
├── SPEC.md               # Final specification
└── tasks.md              # Task breakdown
```

---

## Review Agents

### PM Review Agent

Evaluates business value and user needs:
- Who specifically benefits from this?
- How will we measure success?
- Does this align with current priorities?

### Developer Review Agent

Evaluates technical feasibility:
- How does this fit with existing architecture?
- What's the blast radius if this fails?
- What technical debt does this create or resolve?

### Security Review Agent

Evaluates security implications:
- What data does this expose or collect?
- How could this be abused?
- Does this need security review before deploy?

### Ralph Wiggum Review Agent

Surfaces hidden assumptions:
- What does [jargon term] actually mean?
- Why can't we just [naive alternative]?
- What if [unlikely but possible scenario]?

> "I'm helping!" - Ralph Wiggum

Ralph asks "dumb" questions that reveal implicit assumptions experts skip over.

---

## Example Workflow

```bash
# 1. Submit an idea
$ gt shape "Add user authentication with JWT" --rig gastown --security
Created shaping request: gt-shape-abc
  Rig: gastown
  Reviewers: PM, Developer, Security

# 2. Planner asks questions (check mail or attach)
$ gt mail inbox
  ● [QUESTIONS] gt-shape-abc: 3 questions (from planner)

# 3. Answer questions
$ gt shape answer gt-shape-abc
Questions for gt-shape-abc: Add user authentication

Question 1 of 3 [Scope]:
Should this include OAuth integration or just username/password?
> Just username/password for now

# 4. Review proposal
$ gt shape show gt-shape-abc
# Proposal: Add User Authentication
...

# 5. Check reviews
$ gt shape status gt-shape-abc
✓ PM Review: Complete (APPROVE)
✓ Developer Review: Complete (APPROVE)
✓ Security Review: Complete (APPROVE with notes)

# 6. Approve
$ gt shape approve gt-shape-abc
Approved: gt-shape-abc
Planner will now generate detailed specification.

# 7. Mayor receives handoff and slings to polecats
```

---

## Configuration

### Rig Settings

In `<rig>/.gt/config.json`:

```json
{
  "planner": {
    "enabled": true,
    "require_shaping": false,
    "default_reviewers": ["pm", "developer"],
    "qa_timeout": "24h",
    "max_qa_rounds": 5
  }
}
```

### Requiring Shaping

When `require_shaping: true`, unshaped work cannot be slung:

```bash
$ gt sling raw-idea-abc gastown
Error: Work must be shaped first.

Create shaping request:
  gt shape "description" --rig gastown

Or to bypass (not recommended):
  gt sling raw-idea-abc gastown --allow-unshaped
```

---

## Tips

1. **Be specific in your initial idea** - More detail = fewer Q&A rounds
2. **Request security review for auth/data features** - Use `--security` flag
3. **Don't skip Ralph** - His "dumb" questions often reveal real issues
4. **Iterate quickly** - Answer questions promptly to avoid timeout
5. **Trust the process** - Upfront shaping saves rework later

---

## Bead Types

The Planner creates three new bead types:

| Type | Purpose | Example ID |
|------|---------|------------|
| `shaping` | Active shaping session | `gt-shape-abc` |
| `proposal` | Draft proposal under review | Part of shaping |
| `spec` | Approved specification ready for implementation | `gt-spec-abc` |

---

## See Also

- [Planner Role Proposal](../proposals/planner-role.md) - Design rationale
- [Planner Commands Spec](../specs/planner-commands.spec.md) - Full command reference
- [Planner Beads Spec](../specs/planner-beads.spec.md) - Bead types and lifecycle
