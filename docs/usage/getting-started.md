# Getting Started

## Overview & Features

### Detailed Description

`tasker` is a Go library designed to simplify the management of asynchronous tasks that require access to potentially limited or expensive resources. It provides a highly concurrent and scalable framework for processing jobs, abstracting away the complexities of goroutine management, worker pools, resource lifecycles, and dynamic scaling.

At its core, `tasker` allows you to define custom "resources" (e.g., database connections, API clients, CPU/GPU compute units) and associate them with "workers." These workers pick up tasks from a queue, execute them using an available resource, and return results. The library excels in scenarios where you need to control concurrency, ensure resource availability, prioritize certain tasks, and dynamically adjust processing capacity based on demand.

### Key Features

*   **Generic Task & Resource Management**: Type-safe handling of any custom resource (`R`) and task result (`E`) type.
*   **Customizable Resource Lifecycle**: Define `OnCreate` and `OnDestroy` functions to manage the setup and teardown of your specific resources (e.g., opening/closing database connections).
*   **Worker Pooling**: Maintain a configurable number of base workers to handle a steady load.
*   **Dynamic Worker Scaling (Bursting)**: Automatically scale up (add "burst" workers) when the task queue backlog exceeds a configurable threshold, and scale down when demand subsides.
*   **Priority Queuing**: Submit high-priority tasks to a dedicated queue that is always processed before standard tasks, ensuring critical operations are handled swiftly.
*   **Resource Pooling for Immediate Tasks (`RunTask`)**: Execute urgent, synchronous tasks immediately by borrowing a resource from a pre-allocated pool or creating a temporary one on demand, bypassing queues.
*   **Customizable Health Checks**: Implement a `CheckHealth` function to determine if an error during task execution indicates an unhealthy worker or resource, triggering replacement and task retries.
*   **Task Retries**: Configure maximum retries for tasks that fail due to an unhealthy worker/resource.
*   **Graceful Shutdown**: Safely stop the task manager, allowing in-flight tasks to complete and resources to be properly released.
*   **Real-time Statistics**: Obtain snapshots of active workers, queued tasks, and available resources for monitoring and debugging.

## Installation & Setup

### Prerequisites

*   Go 1.22 or higher

### Installation Steps

To integrate `tasker` into your Go project, use `go get`:

```bash
go get github.com/asaidimu/tasker
```

### Configuration

`tasker` is configured via the `tasker.Config[R]` struct. You instantiate this struct with your desired settings and pass it to `tasker.NewRunner`.

```go
type Config[R any] struct {
	OnCreate        func() (R, error)      // Required: Function to create a new resource.
	OnDestroy       func(R) error          // Required: Function to destroy/cleanup a resource.
	WorkerCount     int                    // Required: Initial and minimum number of base workers. Must be > 0.
	Ctx             context.Context        // Required: Parent context for the Runner; cancellation initiates graceful shutdown.
	CheckHealth     func(error) bool       // Optional: Custom health check for task errors. Default: always healthy.
	BurstTaskThreshold  int                    // Optional: Queue size to trigger bursting. 0 to disable.
	BurstWorkerCount      int                    // Optional: Number of burst workers to add at once. Defaults to 2 if <= 0.
	MaxWorkerCount  int                    // Optional: Maximum total workers allowed. Defaults to WorkerCount + BurstWorkerCount if <= 0.
	BurstInterval   time.Duration          // Optional: Frequency for burst checks. Default: 100ms.
	MaxRetries      int                    // Optional: Max retries for unhealthy tasks. Default: 3.
	ResourcePoolSize int                   // Optional: Size of the pool for `RunTask`. Default: `WorkerCount`.
}
```

## First Tasks: Basic Usage - Simple Calculator

This example demonstrates setting up a basic `tasker.Runner` with two workers for simple arithmetic tasks.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker"
)

// CalculatorResource represents a simple resource, just a placeholder.
type CalculatorResource struct{}

// onCreate for CalculatorResource - no actual setup needed
func createCalcResource() (*CalculatorResource, error) {
	fmt.Println("INFO: Creating CalculatorResource")
	return &CalculatorResource{}, nil
}

// onDestroy for CalculatorResource - no actual cleanup needed
func destroyCalcResource(r *CalculatorResource) error {
	fmt.Println("INFO: Destroying CalculatorResource")
	return nil
}

func main() {
	fmt.Println("--- Basic Usage: Simple Calculator ---")

	ctx := context.Background()

	// Configure the tasker for our CalculatorResource
	config := tasker.Config[*CalculatorResource]{
		OnCreate:    createCalcResource,
		OnDestroy:   destroyCalcResource,
		WorkerCount: 2, // Two base workers
		Ctx:         ctx,
	}

	// Create a new task manager. Tasks will return an int result.
	manager, err := tasker.NewRunner[*CalculatorResource, int](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop() // Ensure the manager is stopped gracefully

	fmt.Println("Queuing a simple addition task...")
	go func() {
		sum, err := manager.QueueTask(func(r *CalculatorResource) (int, error) {
			time.Sleep(50 * time.Millisecond) // Simulate some work
			a, b := 10, 25
			fmt.Printf("Worker processing: %d + %d\n", a, b)
			return a + b, nil
		})

		if err != nil {
			fmt.Printf("Task 1 failed: %v\n", err)
		} else {
			fmt.Printf("Task 1 (Addition) Result: %d\n", sum)
		}
	}()

	fmt.Println("Queuing another subtraction task...")
	go func() {
		difference, err := manager.QueueTask(func(r *CalculatorResource) (int, error) {
			time.Sleep(70 * time.Millisecond) // Simulate some work
			a, b := 100, 40
			fmt.Printf("Worker processing: %d - %d\n", a, b)
			return a - b, nil
		})

		if err != nil {
			fmt.Printf("Task 2 failed: %v\n", err)
		} else {
			fmt.Printf("Task 2 (Subtraction) Result: %d\n", difference)
		}
	}()

	// Allow some time for tasks to complete
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Basic usage example finished.")
}
```

**Expected Output (Illustrative, timing and worker IDs may vary):**

```
--- Basic Usage: Simple Calculator ---
INFO: Creating CalculatorResource
INFO: Creating CalculatorResource
Queuing a simple addition task...
Queuing another subtraction task...
Worker processing: 10 + 25
Worker processing: 100 - 40
Task 1 (Addition) Result: 35
Task 2 (Subtraction) Result: 60
Basic usage example finished.
INFO: Destroying CalculatorResource
INFO: Destroying CalculatorResource
```

**Common Decision Patterns**

*   **When to initialize `tasker.Runner`**: At application startup, once per application lifecycle, or when a specific module requiring concurrent processing is activated.
*   **How to ensure graceful shutdown**: Always pair `tasker.NewRunner` with `defer manager.Stop()` in `main` or the top-level goroutine managing the `tasker` instance.
*   **Choosing `WorkerCount`**: Start with a number that matches your typical concurrent resource needs or CPU cores, then adjust based on profiling.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF application_starts_up THEN initialize_task_manager",
    "IF task_manager_is_created THEN DEFER call_manager.Stop()",
    "IF primary_goal IS asynchronous_processing THEN USE QueueTask(task_func)"
  ],
  "verificationSteps": [
    "Check: `go get github.com/asaidimu/tasker` completes without error -> Expected: Package installed",
    "Check: `manager, err := tasker.NewRunner(...)` returns `nil` for `err` -> Expected: Manager initialized successfully",
    "Check: Output contains 'INFO: Creating CalculatorResource' multiple times -> Expected: Resources are being created for workers"
  ],
  "quickPatterns": [
    "Pattern: Basic `Runner` Initialization\n```go\nimport (\n\t\"context\"\n\t\"github.com/asaidimu/tasker\"\n)\n\ntype MyResource struct{}\nfunc createRes() (*MyResource, error) { return &MyResource{}, nil }\nfunc destroyRes(r *MyResource) error { return nil }\n\nfunc main() {\n\tconfig := tasker.Config[*MyResource]{\n\t\tOnCreate: createRes,\n\t\tOnDestroy: destroyRes,\n\t\tWorkerCount: 2,\n\t\tCtx: context.Background(),\n\t}\n\tmanager, err := tasker.NewRunner[*MyResource, any](config)\n\tif err != nil { /* handle error */ }\n\tdefer manager.Stop()\n}\n```",
    "Pattern: Queueing a Basic Task\n```go\n// manager is an initialized tasker.TaskManager\nresult, err := manager.QueueTask(func(res *MyResource) (string, error) {\n\t// Perform work with 'res'\n\treturn \"Task done\", nil\n})\nif err != nil { /* handle task error */ }\nfmt.Println(result)\n```"
  ],
  "diagnosticPaths": [
    "Error `WorkerCountMustBePositiveError` -> Symptom: `NewRunner` returns error about worker count -> Check: `Config.WorkerCount` is > 0 -> Fix: Set `Config.WorkerCount` to a positive integer.",
    "Error `OnCreateFunctionRequiredError` -> Symptom: `NewRunner` returns error about `OnCreate` -> Check: `Config.OnCreate` is not `nil` -> Fix: Provide a valid `OnCreate` function."
  ]
}
```

---
*Generated using Gemini AI on 6/14/2025, 1:47:53 PM. Review and refine as needed.*