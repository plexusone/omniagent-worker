# omniagent-worker: Roadmap

## Overview

**Project**: omniagent-worker
**Type**: New Package
**Status**: In Progress
**Created**: 2026-07-10
**Updated**: 2026-07-10

## Work Items (In Order)

### Milestone 1: MVP (Current Focus)

- [x] **1.1** Initialize repository with go.mod, LICENSE, README
- [x] **1.2** Implement Worker interface and BaseWorker
- [x] **1.3** Implement Pool for in-process workers
- [ ] **1.4** Add AgentOps integration (basic structure in place)
- [x] **1.5** Implement Coordinator with tracing
- [x] **1.6** Add HTTP server layer (`server/`)
- [x] **1.7** Add WorkerClient for remote calls
- [ ] **1.8** Add Eino workflow integration (`eino/`)
- [ ] **1.9** Create minimal examples
- [ ] **1.10** Complete documentation

### Milestone 2: Production Hardening

- [ ] **2.1** Add circuit breaker for remote workers
- [ ] **2.2** Add retry logic with backoff
- [ ] **2.3** Add metrics (Prometheus)
- [ ] **2.4** Add graceful shutdown handling
- [ ] **2.5** Add connection pooling for remote workers
- [ ] **2.6** Add rate limiting

### Milestone 3: A2A Protocol

- [ ] **3.1** Implement AgentCard generation
- [ ] **3.2** Implement `.well-known/agent-card.json` endpoint
- [ ] **3.3** Implement A2A invocation handler
- [ ] **3.4** Add A2A client for discovery

### Milestone 4: Multi-Agent Spec Integration

- [ ] **4.1** Add Multi-Agent Spec parser
- [ ] **4.2** Implement spec-to-Coordinator renderer
- [ ] **4.3** Support dynamic worker instantiation from spec
- [ ] **4.4** Add validation for spec compliance

### Milestone 5: Advanced Features

- [ ] **5.1** Worker versioning and migration
- [ ] **5.2** Blue-green deployment support
- [ ] **5.3** Worker capability negotiation
- [ ] **5.4** Distributed tracing correlation (full OTEL)
- [ ] **5.5** Worker sandboxing (WASM runtime)

## Future Considerations

### OmniAgent Integration

- OmniAgent can delegate to omniagent-worker Coordinators
- Shared AgentOps traces across assistant and worker teams
- Unified observability dashboard

### Additional Worker Types

- `WebhookWorker` - Triggered by external webhooks
- `ScheduledWorker` - Cron-based execution
- `StreamingWorker` - SSE/WebSocket streaming responses

### Ecosystem Integration

- MCP tool wrapper for workers
- OpenAPI spec generation
- Kubernetes operator for worker deployment

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 0.1.0 | TBD | Initial MVP |
| 0.2.0 | TBD | Production hardening |
| 0.3.0 | TBD | A2A protocol |
| 1.0.0 | TBD | Stable release |

## References

- [PRD.md](PRD.md) - Product requirements
- [TRD.md](TRD.md) - Technical requirements
- [PLAN.md](PLAN.md) - Implementation plan
