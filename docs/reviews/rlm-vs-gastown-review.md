# Review: RLM Architecture vs Gas Town Philosophy

## Executive Summary

This review compares MIT's Recursive Language Models (RLM) architecture with Gas Town's multi-agent orchestration philosophy. While both address LLM context limitations, they represent fundamentally different paradigms:

- **RLM**: Extends single-agent capability through context externalization and recursive self-invocation
- **Gas Town**: Multiplies capability through coordinated multi-agent execution with persistent state

These approaches are **complementary rather than competing**. RLM solves a different problem (ultra-long context within a single query) than Gas Town (coordinated multi-agent software engineering workflows).

## RLM Architecture Overview

### Core Concept

RLM (Zhang et al., MIT CSAIL, 2025) treats long prompts as **external environment** rather than direct neural network input. The key insight:

> "Long prompts should not be fed into the neural network directly but should instead be treated as part of the environment that the LLM can symbolically interact with."

### Mechanism

1. **REPL Environment**: Load prompt as a variable in a Python REPL
2. **Programmatic Access**: LLM writes code to peek into, decompose, and transform the context
3. **Recursive Sub-calling**: LLM spawns sub-LLM instances to handle subproblems
4. **Context Folding**: Never load full context into model window

### Performance Claims

- Handles inputs up to **2 orders of magnitude beyond model context windows**
- On OOLONG-Pairs (quadratic complexity): GPT-5 scores <0.1%, RLM(GPT-5) scores 58%
- Scales to **10M+ tokens** while maintaining comparable costs
- Outperforms context compaction and retrieval-based approaches

### Use Cases

- Deep research over large document corpora
- Information aggregation requiring dense access to many parts of prompt
- Code repository understanding
- Long-horizon reasoning tasks

## Gas Town Philosophy Overview

### Core Concept

Gas Town is a **multi-agent orchestration system** treating AI agent work as structured data. The fundamental insight:

> "Gas Town is a steam engine. Agents are pistons. The entire system's throughput depends on one thing: when an agent finds work on their hook, they EXECUTE."

### Mechanism

1. **Git-Backed Hooks**: Work persists in git worktrees, survives restarts
2. **Propulsion Principle**: Autonomous execution without confirmation
3. **Agent Taxonomy**: Mayor (coordinator), Polecats (workers), Witness (monitor), Refinery (merger)
4. **Capability Ledger**: Every completion recorded, building agent track records
5. **Mail System**: Inter-agent communication for coordination

### Performance Characteristics

- Scales to **20-30 agents** comfortably
- Work state survives crashes and handoffs
- Supports parallel task execution across discrete units
- Built-in accountability and attribution

### Use Cases

- Multi-issue software development
- Batch work across repositories
- Repeatable workflows (formulas/molecules)
- Long-running projects requiring human oversight

## Architectural Comparison

| Dimension | RLM | Gas Town |
|-----------|-----|----------|
| **Paradigm** | Single-agent with recursion | Multi-agent coordination |
| **Context Handling** | Externalize to REPL | Distribute across agents |
| **State Persistence** | None (per-query) | Git-backed hooks |
| **Decomposition** | Code-based recursive calls | Task-based work assignment |
| **Communication** | Function returns | Mail/messaging system |
| **Failure Recovery** | Re-run query | Resume from hook |
| **Scaling Strategy** | Deeper recursion | More agents |
| **Primary Metric** | Accuracy on long inputs | Throughput on discrete tasks |

## Philosophical Differences

### On Context Management

**RLM**: Criticizes context compaction as "rarely expressive enough for tasks that require dense access to many parts of the prompt." Instead, treats context as programmatically addressable data.

**Gas Town**: Embraces context compaction as necessary for multi-agent coordination. Context limits are addressed by distributing work across agents, not by extending single-agent capability.

### On Agent Autonomy

**RLM**: Sub-agents are ephemeral, spawned for specific sub-problems, no identity or history.

**Gas Town**: Agents have persistent identity, tracked work history ("Capability Ledger"), and build reputation over time. "Your CV grows with every completion."

### On Work Organization

**RLM**: Task decomposition happens organically through recursive calls based on model decisions.

**Gas Town**: Task decomposition is explicit through convoys, molecules (workflow definitions), and deliberate work slinging.

### On System Model

**RLM**: Treats the problem as **computational** - how can one LLM process more data?

**Gas Town**: Treats the problem as **organizational** - how can many LLMs collaborate effectively?

## Strengths and Limitations

### RLM Strengths

1. **Massive context scaling**: Handles 10M+ tokens in single query
2. **Dense information access**: Doesn't lose information through summarization
3. **Model-agnostic**: Works with any capable LLM
4. **Task-agnostic**: Same system prompt works across diverse benchmarks

### RLM Limitations

1. **No persistent state**: Each query starts fresh
2. **High variance costs**: Tail-end costs can be 3x+ median
3. **Single-query scope**: Not designed for multi-day projects
4. **No accountability**: Sub-agents are anonymous

### Gas Town Strengths

1. **Crash resilience**: Work survives restarts
2. **Accountability**: Full attribution and work history
3. **Parallelization**: True concurrent execution across agents
4. **Human oversight**: Built-in escalation and monitoring

### Gas Town Limitations

1. **Context-per-agent**: Each agent has standard context limits
2. **Coordination overhead**: Messaging, hook management
3. **Setup complexity**: Requires infrastructure (tmux, git worktrees)
4. **Discrete tasks**: Works best with well-defined, separable work units

## Synthesis: Complementary Approaches

These architectures solve different problems:

**RLM is ideal for**:
- Single queries requiring dense access to massive context
- Deep research and document analysis
- Tasks where the entire input is needed to form an answer
- Situations where you can't pre-decompose the problem

**Gas Town is ideal for**:
- Ongoing software development projects
- Parallelizable work across multiple issues
- Tasks requiring human oversight and approval
- Situations where work must survive agent failures
- Building institutional knowledge through capability tracking

### Potential Integration

Gas Town could potentially incorporate RLM principles:
- **For polecats handling large files**: Use RLM-style REPL interaction for massive codebases
- **For deep research tasks**: Spawn an RLM-capable agent for analysis phases
- **For context-heavy reviews**: RLM approach could help with PR review over large diffs

However, the core Gas Town philosophy of discrete, trackable work units with persistent state remains orthogonal to RLM's single-query optimization.

## Conclusion

RLM and Gas Town represent two valid responses to LLM limitations:

- **RLM**: "Make one agent smarter about context"
- **Gas Town**: "Coordinate many agents with persistent state"

For software engineering workflows, Gas Town's approach aligns better with:
- The discrete nature of software tasks (issues, PRs, features)
- The need for accountability and tracking
- The reality of agent failures and restarts
- The value of building agent capability over time

For analytical tasks requiring massive context (document analysis, deep research), RLM provides capabilities that Gas Town's current architecture doesn't address.

**Recommendation**: These approaches are complementary. Gas Town could benefit from RLM-style context handling for specific sub-tasks, while maintaining its core philosophy of coordinated multi-agent execution with persistent, accountable work tracking.

---

*Review completed by polecat valkyrie, 2026-01-13*
