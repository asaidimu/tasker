# Task-Based Guide

### Resource Management & Health Checks

`tasker` provides robust mechanisms for managing the lifecycle of your application's resources and reacting to failures. This is achieved through the `OnCreate`, `OnDestroy`, and `CheckHealth` functions in the `Config` struct.

*   **`OnCreate func() (R, error)`**: This function is called whenever `tasker` needs a new resource, which happens when a worker starts up or when `RunTask` needs a temporary resource. It's your responsibility to initialize and return a healthy resource here. If `OnCreate` returns an error, the worker will not start (or `RunTask` will fail).

*   **`OnDestroy func(R) error`**: This function is called when a worker shuts down (gracefully or due to an unhealthy condition) or when a temporary resource used by `RunTask` is no longer needed. Use it to release any connections, close files, or perform cleanup.

*   **`CheckHealth func(error) bool`**: This optional but powerful callback allows you to define custom logic for determining if an error returned by a task signifies an "unhealthy" worker or resource. If `CheckHealth(err)` returns `false`:
    1.  The task will be re-queued (up to `Config.MaxRetries`) to be processed by a different worker.
    2.  The worker that returned the unhealthy error will be gracefully shut down and replaced with a new one (which involves calling `OnDestroy` on its old resource and `OnCreate` for a new one). This ensures that a faulty resource or worker doesn't continuously cause task failures.

    If `CheckHealth` is `nil` (default), all task errors are considered "healthy" and do not trigger worker replacement or task retries.

    **Example: Image Processing with Custom Health Checks**

    ```go
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

    	// ... (task queuing and RunTask calls as in example/intermediate/main.go) ...

    	// Give time for tasks and potential worker replacements
    	time.Sleep(1 * time.Second)

    	fmt.Println("Intermediate usage example finished.")
    }
    ```

### Dynamic Worker Scaling (Bursting)

`tasker` can dynamically adjust its worker count based on queue backlog, a feature called "bursting." This allows your application to efficiently handle fluctuating loads without over-provisioning resources during idle times.

*   **`BurstTaskThreshold int`**: If the total number of tasks in the main and priority queues exceeds this value, `tasker` will start adding burst workers.
*   **`BurstWorkerCount int`**: The number of additional burst workers to create each time the threshold is exceeded. This is a batch size for scaling up.
*   **`MaxWorkerCount int`**: The absolute maximum total number of workers (base + burst) that `tasker` will allow. This prevents unbounded scaling.
*   **`BurstInterval time.Duration`**: How frequently `tasker` checks the queue sizes to decide whether to scale up or down. A smaller interval means faster reaction to load changes but more frequent checks.

When the total queued tasks drop significantly below the `BurstTaskThreshold` (specifically, below `BurstTaskThreshold/2`), `tasker` will begin scaling down, stopping burst workers until the load stabilizes or only base workers remain.

**Example: Dynamic Scaling (Bursting)**

(See the `examples/advanced/main.go` file for a complete runnable example demonstrating bursting with 20 heavy computation tasks and real-time stats monitoring.)

### Graceful Shutdown

Properly shutting down `tasker` is critical to ensure that all in-flight tasks complete, no resources are leaked, and all goroutines exit cleanly. Always call `manager.Stop()` when your application is terminating.

`manager.Stop()` performs the following steps:

1.  Signals the burst manager to cease its scaling operations.
2.  Cancels all currently running burst workers, allowing them to finish their current tasks and exit.
3.  Cancels the main context, signaling all base workers to finish their current tasks and exit.
4.  Waits for all workers and the burst manager to complete their shutdown (using a `sync.WaitGroup`).
5.  Cleans up any remaining resources in the internal `RunTask` resource pool by calling `OnDestroy` on them.

`manager.Stop()` blocks until all these steps are complete, guaranteeing a clean exit.

```go
// ... (setup and run tasks) ...

// When it's time to shut down:
err := manager.Stop()
if err != nil {
    log.Printf("Error during graceful shutdown: %v", err)
} else {
    fmt.Println("Task manager successfully shut down.")
}
```

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF task_returns_error AND check_health_returns_false THEN re_queue_task_if_retries_remain AND shutdown_current_worker",
    "IF queue_size_exceeds_threshold AND current_workers_less_than_max THEN start_burst_workers",
    "IF queue_size_below_threshold_half AND burst_workers_exist THEN scale_down_burst_workers",
    "IF application_is_terminating THEN CALL manager.Stop()"
  ],
  "verificationSteps": [
    "Check: `OnCreate` and `OnDestroy` log messages appear correctly for each resource lifecycle event -> Expected: Resource management functions are called.",
    "Check: When a task fails with an unhealthy error and `MaxRetries > 0`, the task is re-queued and a new worker is created (observed by `OnCreate` log) -> Expected: Unhealthy worker replacement occurs.",
    "Check: `manager.Stats().BurstWorkers` increases when queue backlog grows and decreases when it shrinks -> Expected: Dynamic scaling is active.",
    "Check: After `manager.Stop()` completes, `manager.Stats().ActiveWorkers` is 0 and `OnDestroy` logs appear for all resources -> Expected: Graceful shutdown successful."
  ],
  "quickPatterns": [
    "Pattern: Implementing `CheckHealth` for resource failures\n```go\nimport \"errors\"\nfunc myCheckHealth(err error) bool {\n    if err != nil && errors.Is(err, ErrConnectionBroken) { return false }\n    return true\n}\n// Config.CheckHealth = myCheckHealth\n```",
    "Pattern: Configuring Bursting\n```go\nconfig := tasker.Config[*MyResource]{\n    // ... other config\n    BurstTaskThreshold: 10,\n    BurstWorkerCount: 5,\n    MaxWorkerCount: 20,\n    BurstInterval: 500 * time.Millisecond,\n}\n```",
    "Pattern: Graceful Shutdown in `main`\n```go\nfunc main() {\n    // ... setup manager ...\n    defer manager.Stop()\n    // ... run app ...\n}\n```"
  ],
  "diagnosticPaths": [
    "Error `WorkerReplacementFailure` -> Symptom: `OnCreate` fails when replacing an unhealthy worker -> Check: External resource availability or `OnCreate` logic -> Fix: Ensure `OnCreate` is robust.",
    "Error `BurstScalingStuck` -> Symptom: `BurstWorkers` count doesn't change despite queue size fluctuations -> Check: `Config.BurstTaskThreshold`, `Config.BurstWorkerCount`, `Config.MaxWorkerCount` values. Ensure `BurstInterval` is positive. -> Fix: Adjust burst configuration parameters.",
    "Error `ShutdownTimeout` -> Symptom: `manager.Stop()` blocks indefinitely -> Check: Worker task functions respect `context.Context` cancellation; `OnDestroy` functions complete quickly -> Fix: Add context checks in long-running tasks; optimize `OnDestroy` or ensure it handles blocking operations gracefully."
  ]
}
```

---
*Generated using Gemini AI on 6/14/2025, 1:47:53 PM. Review and refine as needed.*