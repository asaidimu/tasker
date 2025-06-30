# Core Operations

### Shutdown Procedures
Tasker provides two distinct methods for shutting down the `TaskManager`, offering control over how currently executing and queued tasks are handled.

#### `Stop() error`
Initiates a graceful shutdown of the `TaskManager`. This is the recommended shutdown method for most applications.

**Behavior:**
1.  Stops accepting new tasks. Any subsequent calls to `QueueTask`, `RunTask`, etc., will immediately return an error.
2.  Signals all existing workers to enter a "drain" mode. Workers will finish processing any tasks currently in their queues (`mainQueue` and `priorityQueue`) before exiting.
3.  Waits for all active and draining workers to complete.
4.  Releases all managed resources by calling `OnDestroy` for each.

**When to Use:**
*   When you need to ensure all outstanding tasks are completed before the application exits or the `TaskManager` is no longer needed.
*   For controlled application shutdowns or during hot reloads where service continuity is important.

#### `Kill() error`
Immediately terminates the `TaskManager` without waiting for queued tasks to complete. This is useful for emergency shutdowns or when rapid termination is prioritized over task completion.

**Behavior:**
1.  Stops accepting new tasks, similar to `Stop()`.
2.  Cancels the underlying `context.Context` that all worker goroutines are derived from, causing them to exit immediately without processing any more queued tasks.
3.  Drops all tasks currently waiting in the `mainQueue` and `priorityQueue`.
4.  Releases all managed resources by calling `OnDestroy` for each.

**When to Use:**
*   During application crashes or unrecoverable error states where immediate resource release is necessary.
*   In testing scenarios where quick teardown is desired.

```go
// Example: Graceful vs. Immediate Shutdown (based on examples/advanced/main.go)
package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/asaidimu/tasker"
)

type HeavyComputeResource struct { ID int }
func createComputeResource() (*HeavyComputeResource, error) { /* ... */ return &HeavyComputeResource{ID: 1}, nil}
func destroyComputeResource(r *HeavyComputeResource) error { /* ... */ return nil}
func checkComputeHealth(err error) bool { return true }

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	config := tasker.Config[*HeavyComputeResource]{
		OnCreate:    createComputeResource,
		OnDestroy:   destroyComputeResource,
		WorkerCount: 2,
		Ctx:         ctx,
	}

	manager, err := tasker.NewTaskManager[*HeavyComputeResource, string](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }

	var tasksSubmitted atomic.Int32
	for i := 0; i < 5; i++ {
		tasksSubmitted.Add(1)
		go func(taskID int) {
			_, err := manager.QueueTask(func(res *HeavyComputeResource) (string, error) {
				fmt.Printf("Worker %d processing Task %d\n", res.ID, taskID)
				time.Sleep(200 * time.Millisecond) // Simulate work
				return fmt.Sprintf("Task %d completed", taskID), nil
			})
			if err != nil { fmt.Printf("Task %d failed: %v\n", taskID, err) }
		}(i)
	}

	fmt.Println("Allowing tasks to start...")
	time.Sleep(100 * time.Millisecond)

	fmt.Println("Initiating graceful shutdown...")
	err = manager.Stop() // This will wait for all 5 tasks to finish
	if err != nil { fmt.Printf("Error during shutdown: %v\n", err) }
	fmt.Println("Task manager gracefully shut down.")

	// For a 'Kill' example, replace manager.Stop() with manager.Kill()
	// and observe tasks being cancelled immediately.
	// manager.Kill()
}
```

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [all in-flight and queued tasks must complete] THEN [call `manager.Stop()`] ELSE [call `manager.Kill()` (to immediately terminate)]",
    "IF [application needs to shut down quickly regardless of task completion] THEN [call `manager.Kill()`] ELSE [call `manager.Stop()` (for graceful termination)]"
  ],
  "verificationSteps": [
    "Check: `manager.Stop()` called → Expected: All queued tasks are eventually processed and complete before `Stop()` returns.",
    "Check: `manager.Kill()` called → Expected: `Kill()` returns quickly, and tasks in queues are not processed, running tasks may be interrupted.",
    "Check: After shutdown, `Stats().ActiveWorkers` is 0 → Expected: All worker goroutines have exited and resources are released."
  ],
  "quickPatterns": [
    "Pattern: Graceful shutdown\n```go\nmanager, _ := tasker.NewTaskManager[MyResource, any](config)\ndefer manager.Stop() // Recommended for most applications\n// ... queue tasks ...\n```",
    "Pattern: Immediate shutdown (e.g., for testing or emergency)\n```go\nmanager, _ := tasker.NewTaskManager[MyResource, any](config)\n// ... queue tasks ...\nerr := manager.Kill() // Forceful termination\nif err != nil { log.Printf(\"Kill failed: %v\", err) }\n```"
  ],
  "diagnosticPaths": [
    "Error `manager.Stop() hangs indefinitely` -> Symptom: Application does not exit after calling `Stop()` -> Check: Ensure all task functions eventually complete or respond to `context.Context` cancellation -> Fix: Add context cancellation checks within long-running tasks or ensure `OnDestroy` is non-blocking.",
    "Error `tasks fail with 'task manager is shutting down' immediately after calling Stop()` -> Symptom: New tasks are rejected immediately -> Check: Confirm `Stop()` was intended to be called before new task submissions ceased -> Fix: Submit tasks only when `TaskManager` is in a running state."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*