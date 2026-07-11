package worker

import (
	"context"
	"testing"
)

func TestCoordinatorBasic(t *testing.T) {
	coord := NewCoordinator(CoordinatorConfig{
		ID: "test-coordinator",
	})

	if coord.ID() != "test-coordinator" {
		t.Fatalf("Expected ID 'test-coordinator', got %q", coord.ID())
	}

	if coord.Pool() == nil {
		t.Fatal("Expected non-nil pool")
	}
}

func TestCoordinatorWithPool(t *testing.T) {
	coord := NewCoordinator(CoordinatorConfig{
		ID: "test-coordinator",
		Workflow: &SequentialWorkflow{
			WorkerIDs: []string{"w1", "w2"},
		},
	})

	// Register workers
	w1 := newTestWorker("w1", "processor")
	w1.executeFunc = func(ctx context.Context, req *Request) (*Response, error) {
		input := req.Input["value"].(float64)
		return NewResponse(req.ID, map[string]any{"value": input * 2}), nil
	}

	w2 := newTestWorker("w2", "processor")
	w2.executeFunc = func(ctx context.Context, req *Request) (*Response, error) {
		input := req.Input["value"].(float64)
		return NewResponse(req.ID, map[string]any{"value": input + 10}), nil
	}

	if err := coord.Pool().Register(w1); err != nil {
		t.Fatalf("Register w1 failed: %v", err)
	}
	if err := coord.Pool().Register(w2); err != nil {
		t.Fatalf("Register w2 failed: %v", err)
	}

	ctx := context.Background()
	resp, err := coord.Execute(ctx, &CoordinatorRequest{
		Input: map[string]any{"value": float64(5)},
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 5 * 2 = 10, 10 + 10 = 20
	expected := float64(20)
	got := resp.Output["value"].(float64)
	if got != expected {
		t.Fatalf("Expected %f, got %f", expected, got)
	}

	if resp.Stats.TaskCount != 2 {
		t.Fatalf("Expected 2 tasks, got %d", resp.Stats.TaskCount)
	}
	if resp.Stats.SuccessCount != 2 {
		t.Fatalf("Expected 2 successes, got %d", resp.Stats.SuccessCount)
	}
}

func TestCoordinatorNoWorkflow(t *testing.T) {
	coord := NewCoordinator(CoordinatorConfig{
		ID: "test-coordinator",
		// No workflow - should just return input
	})

	ctx := context.Background()
	resp, err := coord.Execute(ctx, &CoordinatorRequest{
		Input: map[string]any{"key": "value"},
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp.Output["key"] != "value" {
		t.Fatalf("Expected output to equal input")
	}
}

func TestCoordinatorWorkflowFunc(t *testing.T) {
	coord := NewCoordinator(CoordinatorConfig{
		ID: "test-coordinator",
		Workflow: WorkflowFunc(func(ctx context.Context, input map[string]any, dispatcher WorkerDispatcher) (map[string]any, error) {
			// Call worker directly
			resp, err := dispatcher.Dispatch(ctx, "doubler", NewRequest(input))
			if err != nil {
				return nil, err
			}
			return resp.Output, nil
		}),
	})

	doubler := newTestWorker("doubler", "processor")
	doubler.executeFunc = func(ctx context.Context, req *Request) (*Response, error) {
		v := req.Input["n"].(float64)
		return NewResponse(req.ID, map[string]any{"n": v * 2}), nil
	}

	if err := coord.Pool().Register(doubler); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	ctx := context.Background()
	resp, err := coord.Execute(ctx, &CoordinatorRequest{
		Input: map[string]any{"n": float64(7)},
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	expected := float64(14)
	got := resp.Output["n"].(float64)
	if got != expected {
		t.Fatalf("Expected %f, got %f", expected, got)
	}
}

func TestCoordinatorHealth(t *testing.T) {
	coord := NewCoordinator(CoordinatorConfig{
		ID: "test-coordinator",
	})

	if err := coord.Pool().Register(newTestWorker("w1", "test")); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	ctx := context.Background()
	status := coord.Health(ctx)

	if status.Status != HealthStatusHealthy {
		t.Fatalf("Expected healthy, got %s", status.Status)
	}
}
