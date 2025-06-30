# Task-Based Guide

### Immediate Synchronous Tasks

**Goal**: Execute a task immediately, blocking the caller until completion, without waiting in a queue. Ideal for user-facing, low-latency operations.

**Approach**: Use `RunTask`. This method is designed for urgent, synchronous operations. It attempts to get a resource from the pre-allocated `resourcePool`. If the pool is empty, it temporarily creates a new resource for the task's duration via `OnCreate`, uses it, and then destroys it via `OnDestroy`.

**Example: Generating a Fast Preview**
Imagine a scenario where a user needs an instant preview of an image. This task shouldn't wait in a queue behind other background image processing jobs.

```go
// examples/intermediate/main.go (Simplified)
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker"
)

type ImageProcessor struct { ID int }
func createImageProcessor() (*ImageProcessor, error) { return &ImageProcessor{ID: 1}, nil }
func destroyImageProcessor(p *ImageProcessor) error { return nil }
func checkImageProcessorHealth(err error) bool { return true }

func main() {
	config := tasker.Config[*ImageProcessor]{
		OnCreate:         createImageProcessor,
		OnDestroy:        destroyImageProcessor,
		WorkerCount:      2,
		Ctx:              context.Background(),
		ResourcePoolSize: 1, // Configure a pool for RunTask
	}
	manager, err := tasker.NewTaskManager[*ImageProcessor, string](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }
	defer manager.Stop()

	fmt.Println("Running an immediate task...")
	// This call blocks until the task completes
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

	time.Sleep(50 * time.Millisecond)
}
```

**Outcome**: The `RunTask` call blocks until the `fast_preview.jpg` is generated. The `ImageProcessor` resource is either pulled from the `resourcePool` or created temporarily, ensuring minimal latency.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [caller requires immediate blocking result and task should bypass queues] THEN [use `RunTask`] ELSE [consider `QueueTask` or `QueueTaskWithPriority`]"
  ],
  "verificationSteps": [
    "Check: `RunTask` call blocks and returns immediately upon task completion → Expected: No significant delay beyond task execution time.",
    "Check: `Stats().AvailableResources` decreases when `RunTask` is called and increases after (if from pool) → Expected: Resource pool usage reflects `RunTask` activity.",
    "Check: `OnCreate` is called temporarily if `ResourcePoolSize` is 0 or pool is empty and `RunTask` is called → Expected: Logs confirm temporary resource creation/destruction."
  ],
  "quickPatterns": [
    "Pattern: Synchronous `RunTask` execution\n```go\nresult, err := manager.RunTask(func(res *MyResource) (string, error) {\n    // Your immediate, synchronous task logic\n    return \"Immediate result\", nil\n})\nif err != nil { /* handle error */ }\n// Process result immediately\n```"
  ],
  "diagnosticPaths": [
    "Error `RunTask returns 'failed to create temporary resource'` -> Symptom: No resources are available or `OnCreate` fails for temporary resource -> Check: Ensure `Config.OnCreate` is robust; check if `ResourcePoolSize` is adequate or if resource creation is consistently failing -> Fix: Address `OnCreate` errors, consider increasing `ResourcePoolSize` or improving resource availability."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*