package worker

import (
	"log/slog"
	"time"

	"github.com/plexusone/omniobserve/agentops"
)

// WorkerConfig configures a worker.
type WorkerConfig struct {
	// ID is the unique identifier for this worker instance.
	ID string

	// Type is the worker type (e.g., "research", "synthesis").
	Type string

	// Version is the worker implementation version.
	Version string

	// LLM configures the LLM client for workers that need it.
	// Optional - not all workers require LLM.
	LLM *LLMConfig

	// AgentOps configures observability.
	AgentOps *AgentOpsConfig

	// Logger is the logger for this worker.
	// If nil, a default logger is used.
	Logger *slog.Logger

	// Timeout is the maximum execution time for a single request.
	// Default: 60 seconds.
	Timeout time.Duration
}

// LLMConfig configures the LLM client for workers.
type LLMConfig struct {
	// Provider is the LLM provider (e.g., "anthropic", "openai", "gemini").
	Provider string

	// Model is the model name (e.g., "claude-sonnet-4-20250514").
	Model string

	// APIKey is the API key for the provider.
	// If empty, will be read from environment.
	APIKey string

	// BaseURL is an optional custom base URL for the provider.
	BaseURL string

	// Temperature controls randomness (0.0-1.0).
	Temperature float64

	// MaxTokens is the maximum tokens in the response.
	MaxTokens int
}

// AgentOpsConfig configures AgentOps observability.
type AgentOpsConfig struct {
	// Enabled controls whether AgentOps tracing is active.
	Enabled bool

	// Store is the AgentOps store for persisting traces.
	// If nil and Enabled is true, a no-op store is used.
	Store agentops.Store
}

// CoordinatorConfig configures a coordinator.
type CoordinatorConfig struct {
	// ID is the unique identifier for this coordinator.
	ID string

	// Workflow is the workflow executor for orchestration.
	// If nil, a simple sequential executor is used.
	Workflow WorkflowExecutor

	// AgentOps configures observability.
	AgentOps *AgentOpsConfig

	// Logger is the logger for this coordinator.
	Logger *slog.Logger

	// Timeout is the maximum execution time for a workflow.
	// Default: 5 minutes.
	Timeout time.Duration

	// MaxRetries is the maximum number of retries for failed tasks.
	// Default: 3.
	MaxRetries int
}

// Defaults applies default values to the config.
func (c *WorkerConfig) Defaults() {
	if c.Version == "" {
		c.Version = "1.0.0"
	}
	if c.Timeout == 0 {
		c.Timeout = 60 * time.Second
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
}

// Defaults applies default values to the config.
func (c *CoordinatorConfig) Defaults() {
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Minute
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
}
