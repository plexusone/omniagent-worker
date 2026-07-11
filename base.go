package worker

import (
	"context"
	"log/slog"

	"github.com/plexusone/omniobserve/agentops"
)

// BaseWorker provides common functionality for workers.
// Embed this in your worker implementation to get default behavior.
type BaseWorker struct {
	config   WorkerConfig
	agentOps agentops.Store
	logger   *slog.Logger
}

// NewBaseWorker creates a new BaseWorker with the given configuration.
func NewBaseWorker(cfg WorkerConfig) BaseWorker {
	cfg.Defaults()

	var store agentops.Store
	if cfg.AgentOps != nil && cfg.AgentOps.Enabled && cfg.AgentOps.Store != nil {
		store = cfg.AgentOps.Store
	}

	return BaseWorker{
		config:   cfg,
		agentOps: store,
		logger:   cfg.Logger.With("worker_id", cfg.ID, "worker_type", cfg.Type),
	}
}

// ID returns the worker ID.
func (b *BaseWorker) ID() string {
	return b.config.ID
}

// Type returns the worker type.
func (b *BaseWorker) Type() string {
	return b.config.Type
}

// Version returns the worker version.
func (b *BaseWorker) Version() string {
	return b.config.Version
}

// Config returns the worker configuration.
func (b *BaseWorker) Config() WorkerConfig {
	return b.config
}

// Logger returns the worker's logger.
func (b *BaseWorker) Logger() *slog.Logger {
	return b.logger
}

// AgentOps returns the AgentOps store, or nil if not configured.
func (b *BaseWorker) AgentOps() agentops.Store {
	return b.agentOps
}

// Init is a no-op default implementation.
// Override this in your worker if initialization is needed.
func (b *BaseWorker) Init(ctx context.Context) error {
	b.logger.Info("worker initialized")
	return nil
}

// Shutdown is a no-op default implementation.
// Override this in your worker if cleanup is needed.
func (b *BaseWorker) Shutdown(ctx context.Context) error {
	b.logger.Info("worker shutdown")
	return nil
}

// Health returns a healthy status by default.
// Override this in your worker for custom health checks.
func (b *BaseWorker) Health(ctx context.Context) HealthStatus {
	return HealthStatus{
		Status: HealthStatusHealthy,
	}
}

// Execute is not implemented in BaseWorker.
// You must implement this in your worker.
func (b *BaseWorker) Execute(ctx context.Context, req *Request) (*Response, error) {
	panic("Execute not implemented - override in your worker")
}
