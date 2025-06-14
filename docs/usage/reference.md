# Reference

### `TaskManager[R any, E any]` Interface

This is the primary interface for interacting with the `tasker` library. `R` is the type of resource used by tasks, and `E` is the expected result type of the tasks.

```go
type TaskManager[R any, E any] interface {
	QueueTask(task func(R) (E, error)) (E, error)
	RunTask(task func(R) (E, error)) (E, error)
	QueueTaskWithPriority(task func(R) (E, error)) (E, error)
	Stop() error
	Stats() TaskStats
}
```

### `Config[R any]` Struct

`Config` holds all configuration parameters for creating a new `Runner` instance. It controls worker behavior, resource management, and scaling policies. `R` is the resource type (e.g., `*sql.DB`, `*http.Client`).

```go
type Config[R any] struct {
	OnCreate        func() (R, error)      // Required: Function to create a new resource.
	OnDestroy       func(R) error          // Required: Function to destroy/cleanup a resource.
	WorkerCount     int                    // Required: Initial and minimum number of base workers. Must be > 0.
	Ctx             context.Context        // Required: Parent context for the Runner; cancellation initiates graceful shutdown.
	CheckHealth     func(error) bool       // Optional: Custom health check for task errors. Default: always healthy (func(error) bool { return true }).
	BurstTaskThreshold  int                    // Optional: Queue size to trigger bursting. 0 to disable.
	BurstWorkerCount      int                    // Optional: Number of burst workers to add at once. Defaults to 2 if <= 0.
	MaxWorkerCount  int                    // Optional: Maximum total workers allowed. Defaults to WorkerCount + BurstWorkerCount if <= 0.
	BurstInterval   time.Duration          // Optional: Frequency for burst checks. Default: 100ms if <= 0.
	MaxRetries      int                    // Optional: Max retries for unhealthy tasks. Default: 3 if <= 0.
	ResourcePoolSize int                   // Optional: Size of the pool for `RunTask`. Default: `WorkerCount` if <= 0.
}
```

### `TaskStats` Struct

`TaskStats` provides a snapshot of the task manager's current state and performance metrics.

```go
type TaskStats struct {
	BaseWorkers        int32 // Number of permanently active workers.
	ActiveWorkers      int32 // Total number of currently active workers (base + burst).
	BurstWorkers       int32 // Number of dynamically scaled-up workers.
	QueuedTasks        int32 // Number of tasks currently in the main queue.
	PriorityTasks      int32 // Number of tasks currently in the priority queue.
	AvailableResources int32 // Number of resources currently available in the resource pool.
}
```

### Glossary

*   **Resource (`R`)**: An instance of a dependency required by tasks, managed by `OnCreate` and `OnDestroy`.
*   **Task Result (`E`)**: The type of value returned by a task function upon successful completion.
*   **Worker**: A goroutine responsible for picking up tasks from queues, executing them with a resource.
*   **Base Worker**: A permanently running worker, part of the minimum `WorkerCount`.
*   **Burst Worker**: A dynamically scaled worker, created during high load and scaled down when load subsides.
*   **Main Queue**: The standard queue for `QueueTask` operations.
*   **Priority Queue**: A higher-priority queue for `QueueTaskWithPriority` operations, processed before the main queue.
*   **Resource Pool**: A collection of pre-allocated resources used by `RunTask` for immediate execution.
*   **Unhealthy Error**: An error returned by a task that your `CheckHealth` function identifies as indicating a problematic worker or resource, triggering replacement and task retry.
*   **Graceful Shutdown**: The process of safely stopping the `tasker` manager, allowing in-flight tasks to complete and all resources to be released.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [],
  "verificationSteps": [],
  "quickPatterns": [],
  "diagnosticPaths": []
}
```

---
*Generated using Gemini AI on 6/14/2025, 1:47:53 PM. Review and refine as needed.*