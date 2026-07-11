package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WorkerClient is an HTTP client for calling remote workers.
type WorkerClient struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

// ClientOption configures a WorkerClient.
type ClientOption func(*WorkerClient)

// NewWorkerClient creates a new HTTP client for a remote worker.
func NewWorkerClient(baseURL string, opts ...ClientOption) *WorkerClient {
	c := &WorkerClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		headers: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *WorkerClient) {
		c.httpClient.Timeout = d
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *WorkerClient) {
		c.httpClient = client
	}
}

// WithHeader adds a default header to all requests.
func WithHeader(key, value string) ClientOption {
	return func(c *WorkerClient) {
		c.headers[key] = value
	}
}

// Execute calls the remote worker's execute endpoint.
func (c *WorkerClient) Execute(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/execute", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add default headers
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	// Add AgentOps context headers
	if req.WorkflowID != "" {
		httpReq.Header.Set("X-AgentOps-Workflow-ID", req.WorkflowID)
	}
	if req.TaskID != "" {
		httpReq.Header.Set("X-AgentOps-Task-ID", req.TaskID)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, httpResp.Body)
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("worker returned status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var resp Response
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}

// Health checks the health of the remote worker.
func (c *WorkerClient) Health(ctx context.Context) (HealthStatus, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return HealthStatus{Status: HealthStatusUnhealthy}, fmt.Errorf("failed to create request: %w", err)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return HealthStatus{Status: HealthStatusUnhealthy}, fmt.Errorf("health check failed: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, httpResp.Body)
		_ = httpResp.Body.Close()
	}()

	var status HealthStatus
	if err := json.NewDecoder(httpResp.Body).Decode(&status); err != nil {
		return HealthStatus{Status: HealthStatusUnhealthy}, fmt.Errorf("failed to decode health: %w", err)
	}

	return status, nil
}
