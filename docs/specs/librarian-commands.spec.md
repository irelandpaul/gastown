# Spec: gt librarian Commands

**Task**: hq-qn44
**Author**: Mayor's Aid
**Date**: 2026-01-14
**Status**: Draft

## Overview

This spec defines the command interface for the Librarian agent, including session management and enrichment workflow commands.

---

## 1. Session Management Commands

### gt librarian start

Start the Librarian tmux session.

```bash
gt librarian start [--model <model>]
```

**Flags:**
- `--model`: Model to use (default: gemini)

**Behavior:**
1. Check if Librarian session already exists
2. If exists and running, print "Librarian already running" and exit
3. Create tmux session named `librarian`
4. Set working directory to `<town>/librarian/`
5. Start Claude Code with Librarian context and Gemini model
6. Print "Librarian started. Attach with: gt librarian attach"

**Model Configuration:**
```bash
# Default: Gemini for large context
gt librarian start                    # Uses Gemini

# Override for testing
gt librarian start --model claude     # Uses Claude (more expensive)
```

**Exit Codes:**
- 0: Success
- 1: Already running
- 2: Model not available

**Example:**
```bash
$ gt librarian start
Starting Librarian with Gemini model...
Librarian started. Attach with: gt librarian attach
```

---

### gt librarian attach

Attach to a running Librarian session.

```bash
gt librarian attach
```

**Behavior:**
1. Find Librarian session
2. If not running, start it first
3. Attach to tmux session
4. If already attached elsewhere, show warning

**Exit Codes:**
- 0: Success (attached)
- 1: Session not found and couldn't start

**Example:**
```bash
$ gt librarian attach
Attaching to librarian...
```

---

### gt librarian status

Check Librarian session status.

```bash
gt librarian status [--json]
```

**Flags:**
- `--json`: Output as JSON

**Output (text):**
```
● Librarian session is running
  Model: gemini
  Status: idle
  Current work: none
  Created: 2026-01-14 10:00:00

Attach with: gt librarian attach
```

**Output (json):**
```json
{
  "running": true,
  "model": "gemini",
  "status": "idle",
  "current_work": null,
  "created": "2026-01-14T10:00:00Z"
}
```

**Exit Codes:**
- 0: Running
- 1: Not running

---

### gt librarian restart

Restart the Librarian session.

```bash
gt librarian restart [--model <model>]
```

**Behavior:**
1. Kill existing session if running
2. Wait for clean shutdown
3. Start new session with specified model
4. Print status

**Example:**
```bash
$ gt librarian restart
Stopping librarian...
Starting librarian...
Librarian restarted. Attach with: gt librarian attach
```

---

## 2. Enrichment Commands

### gt librarian enrich

Request context enrichment for a bead.

```bash
gt librarian enrich <bead-id> [flags]
```

**Required:**
- `<bead-id>`: Bead to enrich

**Optional Flags:**
- `--depth <level>`: Research depth (quick, standard, deep) - default: standard
- `--focus <area>`: Focus enrichment on specific area (docs, code, history)
- `--web`: Include web search for external documentation
- `--no-wait`: Don't wait for completion, return immediately

**Behavior:**
1. Verify bead exists and is open
2. Create enrichment request bead (type: `enrichment-request`)
3. Hook request to Librarian
4. Nudge Librarian to start
5. Wait for completion (unless --no-wait)
6. Print enrichment summary

**Output:**
```
Enrichment requested for: gt-abc
  Title: Add user authentication
  Depth: standard
  Focus: all

Librarian researching...

Enrichment complete:
  Files: 5
  Prior beads: 2
  External docs: 3
  Key patterns: 4

Required Reading attached to gt-abc.
```

**Exit Codes:**
- 0: Success
- 1: Bead not found
- 2: Librarian not running
- 3: Enrichment failed

**Examples:**
```bash
# Standard enrichment
$ gt librarian enrich gt-abc

# Quick enrichment (just essentials)
$ gt librarian enrich gt-abc --depth quick

# Deep enrichment with web search
$ gt librarian enrich gt-abc --depth deep --web

# Focus on prior work only
$ gt librarian enrich gt-abc --focus history
```

---

### gt librarian status (enrichment)

Check enrichment status for a bead.

```bash
gt librarian status <bead-id>
```

**Output:**
```
Enrichment Status: gt-abc

  Stage: researching
  Progress:
    ✓ Files searched: 45
    ✓ Prior beads found: 3
    ● Web search: in progress
    ○ Compilation: pending

  Estimated time remaining: ~2 minutes
```

---

## 3. Integration with gt sling

### Enhanced gt sling with --enrich

Add optional enrichment before polecat assignment.

```bash
gt sling <bead-id> <rig> [--enrich] [--enrich-depth <level>]
```

**New Flags:**
- `--enrich`: Enrich bead before slinging
- `--enrich-depth`: Enrichment depth (quick, standard, deep)

**Behavior with --enrich:**
1. Request enrichment from Librarian
2. Wait for enrichment completion
3. Then sling to polecat with enriched context

**Example:**
```bash
# Sling with standard enrichment
$ gt sling gt-abc gastown --enrich
Requesting enrichment for gt-abc...
Enrichment complete (5 files, 2 prior beads).
Slinging to gastown...
Polecat capable assigned.

# Sling with quick enrichment
$ gt sling gt-abc gastown --enrich --enrich-depth quick
```

### Auto-Enrich Configuration

In `<town>/.gt/config.json`:

```json
{
  "librarian": {
    "auto_enrich": false,
    "default_depth": "standard",
    "web_search": false
  }
}
```

When `auto_enrich: true`, all `gt sling` commands automatically enrich first.

---

## 4. bd enrich Command

New beads command for manual enrichment attachment.

### bd enrich

Attach enrichment data to a bead.

```bash
bd enrich <bead-id> [flags]
```

**Flags:**
- `--file <path>`: Read enrichment from markdown file
- `--stdin`: Read enrichment from stdin
- `--clear`: Clear existing enrichment

**Behavior:**
1. Read enrichment content
2. Validate markdown structure
3. Attach to bead's `enrichment` field
4. Update bead timestamp

**Example:**
```bash
# Attach from file
$ bd enrich gt-abc --file enrichment.md

# Attach from stdin
$ echo "## Required Reading\n..." | bd enrich gt-abc --stdin

# Clear enrichment
$ bd enrich gt-abc --clear
```

---

## 5. Viewing Enrichment

### bd show (enhanced)

Show bead with enrichment.

```bash
bd show <bead-id> [--enrichment]
```

**Flags:**
- `--enrichment`: Show full enrichment content (default: summary only)

**Default Output:**
```
? gt-abc · Add user authentication   [● P2 · ENRICHED]
...

Enrichment: 5 files, 2 prior beads, 3 docs
  (use --enrichment to see full Required Reading)
```

**With --enrichment:**
```
? gt-abc · Add user authentication   [● P2 · ENRICHED]
...

## Required Reading

### Files to Read
- `src/auth/handler.go:45-120` - Existing auth handler pattern
- `pkg/jwt/token.go` - JWT implementation
...
```

---

## 6. Configuration

### Town Configuration

In `<town>/.gt/config.json`:

```json
{
  "librarian": {
    "enabled": true,
    "model": "gemini",
    "auto_enrich": false,
    "default_depth": "standard",
    "web_search": false,
    "max_files": 50,
    "max_prior_beads": 10
  }
}
```

### Depth Levels

| Level | Description | Time | Use When |
|-------|-------------|------|----------|
| `quick` | Essential files only | ~30s | Simple tasks, familiar codebase |
| `standard` | Files + prior beads + docs | ~2min | Most tasks |
| `deep` | Comprehensive + web search | ~5min | Complex/unfamiliar tasks |

---

## 7. Directory Structure

```
<town>/
├── librarian/
│   ├── CLAUDE.md           # Librarian context
│   ├── .enrichments/       # Cache of enrichment files
│   │   ├── gt-abc.md
│   │   └── gt-xyz.md
│   └── .research/          # Temporary research artifacts
└── .gt/
    └── config.json         # Librarian config
```

---

## Implementation Notes

### Tmux Session Naming

```
librarian    # Single town-level session
```

### Working Directory

```
<town>/librarian/
```

### Model Selection

Librarian defaults to Gemini via environment variable:
```bash
CLAUDE_MODEL=gemini claude ...
```

### Enrichment Storage

Enrichments stored both:
1. In bead's `enrichment` field (JSONL)
2. As markdown files in `.enrichments/` (for reference)

---

## Testing Checklist

- [ ] `gt librarian start` creates tmux session with Gemini
- [ ] `gt librarian attach` attaches to existing session
- [ ] `gt librarian status` shows correct state
- [ ] `gt librarian enrich` triggers research
- [ ] `gt sling --enrich` enriches before slinging
- [ ] `bd enrich` attaches enrichment to bead
- [ ] `bd show --enrichment` displays full Required Reading
- [ ] Depth levels produce appropriate detail
- [ ] Web search works when enabled
