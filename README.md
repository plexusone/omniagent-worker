# omniagent-worker

Go-first multi-agent worker framework for building task-oriented agent teams.

## Overview

omniagent-worker provides building blocks for creating multi-agent systems in Go:

- ⚙️ **Worker**: Minimal interface for task-oriented agents
- 🎯 **Coordinator**: Manages worker teams with workflow support
- 📦 **Pool**: In-process worker management for embedded use
- 👁️ **AgentOps**: Full observability via OpenTelemetry-compatible tracing

## Installation

```bash
go get github.com/plexusone/omniagent-worker
```

## Quick Start

### In-Process Workers (Embedded)

```go
package main

import (
    "context"

    worker "github.com/plexusone/omniagent-worker"
)

func main() {
    ctx := context.Background()

    // Create coordinator
    coord := worker.NewCoordinator(worker.CoordinatorConfig{
        ID: "my-coordinator",
    })

    // Register workers
    coord.Pool().Register(NewMyWorker())

    // Execute
    resp, err := coord.Execute(ctx, &worker.CoordinatorRequest{
        Input: map[string]any{"topic": "example"},
    })
}
```

### Standalone HTTP Server

```go
package main

import (
    "context"

    worker "github.com/plexusone/omniagent-worker"
    "github.com/plexusone/omniagent-worker/server"
)

func main() {
    ctx := context.Background()

    myWorker := NewMyWorker(worker.WorkerConfig{
        ID:   "my-worker",
        Type: "processor",
    })

    server.Run(ctx, myWorker, server.Config{Port: 8080})
}
```

## Architecture

```
┌─────────────────────────────────────────────┐
│           server/ (optional)                │
│  HTTP endpoints, health checks, A2A         │
├─────────────────────────────────────────────┤
│              Core Package                   │
│  Worker, Coordinator, Pool, AgentOps        │
├─────────────────────────────────────────────┤
│           External Dependencies             │
│  omniobserve/agentops, omnillm, eino        │
└─────────────────────────────────────────────┘
```

## OmniAgent Family

omniagent-worker is part of the OmniAgent family:

| Package | Purpose |
|---------|---------|
| `omniagent` | Full conversational assistant with channels |
| `omniagent-worker` | Task-oriented workers and coordinators |

## Documentation

- [PRD](docs/specs/origin/PRD.md) - Product requirements
- [TRD](docs/specs/origin/TRD.md) - Technical requirements
- [PLAN](docs/specs/origin/PLAN.md) - Implementation plan
- [ROADMAP](docs/specs/origin/ROADMAP.md) - Future work

## License

MIT License - see [LICENSE](LICENSE) for details.
