---
outline: "deep"
lastUpdated: true
editLink: true
prev: true
next: true
title: "Core Operations"
description: "Core Operations documentation and guidance"
---
# Core Operations

### Queueing Tasks
`tasker` provides several methods to queue tasks for asynchronous processing. Tasks in the main queue are processed by available workers on a FIFO basis. For critical operations, you can use a dedicated high-priority queue.

**`QueueTask` (Blocking)**
Adds a task to the standard queue. The call blocks until the task completes and returns its result or error.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

type CalculatorResource struct{}
func createCalcResource() (*CalculatorResource, error) { return &CalculatorResource{}, nil }
func destroyCalcResource(r *CalculatorResource) error { return nil }

func main() {
	ctx := context.Background()
	config := tasker.Config[*CalculatorResource]{
		OnCreate:    createCalcResource,
		OnDestroy:   destroyCalcResource,
		WorkerCount: 2,
		Ctx:         ctx,
	}
	manager, err := tasker.NewTaskManager[*CalculatorResource, int](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	fmt.Println("Queuing a simple Multiplication task...")
	task1Start := time.Now()
	go func() {
		sum, err := manager.QueueTask(func(ctx context.Context, r *CalculatorResource) (int, error) {
			time.Sleep(50 * time.Millisecond)
			a, b := 10, 25
			fmt.Printf("Worker processing: %d * %d\n", a, b)
			return a * b, nil
		})

		if err != nil {
			fmt.Printf("Task 1 failed: %v\n", err)
		} else {
			fmt.Printf("Task 1 (Multiplication) Result: %d (took %s)\n", sum, time.Since(task1Start))
		}
	}()
	// Allow some time for task to complete
	time.Sleep(100 * time.Millisecond)
}
```

**`QueueTaskWithPriority` (Blocking)**
Adds a task to a dedicated high-priority queue. Tasks in this queue are processed before tasks in the main queue.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

type ImageProcessor struct{ ID int }
func createImageProcessor() (*ImageProcessor, error) { return &ImageProcessor{ID: 1}, nil }
func destroyImageProcessor(p *ImageProcessor) error { return nil }

func main() {
	ctx := context.Background()
	config := tasker.Config[*ImageProcessor]{
		OnCreate:    createImageProcessor,
		OnDestroy:   destroyImageProcessor,
		WorkerCount: 1,
		Ctx:         ctx,
	}
	manager, err := tasker.NewTaskManager[*ImageProcessor, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	fmt.Println("Queueing a high-priority thumbnail task...")
	go func() {
		result, err := manager.QueueTaskWithPriority(func(ctx context.Context, proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing HIGH PRIORITY thumbnail\n", proc.ID)
			time.Sleep(50 * time.Millisecond)
			return "video_thumbnail.jpg", nil
		})
		if err != nil {
			fmt.Printf("Priority Thumbnail Failed: %v\n", err)
		} else {
			fmt.Printf("Priority Thumbnail Completed: %s\n", result)
		}
	}()
	time.Sleep(100 * time.Millisecond)
}
```

### Immediate Task Execution (`RunTask`)
`RunTask` executes a task immediately, bypassing the main and priority queues. It attempts to acquire a resource from the internal pool. If no resource is immediately available, it temporarily creates a new one for the task. This method is suitable for urgent tasks that should not be delayed by queuing. It is a synchronous call, blocking until the task finishes.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker/v2"
)

type ImageProcessor struct{ ID int }
func createImageProcessor() (*ImageProcessor, error) { return &ImageProcessor{ID: 1}, nil }
func destroyImageProcessor(p *ImageProcessor) error { return nil }

func main() {
	ctx := context.Background()
	config := tasker.Config[*ImageProcessor]{
		OnCreate:         createImageProcessor,
		OnDestroy:        destroyImageProcessor,
		WorkerCount:      1,
		Ctx:              ctx,
		ResourcePoolSize: 1, // Enable resource pooling for RunTask
	}
	manager, err := tasker.NewTaskManager[*ImageProcessor, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	fmt.Println("Running an immediate task...")
	immediateResult, immediateErr := manager.RunTask(func(ctx context.Context, proc *ImageProcessor) (string, error) {
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

