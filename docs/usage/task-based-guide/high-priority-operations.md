# Task-Based Guide

### High-Priority Operations

**Goal**: Ensure that critical tasks are processed with higher precedence than regular tasks, even under heavy load.

**Approach**: Use `QueueTaskWithPriority`. Tasks submitted via this method are added to a dedicated priority queue, which workers check *before* the main task queue. This guarantees that urgent operations are picked up and executed sooner.

**Example: Urgent Thumbnail Generation**
In an image processing system, generating a thumbnail for a newly uploaded video might be more urgent than resizing a static image. This example demonstrates how `QueueTaskWithPriority` facilitates this prioritization.

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
		OnCreate:    createImageProcessor,
		OnDestroy:   destroyImageProcessor,
		WorkerCount: 2,
		Ctx:         context.Background(),
	}
	manager, err := tasker.NewTaskManager[*ImageProcessor, string](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }
	defer manager.Stop()

	// Queue a normal image resize task (takes longer)
	go func() {
		fmt.Println("Queueing normal resize...")
		result, err := manager.QueueTask(func(proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing normal resize\n", proc.ID)
			time.Sleep(150 * time.Millisecond)
			return "imageA_resized.jpg", nil
		})
		if err != nil { fmt.Printf("Normal Resize Failed: %v\n", err) }
		else { fmt.Printf("Normal Resize Completed: %s\n", result) }
	}()

	// Queue a high-priority thumbnail generation task (faster, but queued after normal)
	go func() {
		fmt.Println("Queueing high-priority thumbnail...")
		result, err := manager.QueueTaskWithPriority(func(proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing HIGH PRIORITY thumbnail\n", proc.ID)
			time.Sleep(50 * time.Millisecond)
			return "video_thumbnail.jpg", nil
		})
		if err != nil { fmt.Printf("Priority Thumbnail Failed: %v\n", err) }
		else { fmt.Printf("Priority Thumbnail Completed: %s\n", result) }
	}()

	time.Sleep(500 * time.Millisecond) // Allow tasks to run
}
```

**Outcome**: Despite being potentially queued later, the "HIGH PRIORITY thumbnail" task is likely to be processed and completed before the "normal resize" task, demonstrating the effect of priority queuing.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [task is critical to user experience or system responsiveness] THEN [use `QueueTaskWithPriority`] ELSE [use `QueueTask`]",
    "IF [multiple tasks are queued concurrently and order of execution matters for urgency] THEN [prioritize critical tasks with `QueueTaskWithPriority`]"
  ],
  "verificationSteps": [
    "Check: High-priority task completes before or significantly faster than a longer normal task queued before it (under load) → Expected: Output order confirms priority execution.",
    "Check: `Stats().PriorityTasks` count increases upon submission and decreases upon processing → Expected: Live stats reflect priority queue utilization."
  ],
  "quickPatterns": [
    "Pattern: High-priority task with result handling\n```go\ngo func() {\n    result, err := manager.QueueTaskWithPriority(func(res *MyResource) (string, error) {\n        // Your high-priority task logic\n        return \"Critical task done\", nil\n    })\n    if err != nil { /* handle error */ }\n    else { /* process result */ }\n}()\n```"
  ],
  "diagnosticPaths": [
    "Error `Priority task not completing faster than normal tasks` -> Symptom: Priority queuing appears ineffective -> Check: Verify enough workers are active to pick up tasks; ensure priority queue is not full preventing submission -> Fix: Adjust `WorkerCount` or `MaxWorkerCount`, check for blocking operations in `OnCreate`."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*