# Task-Based Guide

### Asynchronous Background Jobs

**Goal**: Execute tasks in the background without blocking the main application flow, ensuring they are processed as workers become available.

**Approach**: Use `QueueTask` for standard, asynchronous processing. The call to `QueueTask` itself blocks until the task *completes* and its result/error is returned, making it easy to handle outcomes directly where the task is initiated. To truly run in the background, `QueueTask` calls are typically wrapped in a new goroutine.

**Example: Simple Addition and Subtraction**
This example showcases how `QueueTask` can be used to offload computational work to a managed worker pool.

```go
// examples/basic/main.go (Simplified)
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
	return &CalculatorResource{}, nil
}

func destroyCalcResource(r *CalculatorResource) error {
	return nil
}

func main() {
	config := tasker.Config[*CalculatorResource]{
		OnCreate:    createCalcResource,
		OnDestroy:   destroyCalcResource,
		WorkerCount: 2,
		Ctx:         context.Background(),
	}

	manager, err := tasker.NewTaskManager[*CalculatorResource, int](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	// Task 1: Addition
	go func() {
		sum, err := manager.QueueTask(func(r *CalculatorResource) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 10 + 25, nil
		})
		if err != nil { fmt.Printf("Addition task failed: %v\n", err) }
		else { fmt.Printf("Addition Result: %d\n", sum) }
	}()

	// Task 2: Subtraction
	go func() {
		diff, err := manager.QueueTask(func(r *CalculatorResource) (int, error) {
			time.Sleep(70 * time.Millisecond)
			return 100 - 40, nil
		})
		if err != nil { fmt.Printf("Subtraction task failed: %v\n", err) }
		else { fmt.Printf("Subtraction Result: %d\n", diff) }
	}()

	time.Sleep(1 * time.Second) // Allow tasks to complete
}
```

**Outcome**: The `main` function continues execution, while the addition and subtraction tasks are handled concurrently by the `TaskManager`'s worker pool. Results are printed as tasks complete.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [task does not require immediate blocking execution by the caller] THEN [wrap `QueueTask` call in a new goroutine] ELSE [consider `RunTask`]"
  ],
  "verificationSteps": [
    "Check: `QueueTask` returns a result/error only after the task function completes → Expected: `sum` and `diff` values are correctly calculated.",
    "Check: Multiple `QueueTask` calls are processed concurrently → Expected: Output shows workers handling tasks in parallel, not strictly sequentially from the main goroutine."
  ],
  "quickPatterns": [
    "Pattern: Queue task in a goroutine\n```go\ngo func() {\n    result, err := manager.QueueTask(func(res *MyResource) (string, error) {\n        // Your asynchronous task logic\n        return \"Async done\", nil\n    })\n    if err != nil { /* handle error */ }\n    else { /* process result */ }\n}()\n```"
  ],
  "diagnosticPaths": [
    "Error `QueueTask call blocks indefinitely` -> Symptom: The goroutine calling `QueueTask` stops responding -> Check: Verify the task function itself is not deadlocking or hanging, and that worker count is sufficient -> Fix: Ensure task function completes, increase `WorkerCount` if queues are persistently full."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*