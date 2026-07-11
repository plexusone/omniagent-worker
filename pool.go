package worker

import (
	"context"
	"fmt"
	"sync"
)

// Pool manages a collection of in-process workers.
type Pool struct {
	workers map[string]Worker
	mu      sync.RWMutex
}

// NewPool creates a new worker pool.
func NewPool() *Pool {
	return &Pool{
		workers: make(map[string]Worker),
	}
}

// Register adds a worker to the pool.
// Returns an error if a worker with the same ID already exists.
func (p *Pool) Register(w Worker) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.workers[w.ID()]; exists {
		return fmt.Errorf("worker %q already registered", w.ID())
	}

	p.workers[w.ID()] = w
	return nil
}

// Unregister removes a worker from the pool.
// Returns an error if the worker is not found.
func (p *Pool) Unregister(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.workers[id]; !exists {
		return fmt.Errorf("worker %q not found", id)
	}

	delete(p.workers, id)
	return nil
}

// Get retrieves a worker by ID.
func (p *Pool) Get(id string) (Worker, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w, ok := p.workers[id]
	return w, ok
}

// List returns all registered workers.
func (p *Pool) List() []Worker {
	p.mu.RLock()
	defer p.mu.RUnlock()

	workers := make([]Worker, 0, len(p.workers))
	for _, w := range p.workers {
		workers = append(workers, w)
	}
	return workers
}

// IDs returns the IDs of all registered workers.
func (p *Pool) IDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ids := make([]string, 0, len(p.workers))
	for id := range p.workers {
		ids = append(ids, id)
	}
	return ids
}

// Size returns the number of workers in the pool.
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.workers)
}

// Execute calls a worker in the pool.
// Returns an error if the worker is not found.
func (p *Pool) Execute(ctx context.Context, workerID string, req *Request) (*Response, error) {
	w, ok := p.Get(workerID)
	if !ok {
		return nil, NewNotFoundError(fmt.Sprintf("worker %q not found in pool", workerID))
	}

	return w.Execute(ctx, req)
}

// Init initializes all workers in the pool.
// Stops and returns the first error encountered.
func (p *Pool) Init(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for id, w := range p.workers {
		if err := w.Init(ctx); err != nil {
			return fmt.Errorf("failed to init worker %q: %w", id, err)
		}
	}
	return nil
}

// Shutdown shuts down all workers in the pool.
// Continues on error and returns the last error encountered.
func (p *Pool) Shutdown(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var lastErr error
	for id, w := range p.workers {
		if err := w.Shutdown(ctx); err != nil {
			lastErr = fmt.Errorf("failed to shutdown worker %q: %w", id, err)
		}
	}
	return lastErr
}

// Health returns the aggregate health of all workers.
// Returns "unhealthy" if any worker is unhealthy.
// Returns "degraded" if any worker is degraded.
// Returns "healthy" only if all workers are healthy.
func (p *Pool) Health(ctx context.Context) HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	details := make(map[string]string)
	overallStatus := HealthStatusHealthy

	for id, w := range p.workers {
		status := w.Health(ctx)
		details[id] = status.Status

		switch status.Status {
		case HealthStatusUnhealthy:
			overallStatus = HealthStatusUnhealthy
		case HealthStatusDegraded:
			if overallStatus != HealthStatusUnhealthy {
				overallStatus = HealthStatusDegraded
			}
		}
	}

	return HealthStatus{
		Status:  overallStatus,
		Details: details,
	}
}
