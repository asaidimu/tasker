# Getting Started

### Quick Setup Guide

To begin using Tasker, ensure you have Go 1.24.3 or higher installed. Then, follow these steps to add it to your project and run a basic example.

#### Prerequisites
* Go 1.24.3 or higher

#### Installation
```bash
go get github.com/asaidimu/tasker
```

#### Basic Example: Simple Calculator
This example demonstrates how to set up a `TaskManager` with a `CalculatorResource` and queue basic arithmetic tasks.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker"
)

type CalculatorResource struct{}

func createCalcResource() (*CalculatorResource, error) {
	fmt.Println("INFO: Creating CalculatorResource")
	return &CalculatorResource{}, nil
}

func destroyCalcResource(r *CalculatorResource) error {
	fmt.Println("INFO: Destroying CalculatorResource")
	return nil
}

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

	fmt.Println("Queuing a simple addition task...")
	go func() {
		sum, err := manager.QueueTask(func(r *CalculatorResource) (int, error) {
			time.Sleep(50 * time.Millisecond)
			a, b := 10, 25
			fmt.Printf("Worker processing: %d + %d\n", a, b)
			return a + b, nil
		})

		if err != nil {
			fmt.Printf("Task 1 failed: %v\n", err)
		} else {
			fmt.Printf("Task 1 (Addition) Result: %d\n", sum)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	stats := manager.Stats()
	fmt.Printf("\n--- Current Stats ---\n")
	fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers)
	fmt.Printf("Queued Tasks: %d\n", stats.QueuedTasks)
	fmt.Printf("Available Resources: %d\n", stats.AvailableResources)
	fmt.Println("----------------------")
}
```

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [],
  "verificationSteps": [
    "Check: `go get github.com/asaidimu/tasker` executed successfully → Expected: Package installed, no errors.",
    "Check: `go run main.go` produces output similar to example → Expected: \"INFO: Creating CalculatorResource\" and task results are printed."
  ],
  "quickPatterns": [
    "Pattern: Basic TaskManager Initialization\n```go\n// Define your resource type\ntype MyResource struct{}\n\n// Implement onCreate and onDestroy functions\nfunc createMyResource() (*MyResource, error) { /* ... */ }\nfunc destroyMyResource(r *MyResource) error { /* ... */ }\n\n// Configure and create TaskManager\nconfig := tasker.Config[*MyResource]{\n    OnCreate: createMyResource,\n    OnDestroy: destroyMyResource,\n    WorkerCount: 2,\n    Ctx: context.Background(),\n}\nmanager, err := tasker.NewTaskManager[*MyResource, any](config)\nif err != nil { log.Fatal(err) }\ndefer manager.Stop()\n```"
  ],
  "diagnosticPaths": [
    "Error `Error creating task manager: worker count must be positive` -> Symptom: Manager creation fails with configuration error -> Check: Verify `Config.WorkerCount` is > 0 -> Fix: Set `WorkerCount` to a positive integer.",
    "Error `Error creating task manager: onCreate function is required` -> Symptom: Manager creation fails due to missing resource creator -> Check: Ensure `Config.OnCreate` is assigned a non-nil function -> Fix: Provide a valid `OnCreate` function."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*