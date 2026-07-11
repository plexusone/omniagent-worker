// Package worker provides building blocks for multi-agent systems in Go.
//
// omniagent-worker offers a minimal, composable framework for creating
// task-oriented agent teams with full observability support.
//
// # Core Concepts
//
// Worker is the fundamental interface for task-oriented agents. Workers are
// stateless request/response handlers that can be deployed in-process or as
// HTTP microservices.
//
// Coordinator manages teams of workers, orchestrating workflow execution
// with support for both deterministic (Eino) and LLM-driven coordination.
//
// Pool provides in-process worker management for embedded use cases,
// enabling lightweight deployment without HTTP overhead.
//
// # Observability
//
// All operations integrate with AgentOps (omniobserve) for OpenTelemetry-
// compatible tracing of workflows, tasks, and handoffs.
//
// # Example
//
//	coord := worker.NewCoordinator(worker.CoordinatorConfig{
//	    ID: "stats-team",
//	})
//	coord.Pool().Register(research.NewWorker())
//	coord.Pool().Register(synthesis.NewWorker())
//
//	resp, err := coord.Execute(ctx, &worker.CoordinatorRequest{
//	    Input: map[string]any{"topic": "climate change"},
//	})
package worker
