# Spec: gt planner Commands

**Task**: hq-xmev
**Author**: Mayor's Aid
**Date**: 2026-01-14
**Status**: Draft

## Overview

This spec defines the command interface for the Planner agent, including session management and shaping workflow commands.

---

## 1. Session Management Commands

### gt planner start

Start the Planner tmux session.

```bash
gt planner start [--rig <rig>]
```

**Flags:**
- `--rig`: Rig to start Planner for (default: current rig or gastown)

**Behavior:**
1. Check if Planner session already exists
2. If exists and running, print "Planner already running" and exit
3. Create tmux session named `<rig>-planner`
4. Set working directory to `<rig>/planner/`
5. Start Claude Code with Planner context
6. Print "Planner started. Attach with: gt planner attach"

**Exit Codes:**
- 0: Success
- 1: Already running
- 2: Rig not found

**Example:**
```bash
$ gt planner start --rig gastown
Starting Planner for gastown...
Planner started. Attach with: gt planner attach
```

---

### gt planner attach

Attach to a running Planner session.

```bash
gt planner attach [--rig <rig>]
```

**Flags:**
- `--rig`: Rig whose Planner to attach to

**Behavior:**
1. Find Planner session for rig
2. If not running, start it first
3. Attach to tmux session
4. If already attached elsewhere, show warning

**Exit Codes:**
- 0: Success (attached)
- 1: Session not found and couldn't start

**Example:**
```bash
$ gt planner attach
Attaching to gastown-planner...
```

---

### gt planner status

Check Planner session status.

```bash
gt planner status [--rig <rig>] [--json]
```

**Flags:**
- `--rig`: Rig to check
- `--json`: Output as JSON

**Output (text):**
```
● Planner session is running
  Rig: gastown
  Session: gastown-planner
  Status: attached
  Current work: gt-shape-abc (planning stage)
  Created: 2026-01-14 10:00:00

Attach with: gt planner attach
```

**Output (json):**
```json
{
  "running": true,
  "rig": "gastown",
  "session": "gastown-planner",
  "status": "attached",
  "current_work": "gt-shape-abc",
  "stage": "planning",
  "created": "2026-01-14T10:00:00Z"
}
```

**Exit Codes:**
- 0: Running
- 1: Not running

---

### gt planner restart

Restart the Planner session.

```bash
gt planner restart [--rig <rig>]
```

**Behavior:**
1. Kill existing session if running
2. Wait for clean shutdown
3. Start new session
4. Print status

**Example:**
```bash
$ gt planner restart
Stopping gastown-planner...
Starting gastown-planner...
Planner restarted. Attach with: gt planner attach
```

---

## 2. Shaping Commands

### gt shape

Request shaping for a feature.

```bash
gt shape "<description>" --rig <rig> [flags]
```

**Required:**
- `<description>`: Feature description (quoted string)
- `--rig`: Target rig

**Optional Flags:**
- `--security`: Request Security reviewer
- `--ralph`: Request Ralph Wiggum reviewer
- `--expedite`: Mark as urgent (prioritize reviews)
- `--parent <id>`: Parent epic/feature bead

**Behavior:**
1. Create shaping bead with type `shaping`
2. Create `.specs/<shape-id>/` directory
3. Write `raw-idea.md` with description
4. Hook work to Planner
5. Nudge Planner to start

**Output:**
```
Created shaping request: gt-shape-abc
  Description: Add user authentication
  Rig: gastown
  Reviewers: PM, Developer, Security

Work slung to Planner. Check status: gt shape status gt-shape-abc
```

**Exit Codes:**
- 0: Success
- 1: Invalid rig
- 2: Planner not running (will start)

**Example:**
```bash
$ gt shape "Add user authentication with JWT" --rig gastown --security
Created shaping request: gt-shape-abc
...
```

---

### gt shape status

Check status of a shaping request.

```bash
gt shape status <shape-id> [--json]
```

**Output (text):**
```
Shaping: gt-shape-abc
  Title: Add user authentication with JWT
  Stage: review

  ✓ Planning: Complete (2 Q&A rounds)
  ✓ Proposal: Complete
  ● Review: In progress
    - PM Review: Complete (APPROVE)
    - Developer Review: Complete (REVISE)
    - Security Review: Pending...
  ○ Approval: Waiting
  ○ Spec: Not started
  ○ Handoff: Not started

Artifacts:
  .specs/gt-shape-abc/
  ├── raw-idea.md ✓
  ├── requirements.md ✓
  ├── proposal.md ✓
  └── reviews/
      ├── pm-review.md ✓
      ├── dev-review.md ✓
      └── security-review.md (pending)
```

**Exit Codes:**
- 0: Found
- 1: Not found

---

### gt shape answer

Answer Planner's questions interactively.

```bash
gt shape answer <shape-id>
```

**Behavior:**
1. Read pending questions from shaping bead
2. Display questions one at a time
3. Collect answers (stdin or editor)
4. Save answers to requirements.md
5. Update bead status
6. Nudge Planner to continue

**Interactive Flow:**
```
Questions for gt-shape-abc: Add user authentication

Question 1 of 3 [Scope]:
Should this include OAuth integration or just username/password?
Current assumption: username/password only

Your answer (or 'skip' to leave for later):
> Just username/password for now. OAuth is out of scope.

Question 2 of 3 [Behavior]:
What should happen when a token expires mid-session?
...
```

**Exit Codes:**
- 0: All questions answered
- 1: Partial (some skipped)
- 2: No pending questions

---

### gt shape show

View shaping artifacts.

```bash
gt shape show <shape-id> [--stage <stage>]
```

**Flags:**
- `--stage`: Show specific stage (raw-idea, requirements, proposal, reviews, spec)

**Default Behavior:**
- Shows proposal.md if exists
- Falls back to most recent artifact

**Example:**
```bash
$ gt shape show gt-shape-abc
# Proposal: Add User Authentication

## Summary
Add JWT-based authentication for the API...

## In Scope
- Username/password login
- JWT token generation
- Token refresh
...
```

---

### gt shape approve

Approve a proposal and proceed to spec generation.

```bash
gt shape approve <shape-id> [--message "<note>"]
```

**Behavior:**
1. Verify proposal exists and reviews complete
2. Update bead status to `approved`
3. Create `spec` bead as child
4. Nudge Planner to generate SPEC.md

**Output:**
```
Approved: gt-shape-abc
  Reviews: PM(APPROVE), Developer(REVISE→resolved), Security(APPROVE)

Planner will now generate detailed specification.
Track progress: gt shape status gt-shape-abc
```

**Exit Codes:**
- 0: Approved
- 1: Reviews not complete
- 2: Proposal not ready

---

### gt shape review

Request additional reviewers.

```bash
gt shape review <shape-id> [--security] [--ralph]
```

**Behavior:**
1. Add requested reviewers to bead
2. Nudge Planner to spawn review agents
3. Updates status to `review`

**Example:**
```bash
$ gt shape review gt-shape-abc --ralph
Added Ralph Wiggum reviewer to gt-shape-abc
Planner will run additional review.
```

---

### gt shape cancel

Cancel a shaping request.

```bash
gt shape cancel <shape-id> [--reason "<reason>"]
```

**Behavior:**
1. Update bead status to `cancelled`
2. Archive artifacts (don't delete)
3. Remove from Planner's hook

**Output:**
```
Cancelled: gt-shape-abc
  Reason: Deprioritized
  Artifacts preserved in: .specs/gt-shape-abc/
```

---

## 3. Integration with Existing Commands

### gt sling (enhanced)

When `require_shaping` is enabled:

```bash
$ gt sling raw-idea-abc gastown
Error: Work must be shaped first.

Create shaping request:
  gt shape "description" --rig gastown

Or to bypass (not recommended):
  gt sling raw-idea-abc gastown --allow-unshaped
```

When slinging a spec bead:

```bash
$ gt sling gt-spec-abc gastown
Slinging shaped spec to polecat...
  Spec: gt-spec-abc (Add user authentication)
  Tasks: 12 (from tasks.md)
  Complexity: L

Polecat capable assigned. Work hooked.
```

---

### gt convoy create (enhanced)

```bash
gt convoy create "Auth Sprint" --shape "Add user authentication"
```

**Behavior:**
1. Create convoy bead
2. Create shaping request as child
3. Convoy waits for spec completion
4. Once spec ready, Mayor creates implementation beads

---

## 4. Configuration

### Rig Configuration

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

### Command-Line Overrides

```bash
# Temporarily disable shaping requirement
gt sling <id> <rig> --allow-unshaped

# Force expedited reviews
gt shape "..." --rig gastown --expedite
```

---

## Implementation Notes

### Tmux Session Naming

```
<rig>-planner    # e.g., gastown-planner
```

### Working Directory

```
<town>/<rig>/planner/
```

### Specs Directory

```
<rig>/.specs/<shape-id>/
```

### Bead Prefixes

Shaping beads use rig prefix:
- `gt-shape-abc` (gastown)
- `sc-shape-xyz` (screencoach)

---

## Testing Checklist

- [ ] `gt planner start` creates tmux session
- [ ] `gt planner attach` attaches to existing session
- [ ] `gt planner status` shows correct state
- [ ] `gt shape` creates shaping bead and directories
- [ ] `gt shape status` shows pipeline progress
- [ ] `gt shape answer` handles interactive Q&A
- [ ] `gt shape approve` transitions to spec stage
- [ ] `gt shape cancel` preserves artifacts
- [ ] `gt sling` enforces shaping requirement when configured
