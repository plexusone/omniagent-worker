package worker

import (
	"context"
	"testing"
)

// testWorker is a simple worker for testing.
type testWorker struct {
	BaseWorker
	executeFunc func(ctx context.Context, req *Request) (*Response, error)
}

func newTestWorker(id, workerType string) *testWorker {
	return &testWorker{
		BaseWorker: NewBaseWorker(WorkerConfig{
			ID:      id,
			Type:    workerType,
			Version: "1.0.0",
		}),
	}
}

func (w *testWorker) Execute(ctx context.Context, req *Request) (*Response, error) {
	if w.executeFunc != nil {
		return w.executeFunc(ctx, req)
	}
	return NewResponse(req.ID, map[string]any{"echo": req.Input}), nil
}

func TestPoolRegister(t *testing.T) {
	pool := NewPool()

	w := newTestWorker("test-1", "test")
	if err := pool.Register(w); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Should fail on duplicate
	if err := pool.Register(w); err == nil {
		t.Fatal("Expected error on duplicate registration")
	}
}

func TestPoolGet(t *testing.T) {
	pool := NewPool()

	w := newTestWorker("test-1", "test")
	if err := pool.Register(w); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	got, ok := pool.Get("test-1")
	if !ok {
		t.Fatal("Expected to find worker")
	}
	if got.ID() != "test-1" {
		t.Fatalf("Expected ID 'test-1', got %q", got.ID())
	}

	_, ok = pool.Get("nonexistent")
	if ok {
		t.Fatal("Expected not to find nonexistent worker")
	}
}

func TestPoolExecute(t *testing.T) {
	pool := NewPool()

	w := newTestWorker("test-1", "test")
	w.executeFunc = func(ctx context.Context, req *Request) (*Response, error) {
		return NewResponse(req.ID, map[string]any{"result": "success"}), nil
	}

	if err := pool.Register(w); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	ctx := context.Background()
	req := NewRequest(map[string]any{"key": "value"})

	resp, err := pool.Execute(ctx, "test-1", req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp.Output["result"] != "success" {
		t.Fatalf("Expected result 'success', got %v", resp.Output["result"])
	}
}

func TestPoolExecuteNotFound(t *testing.T) {
	pool := NewPool()
	ctx := context.Background()
	req := NewRequest(map[string]any{})

	_, err := pool.Execute(ctx, "nonexistent", req)
	if err == nil {
		t.Fatal("Expected error for nonexistent worker")
	}

	we, ok := err.(*WorkerError)
	if !ok {
		t.Fatalf("Expected WorkerError, got %T", err)
	}
	if we.Code != ErrCodeNotFound {
		t.Fatalf("Expected NOT_FOUND error, got %s", we.Code)
	}
}

func TestPoolList(t *testing.T) {
	pool := NewPool()

	if err := pool.Register(newTestWorker("w1", "test")); err != nil {
		t.Fatalf("Register w1 failed: %v", err)
	}
	if err := pool.Register(newTestWorker("w2", "test")); err != nil {
		t.Fatalf("Register w2 failed: %v", err)
	}

	workers := pool.List()
	if len(workers) != 2 {
		t.Fatalf("Expected 2 workers, got %d", len(workers))
	}
}

func TestPoolUnregister(t *testing.T) {
	pool := NewPool()

	w := newTestWorker("test-1", "test")
	if err := pool.Register(w); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if err := pool.Unregister("test-1"); err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	if pool.Size() != 0 {
		t.Fatalf("Expected pool size 0, got %d", pool.Size())
	}

	// Should fail on nonexistent
	if err := pool.Unregister("test-1"); err == nil {
		t.Fatal("Expected error on nonexistent unregister")
	}
}

func TestPoolHealth(t *testing.T) {
	pool := NewPool()

	w1 := newTestWorker("w1", "test")
	w2 := newTestWorker("w2", "test")

	if err := pool.Register(w1); err != nil {
		t.Fatalf("Register w1 failed: %v", err)
	}
	if err := pool.Register(w2); err != nil {
		t.Fatalf("Register w2 failed: %v", err)
	}

	ctx := context.Background()
	status := pool.Health(ctx)

	if status.Status != HealthStatusHealthy {
		t.Fatalf("Expected healthy status, got %s", status.Status)
	}
	if len(status.Details) != 2 {
		t.Fatalf("Expected 2 details, got %d", len(status.Details))
	}
}
