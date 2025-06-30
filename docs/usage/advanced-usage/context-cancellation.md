# Advanced Usage

### Context Cancellation and Task Lifecycles

Tasker relies heavily on `context.Context` for managing the lifecycle of its workers and the `TaskManager` itself. Understanding how contexts are used is crucial for proper integration and shutdown.

#### `Config.Ctx`
The `Ctx` field in `tasker.Config` is the parent `context.Context` for the entire `TaskManager`. All worker goroutines and the internal burst manager goroutine derive their contexts from this parent context. 

**Impact of `Config.Ctx` Cancellation:**
*   **Graceful Shutdown (`Stop()`):** When `manager.Stop()` is called, Tasker cancels its internal context (which is derived from `Config.Ctx` by `NewTaskManager`). This signals all workers to enter a drain mode, finishing any queued tasks before exiting.
*   **Immediate Shutdown (`Kill()`):** When `manager.Kill()` is called, Tasker also cancels its internal context, but signals workers to exit immediately without draining queues. This propagates cancellation to any long-running tasks that respect the context.
*   **External Cancellation:** If the `cancel` function associated with the `Config.Ctx` *you provided* is called externally, it will trigger Tasker's graceful shutdown procedure, as if `manager.Stop()` was implicitly called. This is a common pattern for integrating Tasker into larger applications that manage their lifecycles via a root context.

#### Propagating Contexts to Tasks
While `Tasker` manages the worker contexts, your individual task functions (`func(R) (E, error)`) do *not* directly receive a `context.Context`. If your task logic needs to respect external cancellation signals (e.g., for long-running I/O operations), you must pass a context into your task function using a closure or by modifying your resource `R` to carry a context.

**Example: Task with Internal Context Handling**
```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker"
)

type DataProcessor struct{}
func createDataProcessor() (*DataProcessor, error) { return &DataProcessor{}, nil }
func destroyDataProcessor(r *DataProcessor) error { return nil }
func checkDataHealth(err error) bool { return true }

func main() {
	// Create a parent context for the TaskManager
	appCtx, appCancel := context.WithCancel(context.Background())

	config := tasker.Config[*DataProcessor]{
		OnCreate:    createDataProcessor,
		OnDestroy:   destroyDataProcessor,
		WorkerCount: 1,
		Ctx:         appCtx, // TaskManager derives its context from appCtx
	}
	manager, err := tasker.NewTaskManager[*DataProcessor, string](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }
	// Do NOT defer manager.Stop() here, we will call appCancel manually.

	// Task that respects its own context for cancellation
	go func() {
		// Create a task-specific context for a long operation
		taskCtx, taskCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer taskCancel()

		result, err := manager.QueueTask(func(p *DataProcessor) (string, error) {
			fmt.Println("Task: Starting long operation...")
			select {
			case <-time.After(1 * time.Second): // Simulate very long work
				return "Long operation completed", nil
			case <-taskCtx.Done():
				return "", taskCtx.Err() // Task cancelled by its own context
			}
		})

		if err != nil { fmt.Printf("Task finished with error: %v\n", err) }
		else { fmt.Printf("Task result: %s\n", result) }
	}()

	// Simulate external signal to stop the application (and thus the TaskManager)
	time.Sleep(200 * time.Millisecond)
	fmt.Println("Application: Signaling cancellation...")
	appCancel() // Cancelling appCtx will trigger Tasker's graceful shutdown

	time.Sleep(1 * time.Second) // Allow time for TaskManager to shut down
	fmt.Println("Application: Exited.")
}
```

**Outcome**: The `appCancel()` call triggers Tasker's shutdown. The long-running task, if it respects `taskCtx.Done()`, will be cancelled by its *own* context's timeout, and the Tasker will gracefully shut down after the task function returns its error.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [long-running tasks need to be interruptible by application shutdown or explicit cancellation] THEN [pass a dedicated `context.Context` into the task function (e.g., via closure) and ensure task respects `ctx.Done()`] ELSE [tasks will run to completion or only be interrupted by `manager.Kill()`]"
  ],
  "verificationSteps": [
    "Check: Cancelling `Config.Ctx` causes `manager.Stop()` behavior → Expected: Workers drain queues and exit gracefully.",
    "Check: Task function internally checks `context.Context.Done()` → Expected: Long-running task is interrupted when its passed context is cancelled."
  ],
  "quickPatterns": [
    "Pattern: Passing context into a task function via closure\n```go\nfunc processData(ctx context.Context, data []byte) error {\n    select {\n    case <-ctx.Done():\n        return ctx.Err()\n    case <-time.After(100 * time.Millisecond):\n        // Simulate work\n        return nil\n    }\n}\n\n// ... in main/caller ...\n\ntaskCtx, taskCancel := context.WithCancel(context.Background())\ndefer taskCancel()\n\n_, err := manager.QueueTask(func(res *MyResource) (string, error) {\n    err := processData(taskCtx, someData)\n    return \"\", err\n})\n```"
  ],
  "diagnosticPaths": [
    "Error `Tasks not cancelling/exiting during graceful shutdown` -> Symptom: `manager.Stop()` hangs, workers continue processing indefinitely -> Check: Ensure long-running task functions check `context.Context.Done()` at appropriate intervals and return an error -> Fix: Instrument tasks to respond to context cancellation.",
    "Error `TaskManager not shutting down when external root context is cancelled` -> Symptom: Application hangs despite `appCtx.Cancel()` being called -> Check: Confirm the `Config.Ctx` provided to `NewTaskManager` is the correct parent context that is being cancelled -> Fix: Ensure `Config.Ctx` is properly linked to the external cancellation mechanism."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*