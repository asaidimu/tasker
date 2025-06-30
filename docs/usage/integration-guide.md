# Integration Guide

## Environment Requirements

To use Tasker, you need a Go development environment with Go version 1.24.3 or higher. Tasker itself is cross-platform, so it runs wherever Go is supported (Linux, Windows, macOS, etc.). No special compiler settings are required beyond standard Go build practices.

## Initialization Patterns

### Standard initialization of Tasker with custom resource lifecycle functions and basic worker configuration. This pattern demonstrates the minimum required setup for a functional TaskManager.
```[DETECTED_LANGUAGE]
package main

import (
	"context"
	"log"
	"time"

	"github.com/asaidimu/tasker"
)

// Define your custom resource type
type DatabaseConnection struct { ID int }

// onCreate: Function to create a new database connection
func createDBConnection() (*DatabaseConnection, error) {
	log.Println("INFO: Creating DatabaseConnection")
	// Simulate connecting to a database
	time.Sleep(10 * time.Millisecond)
	return &DatabaseConnection{ID: 123}, nil
}

// onDestroy: Function to close the database connection
func destroyDBConnection(conn *DatabaseConnection) error {
	log.Printf("INFO: Destroying DatabaseConnection %d\n", conn.ID)
	// Simulate closing the connection
	return nil
}

func main() {
	// Create a background context for the TaskManager
	ctx := context.Background()

	// Configure the TaskManager
	config := tasker.Config[*DatabaseConnection]{
		OnCreate:    createDBConnection,    // Required: function to create resource
		OnDestroy:   destroyDBConnection,   // Required: function to destroy resource
		WorkerCount: 5,                     // Required: number of base workers
		Ctx:         ctx,                   // Required: parent context for lifecycle
		// Optional fields:
		// MaxWorkerCount: 10,
		// BurstInterval: 100 * time.Millisecond,
		// CheckHealth: func(err error) bool { return true },
		// MaxRetries: 3,
		// ResourcePoolSize: 5,
		// Logger: &myCustomLogger{},
		// Collector: &myCustomMetricsCollector{},
	}

	// Create a new TaskManager instance
	manager, err := tasker.NewTaskManager[*DatabaseConnection, string](config) // Tasks will return string results
	if err != nil {
		log.Fatalf("Failed to create TaskManager: %v", err)
	}

	// Ensure graceful shutdown when main exits
	defer manager.Stop()

	log.Println("TaskManager initialized and running. Add tasks here.")
	// Example task (non-blocking for main)
	go func() {
		_, err := manager.QueueTask(func(db *DatabaseConnection) (string, error) {
			log.Printf("Worker processing database query with connection %d\n", db.ID)
			time.Sleep(50 * time.Millisecond)
			return "Query result", nil
		})
		if err != nil { log.Printf("Task failed: %v\n", err) }
		else { log.Println("Task completed.") }
	}()

	time.Sleep(200 * time.Millisecond) // Allow time for tasks
}

```

## Common Integration Pitfalls

- **Issue**: Blocking operations in `OnCreate` or `OnDestroy`
  - **Solution**: `OnCreate` and `OnDestroy` functions should be non-blocking and execute quickly. Blocking operations can delay worker startup/shutdown, leading to performance issues or graceful shutdown hangs. If an operation *must* block (e.g., waiting for an external service to become available on startup), consider handling it with timeouts or asynchronous initialization outside these functions where possible. Make sure `OnDestroy` never deadlocks or waits indefinitely.

- **Issue**: Task functions that do not respect context cancellation
  - **Solution**: For long-running tasks, if you want them to be interruptible during `manager.Stop()` or `manager.Kill()`, your task function's internal logic needs to periodically check a `context.Context.Done()` channel and return if it's closed. Tasker itself doesn't directly pass a context to your `func(R) (E, error)`, so you must manage context propagation within your application (e.g., by capturing a context in a closure).

- **Issue**: Incorrect `CheckHealth` logic leading to worker thrashing
  - **Solution**: If your `CheckHealth` function returns `false` for every task error (even transient ones), it can cause workers to be constantly replaced (`OnDestroy` then `OnCreate`), leading to high resource churn and degraded performance. Ensure `CheckHealth` returns `false` only for errors that truly indicate an unhealthy, unrecoverable worker or resource state, not for recoverable task-specific failures.

- **Issue**: Queueing tasks after manager shutdown
  - **Solution**: Attempting to call `QueueTask`, `RunTask`, etc., after `manager.Stop()` or `manager.Kill()` has been invoked will immediately return an error (`task manager is shutting down`). Ensure your application's task submission logic is aware of the TaskManager's lifecycle and ceases submissions during shutdown.

## Lifecycle Dependencies

The Tasker's lifecycle is managed by the `context.Context` provided in `Config.Ctx`. When this context is cancelled (either explicitly by your application or implicitly by `manager.Stop()`/`manager.Kill()`), it signals all internal goroutines (workers, burst manager) to begin their shutdown procedures. Workers will call `OnDestroy` on their associated resources as they exit. The `NewTaskManager` function itself calls `OnCreate` to populate the initial `resourcePool` and for each base worker, so `OnCreate` must be ready before `NewTaskManager` is called.



---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*