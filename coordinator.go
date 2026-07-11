package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/plexusone/omniobserve/agentops"
)

// Coordinator manages a team of workers and orchestrates workflow execution.
type Coordinator struct {
	id       string
	pool     *Pool
	clients  map[string]*WorkerClient
	workflow WorkflowExecutor
	agentOps agentops.Store
	logger   *slog.Logger
	config   CoordinatorConfig
}

// NewCoordinator creates a new coordinator.
func NewCoordinator(cfg CoordinatorConfig) *Coordinator {
	cfg.Defaults()

	var store agentops.Store
	if cfg.AgentOps != nil && cfg.AgentOps.Enabled && cfg.AgentOps.Store != nil {
		store = cfg.AgentOps.Store
	}

	return &Coordinator{
		id:       cfg.ID,
		pool:     NewPool(),
		clients:  make(map[string]*WorkerClient),
		workflow: cfg.Workflow,
		agentOps: store,
		logger:   cfg.Logger.With("coordinator_id", cfg.ID),
		config:   cfg,
	}
}

// ID returns the coordinator ID.
func (c *Coordinator) ID() string {
	return c.id
}

// Pool returns the in-process worker pool.
func (c *Coordinator) Pool() *Pool {
	return c.pool
}

// AddRemote adds a remote worker endpoint.
func (c *Coordinator) AddRemote(id, url string, opts ...ClientOption) {
	c.clients[id] = NewWorkerClient(url, opts...)
}

// RemoveRemote removes a remote worker endpoint.
func (c *Coordinator) RemoveRemote(id string) {
	delete(c.clients, id)
}

// CoordinatorRequest is the input to a coordinator.
type CoordinatorRequest struct {
	// WorkflowID is a unique identifier for this workflow instance.
	// If empty, one will be generated.
	WorkflowID string `json:"workflow_id,omitempty"`

	// Input contains workflow-specific input data.
	Input map[string]any `json:"input"`
}

// CoordinatorResponse is the output from a coordinator.
type CoordinatorResponse struct {
	// WorkflowID echoes back the workflow ID.
	WorkflowID string `json:"workflow_id"`

	// Output contains workflow output data.
	Output map[string]any `json:"output"`

	// Stats contains execution statistics.
	Stats WorkflowStats `json:"stats"`

	// Error contains error information if the workflow failed.
	Error *WorkerError `json:"error,omitempty"`
}

// WorkflowStats contains execution statistics.
type WorkflowStats struct {
	// Duration is the total workflow execution time.
	Duration time.Duration `json:"duration"`

	// TaskCount is the number of tasks executed.
	TaskCount int `json:"task_count"`

	// SuccessCount is the number of successful tasks.
	SuccessCount int `json:"success_count"`

	// FailureCount is the number of failed tasks.
	FailureCount int `json:"failure_count"`

	// RetryCount is the number of retried tasks.
	RetryCount int `json:"retry_count"`
}

// Execute runs the workflow with the given input.
func (c *Coordinator) Execute(ctx context.Context, req *CoordinatorRequest) (*CoordinatorResponse, error) {
	startTime := time.Now()

	// Generate workflow ID if not provided
	workflowID := req.WorkflowID
	if workflowID == "" {
		workflowID = generateID()
	}

	c.logger.Info("starting workflow",
		"workflow_id", workflowID,
		"input_keys", mapKeys(req.Input),
	)

	// Create dispatcher for workflow execution
	dispatcher := &tracingDispatcher{
		pool:       c.pool,
		clients:    c.clients,
		agentOps:   c.agentOps,
		workflowID: workflowID,
		logger:     c.logger,
	}

	// Execute workflow
	var output map[string]any
	var workflowErr error

	if c.workflow != nil {
		output, workflowErr = c.workflow.Execute(ctx, req.Input, dispatcher)
	} else {
		// No workflow defined - just return input as output
		output = req.Input
	}

	duration := time.Since(startTime)

	c.logger.Info("workflow completed",
		"workflow_id", workflowID,
		"duration", duration,
		"task_count", dispatcher.stats.TaskCount,
		"success_count", dispatcher.stats.SuccessCount,
		"failure_count", dispatcher.stats.FailureCount,
		"error", workflowErr,
	)

	resp := &CoordinatorResponse{
		WorkflowID: workflowID,
		Output:     output,
		Stats: WorkflowStats{
			Duration:     duration,
			TaskCount:    dispatcher.stats.TaskCount,
			SuccessCount: dispatcher.stats.SuccessCount,
			FailureCount: dispatcher.stats.FailureCount,
			RetryCount:   dispatcher.stats.RetryCount,
		},
	}

	if workflowErr != nil {
		resp.Error = AsWorkerError(workflowErr)
		return resp, workflowErr
	}

	return resp, nil
}

// Init initializes the coordinator and all pool workers.
func (c *Coordinator) Init(ctx context.Context) error {
	c.logger.Info("initializing coordinator")
	return c.pool.Init(ctx)
}

// Shutdown shuts down the coordinator and all pool workers.
func (c *Coordinator) Shutdown(ctx context.Context) error {
	c.logger.Info("shutting down coordinator")
	return c.pool.Shutdown(ctx)
}

// Health returns the aggregate health of the coordinator and workers.
func (c *Coordinator) Health(ctx context.Context) HealthStatus {
	return c.pool.Health(ctx)
}

// tracingDispatcher wraps dispatch calls with tracing.
type tracingDispatcher struct {
	pool       *Pool
	clients    map[string]*WorkerClient
	agentOps   agentops.Store
	workflowID string
	logger     *slog.Logger
	stats      dispatcherStats
}

type dispatcherStats struct {
	TaskCount    int
	SuccessCount int
	FailureCount int
	RetryCount   int
}

// Dispatch calls a worker, first checking the pool, then remote clients.
func (d *tracingDispatcher) Dispatch(ctx context.Context, workerID string, req *Request) (*Response, error) {
	d.stats.TaskCount++

	// Set workflow context on request
	req.WorkflowID = d.workflowID
	req.TaskID = generateID()

	d.logger.Debug("dispatching to worker",
		"worker_id", workerID,
		"request_id", req.ID,
		"task_id", req.TaskID,
	)

	var resp *Response
	var err error

	// Try pool first
	if w, ok := d.pool.Get(workerID); ok {
		resp, err = w.Execute(ctx, req)
	} else if client, ok := d.clients[workerID]; ok {
		// Try remote client
		resp, err = client.Execute(ctx, req)
	} else {
		err = NewNotFoundError(fmt.Sprintf("worker %q not found", workerID))
	}

	if err != nil {
		d.stats.FailureCount++
		d.logger.Error("worker dispatch failed",
			"worker_id", workerID,
			"request_id", req.ID,
			"error", err,
		)
		return nil, err
	}

	if resp.Error != nil {
		d.stats.FailureCount++
		return resp, resp.Error
	}

	d.stats.SuccessCount++
	return resp, nil
}

// mapKeys returns the keys of a map for logging.
func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
