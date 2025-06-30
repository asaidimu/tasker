# Task-Based Guide

### Resilient Error Handling and At-Most-Once Execution

**Goal**: Build a resilient task processing system that automatically recovers from worker/resource failures and correctly handles tasks that should not be retried.

**Approach**: Utilize `Config.CheckHealth` to define what constitutes an "unhealthy" worker/resource error, enabling automatic worker replacement and task retries (`Config.MaxRetries`). For non-idempotent operations, use `QueueTaskOnce` or `QueueTaskWithPriorityOnce` to prevent automatic re-queuing on health-related failures.

**Example: Image Processing with Crash Simulation**
This scenario simulates an image processor crashing due to a bad input, triggering the health check and potentially a retry, or preventing a retry for a sensitive operation.

```go
// examples/intermediate/main.go (Excerpt)
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

type ImageProcessor struct { ID int }
func createImageProcessor() (*ImageProcessor, error) { /* ... */ return &ImageProcessor{ID: 1}, nil}
func destroyImageProcessor(p *ImageProcessor) error { /* ... */ return nil}

// Custom health check: "processor_crash" means the worker is unhealthy and needs replacement.
func checkImageProcessorHealth(err error) bool {
	if err != nil && err.Error() == "processor_crash" {
		fmt.Printf("WARN: Detected unhealthy error: %v. Worker will be replaced.\n", err)
		return false // This error indicates an unhealthy state
	}
	return true // Other errors are just task failures, not worker health issues
}

func main() {
	config := tasker.Config[*ImageProcessor]{
		OnCreate:    createImageProcessor,
		OnDestroy:   destroyImageProcessor,
		WorkerCount: 2,
		Ctx:         context.Background(),
		CheckHealth: checkImageProcessorHealth, // Use custom health check
		MaxRetries:  1,                         // One retry on unhealthy errors
	}
	manager, err := tasker.NewTaskManager[*ImageProcessor, string](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }
	defer manager.Stop()

	// Task that might cause an "unhealthy" error (will be retried once)
	go func() {
		fmt.Println("Queueing task that might crash a worker (retried)...")
		_, err := manager.QueueTask(func(proc *ImageProcessor) (string, error) {
			if rand.Intn(2) == 0 { return "", errors.New("processor_crash") }
			return "processed_retried", nil
		})
		if err != nil { fmt.Printf("Retried task failed: %v\n", err) }
		else { fmt.Println("Retried task completed.") }
	}()

	// Task that might cause an "unhealthy" error (NOT retried - at-most-once)
	go func() {
		fmt.Println("Queueing task that might crash (at-most-once)...")
		_, err := manager.QueueTaskOnce(func(proc *ImageProcessor) (string, error) {
			if rand.Intn(2) == 0 { return "", errors.New("processor_crash") }
			return "processed_once", nil
		})
		if err != nil { fmt.Printf("At-most-once task failed: %v\n", err) }
		else { fmt.Println("At-most-once task completed.") }
	}()

	time.Sleep(2 * time.Second) // Allow time for tasks and potential worker replacements
}
```

**Outcome**: When a task returns a `processor_crash` error:
*   For `QueueTask`, `CheckHealth` identifies it as unhealthy, the worker is replaced, and the task is re-queued for another attempt (up to `MaxRetries`).
*   For `QueueTaskOnce`, `CheckHealth` still identifies it as unhealthy and the worker is replaced, but the task is *not* re-queued, ensuring it's attempted at most once by the `TaskManager`'s retry logic.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [task failure means the worker/resource is broken and needs replacement] THEN [implement `CheckHealth` to return `false`] ELSE [treat as task-specific error only]",
    "IF [operation is non-idempotent and should not be re-executed by the manager if worker fails] THEN [use `QueueTaskOnce` or `QueueTaskWithPriorityOnce`] ELSE [use standard `QueueTask` or `QueueTaskWithPriority`]"
  ],
  "verificationSteps": [
    "Check: `CheckHealth` returns `false` for specific error → Expected: Worker `OnDestroy` is called, new worker `OnCreate` is called, and task is retried (if not `*Once` task).",
    "Check: `QueueTaskOnce` fails due to unhealthy worker → Expected: Task result/error is returned to caller, but task is NOT re-queued internally by Tasker."
  ],
  "quickPatterns": [
    "Pattern: Task with potential unhealthy error\n```go\n_, err := manager.QueueTask(func(res *MyResource) (string, error) {\n    if someCondition { return \"\", errors.New(\"unrecoverable_resource_error\") }\n    return \"ok\", nil\n})\n```",
    "Pattern: At-most-once task\n```go\n_, err := manager.QueueTaskOnce(func(res *MyResource) (string, error) {\n    // Non-idempotent operation\n    return \"ok\", nil\n})\n```"
  ],
  "diagnosticPaths": [
    "Error `Max retries exceeded for task` -> Symptom: Task repeatedly fails and is not processed -> Check: Review `CheckHealth` and task logic. The task is consistently causing an unhealthy state, or `MaxRetries` is too low -> Fix: Debug the root cause of the unhealthy state, increase `MaxRetries` for transient issues, or use `*Once` tasks if re-execution is undesired.",
    "Error `Task not retried as expected` -> Symptom: Task fails but `CheckHealth` indicated unhealthy state, yet no retry occurred -> Check: Ensure the task was not submitted using `*Once` methods; verify `MaxRetries` is greater than 0 -> Fix: Adjust task submission method or `MaxRetries`."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*