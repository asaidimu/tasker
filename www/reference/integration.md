---
outline: "deep"
lastUpdated: true
editLink: true
sidebar: true
title: "Integration Reference"
description: "Documentation on integration aspects, requirements, and patterns"
---
# Integration

## Environment Requirements

Go **1.24.3** or higher is required for building and running applications using `tasker`. The library is platform-agnostic and should run on any operating system supported by Go.

## Initialization Patterns

### The most common way to initialize `tasker` is by providing `OnCreate`, `OnDestroy`, `WorkerCount`, and a `context.Context`.



```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

// MyResource represents a resource for tasks
type MyResource struct{ ID int }

// onCreate simulates resource allocation
func createMyResource() (*MyResource, error) {
	fmt.Println("INFO: Creating MyResource")
	return &MyResource{ID: time.Now().Nanosecond()}, nil
}

// onDestroy simulates resource deallocation
func destroyMyResource(r *MyResource) error {
	fmt.Println("INFO: Destroying MyResource", r.ID)
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := tasker.Config[*MyResource]{
		OnCreate:    createMyResource,
		OnDestroy:   destroyMyResource,
		WorkerCount: 2, // Two base workers
		Ctx:         ctx,
	}

	manager, err := tasker.NewTaskManager[*MyResource, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	fmt.Println("TaskManager initialized and running.")
	// Application logic here
	time.Sleep(500 * time.Millisecond)
}

```


### Integrate custom logging and metrics collection by implementing the `tasker.Logger` and `tasker.MetricsCollector` interfaces.



```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

// MyCustomLogger implements tasker.Logger
type MyCustomLogger struct{}

func (l *MyCustomLogger) Debugf(format string, args ...any) { log.Printf("[DEBUG] "+format, args...) }
func (l *MyCustomLogger) Infof(format string, args ...any)  { log.Printf("[INFO] "+format, args...) }
func (l *MyCustomLogger) Warnf(format string, args ...any)  { log.Printf("[WARN] "+format, args...) }
func (l *MyCustomLogger) Errorf(format string, args ...any) { log.Printf("[ERROR] "+format, args...) }

// MyCustomMetricsCollector implements tasker.MetricsCollector
type MyCustomMetricsCollector struct{}

func (c *MyCustomMetricsCollector) RecordArrival()       { fmt.Println("Metric: Task Arrived") }
func (c *MyCustomMetricsCollector) RecordCompletion(s tasker.TaskLifecycleTimestamps) { fmt.Printf("Metric: Task Completed (exec: %v)\n", s.FinishedAt.Sub(s.StartedAt)) }
func (c *MyCustomMetricsCollector) RecordFailure(s tasker.TaskLifecycleTimestamps)    { fmt.Println("Metric: Task Failed") }
func (c *MyCustomMetricsCollector) RecordRetry()         { fmt.Println("Metric: Task Retried") }
func (c *MyCustomMetricsCollector) Metrics() tasker.TaskMetrics { return tasker.TaskMetrics{} }

// MyResource and lifecycle functions (same as above)
type MyResource struct{ ID int }
func createMyResource() (*MyResource, error) { fmt.Println("INFO: Creating MyResource"); return &MyResource{ID: time.Now().Nanosecond()}, nil }
func destroyMyResource(r *MyResource) error { fmt.Println("INFO: Destroying MyResource", r.ID); return nil }

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := tasker.Config[*MyResource]{
		OnCreate:    createMyResource,
		OnDestroy:   destroyMyResource,
		WorkerCount: 1,
		Ctx:         ctx,
		Logger:      &MyCustomLogger{}, // Provide custom logger
		Collector:   &MyCustomMetricsCollector{}, // Provide custom collector
	}

	manager, err := tasker.NewTaskManager[*MyResource, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	_, _ = manager.QueueTask(func(ctx context.Context, res *MyResource) (string, error) {
		fmt.Printf("Worker %d processing a task with custom logging/metrics\n", res.ID)
		return "done", nil
	})

	time.Sleep(100 * time.Millisecond)
}

```


## Common Pitfalls

### Not handling `context.Context` cancellation within task functions.

**Solution**: Tasks should periodically check `ctx.Done()` or use `select { case <-ctx.Done(): return nil, ctx.Err() ... }` to respond to manager shutdowns or timeouts.

### Blocking indefinitely in `OnCreate`, `OnDestroy`, or task functions.

**Solution**: Ensure these functions complete in a timely manner. Long-running initialization/cleanup should be handled asynchronously or outside `tasker`'s direct control if it risks blocking the pool.

### Ignoring errors returned by `QueueTask` or `RunTask`.

**Solution**: Always check the `error` return value to handle potential submission failures (e.g., manager shutting down) or task execution errors.

## Lifecycle Dependencies

The `TaskManager` relies on a parent `context.Context` for its overall lifecycle. Cancelling this context (passed via `Config.Ctx`) initiates a graceful shutdown (`Stop()` behavior). Users must ensure the `TaskManager`'s `Stop()` or `Kill()` method is called when the application exits to properly release resources and wait for tasks (for `Stop()`). Resource creation (`OnCreate`) and destruction (`OnDestroy`) are directly tied to worker lifecycles and the `RunTask` operation. Workers will not start if `OnCreate` fails, and resources are destroyed when workers exit or the manager shuts down.

