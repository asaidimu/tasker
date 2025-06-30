# Tasker: A Robust Concurrent Task Management Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/asaidimu/tasker.svg)](https://pkg.go.dev/github.com/asaidimu/tasker)
[![Build Status](https://github.com/asaidimu/tasker/actions/workflows/test.yml/badge.svg)](https://github.com/asaidimu/tasker/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Tasker** is a powerful and flexible Go library designed for efficient management of concurrent tasks. It provides a highly customizable worker pool, dynamic scaling (bursting), priority queuing, and robust resource lifecycle management, making it ideal for processing background jobs, handling I/O-bound operations, or managing CPU-intensive computations with controlled concurrency.

## 🚀 Quick Links
- [Overview & Features](#overview--features)
- [Installation & Setup](#installation--setup)
- [Usage Documentation](#usage-documentation)
- [Project Architecture](#project-architecture)
- [Development & Contributing](#development--contributing)
- [Additional Information](#additional-information)

---

## Overview & Features

In modern applications, efficiently managing concurrent tasks and shared resources is critical. `tasker` addresses this by providing a comprehensive solution that abstracts away the complexities of goroutine management, worker pools, and resource lifecycles. It allows developers to define tasks that operate on specific resources (e.g., database connections, external API clients, custom compute units) and then queue these tasks for execution, letting `tasker` handle the underlying concurrency, scaling, and error recovery.

This library is particularly useful for applications that:
*   Need to process a high volume of background jobs reliably.
*   Perform operations on limited or expensive shared resources.
*   Require dynamic adjustment of processing capacity based on load.
*   Demand prioritization of certain critical tasks over others.

### Key Features

*   **Concurrent Task Execution**: Manages a pool of workers to execute tasks concurrently, ensuring optimal utilization of system resources.
*   **Generic Resource Management**: Define custom `OnCreate` and `OnDestroy` functions for any resource type (`R`), guaranteeing proper setup and cleanup of external dependencies or expensive objects.
*   **Rate-Based Dynamic Worker Scaling**: Automatically scales the number of workers up or down based on the real-time task arrival and completion rates. This ensures that the system's throughput dynamically matches the incoming workload, optimizing resource utilization and responsiveness without manual tuning.
*   **Priority Queues**: Supports both standard (`QueueTask`) and high-priority (`QueueTaskWithPriority`) task queues, allowing critical operations to bypass regular tasks and get processed faster.
*   **Immediate Task Execution with Resource Pooling (`RunTask`)**: Execute tasks synchronously, either by acquiring a resource from a pre-allocated pool or by temporarily creating a new one. This is ideal for urgent, low-latency operations that should not be delayed by queuing.
*   **Customizable Health Checks & Retries**: Define custom logic (`CheckHealth`) to determine if a given error indicates an "unhealthy" state for a worker or its associated resource, enabling automatic worker replacement and configurable task retries for transient failures.
*   **"At-Most-Once" Task Execution**: Provides `QueueTaskOnce` and `QueueTaskWithPriorityOnce` methods for non-idempotent operations, preventing re-queuing on health-related failures.
*   **Graceful & Immediate Shutdown**: Offers both `Stop()` for graceful shutdown (waiting for active tasks to complete) and `Kill()` for immediate termination (cancelling all tasks), giving control over shutdown behavior.
*   **Real-time Performance Metrics**: Access live statistics (`Stats()`) and comprehensive performance metrics (`Metrics()`), including task arrival/completion rates, average/min/max/percentile execution times, average wait times, and success/failure rates, for robust monitoring and debugging.
*   **Custom Logging**: Integrate with your preferred logging solution by providing an implementation of the `tasker.Logger` interface.

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

You can verify the installation and see `tasker` in action by building and running the provided examples:

```bash
# Navigate to the examples directory within your Go module path
# This assumes your GOPATH is set correctly, typically within your user home directory.
cd "$(go env GOPATH)/src/github.com/asaidimu/tasker/examples/basic"
go run main.go

cd "$(go env GOPATH)/src/github.com/asaidimu/tasker/examples/intermediate"
go run main.go

cd "$(go env GOPATH)/src/github.com/asaidimu/tasker/examples/advanced"
go run main.go
```

---

## Usage Documentation

`tasker` is designed to be highly configurable and flexible. All interactions happen through the `tasker.NewTaskManager` constructor and the returned `TaskManager` interface.

### Core Concepts

*   **Resource (`R`)**: This is the generic type of resource your tasks will operate on. It could be anything: a database connection, an HTTP client, a custom processing struct, a CPU/GPU compute unit, or any other external dependency or expensive object that needs managed lifecycle. `tasker` manages the creation, use, and destruction of these resources.
*   **Task Result (`E`)**: This is the generic type of value your tasks will return upon successful completion. This allows `tasker` to be type-safe for various task outputs.
*   **Task Function**: Your actual work logic is defined as a `func(resource R) (result E, err error)`. This function receives an instance of your defined `R` resource type. It's expected to return the result `E` or an `error`.
*   **`tasker.Config[R]`**: A struct used to configure the `TaskManager` with essential parameters such as resource lifecycle functions (`OnCreate`, `OnDestroy`), initial worker counts, dynamic scaling parameters, optional health check logic, and custom logging/metrics.
*   **`tasker.Manager[R, E]`**: The concrete implementation of the `TaskManager[R, E]` interface, providing the core task management capabilities. You instantiate this via `tasker.NewTaskManager`.
*   **`QueueTask(func(R) (E, error)) (E, error)`**: Adds a task to the standard queue for asynchronous processing. The call blocks until the task completes and returns its result or error. Suitable for background processing.
*   **`QueueTaskWithPriority(func(R) (E, error)) (E, error)`**: Adds a task to a dedicated high-priority queue. Tasks in this queue are processed before tasks in the main queue, ensuring faster execution for critical operations. This call also blocks until completion.
*   **`QueueTaskOnce(func(R) (E, error)) (E, error)`**: Similar to `QueueTask`, but if the task fails and `CheckHealth` indicates an unhealthy worker, this task will *not* be retried. Use for non-idempotent operations where "at-most-once" processing by `tasker` is desired.
*   **`QueueTaskWithPriorityOnce(func(R) (E, error)) (E, error)`**: Combines high priority with "at-most-once" execution semantics.
*   **`RunTask(func(R) (E, error)) (E, error)`**: Executes a task immediately. It attempts to acquire a resource from a pre-allocated pool first. If the pool is empty, it temporarily creates a new resource for the task. This is a synchronous call, blocking until the task finishes. Ideal for urgent, low-latency operations that should not be delayed by queuing.
*   **`Stop() error`**: Initiates a graceful shutdown of the manager. It stops accepting new tasks and waits for all currently queued and executing tasks to complete before releasing resources.
*   **`Kill() error`**: Initiates an immediate shutdown of the manager. It cancels all running tasks and drops all queued tasks without waiting for them to complete, then releases resources.
*   **`Stats() TaskStats`**: Returns a `TaskStats` struct containing real-time operational statistics, such as active worker counts, queued task counts, and available resources.
*   **`Metrics() TaskMetrics`**: Returns a `TaskMetrics` struct providing comprehensive performance metrics, including task arrival/completion rates, various execution time percentiles (P95, P99), average wait times, and success/failure rates.

---

### Basic Usage: Simple Calculator

This example demonstrates the fundamental setup for `tasker` using a simple `CalculatorResource` with two base workers.

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
	manager, err := tasker.NewTaskManager[*CalculatorResource, int](config) // Tasks will return an int result
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

**Expected Output (Illustrative, timings may vary due to concurrency):**
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


## Project Architecture

`tasker` is designed with modularity and extensibility in mind, built around goroutines and channels for efficient concurrency management.

### Core Components

*   **`Manager[R, E]`**: The central orchestrator, implementing the `TaskManager` interface. It holds the `Config`, manages task queues, synchronizes operations, and oversees the lifecycle of workers and resources. It's the primary interface for users to interact with `tasker`.
*   **Worker Goroutines**: These are the workhorses of `tasker`. Each worker runs in its own goroutine, typically holding a single long-lived resource (`R`). Workers continuously pull tasks from the `priorityQueue` or `mainQueue`, execute them, handle errors, and manage retries based on the `CheckHealth` function.
*   **Task Queues (`mainQueue`, `priorityQueue`)**: These are Go channels (`chan *Task[R, E]`) acting as FIFO queues. The `priorityQueue` is always checked first by workers, ensuring high-priority tasks are processed ahead of standard tasks.
*   **Resource Pool (`resourcePool`)**: A buffered channel holding pre-allocated `R` type resources. This pool is primarily used by `RunTask` operations to provide immediate resource access for synchronous, low-latency task execution, avoiding the main queues.
*   **Burst Manager Goroutine**: This background goroutine is responsible for dynamic worker scaling. It periodically fetches performance `Metrics` (specifically `TaskArrivalRate` and `TaskCompletionRate`). If the arrival rate significantly exceeds the completion rate, it instructs the `Manager` to start new burst workers. Conversely, if the system is over-provisioned (completion rate much higher than arrival rate), it signals idle burst workers to shut down.
*   **`Task[R, E]`**: An internal struct that encapsulates a task's executable function (`run`), channels for returning its result (`result chan E`) and any error (`err chan error`) to the caller, and a retry counter (`retries`) for fault tolerance. It also stores `queuedAt` for metrics calculations.
*   **`Config[R]`**: A comprehensive struct that defines the operational parameters for creating a `Manager` instance. This includes essential functions for resource lifecycle management (`OnCreate`, `OnDestroy`), `WorkerCount`, `MaxWorkerCount`, dynamic scaling parameters (`BurstInterval`), retry policy (`MaxRetries`), and optional custom logging (`Logger`) and metrics collection (`Collector`).
*   **`TaskStats` & `TaskMetrics`**: These structs provide real-time snapshots of the `Manager`'s current operational state (`Stats()`) and in-depth performance statistics (`Metrics()`), respectively. `TaskMetrics` includes values like average, min, max, P95, and P99 execution times, wait times, arrival/completion rates, and success/failure rates.
*   **`Logger` and `MetricsCollector` Interfaces**: These interfaces allow for flexible integration with external logging and monitoring systems.

### Data Flow

1.  **Initialization**: Upon calling `NewTaskManager`, `tasker` creates the specified number of base workers, initializes the resource pool by calling `OnCreate` for each pooled resource, and starts the dedicated burst manager goroutine.
2.  **Task Submission**:
    *   Calls to `QueueTask` (and `QueueTaskOnce`) send `*Task[R, E]` instances to the `mainQueue`.
    *   Calls to `QueueTaskWithPriority` (and `QueueTaskWithPriorityOnce`) send `*Task[R, E]` instances to the `priorityQueue`.
    *   Calls to `RunTask` first attempt to acquire a resource from the `resourcePool`. If successful, the task is executed directly with that resource. If the pool is empty, a temporary resource is created via `OnCreate`, used for the task, and then destroyed via `OnDestroy`.
    *   All task submissions are recorded by the `MetricsCollector` to track arrival rates.
3.  **Task Execution**: Worker goroutines continuously select tasks from the `priorityQueue` (preferentially) or the `mainQueue`. Once a task is acquired, the worker executes the task's `run` function with its dedicated resource. The `MetricsCollector` records task start and completion times.
4.  **Health Checks & Retries**: If a task execution returns an error, `tasker` invokes the `CheckHealth` function (if provided). If `CheckHealth` returns `false`, indicating an unhealthy state of the worker or its resource, the worker goroutine processing that task is terminated, its resource destroyed (`OnDestroy`), and a new worker is created to replace it. The original task that caused the unhealthy error will then be re-queued (up to `MaxRetries`) to be processed by a newly healthy worker. Tasks submitted via `*Once` methods are not re-queued. Failures and retries are recorded by the `MetricsCollector`.
5.  **Dynamic Scaling**: The burst manager periodically analyzes the `TaskArrivalRate` versus the `TaskCompletionRate` from the `MetricsCollector`. If demand outstrips capacity, it dynamically starts additional "burst" workers. If demand subsides, it signals idle burst workers to gracefully shut down, optimizing resource usage.
6.  **Graceful Shutdown (`Stop`)**: When `Stop()` is called, the `Manager` transitions to a "stopping" state, preventing new tasks from being queued. It then cancels its primary `context.Context`, signaling all worker goroutines and the burst manager to stop processing new tasks. Workers in "stopping" mode will first finish processing any tasks remaining in their queues before exiting. `tasker` then waits for all goroutines to complete and drains/destroys all managed resources via `OnDestroy`.
7.  **Immediate Shutdown (`Kill`)**: When `Kill()` is called, the `Manager` transitions to a "killed" state. It immediately cancels its primary `context.Context`, causing all active worker goroutines to terminate without waiting for tasks to complete. All queued tasks are dropped, and all resources are destroyed.

### Extension Points

`tasker` is designed to be highly pluggable through its function-based configuration, allowing you to seamlessly integrate it into various application contexts:

*   **`OnCreate func() (R, error)`**: This essential function defines how to instantiate your specific resource `R`. This could involve connecting to a database, initializing an external SDK client, setting up a specialized compute object, or any other resource provisioning logic.
*   **`OnDestroy func(R) error`**: This crucial function defines how to clean up your resource `R` when a worker shuts down, a temporary resource is no longer needed, or the `TaskManager` itself is shutting down. It's vital for releasing connections, closing files, deallocating memory, or performing other necessary resource finalization.
*   **`CheckHealth func(error) bool`**: Implement custom logic here to determine if a task's returned error indicates an underlying problem with the worker or its resource. Returning `false` from this function will cause `tasker` to consider the worker unhealthy, leading to its replacement and a potential retry of the task, significantly enhancing system resilience.
*   **`Logger Logger`**: Provide your own implementation of the `tasker.Logger` interface to integrate `tasker`'s internal logging messages with your application's preferred logging framework (e.g., `logrus`, `zap`). The default is a no-op logger.
*   **`Collector MetricsCollector`**: Supply a custom implementation of the `tasker.MetricsCollector` interface to integrate `tasker`'s rich performance metrics with your existing monitoring and observability stack (e.g., Prometheus, Datadog). A default internal collector is provided if none is specified.

---

## Development & Contributing

We welcome contributions to `tasker`! Whether it's bug reports, feature requests, or code contributions, your input is valuable.

### Development Setup

To set up your local development environment:

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
2.  **Keep your code clean and well-documented.** Follow Go's idiomatic style and best practices.
3.  **Write clear, concise commit messages** that describe your changes. We generally follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) (e.g., `feat: add new feature`, `fix: resolve bug`).
4.  **Ensure your changes are tested.** New features require new tests, and bug fixes should include a test that reproduces the bug.
5.  **Open a Pull Request** against the `main` branch. Provide a detailed description of your changes, including context, problem solved, and how it was solved.

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

*   **"Task manager is shutting down: cannot queue task"**: This error occurs if you attempt to `QueueTask`, `QueueTaskWithPriority`, `QueueTaskOnce`, `QueueTaskWithPriorityOnce`, or `RunTask` after `manager.Stop()` or `manager.Kill()` has been called, or if the `Ctx` provided during `NewTaskManager` initialization has been cancelled. Ensure tasks are only submitted while the manager is actively running.
*   **Resource Creation Failure**: If your `OnCreate` function returns an error, `tasker` will log it, and the associated worker will not start (or a `RunTask` call will fail). Ensure your resource creation logic is robust and handles transient issues gracefully.
*   **Workers Not Starting/Stopping as Expected**:
    *   Verify your `WorkerCount` and `MaxWorkerCount` settings in `tasker.Config`.
    *   For dynamic scaling (bursting), ensure `BurstInterval` is configured and not set to 0.
    *   Check that your main `context.Context` (passed as `Ctx` in `Config`) and its `cancel` function are managed correctly, especially for graceful shutdown scenarios.
    *   Review your `CheckHealth` logic; an incorrect implementation might lead to workers constantly restarting (thrashing) or not restarting when they should.
*   **Deadlocks/Goroutine Leaks**: While `tasker` is designed to prevent these within its core logic, improper usage (e.g., blocking indefinitely within your `OnCreate`, `OnDestroy`, or task functions, or not using buffered channels for task results outside the library) can lead to such issues. Always ensure your custom functions (`OnCreate`, `OnDestroy`, `taskFunc`) do not block indefinitely.

### FAQ

*   **When should I use `QueueTask` vs. `RunTask`?**
    *   Use `QueueTask` (or `QueueTaskWithPriority`) for asynchronous, background tasks that can wait in a queue for an available worker. This is ideal for high-throughput, batch processing, or any operation where immediate synchronous completion isn't strictly necessary.
    *   Use `RunTask` for synchronous, immediate tasks that require low latency and might need a resource right away, bypassing the queues. It's suitable for user-facing requests or critical operations that can't afford any queuing delay, such as generating a quick preview.
*   **How does `CheckHealth` affect workers and tasks?**
    *   If `CheckHealth` returns `false` for an error returned by a task, `tasker` considers the worker (or its underlying resource) to be in an unhealthy state. This unhealthy worker will be gracefully shut down, its resource destroyed (`OnDestroy`), and a new worker will be created to replace it. The original task that caused the unhealthy error will also be re-queued (up to `MaxRetries`) to be processed by a newly healthy worker. If `CheckHealth` returns `true` (or is `nil`), the error is considered a task-specific failure, not a worker health issue, and the worker continues operating.
*   **What happens if a task panics?**
    *   It is generally recommended that your task functions (the `func(R) (E, error)` you pass to `QueueTask` etc.) internally recover from panics and convert them into errors. If a panic occurs and is *not* recovered within the task function itself, it will crash the specific worker goroutine that was executing it. While `tasker` will detect the worker's exit and attempt to replace it, unhandled panics can lead to unexpected behavior, lost task results, and potential resource leaks if `OnDestroy` is not called due to an abrupt exit.
*   **Is `tasker` suitable for long-running tasks?**
    *   Yes, `tasker` can handle long-running tasks. However, be mindful of the parent `context.Context` passed to `NewTaskManager`. If that context is cancelled, all workers will attempt to gracefully shut down, potentially interrupting very long tasks. For indefinite tasks, consider managing them using separate Goroutines outside of `tasker` or ensure a very long-lived context for your `TaskManager` instance.

### Changelog / Roadmap

*   [**CHANGELOG.md**](CHANGELOG.md): See the project's history of changes and version releases.

### License

`tasker` is distributed under the MIT License. See [LICENSE.md](LICENSE.md) for the full text.

### Acknowledgments

This project is inspired by common worker pool patterns and the need for robust, flexible concurrency management in modern Go applications. It aims to provide a reliable foundation for building scalable backend services.
