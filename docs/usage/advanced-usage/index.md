# Advanced Usage

### Graceful Shutdown (`Stop`)
Initiates a graceful shutdown of the task manager. It stops accepting new tasks, waits for all currently queued tasks to be completed, and then releases all managed resources. This is the recommended way to shut down your `tasker` instance in production environments to ensure no in-flight tasks are lost.

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

### Immediate Shutdown (`Kill`)
Immediately terminates the task manager. It cancels all running tasks, drops all queued tasks, and releases resources without waiting for tasks to complete. Use `Kill()` for scenarios where immediate termination is required, even at the cost of losing in-flight tasks (e.g., during critical error recovery or unit testing).

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
	// Expect 0 active workers and the long task to be cancelled
}
```

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF application_exit_strategy_requires_task_completion THEN USE Stop() ELSE USE Kill()",
    "IF graceful_shutdown_required THEN CALL Stop() AND AWAIT_COMPLETION ELSE CALL Kill()",
    "IF immediate_resource_release_needed THEN CALL Kill() ELSE (IF resources_have_stateful_cleanup THEN CALL Stop() AND WAIT_FOR_DRAIN ELSE CALL Stop() WITHOUT_WAIT)"
  ],
  "verificationSteps": [
    "Check: `manager.Stop()` returns nil error AND `manager.Stats().ActiveWorkers` is 0 AND `manager.Stats().QueuedTasks` is 0 -> Expected: All tasks completed, workers shut down",
    "Check: `manager.Kill()` returns nil error AND `manager.Stats().ActiveWorkers` is 0 -> Expected: All workers immediately terminated, queued tasks dropped"
  ],
  "quickPatterns": [
    "Pattern: Graceful_shutdown\n```go\n// On application exit or signal handler\nif err := manager.Stop(); err != nil {\n    log.Printf(\"Shutdown error: %v\", err)\n}\nlog.Println(\"Manager gracefully stopped.\")\n```",
    "Pattern: Immediate_shutdown\n```go\n// On critical error or immediate termination signal\nif err := manager.Kill(); err != nil {\n    log.Printf(\"Kill error: %v\", err)\n}\nlog.Println(\"Manager immediately killed.\")\n```"
  ],
  "diagnosticPaths": [
    "Error: Tasks are lost during shutdown -> Symptom: `manager.Stats().QueuedTasks` is non-zero after `Stop()` or `Kill()` was called -> Check: If `Kill()` was used, this is expected. If `Stop()` was used, verify tasks are properly consuming `ctx.Done()` or are not blocking indefinitely. -> Fix: Use `Stop()` for graceful shutdown. Ensure long-running tasks inside `taskFunc` check `context.Context` for cancellation.",
    "Error: Application hangs on shutdown -> Symptom: `manager.Stop()` or `manager.Kill()` does not return -> Check: Verify all worker goroutines eventually exit, check for deadlocks in `OnCreate`, `OnDestroy`, or `taskFunc` -> Fix: Ensure all custom functions are non-blocking and return in a timely manner. If tasks are truly infinite, manage them outside of `tasker` or provide a very long-lived context."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 9:35:09 PM. Review and refine as needed.*