# omniagent-worker: Product Requirements Document

## Overview

**Project**: omniagent-worker
**Type**: New Package
**Status**: Planning
**Created**: 2026-07-10

## Problem Statement

PlexusOne needs a Go-first multi-agent framework that:

1. Provides building blocks for task-oriented worker agents
2. Supports both in-process and distributed deployment
3. Integrates with existing PlexusOne observability (AgentOps)
4. Remains lightweight enough for embedding in skills (omniskill)
5. Is maintained and driven by internal use cases, not external adoption

Current state:

- `agent-team-stats` implements custom agent patterns that should be reusable
- `omniagent` is a full conversational assistant, too heavy for task workers
- No Go equivalent to CrewAI for multi-agent coordination
- AgentOps observability exists but isn't used for agent lifecycle tracing

## Goals

### Primary Goals

1. **Provide Worker abstraction** - Minimal interface for task-oriented agents
2. **Provide Coordinator abstraction** - Manages worker teams with workflow support
3. **Support dual deployment** - In-process (Pool) and remote (HTTP) workers
4. **Integrate AgentOps** - Full observability for workflows, tasks, handoffs
5. **Enable omniskill integration** - Lightweight enough for skill embedding

### Non-Goals

1. Not competing with CrewAI or other external frameworks
2. Not replacing omniagent for conversational assistants
3. Not supporting non-Go languages
4. Not implementing Multi-Agent Spec rendering (future work)

## Users

### Primary User

- PlexusOne internal development (ourselves)
- Projects: agent-team-stats, future agent teams

### Secondary Users

- Other Go projects in the grokify ecosystem
- Developers wanting lightweight multi-agent coordination

## Requirements

### Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| F1 | Worker interface with Init/Shutdown lifecycle | Must |
| F2 | Coordinator for managing worker teams | Must |
| F3 | Pool for in-process worker management | Must |
| F4 | AgentOps integration for tracing | Must |
| F5 | HTTP server layer for standalone deployment | Must |
| F6 | Health check endpoints | Must |
| F7 | Eino workflow integration | Should |
| F8 | A2A protocol support | Could |
| F9 | Remote worker discovery | Could |

### Non-Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| N1 | Core interfaces have zero external dependencies | Must |
| N2 | In-process mode adds no HTTP overhead | Must |
| N3 | Clear separation between core and server layers | Must |
| N4 | Compatible with omniskill embedding | Must |
| N5 | Follows PlexusOne naming conventions | Must |

## Success Criteria

1. agent-team-stats successfully migrates to omniagent-worker
2. AgentOps traces show complete workflow visibility
3. Stats verification works as omniskill in omniagent
4. In-process mode has negligible overhead vs current implementation

## Dependencies

### Upstream Dependencies

- `github.com/plexusone/omniobserve` - AgentOps tracing
- `github.com/plexusone/omnillm` - LLM abstraction (for workers that need LLM)
- `github.com/cloudwego/eino` - Workflow orchestration

### Downstream Dependents

- `github.com/plexusone/agent-team-stats` - First consumer
- `github.com/plexusone/omniagent` - Potential integration for team delegation

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Scope creep to match CrewAI | Medium | High | Stay focused on internal needs |
| Over-engineering the interface | Medium | Medium | Start minimal, add as needed |
| AgentOps integration complexity | Low | Medium | Leverage existing middleware |

## Timeline

See PLAN.md for detailed phases.

## References

- [IDEATION_CHAT.md](../../../IDEATION_CHAT.md) - Original discussion
- [omniobserve/agentops](https://github.com/plexusone/omniobserve) - AgentOps package
- [omniskill](https://github.com/plexusone/omniskill) - Skill interface
