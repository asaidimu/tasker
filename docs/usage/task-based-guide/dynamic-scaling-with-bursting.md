# Task-Based Guide

### Dynamic Scaling (Bursting)

**Goal**: Automatically adjust the number of active workers based on real-time task load to optimize resource utilization and maintain throughput.

**Approach**: Configure `Config.BurstInterval`. Tasker's internal burst manager periodically monitors the `TaskArrivalRate` and `TaskCompletionRate` (from `Metrics()`).

*   If `TaskArrivalRate > TaskCompletionRate`: The system is falling behind, so the burst manager dynamically starts additional "burst" workers (up to `MaxWorkerCount`) to meet demand.
*   If `TaskCompletionRate > TaskArrivalRate`: The system is over-provisioned, and the burst manager signals idle burst workers to gracefully shut down, reducing resource consumption.

**Example: Handling Burst Loads for Heavy Computation**
This advanced example demonstrates a scenario where a large number of computational tasks are queued rapidly, forcing Tasker to scale up workers, and then scaling down once the load subsides.

```go
// examples/advanced/main.go (Simplified)
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/asaidimu/tasker"
)

type HeavyComputeResource struct { ID int }
func createComputeResource() (*HeavyComputeResource, error) { /* ... */ return &HeavyComputeResource{ID: rand.Intn(1000)}, nil }
func destroyComputeResource(r *HeavyComputeResource) error { /* ... */ return nil}
func checkComputeHealth(err error) bool { return true }

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := tasker.Config[*HeavyComputeResource]{
		OnCreate: createComputeResource,
		OnDestroy: destroyComputeResource,
		WorkerCount: 2,                     // Start with 2 base workers
		Ctx: ctx,
		BurstInterval: 200 * time.Millisecond, // Check every 200ms
		MaxRetries: 0,
	}

	manager, err := tasker.NewTaskManager[*HeavyComputeResource, string](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }

	var tasksSubmitted atomic.Int32
	for i := 0; i < 20; i++ { // Aggressively queue 20 tasks
		tasksSubmitted.Add(1)
		go func(taskID int) {
			_, _ = manager.QueueTask(func(res *HeavyComputeResource) (string, error) {
				time.Sleep(time.Duration(100 + rand.Intn(300)) * time.Millisecond)
				return fmt.Sprintf("Task %d completed", taskID), nil
			})
		}(i)
	}

	fmt.Println("Monitoring stats (watch for BurstWorkers)...")
	ticker := time.NewTicker(250 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				stats := manager.Stats()
				metrics := manager.Metrics()
				fmt.Printf("Stats: Active=%d (Base=%d, Burst=%d), Queued=%d, ArrivalRate=%.2f, CompletionRate=%.2f\n",
					stats.ActiveWorkers, stats.BaseWorkers, stats.BurstWorkers,
					stats.QueuedTasks + stats.PriorityTasks,
					metrics.TaskArrivalRate, metrics.TaskCompletionRate)
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	time.Sleep(5 * time.Second) // Observe scaling up and down
	close(done) // Stop monitoring

	fmt.Println("Initiating graceful shutdown...")
	_ = manager.Stop()
}
```

**Outcome**: The `Stats()` output will show `BurstWorkers` dynamically increasing when the `QueuedTasks` grow and `TaskArrivalRate` exceeds `TaskCompletionRate`, then decreasing once tasks are processed and throughput catches up with or exceeds demand. This demonstrates efficient resource scaling.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [workload is highly variable with unpredictable spikes] THEN [enable dynamic scaling by setting `BurstInterval` > 0 and `MaxWorkerCount` appropriately] ELSE [rely on fixed `WorkerCount`]",
    "IF [resource costs are high per worker] THEN [configure dynamic scaling to optimize worker count based on demand] ELSE [maintain a static worker pool]"
  ],
  "verificationSteps": [
    "Check: Under high load, `Stats().BurstWorkers` increases from 0 → Expected: New worker creation logs appear, `ActiveWorkers` count grows.",
    "Check: After load subsides, `Stats().BurstWorkers` decreases to 0 → Expected: Burst worker destruction logs appear, `ActiveWorkers` count returns to `BaseWorkers`.",
    "Check: `TaskArrivalRate` vs `TaskCompletionRate` correlation with `BurstWorkers` count changes → Expected: System scales up when arrival rate > completion rate, scales down when arrival rate < completion rate."
  ],
  "quickPatterns": [
    "Pattern: Enable dynamic scaling\n```go\nconfig := tasker.Config[*MyResource]{\n    // ... other config ...\n    WorkerCount:   2,  // Base workers\n    MaxWorkerCount: 10, // Max total workers\n    BurstInterval: 100 * time.Millisecond, // Check every 100ms\n}\nmanager, _ := tasker.NewTaskManager[*MyResource, any](config)\n```"
  ],
  "diagnosticPaths": [
    "Error `System not scaling up during high load` -> Symptom: `QueuedTasks` remains high, `ActiveWorkers` does not increase -> Check: Verify `BurstInterval` is set and `MaxWorkerCount` is greater than `WorkerCount`; ensure `MetricsCollector` is active -> Fix: Adjust `BurstInterval` (e.g., lower it) or increase `MaxWorkerCount`.",
    "Error `System thrashing (constantly scaling up/down)` -> Symptom: `BurstWorkers` rapidly fluctuates -> Check: `BurstInterval` might be too low; rapid changes in `TaskArrivalRate` or `TaskCompletionRate` -> Fix: Increase `BurstInterval` to smooth out scaling decisions; analyze workload patterns."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*