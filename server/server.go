// Package server provides HTTP server functionality for workers and coordinators.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	worker "github.com/plexusone/omniagent-worker"
)

// Config configures the HTTP server.
type Config struct {
	// Port is the port to listen on.
	Port int

	// ReadTimeout is the maximum duration for reading the request.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration for writing the response.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum duration to wait for the next request.
	IdleTimeout time.Duration

	// Logger is the logger for the server.
	Logger *slog.Logger
}

// Defaults applies default values to the config.
func (c *Config) Defaults() {
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 30 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 60 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 120 * time.Second
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
}

// Run starts an HTTP server for a worker or coordinator.
func Run(ctx context.Context, handler any, cfg Config) error {
	cfg.Defaults()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.WriteTimeout))

	// Register routes based on handler type
	switch h := handler.(type) {
	case worker.Worker:
		registerWorkerRoutes(r, h, cfg.Logger)
	case *worker.Coordinator:
		registerCoordinatorRoutes(r, h, cfg.Logger)
	default:
		return fmt.Errorf("unsupported handler type: %T", handler)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	cfg.Logger.Info("starting server", "port", cfg.Port)

	// Handle graceful shutdown
	errChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		cfg.Logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

// registerWorkerRoutes registers HTTP routes for a worker.
func registerWorkerRoutes(r chi.Router, w worker.Worker, logger *slog.Logger) {
	r.Get("/health", func(rw http.ResponseWriter, req *http.Request) {
		status := w.Health(req.Context())
		writeJSON(rw, http.StatusOK, status)
	})

	r.Get("/ready", func(rw http.ResponseWriter, req *http.Request) {
		status := w.Health(req.Context())
		if status.Status == worker.HealthStatusUnhealthy {
			writeJSON(rw, http.StatusServiceUnavailable, status)
			return
		}
		writeJSON(rw, http.StatusOK, status)
	})

	r.Post("/execute", func(rw http.ResponseWriter, req *http.Request) {
		var workerReq worker.Request
		if err := json.NewDecoder(req.Body).Decode(&workerReq); err != nil {
			writeJSON(rw, http.StatusBadRequest, worker.NewValidationError("invalid request body"))
			return
		}

		// Extract AgentOps context from headers
		if wfID := req.Header.Get("X-AgentOps-Workflow-ID"); wfID != "" {
			workerReq.WorkflowID = wfID
		}
		if taskID := req.Header.Get("X-AgentOps-Task-ID"); taskID != "" {
			workerReq.TaskID = taskID
		}

		resp, err := w.Execute(req.Context(), &workerReq)
		if err != nil {
			logger.Error("execute failed", "error", err)
			writeJSON(rw, http.StatusInternalServerError, worker.AsWorkerError(err))
			return
		}

		writeJSON(rw, http.StatusOK, resp)
	})

	r.Get("/info", func(rw http.ResponseWriter, req *http.Request) {
		writeJSON(rw, http.StatusOK, map[string]string{
			"id":      w.ID(),
			"type":    w.Type(),
			"version": w.Version(),
		})
	})
}

// registerCoordinatorRoutes registers HTTP routes for a coordinator.
func registerCoordinatorRoutes(r chi.Router, c *worker.Coordinator, logger *slog.Logger) {
	r.Get("/health", func(rw http.ResponseWriter, req *http.Request) {
		status := c.Health(req.Context())
		writeJSON(rw, http.StatusOK, status)
	})

	r.Get("/ready", func(rw http.ResponseWriter, req *http.Request) {
		status := c.Health(req.Context())
		if status.Status == worker.HealthStatusUnhealthy {
			writeJSON(rw, http.StatusServiceUnavailable, status)
			return
		}
		writeJSON(rw, http.StatusOK, status)
	})

	r.Post("/execute", func(rw http.ResponseWriter, req *http.Request) {
		var coordReq worker.CoordinatorRequest
		if err := json.NewDecoder(req.Body).Decode(&coordReq); err != nil {
			writeJSON(rw, http.StatusBadRequest, worker.NewValidationError("invalid request body"))
			return
		}

		resp, err := c.Execute(req.Context(), &coordReq)
		if err != nil {
			logger.Error("execute failed", "error", err)
			// Still return the response with error info
			if resp != nil {
				writeJSON(rw, http.StatusInternalServerError, resp)
				return
			}
			writeJSON(rw, http.StatusInternalServerError, worker.AsWorkerError(err))
			return
		}

		writeJSON(rw, http.StatusOK, resp)
	})

	r.Get("/info", func(rw http.ResponseWriter, req *http.Request) {
		writeJSON(rw, http.StatusOK, map[string]any{
			"id":      c.ID(),
			"workers": c.Pool().IDs(),
		})
	})
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}
