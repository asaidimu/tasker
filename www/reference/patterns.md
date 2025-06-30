---
outline: "deep"
lastUpdated: true
editLink: true
sidebar: true
title: "Patterns Reference"
description: "Common patterns, examples, and validation approaches"
---
# Examples

## `Basic tasker setup`

Demonstrates the fundamental setup for `tasker` using a simple `CalculatorResource` with two base workers.



```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

// CalculatorResource represents a simple resource,
// in this case, just a placeholder.
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
		// No specific health check or burst settings for this basic example
	}

	// Create a new task manager
	manager, err := tasker.NewTaskManager[*CalculatorResource, int](config) // Tasks will return an int result
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop() // Ensure the manager is stopped gracefully

	fmt.Println("Queuing a simple Multiplication task...")
	task1Start := time.Now()
	// Queue a task to perform addition
	go func() {
		sum, err := manager.QueueTask(func(ctx context.Context, r *CalculatorResource) (int, error) {
			// In a real scenario, 'r' could be a connection to a math service
			time.Sleep(50 * time.Millisecond) // Simulate some work
			a, b := 10, 25
			fmt.Printf("Worker processing: %d * %d\n", a, b)
			return a * b, nil
		})

		if err != nil {
			fmt.Printf("Task 1 failed: %v\n", err)
		} else {
			fmt.Printf("Task 1 (Multiplication) Result: %d (took %s)\n", sum, time.Since(task1Start))
		}
	}()

	fmt.Println("Queuing another addition task...")
	task2Start := time.Now()
	manager.QueueTaskWithCallback(
		func(ctx context.Context, r *CalculatorResource) (int, error) {
			time.Sleep(50 * time.Millisecond) // Simulate some work
			a, b := 10, 25
			fmt.Printf("Worker processing: %d + %d\n", a, b)
			return a + b, nil
		},
		func(sum int, err error) { // do something with the results
			if err != nil {
				fmt.Printf("Task 2 failed: %v\n", err)
			} else {
				fmt.Printf("Task 2 (Addition) Result: %d (took %s)\n", sum, time.Since(task2Start))
			}
		},
	)

	fmt.Println("Queuing another subtraction task...")

	task3Start := time.Now()
	differencech, errch := manager.QueueTaskAsync(func(ctx context.Context, r *CalculatorResource) (int, error) {
		time.Sleep(70 * time.Millisecond) // Simulate some work
		a, b := 100, 40
		fmt.Printf("Worker processing: %d - %d\n", a, b)
		return a - b, nil
	})

	difference := <-differencech
	err = <-errch

	if err != nil {
		fmt.Printf("Task 3 failed: %v\n", err)
	} else {
		fmt.Printf("Task 3 (Subtraction) Result: %d (took %s)\n", difference, time.Since(task3Start))
	}

	// Allow some time for tasks to complete
	time.Sleep(500 * time.Millisecond)

	stats := manager.Stats()
	fmt.Printf("\n--- Current Stats ---\n")
	fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers)
	fmt.Printf("Queued Tasks: %d\n", stats.QueuedTasks)
	fmt.Printf("Available Resources: %d\n", stats.AvailableResources)
	fmt.Println("----------------------")

	fmt.Println("Basic usage example finished.")
}

```


Output log shows 'Creating CalculatorResource' twice, tasks being processed, and correct calculation results. Final stats show 2 active workers and 0 queued tasks.

**Related Methods**: `NewTaskManager`, `QueueTask`, `QueueTaskWithCallback`, `QueueTaskAsync`, `Stats`, `Stop`

---

## `Custom health check`

Demonstrates defining a custom health check function to identify unhealthy worker/resource states and trigger worker replacement and task retries.



```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/asaidimu/tasker/v2"
)

// ImageProcessor represents a resource for image manipulation.
type ImageProcessor struct {
	ID        int
	IsHealthy bool
}

// onCreate for ImageProcessor: simulates creating a connection to an image processing service.
func createImageProcessor() (*ImageProcessor, error) {
	id := rand.Intn(1000)
	fmt.Printf("INFO: Creating ImageProcessor %d\n", id)
	return &ImageProcessor{ID: id, IsHealthy: true}, nil
}

// onDestroy for ImageProcessor: simulates closing the connection.
func destroyImageProcessor(p *ImageProcessor) error {
	fmt.Printf("INFO: Destroying ImageProcessor %d\n", p.ID)
	return nil
}

// checkImageProcessorHealth: Custom health check.
// If the error is "processor_crash", consider the worker/resource unhealthy.
func checkImageProcessorHealth(err error) bool {
	if err != nil && err.Error() == "processor_crash" {
		fmt.Printf("WARN: Detected unhealthy error: %v. Worker will be replaced.\n", err)
		return false // This error indicates an unhealthy state
	}
	return true // Other errors are just task failures, not worker health issues
}

func main() {
	fmt.Println("\n--- Intermediate Usage: Image Processing ---")

	ctx := context.Background()

	config := tasker.Config[*ImageProcessor]{
		OnCreate:         createImageProcessor,
		OnDestroy:        destroyImageProcessor,
		WorkerCount:      2,
		Ctx:              ctx,
		CheckHealth:      checkImageProcessorHealth, // Use custom health check
		MaxRetries:       1,
		ResourcePoolSize: 1,
	}

	manager, err := tasker.NewTaskManager[*ImageProcessor, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	fmt.Println("Queueing a task that might crash a worker (unhealthy error)...")
	go func() {
		result, err := manager.QueueTask(func(ctx context.Context, proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing problematic image (might crash)\n", proc.ID)
			time.Sleep(100 * time.Millisecond)
			if rand.Intn(2) == 0 {
				return "", errors.New("processor_crash") // This triggers CheckHealth to return false
			}
			return "problematic_image_processed.jpg", nil
		})
		if err != nil {
			fmt.Printf("Problematic Image Task Failed: %v\n", err)
		} else {
			fmt.Printf("Problematic Image Task Completed: %s\n", result)
		}
	}()

	time.Sleep(500 * time.Millisecond)
}

```


When a task returns "processor_crash" error, the log should show 'WARN: Detected unhealthy error: processor_crash. Worker will be replaced.' followed by 'INFO: Destroying ImageProcessor [ID]' and 'INFO: Creating ImageProcessor [New ID]', indicating worker replacement. The task may be retried or fail with 'max retries exceeded'.

**Related Methods**: `QueueTask`

**Related Errors**: `processor_crash`, `max retries exceeded`

---

## `Graceful shutdown`

Demonstrates initiating a graceful shutdown of the `TaskManager` using `Stop()`, ensuring all pending tasks are completed before resources are released.



```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

type HeavyComputeResource struct{ ID int }
func createComputeResource() (*HeavyComputeResource, error) { fmt.Printf("INFO: Creating ComputeResource\n"); return &HeavyComputeResource{ID: 1}, nil }
func destroyComputeResource(r *HeavyComputeResource) error { fmt.Printf("INFO: Destroying ComputeResource\n"); return nil }

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := tasker.Config[*HeavyComputeResource]{
		OnCreate:    createComputeResource,
		OnDestroy:   destroyComputeResource,
		WorkerCount: 2,
		Ctx:         ctx,
	}

	manager, err := tasker.NewTaskManager[*HeavyComputeResource, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}

	// Queue some tasks
	for i := 0; i < 5; i++ {
		taskID := i
		go func() {
			_, _ = manager.QueueTask(func(ctx context.Context, res *HeavyComputeResource) (string, error) {
				fmt.Printf("Worker %d processing Task %d\n", res.ID, taskID)
				time.Sleep(100 * time.Millisecond)
				return fmt.Sprintf("Task %d completed", taskID), nil
			})
		}()
	}

	// Allow some tasks to start, then initiate graceful shutdown
	time.Sleep(200 * time.Millisecond)
	fmt.Println("\nInitiating graceful shutdown...")
	err = manager.Stop()
	if err != nil {
		fmt.Printf("Error during graceful shutdown: %v\n", err)
	} else {
		fmt.Println("Task manager gracefully shut down.")
	}
	fmt.Printf("Final Active Workers: %d\n", manager.Stats().ActiveWorkers)
}

```


All 5 tasks submitted are eventually reported as 'completed'. Logs indicate workers are destroyed only after tasks finish. Final 'Active Workers' should be 0. `Task manager gracefully shut down.` should be printed.

**Related Methods**: `Stop`

**Related Errors**: `task manager already stopping or killed`

---

## `Immediate shutdown`

Demonstrates initiating an immediate shutdown of the `TaskManager` using `Kill()`, which cancels active tasks and drops queued ones without waiting.



```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

type HeavyComputeResource struct{ ID int }
func createComputeResource() (*HeavyComputeResource, error) { fmt.Printf("INFO: Creating ComputeResource\n"); return &HeavyComputeResource{ID: 1}, nil }
func destroyComputeResource(r *HeavyComputeResource) error { fmt.Printf("INFO: Destroying ComputeResource\n"); return nil }

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := tasker.Config[*HeavyComputeResource]{
		OnCreate:    createComputeResource,
		OnDestroy:   destroyComputeResource,
		WorkerCount: 2,
		Ctx:         ctx,
	}

	manager, err := tasker.NewTaskManager[*HeavyComputeResource, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}

	// Queue a long-running task
	go func() {
		_, _ = manager.QueueTask(func(ctx context.Context, res *HeavyComputeResource) (string, error) {
			fmt.Printf("Worker %d processing long Task...\n", res.ID)
			select {
			case <-ctx.Done():
				fmt.Printf("Task on Worker %d cancelled due to shutdown.\n", res.ID)
				return "", ctx.Err()
			case <-time.After(5 * time.Second): // Simulate long work
				return "long task completed", nil
			}
		})
	}()

	// Immediately kill the manager
	time.Sleep(50 * time.Millisecond)
	fmt.Println("\nInitiating immediate shutdown (Kill)...")
	err = manager.Kill()
	if err != nil {
		fmt.Printf("Error during immediate shutdown: %v\n", err)
	} else {
		fmt.Println("Task manager immediately shut down.")
	}
	fmt.Printf("Final Active Workers: %d\n", manager.Stats().ActiveWorkers)
}

```


The long-running task should be reported as 'cancelled'. Logs should show workers being destroyed immediately. Final 'Active Workers' should be 0. `Task manager immediately shut down.` should be printed.

**Related Methods**: `Kill`

**Related Errors**: `task manager already killed`

---

## `Get live stats`

Demonstrates how to retrieve real-time operational statistics of the `TaskManager`.



```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

type CalcResource struct{}
func createCalcResource() (*CalcResource, error) { return &CalcResource{}, nil }
func destroyCalcResource(r *CalcResource) error { return nil }

func main() {
	ctx := context.Background()
	config := tasker.Config[*CalcResource]{
		OnCreate:    createCalcResource,
		OnDestroy:   destroyCalcResource,
		WorkerCount: 1,
		Ctx:         ctx,
	}
	manager, err := tasker.NewTaskManager[*CalcResource, int](config)
	if err != nil {
		log.Fatalf("Error creating manager: %v", err)
	}
	defer manager.Stop()

	// Queue a task to keep a worker busy
	go func() {
		_, _ = manager.QueueTask(func(ctx context.Context, r *CalcResource) (int, error) {
			time.Sleep(200 * time.Millisecond)
			return 0, nil
		})
	}()

	time.Sleep(50 * time.Millisecond) // Allow worker to pick up task
	stats := manager.Stats()
	fmt.Printf("Current Stats: Active Workers: %d, Queued Tasks: %d, Available Resources: %d\n",
		stats.ActiveWorkers, stats.QueuedTasks, stats.AvailableResources)

	time.Sleep(200 * time.Millisecond) // Wait for task to complete
	stats = manager.Stats()
	fmt.Printf("Stats after task completion: Active Workers: %d, Queued Tasks: %d, Available Resources: %d\n",
		stats.ActiveWorkers, stats.QueuedTasks, stats.AvailableResources)
}

```


Output shows 'Active Workers: 1' and 'Queued Tasks: 0' initially (if task immediately picked up), and then 'Active Workers: 1' and 'Queued Tasks: 0' after task completion. The 'Available Resources' should reflect pool state.

**Related Methods**: `Stats`

---

## `Get performance metrics`

Demonstrates how to retrieve comprehensive performance metrics from the `TaskManager`.



```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

type CalcResource struct{}
func createCalcResource() (*CalcResource, error) { return &CalcResource{}, nil }
func destroyCalcResource(r *CalcResource) error { return nil }

func main() {
	ctx := context.Background()
	config := tasker.Config[*CalcResource]{
		OnCreate:    createCalcResource,
		OnDestroy:   destroyCalcResource,
		WorkerCount: 1,
		Ctx:         ctx,
	}
	manager, err := tasker.NewTaskManager[*CalcResource, int](config)
	if err != nil {
		log.Fatalf("Error creating manager: %v", err)
	}
	defer manager.Stop()

	// Queue multiple tasks to generate metrics data
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = manager.QueueTask(func(ctx context.Context, r *CalcResource) (int, error) {
				time.Sleep(time.Duration(10 + i*5) * time.Millisecond) // Varying work
				return 0, nil
			})
		}()
	}

	time.Sleep(500 * time.Millisecond) // Allow tasks to process

	metrics := manager.Metrics()
	fmt.Printf("\n--- Performance Metrics ---\n")
	fmt.Printf("Total Tasks Completed: %d\n", metrics.TotalTasksCompleted)
	fmt.Printf("Average Execution Time: %v\n", metrics.AverageExecutionTime)
	fmt.Printf("P95 Execution Time: %v\n", metrics.P95ExecutionTime)
	fmt.Printf("Task Completion Rate: %.2f/sec\n", metrics.TaskCompletionRate)
	fmt.Printf("Success Rate: %.2f\n", metrics.SuccessRate)
	fmt.Println("-----------------------------")
}

```


Output shows various metrics like `Total Tasks Completed`, `Average Execution Time`, `P95 Execution Time`, `Task Completion Rate`, and `Success Rate`. Values should be non-zero after tasks have completed.

**Related Methods**: `Metrics`

---

## `Queue and wait`

Template for queuing a task to the main queue and blocking the caller until its result is available.



```go
result, err := manager.QueueTask(func(ctx context.Context, res *MyResource) (string, error) {
    // Task logic here
    return "task_done", nil
})
if err != nil {
    fmt.Printf("Task failed: %v\n", err)
} else {
    fmt.Printf("Task completed with result: %s\n", result)
}

```


The `result` variable holds the correct task output, and `err` is nil if the task was successful. The `fmt.Printf` will execute after the task is truly finished.

**Related Methods**: `QueueTask`

---

## `Queue with callback`

Template for queuing a task to the main queue and receiving its result asynchronously via a callback function.



```go
manager.QueueTaskWithCallback(func(ctx context.Context, res *MyResource) (string, error) {
    // Task logic here
    return "callback_task_done", nil
}, func(result string, err error) {
    if err != nil {
        fmt.Printf("Callback task failed: %v\n", err)
    } else {
        fmt.Printf("Callback task completed with result: %s\n", result)
    }
})
fmt.Println("Submitted callback task, not blocking.")

```


The line 'Submitted callback task, not blocking.' prints immediately. The callback function prints its result message later, once the task is processed.

**Related Methods**: `QueueTaskWithCallback`

---

## `Queue asynchronously with channels`

Template for queuing a task to the main queue and receiving its result and error via Go channels, allowing non-blocking submission and flexible result consumption.



```go
resultChan, errChan := manager.QueueTaskAsync(func(ctx context.Context, res *MyResource) (string, error) {
    // Task logic here
    return "channel_task_done", nil
})
fmt.Println("Submitted async channel task, not blocking.")

go func() {
    result := <-resultChan
    err := <-errChan
    if err != nil {
        fmt.Printf("Async channel task failed: %v\n", err)
    } else {
        fmt.Printf("Async channel task completed with result: %s\n", result)
    }
}()

```


The line 'Submitted async channel task, not blocking.' prints immediately. The result/error message from the goroutine receiving from channels prints later, once the task is processed.

**Related Methods**: `QueueTaskAsync`

---

