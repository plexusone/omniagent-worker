package worker

import (
	"context"
)

// WorkflowExecutor defines how workflows are executed.
// Implement this interface to provide custom orchestration logic.
type WorkflowExecutor interface {
	// Execute runs the workflow with the given input.
	// The dispatcher is used to call workers within the workflow.
	Execute(ctx context.Context, input map[string]any, dispatcher WorkerDispatcher) (map[string]any, error)
}

// WorkerDispatcher is used by workflows to call workers.
type WorkerDispatcher interface {
	// Dispatch calls a worker by ID with the given request.
	Dispatch(ctx context.Context, workerID string, req *Request) (*Response, error)
}

// SequentialWorkflow executes workers in sequence.
// This is the default workflow when none is specified.
type SequentialWorkflow struct {
	// WorkerIDs is the ordered list of workers to execute.
	WorkerIDs []string

	// InputMapper maps workflow input to worker input.
	// If nil, the workflow input is passed directly to each worker.
	InputMapper func(workerID string, workflowInput, previousOutput map[string]any) map[string]any
}

// Execute runs workers sequentially, passing output from one to the next.
func (w *SequentialWorkflow) Execute(ctx context.Context, input map[string]any, dispatcher WorkerDispatcher) (map[string]any, error) {
	currentOutput := input

	for _, workerID := range w.WorkerIDs {
		workerInput := currentOutput
		if w.InputMapper != nil {
			workerInput = w.InputMapper(workerID, input, currentOutput)
		}

		req := NewRequest(workerInput)
		resp, err := dispatcher.Dispatch(ctx, workerID, req)
		if err != nil {
			return nil, err
		}
		if resp.Error != nil {
			return nil, resp.Error
		}

		currentOutput = resp.Output
	}

	return currentOutput, nil
}

// WorkflowFunc is a function that implements WorkflowExecutor.
type WorkflowFunc func(ctx context.Context, input map[string]any, dispatcher WorkerDispatcher) (map[string]any, error)

// Execute implements WorkflowExecutor.
func (f WorkflowFunc) Execute(ctx context.Context, input map[string]any, dispatcher WorkerDispatcher) (map[string]any, error) {
	return f(ctx, input, dispatcher)
}
