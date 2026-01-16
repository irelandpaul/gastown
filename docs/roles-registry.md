# Gas Town Roles Registry

Complete reference for all Gas Town agent roles. Each role has specific responsibilities, authority boundaries, and lifecycle commands.

## Quick Reference

| Role | Scope | Persistence | Primary Function |
|------|-------|-------------|------------------|
| **Mayor** | Town | Persistent | Global coordination, convoy creation, escalation handling |
| **Aid** | Town | Persistent | Tactical execution for Mayor's work |
| **Deacon** | Town | Persistent | System health monitoring, watchdog |
| **Librarian** | Town | Persistent | Research and bead enrichment |
| **Witness** | Rig | Persistent | Polecat monitoring and recovery |
| **Refinery** | Rig | Persistent | Merge queue processing |
| **Polecat** | Rig | Ephemeral | Implementation work on specific issues |
| **Planner** | N/A | Stateless | Structured spec planning workflow |

---

## Town-Level Roles

### Mayor

**Role**: Chief-of-staff agent responsible for global coordination across all rigs.

**Responsibilities**:
- Create and manage convoys (batches of related work)
- Distribute work to polecats via `gt sling`
- Handle escalations from other agents
- Coordinate cross-rig communication
- Notify users of important events

**Authority Boundaries**:
- Can sling work to any polecat in any rig
- Can create convoys and assign priorities
- Cannot directly modify code (delegates to polecats)
- Coordinates but does not implement

**Commands**:
```bash
gt mayor start       # Start the Mayor session
gt mayor stop        # Stop the Mayor session
gt mayor attach      # Attach to the Mayor session
gt mayor status      # Check Mayor session status
gt mayor restart     # Restart the Mayor session
```

**Communication**:
- Receives mail via `gt mail`
- Sends nudges via `gt nudge`
- Creates convoys via `gt convoy create`

---

### Aid (Mayor's Aid)

**Role**: Tactical executor that works alongside the Mayor, implementing bead-based work.

**Responsibilities**:
- Execute beads slung by Mayor
- Maintain focused execution context
- Mail Mayor for review when work is complete
- Handle implementation tasks Mayor delegates

**Authority Boundaries**:
- Works only on beads slung by Mayor
- Can compact context freely (focused execution)
- Mails Mayor for review, doesn't self-approve
- Bead-only work constraint

**Commands**:
```bash
gt aid start         # Start the Aid session
gt aid stop          # Stop the Aid session
gt aid attach        # Attach to the Aid session
gt aid status        # Check Aid session status
gt aid restart       # Restart the Aid session
```

**Communication**:
- Receives beads via Mayor's sling
- Mails Mayor when work is complete

---

### Deacon

**Role**: Daemon beacon and hierarchical health-check orchestrator.

**Responsibilities**:
- Monitor Mayor and all Witnesses
- Handle lifecycle requests (spawn, kill agents)
- Maintain system heartbeat
- Trigger recovery for unresponsive agents
- Manage dogs (helper workers)

**Authority Boundaries**:
- Can force-kill unresponsive agents
- Can trigger polecat spawns
- Cannot modify code directly
- Monitors but does not implement

**Commands**:
```bash
gt deacon start              # Start the Deacon session
gt deacon stop               # Stop the Deacon session
gt deacon attach             # Attach to the Deacon session
gt deacon status             # Check Deacon session status
gt deacon restart            # Restart the Deacon session
gt deacon pause              # Pause patrol actions
gt deacon resume             # Resume patrol actions
gt deacon health-check       # Send health ping to an agent
gt deacon health-state       # Show health state for all agents
gt deacon force-kill         # Force-kill an unresponsive agent
gt deacon stale-hooks        # Find and unhook stale beads
gt deacon trigger-pending    # Trigger pending polecat spawns
gt deacon heartbeat          # Update Deacon heartbeat
```

**Communication**:
- Receives heartbeats from all agents
- Sends health check pings
- Triggers recovery actions

---

### Librarian

**Role**: Research-focused agent that enriches beads with context before work begins.

**Responsibilities**:
- Research documentation, prior work, and patterns
- Attach "Required Reading" to beads
- Review completed beads and capture observations
- Synthesize observations into axioms
- Manage skill injection for context

**Authority Boundaries**:
- Enriches beads with context, doesn't implement
- Can comment on beads with research findings
- Cannot modify code directly
- Front-loads context for polecats

**Commands**:
```bash
gt librarian start           # Start (uses Gemini by default)
gt librarian start --agent claude  # Start with Claude
gt librarian stop            # Stop the Librarian session
gt librarian attach          # Attach to the session
gt librarian status          # Check session status
gt librarian restart         # Restart the session
gt librarian enrich <bead>   # Enrich a bead with context
gt librarian review <bead>   # Review completed work
gt librarian summarize       # Synthesize observations
gt librarian skills          # List available skills
gt librarian inject <bead>   # Inject skills into enrichment
gt librarian match <bead>    # Preview matching skills
```

**Communication**:
- Receives enrichment requests
- Comments on beads with Required Reading

---

## Rig-Level Roles

### Witness

**Role**: Patrol agent that oversees polecats within a rig.

**Responsibilities**:
- Monitor polecat health and progress
- Detect stuck or crashed polecats
- Nudge blocked polecats
- Report status to Mayor
- Handle orphaned worktrees

**Authority Boundaries**:
- Monitors polecats within its rig only
- Can nudge polecats to resume work
- Can trigger recovery for stuck agents
- Cannot implement, only monitors

**Commands**:
```bash
gt witness start      # Start the witness for current rig
gt witness stop       # Stop the witness
gt witness attach     # Attach to witness session
gt witness status     # Show witness status
gt witness restart    # Restart the witness
```

**Communication**:
- Monitors polecat sessions
- Reports to Mayor via mail
- Nudges polecats when stuck

---

### Refinery

**Role**: Merge queue processor that handles polecat merge requests.

**Responsibilities**:
- Process merge requests from polecats
- Merge changes to integration branches
- Handle merge conflicts
- Ensure code quality before main
- Manage the merge queue

**Authority Boundaries**:
- Processes MRs within its rig only
- Can merge polecat work to main
- Cannot create new features, only merges
- Claims and releases MRs from queue

**Commands**:
```bash
gt refinery start     # Start the refinery
gt refinery stop      # Stop the refinery
gt refinery attach    # Attach to refinery session
gt refinery status    # Show refinery status
gt refinery restart   # Restart the refinery
gt refinery queue     # Show the merge queue
gt refinery ready     # List MRs ready for processing
gt refinery claim     # Claim an MR for processing
gt refinery release   # Release a claimed MR
gt refinery blocked   # List MRs blocked by open tasks
gt refinery unclaimed # List unclaimed MRs
```

**Communication**:
- Receives MRs from polecats
- Reports merge results

---

### Polecat

**Role**: Ephemeral worker agent that implements specific issues.

**Responsibilities**:
- Implement features, fixes, and tasks
- Work in isolated git worktrees
- Create merge requests when done
- Self-cleanup after completion
- Follow GUPP (work on hook = execute)

**Authority Boundaries**:
- Works only on assigned issues (hooked work)
- Operates in isolated worktree (no conflicts)
- Creates MRs, doesn't merge directly
- Ephemeral - destroyed after work completion

**Commands**:
```bash
gt polecat list           # List polecats in rig
gt polecat status <name>  # Show polecat status
gt polecat nuke <name>    # Destroy polecat completely
gt polecat remove <name>  # Remove polecat from rig
gt polecat stale          # Detect stale polecats
gt polecat gc             # Garbage collect stale branches
gt polecat sync <name>    # Sync beads for polecat
gt polecat identity       # Manage polecat identities
gt polecat git-state      # Show git state for verification
gt polecat check-recovery # Check if recovery needed
```

**Lifecycle**:
- Spawned via `gt sling` with work assignment
- Works in isolated worktree
- Signals completion via `gt done`
- Self-nukes after successful MR

**Communication**:
- Receives work via hook (GUPP)
- Mails when blocked or done
- Signals completion to refinery

---

## Stateless Workflow

### Planner

**Role**: Structured planning workflow for spec creation (not a persistent agent).

**Responsibilities**:
- Guide users through spec planning
- Ask clarifying questions
- Generate proposals from ideas
- Produce final specifications

**Authority Boundaries**:
- Stateless workflow, not persistent agent
- Creates specs, doesn't implement
- Interactive Q&A process

**Commands**:
```bash
gt planner new <idea>     # Start planning session
gt planner status         # Show session status
gt planner answer         # Answer clarifying question
gt planner show           # Show session details
gt planner list           # List all planning sessions
gt planner cancel         # Cancel a session
gt planner --rig <name>   # Target specific rig
```

**Workflow**:
1. Start with `gt planner new "your idea"`
2. Answer clarifying questions
3. Review generated proposal
4. Approve to generate final spec

---

## Role Hierarchy

```
                    ┌─────────────┐
                    │   DEACON    │  (Watchdog)
                    │ Monitors All│
                    └──────┬──────┘
                           │
           ┌───────────────┼───────────────┐
           │               │               │
    ┌──────┴──────┐ ┌──────┴──────┐ ┌──────┴──────┐
    │    MAYOR    │ │  LIBRARIAN  │ │   WITNESS   │
    │ Coordinates │ │ Researches  │ │  Monitors   │
    └──────┬──────┘ └─────────────┘ └──────┬──────┘
           │                               │
    ┌──────┴──────┐                ┌───────┴───────┐
    │     AID     │                │   POLECATS    │
    │  Executes   │                │  Implement    │
    └─────────────┘                └───────┬───────┘
                                           │
                                   ┌───────┴───────┐
                                   │   REFINERY    │
                                   │    Merges     │
                                   └───────────────┘
```

---

## Starting the System

Full startup sequence:

```bash
# 1. Start town-level agents
gt deacon start     # Watchdog first
gt mayor start      # Coordinator
gt aid start        # Mayor's executor
gt librarian start  # Research agent

# 2. Start rig-level agents (in each rig)
gt witness start    # Polecat monitor
gt refinery start   # Merge processor

# 3. Polecats are spawned automatically via gt sling
```

Or use the unified command:
```bash
gt up               # Start all Gas Town services
gt down             # Stop all services
```

---

## See Also

- [architecture.md](design/architecture.md) - Technical architecture details
- [glossary.md](glossary.md) - Gas Town terminology
- [convoy.md](concepts/convoy.md) - Work batching with convoys
- [molecules.md](concepts/molecules.md) - Workflow molecules
