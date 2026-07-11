package worker

import (
	"context"
)

// Worker is the base interface for task-oriented agents.
// Workers are stateless request/response handlers that can be
// deployed in-process or as HTTP microservices.
type Worker interface {
	// ID returns the unique identifier for this worker instance.
	ID() string

	// Type returns the worker type (e.g., "research", "synthesis").
	// Multiple workers can have the same type.
	Type() string

	// Version returns the worker implementation version.
	Version() string

	// Init initializes the worker. Called once before Execute.
	Init(ctx context.Context) error

	// Shutdown gracefully shuts down the worker.
	Shutdown(ctx context.Context) error

	// Health returns the current health status of the worker.
	Health(ctx context.Context) HealthStatus

	// Execute performs the worker's task.
	Execute(ctx context.Context, req *Request) (*Response, error)
}

// Request is the input to a worker.
type Request struct {
	// ID is a unique identifier for this request.
	ID string `json:"id"`

	// TaskID is the AgentOps task ID for tracing.
	TaskID string `json:"task_id,omitempty"`

	// WorkflowID is the workflow this request belongs to.
	WorkflowID string `json:"workflow_id,omitempty"`

	// Input contains worker-specific input data.
	Input map[string]any `json:"input"`
}

// Response is the output from a worker.
type Response struct {
	// RequestID echoes back the request ID.
	RequestID string `json:"request_id"`

	// Output contains worker-specific output data.
	Output map[string]any `json:"output"`

	// Error contains structured error information if the request failed.
	Error *WorkerError `json:"error,omitempty"`
}

// HealthStatus represents the health of a worker.
type HealthStatus struct {
	// Status is one of "healthy", "degraded", or "unhealthy".
	Status string `json:"status"`

	// Details contains additional health information.
	Details map[string]string `json:"details,omitempty"`
}

// Health status constants.
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
)

// NewRequest creates a new request with a generated ID.
func NewRequest(input map[string]any) *Request {
	return &Request{
		ID:    generateID(),
		Input: input,
	}
}

// NewResponse creates a successful response.
func NewResponse(requestID string, output map[string]any) *Response {
	return &Response{
		RequestID: requestID,
		Output:    output,
	}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(requestID string, err *WorkerError) *Response {
	return &Response{
		RequestID: requestID,
		Error:     err,
	}
}
