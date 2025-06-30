# Advanced Usage

### Performance Tuning and Monitoring

Tasker provides comprehensive statistics and metrics to help you monitor its operational state and performance, enabling effective tuning and troubleshooting.

#### Real-time Statistics with `Stats()`
`manager.Stats()` returns a `TaskStats` struct, providing an immediate snapshot of the `TaskManager`'s current operational state:

*   `BaseWorkers`: Number of permanently active workers.
*   `ActiveWorkers`: Total currently active workers (base + burst).
*   `BurstWorkers`: Number of dynamically scaled-up workers.
*   `QueuedTasks`: Number of tasks in the main queue.
*   `PriorityTasks`: Number of tasks in the priority queue.
*   `AvailableResources`: Number of resources in the `RunTask` resource pool.

These statistics are useful for quick checks on worker counts and queue backlogs.

#### Comprehensive Performance Metrics with `Metrics()`
`manager.Metrics()` returns a `TaskMetrics` struct, offering deeper insights into the system's performance, throughput, and reliability over time. This data is collected by an internal `MetricsCollector` (or your custom one).

Key `TaskMetrics` fields include:

*   **Latency & Duration**: `AverageExecutionTime`, `MinExecutionTime`, `MaxExecutionTime`, `P95ExecutionTime`, `P99ExecutionTime`, `AverageWaitTime`.
*   **Throughput & Volume**: `TaskArrivalRate`, `TaskCompletionRate`, `TotalTasksCompleted`.
*   **Reliability & Error**: `TotalTasksFailed`, `TotalTasksRetried`, `SuccessRate`, `FailureRate`.

These metrics are essential for identifying performance bottlenecks, understanding system load, and diagnosing issues related to task failures or retries.

**Example: Monitoring Stats and Metrics**
```go
// examples/advanced/main.go (Excerpt)
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
	defer cancel()

	config := tasker.Config[*HeavyComputeResource]{
		OnCreate: createComputeResource,
		OnDestroy: destroyComputeResource,
		WorkerCount: 2,
		Ctx: ctx,
		BurstInterval: 200 * time.Millisecond,
	}

	manager, err := tasker.NewTaskManager[*HeavyComputeResource, string](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }

	// Queue some tasks
	var tasksSubmitted atomic.Int32
	for i := 0; i < 5; i++ {
		tasksSubmitted.Add(1)
		go func(taskID int) {
			_, _ = manager.QueueTask(func(res *HeavyComputeResource) (string, error) {
				time.Sleep(200 * time.Millisecond)
				return fmt.Sprintf("Task %d completed", taskID), nil
			})
		}(i)
	}

	// Monitor stats and metrics periodically
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

	time.Sleep(1 * time.Second) // Allow some monitoring
	close(done) // Stop monitoring

	_ = manager.Stop()
}
```

**Outcome**: The console output will show real-time updates of worker counts, queued tasks, and calculated arrival/completion rates, allowing you to observe the `TaskManager`'s performance characteristics.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [immediate operational status (worker count, queue size) is needed] THEN [use `manager.Stats()`] ELSE [for deeper performance analysis, use `manager.Metrics()`]",
    "IF [long-term trends or detailed latency breakdowns (percentiles) are required] THEN [regularly poll `manager.Metrics()` and push to external monitoring system] ELSE [use `manager.Stats()` for quick checks]"
  ],
  "verificationSteps": [
    "Check: `Stats().ActiveWorkers` matches expected base + burst workers → Expected: Worker count is accurate.",
    "Check: `Stats().QueuedTasks` increases upon task submission and decreases upon processing → Expected: Queue size reflects task backlog.",
    "Check: `Metrics().TaskArrivalRate` and `Metrics().TaskCompletionRate` reflect actual task flow → Expected: Rates align with submitted and completed tasks."
  ],
  "quickPatterns": [
    "Pattern: Polling stats and metrics\n```go\nticker := time.NewTicker(5 * time.Second)\ndefer ticker.Stop()\nfor range ticker.C {\n    stats := manager.Stats()\n    metrics := manager.Metrics()\n    fmt.Printf(\"Active: %d, Queued: %d, Avg Exec: %s\\n\", stats.ActiveWorkers, stats.QueuedTasks, metrics.AverageExecutionTime)\n}\n```"
  ],
  "diagnosticPaths": [
    "Error `High QueuedTasks, low ActiveWorkers` -> Symptom: Tasks are backing up, but not enough workers are active -> Check: Verify `WorkerCount` and `MaxWorkerCount` in `Config`; check if dynamic scaling is enabled (`BurstInterval` > 0) -> Fix: Increase `WorkerCount` or `MaxWorkerCount`, adjust `BurstInterval`.",
    "Error `AverageWaitTime is consistently high` -> Symptom: Tasks spend too much time in queues -> Check: `ActiveWorkers` might be too low relative to `TaskArrivalRate`; tasks might be very long-running -> Fix: Increase worker capacity, optimize task execution time."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*