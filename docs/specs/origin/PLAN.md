# omniagent-worker: Implementation Plan

## Overview

**Project**: omniagent-worker
**Type**: New Package
**Status**: Planning
**Created**: 2026-07-10

## Phases

### Phase 1: Core Interfaces

**Goal**: Establish the foundational types and interfaces.

**Tasks**:

1. Create repository structure
   - [ ] Initialize go.mod
   - [ ] Create directory structure
   - [ ] Add LICENSE (MIT + Apache-2.0)
   - [ ] Add README.md stub

2. Implement core types
   - [ ] `worker.go` - Worker interface, Request, Response, HealthStatus
   - [ ] `base.go` - BaseWorker implementation
   - [ ] `config.go` - WorkerConfig, LLMConfig, AgentOpsConfig
   - [ ] `errors.go` - WorkerError, error codes
   - [ ] `doc.go` - Package documentation

3. Implement Pool
   - [ ] `pool.go` - Pool struct, Register, Get, Execute, Init, Shutdown

4. Add tests
   - [ ] `worker_test.go` - Interface compliance tests
   - [ ] `pool_test.go` - Pool functionality tests

**Exit Criteria**: Core interfaces compile and pass unit tests.

### Phase 2: AgentOps Integration

**Goal**: Integrate with omniobserve/agentops for observability.

**Tasks**:

1. Add AgentOps integration
   - [ ] `agentops.go` - Store initialization, context helpers
   - [ ] Tracing wrappers for Pool.Execute

2. Implement Coordinator
   - [ ] `coordinator.go` - Coordinator struct, Execute with tracing
   - [ ] `dispatcher.go` - TracingDispatcher implementation

3. Add tests
   - [ ] `agentops_test.go` - Verify traces are created
   - [ ] `coordinator_test.go` - Coordinator with mock workers

**Exit Criteria**: Coordinator execution creates proper AgentOps traces.

### Phase 3: Server Layer

**Goal**: Add HTTP server for standalone deployment.

**Tasks**:

1. Implement server package
   - [ ] `server/server.go` - Run function, Config
   - [ ] `server/handlers.go` - Execute, Health handlers
   - [ ] `server/middleware.go` - AgentOps HTTP middleware

2. Add client for remote workers
   - [ ] `client.go` - WorkerClient, HTTP calls with context propagation

3. Add tests
   - [ ] `server/server_test.go` - HTTP endpoint tests
   - [ ] `client_test.go` - Client tests

**Exit Criteria**: Workers can be deployed standalone and called remotely.

### Phase 4: Eino Integration

**Goal**: Support Eino workflows for deterministic orchestration.

**Tasks**:

1. Implement Eino adapters
   - [ ] `eino/workflow.go` - WorkflowExecutor wrapper
   - [ ] `eino/nodes.go` - Lambda node helpers

2. Add example workflow
   - [ ] `examples/coordinator/main.go` - Example with Eino workflow

**Exit Criteria**: Coordinator can use Eino graphs for orchestration.

### Phase 5: Documentation & Examples

**Goal**: Complete documentation and examples.

**Tasks**:

1. Documentation
   - [ ] Update README.md with full documentation
   - [ ] Add GoDoc comments to all exported types
   - [ ] Create architecture diagram

2. Examples
   - [ ] `examples/minimal/main.go` - Minimal worker
   - [ ] `examples/coordinator/main.go` - Coordinator with pool
   - [ ] `examples/standalone/main.go` - Standalone HTTP deployment

**Exit Criteria**: Package is documented and has working examples.

## Dependencies Between Phases

```
Phase 1 (Core Interfaces)
    │
    ▼
Phase 2 (AgentOps Integration)
    │
    ├──────────────────┐
    ▼                  ▼
Phase 3 (Server)    Phase 4 (Eino)
    │                  │
    └────────┬─────────┘
             ▼
     Phase 5 (Documentation)
```

## Timeline Estimates

| Phase | Estimated Effort |
|-------|------------------|
| Phase 1 | 1-2 days |
| Phase 2 | 1-2 days |
| Phase 3 | 1-2 days |
| Phase 4 | 1 day |
| Phase 5 | 1 day |
| **Total** | **5-8 days** |

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| AgentOps API changes | Pin to specific version, wrap access |
| Eino complexity | Make Eino optional, support custom WorkflowExecutor |
| Scope creep | Stick to PRD requirements |

## Success Metrics

1. All tests pass
2. golangci-lint clean
3. agent-team-stats can import and use
4. AgentOps traces visible in observability backend
