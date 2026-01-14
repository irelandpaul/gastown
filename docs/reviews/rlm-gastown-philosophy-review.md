# Review: RLM + Librarian Architecture vs Gas Town Philosophy

**Reviewer**: polecat valkyrie
**Task**: hq-qndw.1
**Date**: 2026-01-13

## Summary

This review evaluates the proposed "RLM + Librarian Knowledge Graph Architecture" against Gas Town's core philosophical principles. The proposal aims to transform Beads from a task tracker into a unified knowledge graph with execution capabilities, incorporating MIT's Recursive Language Model (RLM) patterns and a new "Librarian" agent role.

**Verdict**: The architecture contains valuable concepts but introduces tension with Gas Town's core philosophy. Recommend **partial adoption** with significant modifications.

---

## 1. RLM vs The Propulsion Principle

### The Propulsion Principle

> "Gas Town is a steam engine. Agents are pistons. When an agent finds work on their hook, they EXECUTE."

Key characteristics:
- **Immediate execution** - No waiting, no confirmation
- **Work on hook = assignment** - The hook IS the contract
- **Discrete units** - Clear start and end points
- **Autonomous momentum** - Agent drives forward independently

### RLM's Approach

RLM treats context as an **external environment** that the model interacts with through a REPL. The model:
- Loads prompts as variables
- Writes code to decompose and explore context
- Spawns sub-LLM instances recursively
- Discovers the answer through iterative exploration

### Tension Points

| Propulsion Principle | RLM Pattern |
|---------------------|-------------|
| Execute immediately | Explore iteratively |
| Clear work unit (hook) | Emergent decomposition |
| Piston fires once | Recursive self-invocation |
| Work survives restart | Per-query, stateless |

**Core conflict**: RLM's exploratory, recursive nature conflicts with Gas Town's "fire and forget" piston model. A polecat with RLM-style execution would be "exploring" rather than "executing" - a fundamental mismatch.

### Compatibility Path

RLM could align with propulsion if used **within** a discrete task, not as the task orchestration model:
- Hook assigns: "Analyze this 500KB log file"
- Polecat uses RLM-style REPL interaction internally
- Polecat completes, submits to merge queue

**Verdict**: RLM as internal technique = OK. RLM as orchestration model = conflicts with propulsion.

---

## 2. Librarian vs Existing Roles

### Proposed Librarian Role

The proposal introduces a "Librarian" agent that:
- Summarizes knowledge into "Axioms" (patterns/truths)
- Maintains the knowledge graph
- Provides context to other agents

### Existing Role Analysis

| Role | Purpose | Lifecycle |
|------|---------|-----------|
| **Witness** | Monitors polecat health, handles nudges/cleanup | Persistent, per-rig |
| **Refinery** | Processes merge queue, verification | Persistent, per-rig |
| **Mayor** | Global coordination, cross-rig communication | Persistent, singleton |
| **Deacon** | Daemon beacon, heartbeats, monitoring | Persistent, singleton |

### Alignment Assessment

**Positive alignment**:
- Like Witness, Librarian is a background, persistent role
- Like Refinery, Librarian processes and transforms data
- Gas Town already has "persistent infrastructure" roles

**Concerns**:
1. **Overlaps with Mayor**: Mayor already coordinates cross-rig information
2. **Knowledge vs Work**: Gas Town tracks *work* (tasks, issues, MRs). Knowledge is secondary.
3. **Axiom generation**: Who validates "truths"? This introduces subjective judgment into a system designed for objective execution.

### Role Collision Risk

The Librarian could easily become:
- A "Mayor-lite" (coordination without authority)
- A "context compaction" service (which RLM's own paper critiques)
- An accountability gap (who owns axioms?)

**Verdict**: Librarian introduces role confusion. Consider merging knowledge functions into Mayor or creating a simpler "Index" service without agent autonomy.

---

## 3. Knowledge-as-Beads Philosophy

### Current Beads Philosophy

Beads are:
- **Work units** (bugs, tasks, features)
- **Structured data** (JSON in SQLite)
- **Git-backed** (survives failures)
- **Attributable** (who created, who completed)
- **Finite** (created → worked → closed)

### Proposed Knowledge-as-Beads

The proposal wants beads to also be:
- **Axioms** (patterns, truths)
- **Summaries** (compressed knowledge)
- **Persistent knowledge** (never closed?)

### Philosophical Tension

| Work Beads | Knowledge Beads |
|------------|-----------------|
| Created by intent | Derived by analysis |
| Clearly scoped | Emergent boundaries |
| Finite lifecycle | Potentially infinite |
| Accountable actor | Machine-generated |
| Actionable | Informational |

**Core issue**: Gas Town's philosophy is about **accountable execution**. Every bead represents work someone is responsible for. Knowledge beads blur this:
- Who is accountable for an incorrect "axiom"?
- How do you close a "truth"?
- What happens when axioms conflict?

### The Capability Ledger Problem

> "Every completion is recorded. Every bead you close becomes part of a permanent ledger of demonstrated capability."

Knowledge beads don't fit this model:
- You don't "complete" knowledge
- No CV credit for generating axioms
- Breaks the incentive structure

**Verdict**: Knowledge-as-beads dilutes the beads philosophy. Consider labels/tags for knowledge classification instead of new bead types.

---

## 4. Recommendations

### Adopt (with modifications)

**RLM-style internal processing**:
- Allow polecats to use REPL-based context exploration for large inputs
- Keep it internal to the task, not the orchestration model
- Preserve hook-based work assignment

### Simplify or Skip

**Librarian agent**:
- Do not create a new agent role
- If knowledge indexing is needed, make it a service (daemon plugin) not an agent
- Or extend Mayor's capabilities

**Knowledge beads**:
- Do not create new bead types for knowledge
- Use labels/tags on existing beads: `knowledge:pattern`, `knowledge:axiom`
- Keep beads as work units with clear lifecycle

### Preserve

**Core philosophy**:
- Propulsion principle (execute immediately)
- Capability ledger (accountable completions)
- Hook-based work assignment
- Discrete, finite work units

---

## Conclusion

The RLM + Librarian proposal contains useful ideas (handling large context, capturing knowledge) but implements them in ways that conflict with Gas Town's philosophical foundations.

**Gas Town is about execution, not exploration.**

The proposal shifts toward:
- Recursive exploration (vs. discrete execution)
- Emergent knowledge (vs. assigned work)
- New agent roles (vs. extending existing ones)

These shifts would make Gas Town slower, harder to reason about, and reduce accountability.

**Final recommendation**: Extract the valuable parts (REPL-style context handling, knowledge tagging) and implement them within the existing philosophical framework rather than adding new orchestration patterns or agent roles.

---

*Review completed by polecat valkyrie, 2026-01-13*
