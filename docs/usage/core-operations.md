# Core Operations

### QueueTask

`QueueTask` is used to add a task to the main queue for asynchronous processing. The task will be picked up by an available worker, executed, and its result (or error) will be returned to the caller via a blocking channel once completed. This is the primary method for submitting background jobs.

```go
result, err := manager.QueueTask(func(resource *MyResource) (MyResultType, error) {
    // Task logic here, using 'resource'
    return someResult, nil
})
if err != nil {
    // Handle queuing or task execution error
}
// result contains the returned value of type MyResultType
```

### QueueTaskWithPriority

`QueueTaskWithPriority` functions identically to `QueueTask`, but adds the task to a dedicated high-priority queue. Workers will always attempt to process tasks from the priority queue before moving to the main queue. This is crucial for time-sensitive operations.

```go
urgentResult, urgentErr := manager.QueueTaskWithPriority(func(resource *MyResource) (UrgentResultType, error) {
    // High-priority task logic
    return urgentData, nil
})
```

### RunTask

`RunTask` executes a task immediately and synchronously, bypassing both the main and priority queues. It attempts to acquire a resource from the internal pool. If no resource is immediately available, it temporarily creates a new one using your `OnCreate` function, which is then destroyed via `OnDestroy` after the task completes. This method is suitable for urgent tasks that should not be delayed by queueing, or for scenarios requiring a direct, blocking result.

```go
previewResult, previewErr := manager.RunTask(func(resource *MyResource) (PreviewResultType, error) {
    // Immediate, synchronous task logic
    return fastPreview, nil
})
```

### Stats

`manager.Stats()` provides real-time operational statistics about the `tasker` instance. This snapshot includes information on the number of active workers, tasks in queues, and available resources in the pool, useful for monitoring and debugging.

```go
stats := manager.Stats()
fmt.Printf("Active Workers: %d (Base: %d, Burst: %d)\n", stats.ActiveWorkers, stats.BaseWorkers, stats.BurstWorkers)
fmt.Printf("Queued Tasks: %d (Main: %d, Priority: %d)\n", stats.QueuedTasks + stats.PriorityTasks, stats.QueuedTasks, stats.PriorityTasks)
fmt.Printf("Available Resources in Pool: %d\n", stats.AvailableResources)
```

The `TaskStats` struct provides:

*   `BaseWorkers`: Configured permanent workers.
*   `ActiveWorkers`: Total workers currently running (base + burst).
*   `BurstWorkers`: Dynamically scaled workers.
*   `QueuedTasks`: Tasks in the standard queue.
*   `PriorityTasks`: Tasks in the high-priority queue.
*   `AvailableResources`: Resources ready in the `RunTask` pool.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF task_is_background_and_can_wait THEN USE QueueTask",
    "IF task_is_critical_and_time_sensitive THEN USE QueueTaskWithPriority",
    "IF task_needs_immediate_synchronous_execution THEN USE RunTask",
    "IF system_state_monitoring_is_needed THEN CALL Stats()"
  ],
  "verificationSteps": [
    "Check: `manager.QueueTask(...)` returns `MyResultType` -> Expected: Asynchronous task successfully processed.",
    "Check: `manager.RunTask(...)` returns `MyResultType` instantly -> Expected: Synchronous task processed immediately.",
    "Check: `manager.Stats().QueuedTasks` decreases after task submission -> Expected: Tasks are being dequeued by workers.",
    "Check: `manager.Stats().ActiveWorkers` matches configured `WorkerCount` (initially) -> Expected: Base workers are running."
  ],
  "quickPatterns": [
    "Pattern: Queue Task with Error Handling\n```go\nresp, err := manager.QueueTask(func(r *MyRes) (string, error) { return \"ok\", nil })\nif err != nil { /* error handling */ }\nfmt.Println(resp)\n```",
    "Pattern: Immediate Task with Resource Borrowing\n```go\nresp, err := manager.RunTask(func(r *MyRes) (string, error) { return \"fast_ok\", nil })\nif err != nil { /* error handling */ }\nfmt.Println(resp)\n```",
    "Pattern: Get Current Statistics\n```go\nstats := manager.Stats()\nfmt.Printf(\"Active Workers: %d, Queued: %d\\n\", stats.ActiveWorkers, stats.QueuedTasks)\n```"
  ],
  "diagnosticPaths": [
    "Error `TaskQueuingFailedError` -> Symptom: `QueueTask` or `QueueTaskWithPriority` returns an error `task manager is shutting down` -> Check: Application shutdown sequence or manager state -> Fix: Ensure tasks are queued before `manager.Stop()` is called.",
    "Error `ImmediateTaskResourceCreationError` -> Symptom: `RunTask` returns an error about temporary resource creation -> Check: `Config.OnCreate` function for issues or external resource availability -> Fix: Debug `OnCreate` for resource allocation failures."
  ]
}
```

---
*Generated using Gemini AI on 6/14/2025, 1:47:53 PM. Review and refine as needed.*