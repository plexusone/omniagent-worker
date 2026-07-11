# OmniAgent Worker

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omniagent-worker/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omniagent-worker/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omniagent-worker/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omniagent-worker/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omniagent-worker/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omniagent-worker/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omniagent-worker
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omniagent-worker
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/omniagent-worker
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomniagent-worker
 [loc-svg]: https://tokei.rs/b1/github/plexusone/omniagent-worker
 [repo-url]: https://github.com/plexusone/omniagent-worker
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omniagent-worker/blob/main/LICENSE

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

- [CHANGELOG](CHANGELOG.md) - Version history
- [v0.1.0 Release Notes](docs/releases/v0.1.0.md) - Latest release
- [PRD](docs/specs/origin/PRD.md) - Product requirements
- [TRD](docs/specs/origin/TRD.md) - Technical requirements
- [PLAN](docs/specs/origin/PLAN.md) - Implementation plan
- [ROADMAP](docs/specs/origin/ROADMAP.md) - Future work

## License

MIT License - see [LICENSE](LICENSE) for details.
