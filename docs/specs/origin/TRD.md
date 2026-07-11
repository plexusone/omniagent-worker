# omniagent-worker: Technical Requirements Document

## Overview

**Project**: omniagent-worker
**Type**: New Package
**Status**: Planning
**Created**: 2026-07-10

## Architecture

### Package Structure

```
github.com/plexusone/omniagent-worker/
├── worker.go           # Worker interface
├── coordinator.go      # Coordinator implementation
├── pool.go             # In-process worker pool
├── client.go           # HTTP client for remote workers
├── agentops.go         # AgentOps integration
├── config.go           # Configuration types
├── health.go           # Health check interface
├── doc.go              # Package documentation
│
├── server/             # HTTP server layer (optional)
│   ├── server.go       # HTTP server setup
│   ├── handlers.go     # HTTP handlers
│   ├── middleware.go   # AgentOps HTTP middleware
│   ├── health.go       # Health endpoints
│   └── a2a.go          # A2A protocol (future)
│
├── eino/               # Eino workflow integration
│   ├── workflow.go     # Workflow builder
│   └── nodes.go        # Lambda node helpers
│
└── examples/
    ├── minimal/        # Minimal worker example
    └── coordinator/    # Coordinator example
```

### Layer Separation

```
┌─────────────────────────────────────────────┐
│           server/ (optional)                │
│  HTTP, health endpoints, A2A protocol       │
├─────────────────────────────────────────────┤
│              Core Package                   │
│  Worker, Coordinator, Pool, AgentOps        │
├─────────────────────────────────────────────┤
│           External Dependencies             │
│  omniobserve/agentops, omnillm, eino        │
└─────────────────────────────────────────────┘
```

## Core Interfaces

### Worker Interface

```go
// Worker is the base interface for task-oriented agents.
// Workers are stateless request/response handlers that can be
// deployed in-process or as HTTP microservices.
type Worker interface {
    // Identity
    ID() string
    Type() string
    Version() string

    // Lifecycle
    Init(ctx context.Context) error
    Shutdown(ctx context.Context) error

    // Health
    Health(ctx context.Context) HealthStatus

    // Execution - the core work interface
    Execute(ctx context.Context, req *Request) (*Response, error)
}

// Request is the input to a worker
type Request struct {
    ID       string         // Unique request ID
    TaskID   string         // AgentOps task ID (for tracing)
    Workflow string         // Workflow name (for context)
    Input    map[string]any // Worker-specific input
}

// Response is the output from a worker
type Response struct {
    RequestID string         // Echo back request ID
    Output    map[string]any // Worker-specific output
    Error     *WorkerError   // Structured error if failed
}

// HealthStatus represents worker health
type HealthStatus struct {
    Status  string            // "healthy", "degraded", "unhealthy"
    Details map[string]string // Additional details
}
```

### BaseWorker Implementation

```go
// BaseWorker provides common functionality for workers
type BaseWorker struct {
    id        string
    workerType string
    version   string
    config    WorkerConfig
    agentOps  agentops.Store
    logger    *slog.Logger
}

// WorkerConfig configures a worker
type WorkerConfig struct {
    ID      string
    Type    string
    Version string

    // LLM configuration (optional - not all workers need LLM)
    LLM *LLMConfig

    // Observability
    AgentOps *AgentOpsConfig

    // Logging
    Logger *slog.Logger
}

// LLMConfig for workers that need LLM
type LLMConfig struct {
    Provider string // "anthropic", "openai", "gemini", etc.
    Model    string
    APIKey   string
}

// AgentOpsConfig for observability
type AgentOpsConfig struct {
    Store   agentops.Store
    Enabled bool
}
```

### Coordinator Interface

```go
// Coordinator manages a team of workers
type Coordinator struct {
    id       string
    pool     *Pool
    clients  map[string]*WorkerClient
    workflow WorkflowExecutor
    agentOps agentops.Store
    config   CoordinatorConfig
}

// CoordinatorConfig configures the coordinator
type CoordinatorConfig struct {
    ID       string
    Workflow WorkflowExecutor // Eino graph or custom
    AgentOps *AgentOpsConfig
    Timeout  time.Duration
}

// NewCoordinator creates a new coordinator
func NewCoordinator(cfg CoordinatorConfig) *Coordinator

// Pool returns the in-process worker pool
func (c *Coordinator) Pool() *Pool

// AddRemote adds a remote worker endpoint
func (c *Coordinator) AddRemote(id, url string, opts ...ClientOption)

// Execute runs the workflow with the given input
func (c *Coordinator) Execute(ctx context.Context, req *CoordinatorRequest) (*CoordinatorResponse, error)

// CoordinatorRequest is input to the coordinator
type CoordinatorRequest struct {
    WorkflowID string         // Unique workflow instance ID
    Input      map[string]any // Workflow input
}

// CoordinatorResponse is output from the coordinator
type CoordinatorResponse struct {
    WorkflowID string         // Echo back workflow ID
    Output     map[string]any // Workflow output
    Stats      WorkflowStats  // Execution statistics
}

// WorkflowStats contains execution metrics
type WorkflowStats struct {
    Duration      time.Duration
    TaskCount     int
    SuccessCount  int
    FailureCount  int
    RetryCount    int
}
```

### Pool Interface

```go
// Pool manages in-process workers
type Pool struct {
    workers map[string]Worker
    mu      sync.RWMutex
}

// NewPool creates a new worker pool
func NewPool() *Pool

// Register adds a worker to the pool
func (p *Pool) Register(w Worker) error

// Unregister removes a worker from the pool
func (p *Pool) Unregister(id string) error

// Get retrieves a worker by ID
func (p *Pool) Get(id string) (Worker, bool)

// List returns all registered workers
func (p *Pool) List() []Worker

// Execute calls a worker directly (in-process)
func (p *Pool) Execute(ctx context.Context, workerID string, req *Request) (*Response, error)

// Init initializes all workers
func (p *Pool) Init(ctx context.Context) error

// Shutdown shuts down all workers
func (p *Pool) Shutdown(ctx context.Context) error
```

### WorkflowExecutor Interface

```go
// WorkflowExecutor defines how workflows are executed
type WorkflowExecutor interface {
    Execute(ctx context.Context, input map[string]any, dispatcher WorkerDispatcher) (map[string]any, error)
}

// WorkerDispatcher is passed to workflows for calling workers
type WorkerDispatcher interface {
    Dispatch(ctx context.Context, workerID string, req *Request) (*Response, error)
}
```

## AgentOps Integration

### Automatic Tracing

```go
// Coordinator.Execute automatically creates:
// 1. Workflow span (gen_ai.agent.workflow.*)
// 2. Task spans for each worker call (gen_ai.agent.task.*)
// 3. Handoff spans between workers (gen_ai.agent.handoff.*)

func (c *Coordinator) Execute(ctx context.Context, req *CoordinatorRequest) (*CoordinatorResponse, error) {
    // Start workflow
    ctx, workflow, _ := middleware.StartWorkflow(ctx, c.agentOps, req.WorkflowID)
    defer middleware.CompleteWorkflow(ctx, c.agentOps, workflow)

    // Workflow executor gets a dispatcher that traces calls
    dispatcher := &tracingDispatcher{
        pool:     c.pool,
        clients:  c.clients,
        agentOps: c.agentOps,
        workflow: workflow,
    }

    return c.workflow.Execute(ctx, req.Input, dispatcher)
}
```

### Context Propagation

```go
// X-AgentOps-* headers propagate context to remote workers
// - X-AgentOps-Workflow-ID
// - X-AgentOps-Task-ID
// - X-AgentOps-Agent-ID
// - X-AgentOps-Trace-ID
```

## Server Layer

### HTTP Server

```go
// server/server.go

// Run starts an HTTP server for the coordinator or worker
func Run(ctx context.Context, handler any, cfg Config) error

// Config for the HTTP server
type Config struct {
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

// Supports both:
// - Single worker: server.Run(ctx, myWorker, cfg)
// - Coordinator: server.Run(ctx, myCoordinator, cfg)
```

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/execute` | POST | Execute worker/coordinator |
| `/health` | GET | Health check |
| `/ready` | GET | Readiness check |
| `/.well-known/agent-card.json` | GET | A2A discovery (future) |

## Eino Integration

```go
// eino/workflow.go

// NewStatefulWorkflow wraps an Eino graph as a WorkflowExecutor
func NewStatefulWorkflow(graph *compose.Graph[Input, Output]) WorkflowExecutor

// Lambda helpers for common patterns
func WorkerNode(workerID string) compose.Lambda[Input, Output]
func QualityCheckNode(checker QualityChecker) compose.Lambda[Input, Output]
func RetryNode(workerID string, maxRetries int) compose.Lambda[Input, Output]
```

## Configuration

### Environment Variables

```bash
# Worker identity
WORKER_ID=synthesis
WORKER_TYPE=synthesis
WORKER_VERSION=1.0.0

# LLM (optional)
LLM_PROVIDER=anthropic
LLM_MODEL=claude-sonnet-4-20250514
LLM_API_KEY=${ANTHROPIC_API_KEY}

# AgentOps
AGENTOPS_ENABLED=true
AGENTOPS_DSN=postgres://...

# Server
SERVER_PORT=8004
SERVER_READ_TIMEOUT=30s
```

### Programmatic Configuration

```go
worker := NewMyWorker(WorkerConfig{
    ID:      "synthesis",
    Type:    "synthesis",
    Version: "1.0.0",
    LLM: &LLMConfig{
        Provider: "anthropic",
        Model:    "claude-sonnet-4-20250514",
    },
    AgentOps: &AgentOpsConfig{
        Enabled: true,
        Store:   agentOpsStore,
    },
})
```

## Dependencies

### Direct Dependencies

```go
require (
    github.com/plexusone/omniobserve v0.11.0  // AgentOps
    github.com/plexusone/omnillm v0.17.0      // LLM (optional)
    github.com/cloudwego/eino v0.9.12         // Workflows (optional)
    github.com/go-chi/chi/v5 v5.3.0           // HTTP router (server only)
)
```

### Dependency Rules

1. Core package MUST NOT import `server/`
2. Core package MAY import `eino/` (optional)
3. `server/` imports core package
4. AgentOps is the only required external dependency

## Error Handling

```go
// WorkerError provides structured error information
type WorkerError struct {
    Code    string // Error code (e.g., "VALIDATION_ERROR")
    Message string // Human-readable message
    Details any    // Additional context
    Cause   error  // Underlying error
}

// Standard error codes
const (
    ErrCodeValidation = "VALIDATION_ERROR"
    ErrCodeExecution  = "EXECUTION_ERROR"
    ErrCodeTimeout    = "TIMEOUT_ERROR"
    ErrCodeLLM        = "LLM_ERROR"
    ErrCodeNotFound   = "NOT_FOUND"
)
```

## Testing

### Unit Tests

```go
// Mock worker for testing
type MockWorker struct {
    BaseWorker
    ExecuteFunc func(ctx context.Context, req *Request) (*Response, error)
}

// Mock AgentOps store
type MockAgentOpsStore struct {
    // Records all operations for verification
}
```

### Integration Tests

```go
// Test coordinator with in-process pool
func TestCoordinatorWithPool(t *testing.T) {
    pool := NewPool()
    pool.Register(NewMockWorker("research"))
    pool.Register(NewMockWorker("synthesis"))

    coord := NewCoordinator(CoordinatorConfig{
        Workflow: testWorkflow,
    })
    coord.pool = pool

    resp, err := coord.Execute(ctx, req)
    // Verify workflow execution
}
```

## Security Considerations

1. **Input Validation** - All Request inputs are validated
2. **Timeout Enforcement** - Workers have execution timeouts
3. **Error Sanitization** - Internal errors not exposed in responses
4. **No Credential Logging** - LLM API keys never logged

## References

- [omniobserve/agentops](https://github.com/plexusone/omniobserve) - AgentOps package
- [Eino Documentation](https://github.com/cloudwego/eino) - Workflow framework
- [omniskill](https://github.com/plexusone/omniskill) - Skill interface
