# Tasker: A Robust Concurrent Task Management Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/asaidimu/tasker.svg)](https://pkg.go.dev/github.com/asaidimu/tasker)
[![Build Status](https://github.com/asaidimu/tasker/workflows/Test%20Workflow/badge.svg)](https://github.com/asaidimu/tasker/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Tasker** is a powerful and flexible Go library designed for efficient management of concurrent tasks. It provides a highly customizable worker pool, dynamic scaling (bursting), priority queuing, and robust resource lifecycle management, making it ideal for processing background jobs, handling I/O-bound operations, or managing CPU-intensive computations with controlled concurrency.

## 🚀 Quick Links
- [Overview & Features](#overview--features)
- [Installation & Setup](#installation--setup)
- [Usage Documentation](#usage-documentation)
  - [Basic Usage: Simple Calculator](#basic-usage-simple-calculator)
  - [Intermediate Usage: Image Processing with Health Checks & Pooling](#intermediate-usage-image-processing-with-health-checks--pooling)
  - [Advanced Usage: Dynamic Scaling (Bursting) & Graceful Shutdown](#advanced-usage-dynamic-scaling-bursting--graceful-shutdown)
- [Project Architecture](#project-architecture)
- [Development & Contributing](#development--contributing)
- [Additional Information](#additional-information)

---

## Overview & Features

In modern applications, efficiently managing concurrent tasks and shared resources is critical. `tasker` addresses this by providing a comprehensive solution that abstracts away the complexities of goroutine management, worker pools, and resource lifecycles. It allows developers to define tasks that operate on specific resources (e.g., database connections, external API clients, custom compute units) and then queue these tasks for execution, letting `tasker` handle the underlying concurrency, scaling, and error recovery.

### Key Features

*   **Concurrent Task Execution**: Manages a pool of workers to execute tasks concurrently.
*   **Generic Resource Management**: Define custom `OnCreate` and `OnDestroy` functions for any resource type (`R`), ensuring proper setup and cleanup.
*   **Dynamic Worker Scaling (Bursting)**: Automatically scales up and down the number of workers based on queue backlog, ensuring optimal resource utilization during peak loads.
*   **Priority Queues**: Supports both standard and high-priority task queues, allowing critical operations to bypass regular tasks.
*   **Immediate Task Execution with Resource Pooling (`RunTask`)**: Execute tasks synchronously, either by acquiring a resource from a pre-allocated pool or by temporarily creating a new one, ideal for urgent, low-latency operations.
*   **Customizable Health Checks & Retries**: Define custom logic (`CheckHealth`) to determine if a worker/resource is unhealthy, enabling automatic worker replacement and configurable task retries for transient failures.
*   **Graceful Shutdown**: Ensures all active tasks complete and resources are properly released during shutdown, preventing data loss or resource leaks.
*   **Real-time Performance Metrics**: Access live statistics on active workers, queued tasks, and available resources for monitoring and debugging.

---

## Installation & Setup

### Prerequisites

*   Go **1.24.3** or higher

### Installation Steps

To add `tasker` to your Go project, use `go get`:

```bash
go get github.com/asaidimu/tasker
```

### Verification

You can verify the installation by building and running the provided examples:

```bash
cd $GOPATH/src/github.com/asaidimu/tasker/examples/basic
go run main.go

cd $GOPATH/src/github.com/asaidimu/tasker/examples/intermediate
go run main.go

cd $GOPATH/src/github.com/asaidimu/tasker/examples/advanced
go run main.go
```

---

## Usage Documentation

`tasker` is designed to be highly configurable and flexible. All interactions happen through the `tasker.NewRunner` constructor and the returned `TaskManager` interface.

### Core Concepts

*   **Resource (R)**: This is the type of resource your tasks will operate on. It could be anything: a database connection, an HTTP client, a custom processing struct, etc. `tasker` manages the lifecycle of these resources.
*   **Task Result (E)**: This is the type of value your tasks will return upon successful completion.
*   **Task Function**: Your actual work logic is defined as a `func(resource R) (result E, err error)`.
*   **`tasker.Config[R]`**: Configures the `Runner` with resource lifecycle functions (`OnCreate`, `OnDestroy`), worker counts, scaling parameters, and more.
*   **`tasker.Runner[R, E]`**: The core task manager that implements `TaskManager[R, E]`.
*   **`QueueTask`**: Adds a task to the standard queue for asynchronous processing. Returns the result once the task completes.
*   **`QueueTaskWithPriority`**: Adds a task to a dedicated high-priority queue. These tasks are picked up before standard queued tasks.
*   **`RunTask`**: Executes a task immediately. It tries to get a resource from a pre-allocated pool or creates a temporary one if needed. This is a synchronous call.
*   **`Stop()`**: Gracefully shuts down the manager, allowing active tasks to complete.
*   **`Stats()`**: Provides real-time operational statistics.

---

### Basic Usage: Simple Calculator

This example demonstrates the fundamental setup for `tasker` using a simple `CalculatorResource`.

```go
// examples/basic/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker"
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
	manager, err := tasker.NewRunner[*CalculatorResource, int](config) // Tasks will return an int result
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop() // Ensure the manager is stopped gracefully

	fmt.Println("Queuing a simple addition task...")
	task1Start := time.Now()
	// Queue a task to perform addition
	go func() {
		sum, err := manager.QueueTask(func(r *CalculatorResource) (int, error) {
			// In a real scenario, 'r' could be a connection to a math service
			time.Sleep(50 * time.Millisecond) // Simulate some work
			a, b := 10, 25
			fmt.Printf("Worker processing: %d + %d\n", a, b)
			return a + b, nil
		})

		if err != nil {
			fmt.Printf("Task 1 failed: %v\n", err)
		} else {
			fmt.Printf("Task 1 (Addition) Result: %d (took %s)\n", sum, time.Since(task1Start))
		}
	}()

	fmt.Println("Queuing another subtraction task...")
	task2Start := time.Now()
	// Queue another task for subtraction
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
			fmt.Printf("Task 2 (Subtraction) Result: %d (took %s)\n", difference, time.Since(task2Start))
		}
	}()

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

**Expected Output (Illustrative, timings may vary):**
```
--- Basic Usage: Simple Calculator ---
INFO: Creating CalculatorResource
INFO: Creating CalculatorResource
Queuing a simple addition task...
Queuing another subtraction task...
Worker processing: 100 - 40
Worker processing: 10 + 25
Task 2 (Subtraction) Result: 60 (took 70.xxxms)
Task 1 (Addition) Result: 35 (took 70.xxxms)

--- Current Stats ---
Active Workers: 2
Queued Tasks: 0
Available Resources: 0
----------------------
Basic usage example finished.
```

---

### Intermediate Usage: Image Processing with Health Checks & Pooling

This example introduces more advanced features like custom health checks (`CheckHealth`), task retries (`MaxRetries`), high-priority queuing (`QueueTaskWithPriority`), and immediate execution with resource pooling (`RunTask`).

```go
// examples/intermediate/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/asaidimu/tasker"
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
		WorkerCount:      2,           // Two base workers
		Ctx:              ctx,
		CheckHealth:      checkImageProcessorHealth, // Use custom health check
		MaxRetries:       1,           // One retry on unhealthy errors
		ResourcePoolSize: 1,           // Small pool for RunTask
	}

	manager, err := tasker.NewRunner[*ImageProcessor, string](config) // Tasks return string (e.g., image path)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	// --- 1. Queue a normal image resize task ---
	fmt.Println("\nQueueing a normal image resize task...")
	go func() {
		result, err := manager.QueueTask(func(proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing normal resize for imageA.jpg\n", proc.ID)
			time.Sleep(150 * time.Millisecond)
			if rand.Intn(10) < 3 { // Simulate a healthy but non-fatal processing error 30% of the time
				return "", errors.New("image_corrupted")
			}
			return "imageA_resized.jpg", nil
		})
		if err != nil {
			fmt.Printf("Normal Resize Failed: %v\n", err)
		} else {
			fmt.Printf("Normal Resize Completed: %s\n", result)
		}
	}()

	// --- 2. Queue a high-priority thumbnail generation task ---
	fmt.Println("Queueing a high-priority thumbnail task...")
	go func() {
		result, err := manager.QueueTaskWithPriority(func(proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing HIGH PRIORITY thumbnail for video.mp4\n", proc.ID)
			time.Sleep(50 * time.Millisecond) // Faster processing
			return "video_thumbnail.jpg", nil
		})
		if err != nil {
			fmt.Printf("Priority Thumbnail Failed: %v\n", err)
		} else {
			fmt.Printf("Priority Thumbnail Completed: %s\n", result)
		}
	}()

	// --- 3. Simulate a task that causes an "unhealthy" error ---
	fmt.Println("Queueing a task that might crash a worker (unhealthy error)...")
	go func() {
		result, err := manager.QueueTask(func(proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing problematic image (might crash)\n", proc.ID)
			time.Sleep(100 * time.Millisecond)
			if rand.Intn(2) == 0 { // 50% chance to simulate a crash
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

	// --- 4. Run an immediate task (e.g., generate a preview for a user) ---
	fmt.Println("\nRunning an immediate task (using resource pool or temporary resource)...")
	immediateResult, immediateErr := manager.RunTask(func(proc *ImageProcessor) (string, error) {
		fmt.Printf("IMMEDIATE Task processing fast preview with processor %d\n", proc.ID)
		time.Sleep(20 * time.Millisecond)
		return "fast_preview.jpg", nil
	})
	if immediateErr != nil {
		fmt.Printf("Immediate Task Failed: %v\n", immediateErr)
	} else {
		fmt.Printf("Immediate Task Completed: %s\n", immediateResult)
	}

	// Give time for tasks and potential worker replacements
	time.Sleep(1 * time.Second)

	stats := manager.Stats()
	fmt.Printf("\n--- Current Stats ---\n")
	fmt.Printf("Base Workers: %d\n", stats.BaseWorkers)
	fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers)
	fmt.Printf("Queued Tasks: %d\n", stats.QueuedTasks)
	fmt.Printf("Priority Tasks: %d\n", stats.PriorityTasks)
	fmt.Printf("Available Resources: %d\n", stats.AvailableResources)
	fmt.Println("----------------------")

	fmt.Println("Intermediate usage example finished.")
}
```

**Expected Output (Illustrative, order and IDs may vary due to concurrency and random failures):**
```
--- Intermediate Usage: Image Processing ---
INFO: Creating ImageProcessor 719
INFO: Creating ImageProcessor 409
INFO: Creating ImageProcessor 981
Queueing a normal image resize task...
Queueing a high-priority thumbnail task...
Queueing a task that might crash a worker (unhealthy error)...

Running an immediate task (using resource pool or temporary resource)...
IMMEDIATE Task processing fast preview with processor 981
Worker 719 processing HIGH PRIORITY thumbnail for video.mp4
Immediate Task Completed: fast_preview.jpg
Priority Thumbnail Completed: video_thumbnail.jpg
Worker 409 processing normal resize for imageA.jpg
Worker 719 processing problematic image (might crash)
Normal Resize Completed: imageA_resized.jpg
WARN: Detected unhealthy error: processor_crash. Worker will be replaced.
INFO: Destroying ImageProcessor 719
INFO: Creating ImageProcessor 507
Problematic Image Task Failed: max retries exceeded for task: processor_crash

--- Current Stats ---
Base Workers: 2
Active Workers: 2
Queued Tasks: 0
Priority Tasks: 0
Available Resources: 1
----------------------
Intermediate usage example finished.
```
Notice how `processor_crash` leads to a `WARN` and the worker being replaced, showcasing the `CheckHealth` functionality.

---

### Advanced Usage: Dynamic Scaling (Bursting) & Graceful Shutdown

This example showcases `tasker`'s dynamic scaling capabilities (`BurstTaskThreshold`, `BurstWorkerCount`, `BurstInterval`) to handle fluctuating loads and demonstrates a comprehensive graceful shutdown procedure.

```go
// examples/advanced/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/asaidimu/tasker"
)

// HeavyComputeResource represents a resource for intensive computations.
type HeavyComputeResource struct {
	ID int
}

// onCreate for HeavyComputeResource: simulates allocating CPU/GPU resources.
func createComputeResource() (*HeavyComputeResource, error) {
	id := rand.Intn(1000)
	fmt.Printf("INFO: Creating HeavyComputeResource %d\n", id)
	return &HeavyComputeResource{ID: id}, nil
}

// onDestroy for HeavyComputeResource: simulates releasing CPU/GPU resources.
func destroyComputeResource(r *HeavyComputeResource) error {
	fmt.Printf("INFO: Destroying HeavyComputeResource %d\n", r.ID)
	return nil
}

// checkComputeHealth: All errors are considered healthy (task-specific, not worker-specific)
func checkComputeHealth(err error) bool {
	return true
}

func main() {
	fmt.Println("\n--- Advanced Usage: Dynamic Scaling (Bursting) ---")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the main context is cancelled on exit

	config := tasker.Config[*HeavyComputeResource]{
		OnCreate:        createComputeResource,
		OnDestroy:       destroyComputeResource,
		WorkerCount:     2,                   // Start with 2 base workers
		Ctx:             ctx,
		CheckHealth:     checkComputeHealth,
		BurstTaskThreshold:  5,                   // If 5+ tasks in queue, start bursting
		BurstWorkerCount:      2,                   // Add 2 burst workers at a time
		BurstInterval:   200 * time.Millisecond, // Check every 200ms
		MaxRetries:      0,                   // No retries for this example
		ResourcePoolSize: 0,                  // Not using RunTask heavily here
	}

	manager, err := tasker.NewRunner[*HeavyComputeResource, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	// Do NOT defer manager.Stop() here. We will call it manually with a delay to observe shutdown.

	var tasksSubmitted atomic.Int32
	var tasksCompleted atomic.Int32

	// --- 1. Aggressively queue a large number of tasks ---
	fmt.Println("\nAggressively queuing 20 heavy computation tasks...")
	for i := range 20 {
		taskID := i
		tasksSubmitted.Add(1)
		go func() {
			result, err := manager.QueueTask(func(res *HeavyComputeResource) (string, error) {
				processingTime := time.Duration(100 + rand.Intn(300)) * time.Millisecond // 100-400ms work
				fmt.Printf("Worker %d processing Task %d for %s\n", res.ID, taskID, processingTime)
				time.Sleep(processingTime)
				if rand.Intn(10) == 0 { // 10% chance of failure
					return "", errors.New(fmt.Sprintf("computation_error_task_%d", taskID))
				}
				return fmt.Sprintf("Task %d completed", taskID), nil
			})
			if err != nil {
				fmt.Printf("Task %d failed: %v\n", taskID, err)
			} else {
				fmt.Printf("Task %d result: %s\n", taskID, result)
				tasksCompleted.Add(1)
			}
		}()
	}

	// --- 2. Monitor stats to observe bursting ---
	fmt.Println("\nMonitoring stats (watch for BurstWorkers)...")
	ticker := time.NewTicker(250 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				stats := manager.Stats()
				fmt.Printf("Stats: Active=%d (Base=%d, Burst=%d), Queued=%d (Main=%d, Prio=%d), AvailRes=%d, Comp=%d/%d\n",
					stats.ActiveWorkers, stats.BaseWorkers, stats.BurstWorkers,
					stats.QueuedTasks+stats.PriorityTasks, stats.QueuedTasks, stats.PriorityTasks,
					stats.AvailableResources, tasksCompleted.Load(), tasksSubmitted.Load())
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	// Let the system run for a while to observe scaling up and down
	time.Sleep(5 * time.Second)

	// Trigger a high-priority task during burst
	fmt.Println("\nQueueing a high priority task during active bursting...")
	go func() {
		result, err := manager.QueueTaskWithPriority(func(res *HeavyComputeResource) (string, error) {
			fmt.Printf("Worker %d processing HIGH PRIORITY urgent report!\n", res.ID)
			time.Sleep(50 * time.Millisecond)
			return "Urgent report generated!", nil
		})
		if err != nil {
			fmt.Printf("Urgent task failed: %v\n", err)
		} else {
			fmt.Printf("Urgent task result: %s\n", result)
			tasksCompleted.Add(1)
		}
	}()

	time.Sleep(2 * time.Second) // Allow priority task and more processing

	// --- 3. Graceful Shutdown ---
	fmt.Println("\nInitiating graceful shutdown...")
	close(done) // Stop the stats monitoring goroutine
	err = manager.Stop()
	if err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	} else {
		fmt.Println("Task manager gracefully shut down.")
	}

	stats := manager.Stats()
	fmt.Printf("\n--- Final Stats After Shutdown ---\n")
	fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers) // Should be 0
	fmt.Printf("Burst Workers: %d\n", stats.BurstWorkers)   // Should be 0
	fmt.Printf("Queued Tasks: %d\n", stats.QueuedTasks)     // Should be 0
	fmt.Printf("Priority Tasks: %d\n", stats.PriorityTasks) // Should be 0
	fmt.Printf("Available Resources: %d\n", stats.AvailableResources) // Should be 0
	fmt.Printf("Tasks Completed: %d / %d\n", tasksCompleted.Load(), tasksSubmitted.Load())
	fmt.Println("---------------------------------")

	fmt.Println("Advanced usage example finished.")
}
```

**Expected Output (Illustrative, IDs and timings vary greatly):**
```
--- Advanced Usage: Dynamic Scaling (Bursting) ---
INFO: Creating HeavyComputeResource 578
INFO: Creating HeavyComputeResource 662

Aggressively queuing 20 heavy computation tasks...

Monitoring stats (watch for BurstWorkers)...
Stats: Active=2 (Base=2, Burst=0), Queued=20 (Main=20, Prio=0), AvailRes=0, Comp=0/20
Worker 578 processing Task 0 for 191ms
Stats: Active=2 (Base=2, Burst=0), Queued=19 (Main=19, Prio=0), AvailRes=0, Comp=0/20
Worker 662 processing Task 1 for 142ms
Stats: Active=2 (Base=2, Burst=0), Queued=18 (Main=18, Prio=0), AvailRes=0, Comp=0/20
INFO: Creating HeavyComputeResource 811
INFO: Creating HeavyComputeResource 824
Stats: Active=4 (Base=2, Burst=2), Queued=16 (Main=16, Prio=0), AvailRes=0, Comp=0/20
Worker 578 processing Task 2 for 211ms
Task 0 result: Task 0 completed
Worker 662 processing Task 3 for 308ms
Task 1 result: Task 1 completed
Stats: Active=4 (Base=2, Burst=2), Queued=14 (Main=14, Prio=0), AvailRes=0, Comp=2/20
... (more workers spin up, tasks complete) ...
Stats: Active=6 (Base=2, Burst=4), Queued=3 (Main=3, Prio=0), AvailRes=0, Comp=17/20
Stats: Active=4 (Base=2, Burst=2), Queued=1 (Main=1, Prio=0), AvailRes=0, Comp=18/20

Queueing a high priority task during active bursting...
Stats: Active=2 (Base=2, Burst=0), Queued=1 (Main=0, Prio=1), AvailRes=0, Comp=19/20
Worker 578 processing HIGH PRIORITY urgent report!
Urgent task result: Urgent report generated!
Task 19 result: Task 19 completed

Initiating graceful shutdown...
Stats: Active=2 (Base=2, Burst=0), Queued=0 (Main=0, Prio=0), AvailRes=0, Comp=20/20
INFO: Destroying HeavyComputeResource 578
INFO: Destroying HeavyComputeResource 662
Task manager gracefully shut down.

--- Final Stats After Shutdown ---
Active Workers: 0
Burst Workers: 0
Queued Tasks: 0
Priority Tasks: 0
Available Resources: 0
Tasks Completed: 20 / 20
---------------------------------
Advanced usage example finished.
```
This output clearly shows `BurstWorkers` increasing and then decreasing as the queue size changes, and ultimately all resources being destroyed during graceful shutdown.

---

## Project Architecture

`tasker` is designed with modularity and extensibility in mind, built around goroutines and channels for efficient concurrency management.

### Core Components

*   **`Runner[R, E]`**: The central orchestrator. It manages the lifecycle of workers, resources, and tasks. It holds the configuration, queues, and synchronizes operations.
*   **Worker Goroutines**: These are the workhorses. Each worker holds a single resource (`R`) and continuously picks tasks from the priority queue or main queue. They handle task execution, error reporting, and retry logic based on `CheckHealth`.
*   **Task Queues (`mainQueue`, `priorityQueue`)**: Unbuffered channels (`chan *Task[R, E]`) acting as FIFO queues for tasks awaiting execution. `priorityQueue` is always checked first by workers.
*   **Resource Pool (`resourcePool`)**: A buffered channel holding pre-allocated `R` type resources, primarily used by `RunTask` for immediate, low-latency task execution without waiting for a dedicated worker.
*   **Burst Manager Goroutine**: (If configured) This goroutine periodically monitors the task queue backlog. If the backlog exceeds `BurstTaskThreshold`, it spins up additional "burst" workers. If the backlog shrinks, it gracefully scales down burst workers.
*   **`Task[R, E]`**: A struct encapsulating the task function, channels for returning results (`result chan E`, `err chan error`), and a retry counter.
*   **`Config[R]`**: Defines the behavior of the `Runner`, including resource creation/destruction, worker counts, scaling parameters, and health check logic.
*   **`TaskStats`**: Provides a snapshot of the `Runner`'s current operational state, including worker counts and queue sizes.

### Data Flow

1.  **Initialization**: `NewRunner` creates base workers, initializes the resource pool, and starts the optional burst manager. `OnCreate` is called for each resource.
2.  **Task Submission**:
    *   `QueueTask` sends `*Task[R, E]` to `mainQueue`.
    *   `QueueTaskWithPriority` sends `*Task[R, E]` to `priorityQueue`.
    *   `RunTask` directly attempts to retrieve a resource from `resourcePool` or creates a temporary one, executes the task, and returns the resource to the pool (or destroys it).
3.  **Task Execution**: Workers prioritize tasks from `priorityQueue`, then `mainQueue`. They acquire a resource (which they own for their lifetime, or a temporary one for `RunTask`), execute the task's `run` function, and send results/errors back to the caller via channels embedded in the `Task` struct.
4.  **Health Checks & Retries**: If a task returns an error and `CheckHealth` indicates an unhealthy state, the worker may be terminated and recreated, and the task re-queued (up to `MaxRetries`).
5.  **Dynamic Scaling**: The burst manager continuously monitors queue lengths. If tasks accumulate, it instructs the `Runner` to start new burst workers. If queues clear, it signals burst workers to shut down gracefully.
6.  **Graceful Shutdown (`Stop`)**: The main context is cancelled, signaling all workers and the burst manager to stop processing new tasks. Active tasks are allowed to complete. All resources (worker-owned and pool-based) are destroyed via `OnDestroy`. The system waits for all goroutines to exit before returning.

### Extension Points

`tasker` is designed to be highly pluggable through its function-based configuration:

*   **`OnCreate func() (R, error)`**: Define how to instantiate your specific resource `R`. This could involve connecting to a database, initializing an external SDK, or setting up a complex object.
*   **`OnDestroy func(R) error`**: Define how to clean up your resource `R` when a worker shuts down or a temporary resource is no longer needed. This is crucial for releasing connections, closing files, or deallocating memory.
*   **`CheckHealth func(error) bool`**: Implement custom logic to determine if a task's error indicates an underlying problem with the worker or its resource. Returning `false` here will cause `tasker` to replace the worker and potentially retry the task, enhancing system resilience.

---

## Development & Contributing

We welcome contributions to `tasker`!

### Development Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/asaidimu/tasker.git
    cd tasker
    ```
2.  **Ensure Go Modules are tidy:**
    ```bash
    go mod tidy
    ```
3.  **Build the project:**
    ```bash
    go build -v ./...
    ```

### Scripts

The project includes a `Makefile` for common development tasks:

*   `make build`: Compiles the main package and its dependencies.
*   `make test`: Runs all unit tests with verbose output.
*   `make clean`: Removes generated executable files.

### Testing

To run the test suite:

```bash
go test -v ./...
```

All contributions are expected to pass existing tests and maintain a high level of code coverage. New features or bug fixes should come with appropriate tests.

### Contributing Guidelines

We appreciate your interest in contributing! To ensure a smooth contribution process, please follow these guidelines:

1.  **Fork the repository** and create your branch from `main`.
2.  **Keep your code clean and well-documented.** Follow Go's idiomatic style.
3.  **Write clear, concise commit messages** that describe your changes.
4.  **Ensure your changes are tested.** New features require new tests, and bug fixes should include a test that reproduces the bug.
5.  **Open a Pull Request** against the `main` branch. Provide a detailed description of your changes.

### Issue Reporting

If you encounter any bugs or have feature requests, please open an issue on the [GitHub Issues page](https://github.com/asaidimu/tasker/issues).

When reporting a bug, please include:
*   A clear, concise description of the problem.
*   Steps to reproduce the behavior.
*   Expected behavior.
*   Actual behavior.
*   Any relevant error messages or logs.
*   Your Go version and operating system.

---

## Additional Information

### Troubleshooting

*   **"Task manager is shutting down: cannot queue task"**: This error occurs if you try to `QueueTask` or `RunTask` after `manager.Stop()` has been called or its parent context has been cancelled. Ensure tasks are only submitted while the manager is active.
*   **Resource Creation Failure**: If your `OnCreate` function returns an error, `tasker` will log it and the associated worker will not start (or a `RunTask` call will fail). Ensure your resource creation logic is robust.
*   **Workers Not Starting/Stopping as Expected**:
    *   Check your `WorkerCount` and `MaxWorkerCount` settings.
    *   For bursting, ensure `BurstTaskThreshold`, `BurstWorkerCount`, and `BurstInterval` are correctly configured.
    *   Verify your `Ctx` and `cancel` functions are being managed correctly, especially for graceful shutdown.
    *   Review `CheckHealth` logic; an incorrect implementation might cause workers to constantly recreate.
*   **Deadlocks/Goroutine Leaks**: While `tasker` is designed to prevent these, improper usage (e.g., blocking in `OnCreate`/`OnDestroy`, or not using buffered channels for task results outside the library) can lead to issues. Always ensure your custom functions (`OnCreate`, `OnDestroy`, task functions) don't block indefinitely.

### FAQ

*   **When should I use `QueueTask` vs. `RunTask`?**
    *   Use `QueueTask` for asynchronous, background tasks that can wait in a queue for an available worker. Ideal for high-throughput, batch processing.
    *   Use `RunTask` for synchronous, immediate tasks that require low latency and might need a resource right away, bypassing the queues. It's suitable for user-facing requests or critical operations that can't afford queueing delay.
*   **How does `CheckHealth` affect workers?**
    *   If `CheckHealth` returns `false` for an error, `tasker` considers the worker (or its resource) to be in an unhealthy state. This worker will be shut down, its resource destroyed (`OnDestroy`), and a new worker will be created to replace it. The original task that caused the unhealthy error will also be re-queued (up to `MaxRetries`).
*   **What happens if a task panics?**
    *   It is generally recommended that your task functions (the `func(R) (E, error)` you pass to `QueueTask` etc.) recover from panics and return an error instead. If a panic occurs and is not recovered within the task, it will crash the worker goroutine. While `tasker` will attempt to replace the worker, unhandled panics can lead to unexpected behavior and resource leaks.
*   **Is `tasker` suitable for long-running tasks?**
    *   Yes, `tasker` can handle long-running tasks. However, be mindful of the `Ctx` passed to `NewRunner`. If that context is cancelled, all workers will attempt to gracefully shut down, potentially interrupting very long tasks. For indefinite tasks, consider separate Goroutines outside of `tasker` or ensure a very long-lived context.

### Changelog / Roadmap

*   [**CHANGELOG.md**](CHANGELOG.md): See the project's history of changes.
*   **Roadmap**: Future plans may include deeper metrics integration (e.g., Prometheus exporters), more sophisticated queuing strategies, and enhanced observability features.

### License

`tasker` is distributed under the MIT License. See [LICENSE.md](LICENSE.md) for the full text.

### Acknowledgments

This project is inspired by common worker pool patterns and the need for robust, flexible concurrency management in modern Go applications.
