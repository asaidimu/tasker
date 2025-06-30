# Core Operations

### Resource Management and Health Checks
Tasker provides robust mechanisms for managing the lifecycle of your task resources and reacting to their health status.

#### `OnCreate func() (R, error)`
This function is crucial for initializing and preparing your resources. It's called when a worker starts up or when `RunTask` needs a temporary resource. `OnCreate` should establish connections, load configurations, or perform any setup necessary for your resource `R` to be operational.

#### `OnDestroy func(R) error`
This function handles the cleanup and deallocation of your resources. It's called when a worker shuts down, a temporary `RunTask` resource is no longer needed, or the entire `TaskManager` is stopping/killing. `OnDestroy` should close connections, release memory, or perform any finalization to prevent resource leaks.

#### `CheckHealth func(error) bool`
An optional but highly recommended function for defining custom logic to determine if an error returned by a task indicates an "unhealthy" state for the worker or its associated resource. If `CheckHealth` returns `false` (meaning unhealthy):

1.  The worker processing that task is considered faulty and will be shut down.
2.  Its resource will be destroyed via `OnDestroy`.
3.  A new worker will be created to replace the unhealthy one.
4.  The original task that caused the unhealthy error will be re-queued (up to `MaxRetries` times) to be processed by a newly healthy worker.

If `CheckHealth` returns `true` (or if the function is `nil`), the error is considered a task-specific failure that does not impact worker health, and the worker continues operating.

```go
// Example: Image Processing with custom health check (from examples/intermediate/main.go)
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

type ImageProcessor struct {
	ID        int
	IsHealthy bool
}

func createImageProcessor() (*ImageProcessor, error) {
	id := rand.Intn(1000)
	fmt.Printf("INFO: Creating ImageProcessor %d\n", id)
	return &ImageProcessor{ID: id, IsHealthy: true}, nil
}

func destroyImageProcessor(p *ImageProcessor) error {
	fmt.Printf("INFO: Destroying ImageProcessor %d\n", p.ID)
	return nil
}

func checkImageProcessorHealth(err error) bool {
	if err != nil && err.Error() == "processor_crash" {
		fmt.Printf("WARN: Detected unhealthy error: %v. Worker will be replaced.\n", err)
		return false // This error indicates an unhealthy state
	}
	return true // Other errors are just task failures, not worker health issues
}

func main() {
	ctx := context.Background()

	config := tasker.Config[*ImageProcessor]{
		OnCreate:    createImageProcessor,
		OnDestroy:   destroyImageProcessor,
		WorkerCount: 2,
		Ctx:         ctx,
		CheckHealth: checkImageProcessorHealth, // Custom health check applied here
		MaxRetries:  1,
	}

	manager, err := tasker.NewTaskManager[*ImageProcessor, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	// Simulate a task that might crash a worker
	go func() {
		_, err := manager.QueueTask(func(proc *ImageProcessor) (string, error) {
			if rand.Intn(2) == 0 { // 50% chance to simulate a crash
				return "", errors.New("processor_crash") // Triggers CheckHealth to return false
			}
			return "processed", nil
		})
		if err != nil { fmt.Printf("Problematic task failed: %v\n", err) }
	}()

	time.Sleep(1 * time.Second) // Allow time for potential worker replacement
}
```

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [task failure indicates resource/worker malfunction (e.g., connection lost)] THEN [implement `CheckHealth` to return `false` for that error] ELSE [allow `CheckHealth` to return `true` (default behavior)]"
  ],
  "verificationSteps": [
    "Check: `OnCreate` is called upon worker startup or `RunTask` with empty pool → Expected: Resource initialization logs appear.",
    "Check: `OnDestroy` is called upon worker shutdown or `RunTask` resource disposal → Expected: Resource cleanup logs appear.",
    "Check: Task fails with error that `CheckHealth` marks as unhealthy → Expected: Worker is replaced, `OnDestroy` and `OnCreate` are called again for the replacement worker, and the task might be retried."
  ],
  "quickPatterns": [
    "Pattern: Basic `OnCreate`/`OnDestroy`\n```go\nfunc onCreateDB() (*sql.DB, error) {\n    db, err := sql.Open(\"postgres\", \"conn_str\")\n    // Handle err, ping db, etc.\n    return db, err\n}\n\nfunc onDestroyDB(db *sql.DB) error {\n    return db.Close()\n}\n```",
    "Pattern: Custom `CheckHealth` for specific errors\n```go\nfunc checkDBHealth(err error) bool {\n    if err != nil && strings.Contains(err.Error(), \"connection refused\") {\n        return false // Unhealthy, worker should be replaced\n    }\n    return true // Other errors are just task failures\n}\n```"
  ],
  "diagnosticPaths": [
    "Error `Worker failed to create resource: [details]` -> Symptom: Workers fail to start -> Check: Debug `Config.OnCreate` function for logic errors, network issues, or invalid credentials -> Fix: Correct `OnCreate` implementation, ensure external dependencies are reachable.",
    "Error `Error destroying resource: [details]` -> Symptom: Resource leaks or warnings on worker exit -> Check: Debug `Config.OnDestroy` function for cleanup failures -> Fix: Ensure `OnDestroy` properly releases all resource components (e.g., closing file descriptors, network connections).",
    "Error `Unhealthy worker is exiting.` -> Symptom: Workers are frequently created/destroyed (thrashing) -> Check: Review `CheckHealth` logic; ensure it only returns `false` for genuinely unhealthy, unrecoverable worker states, not transient task errors -> Fix: Refine `CheckHealth` to be more precise."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*