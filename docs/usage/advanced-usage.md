# Advanced Usage

### Context Management in Tasks

While `tasker` itself uses `context.Context` for managing worker lifecycles, your individual tasks can (and often should) also accept a `context.Context` parameter. This allows you to implement timeouts, cancellation, or propagate request-scoped values within your task logic. The `tasker` library does not automatically pass its internal worker context to your task function's signature. Instead, you'd typically pass a derived context explicitly when queuing the task.

```go
func myLongRunningTask(ctx context.Context, res *MyResource) (string, error) {
    select {
    case <-ctx.Done():
        return "", ctx.Err() // Context cancelled or timed out
    case <-time.After(100 * time.Millisecond):
        // Simulate work
    }
    return "Done", nil
}

// Example of queuing with a context and timeout
func queueWithTimeout(manager tasker.TaskManager[*MyResource, string]) {
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    result, err := manager.QueueTask(func(res *MyResource) (string, error) {
        return myLongRunningTask(ctx, res) // Pass the derived context to your actual task logic
    })

    if err != nil {
        fmt.Printf("Task with timeout failed: %v\n", err)
    } else {
        fmt.Printf("Task with timeout completed: %s\n", result)
    }
}
```

### Error Handling Strategies (beyond just `CheckHealth`)

`CheckHealth` is for identifying *unhealthy* worker/resource states. For general task failures that don't indicate a system-level problem, you should handle errors returned directly from `QueueTask`, `QueueTaskWithPriority`, or `RunTask`.

*   **Task-specific errors**: Your task function `func(R) (E, error)` can return any `error`. If `CheckHealth` returns `true` for this error, the worker is kept, and the error is simply passed back to the caller.
    ```go
    result, err := manager.QueueTask(func(r *MyResource) (string, error) {
        if someCondition { return "", errors.New("business_logic_error") }
        return "success", nil
    })
    if err != nil {
        if errors.Is(err, errors.New("business_logic_error")) {
            // Handle specific business error
        }
    }
    ```

*   **Context errors**: If the `tasker` manager's main context (`Config.Ctx`) is cancelled, subsequent `QueueTask`/`RunTask` calls will immediately return an error indicating shutdown. Your application should gracefully handle this.

### Performance Considerations

*   **`WorkerCount`**: Setting this too low might lead to tasks queueing up; too high might consume excessive resources. Tune based on profiling.
*   **`ResourcePoolSize`**: For `RunTask`, a larger pool reduces the overhead of `OnCreate`/`OnDestroy` for immediate tasks but consumes more idle resources. Tune based on the frequency and criticality of `RunTask` calls.
*   **`BurstInterval`**: A shorter interval makes bursting more reactive but increases the overhead of the burst manager checks. A longer interval makes it less reactive. Find a balance that suits your load patterns.
*   **`MaxRetries`**: Setting this too high for `unhealthy` errors can lead to tasks repeatedly failing and consuming resources if the underlying issue isn't resolved (e.g., a perpetually failing `OnCreate`).
*   **Heavy `OnCreate`/`OnDestroy`**: If your resource creation/destruction is very expensive, it can impact worker startup/shutdown times. Consider optimizing these functions.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF task_is_long_running OR task_has_external_dependencies THEN embed_context_in_task_logic",
    "IF task_returns_non_unhealthy_error THEN handle_error_returned_by_queue_task_caller",
    "IF application_experiences_bottlenecks THEN review_worker_count_and_resource_pool_size",
    "IF burst_manager_overhead_is_high THEN increase_burst_interval"
  ],
  "verificationSteps": [
    "Check: Task respects `context.Done()` and returns `ctx.Err()` -> Expected: Task stops early when context is cancelled.",
    "Check: `QueueTask` returns specific business logic errors, not wrapped in `tasker` errors (if `CheckHealth` passes) -> Expected: Custom error types are propagated.",
    "Check: Application performance under load is optimal -> Expected: Worker and pool sizes are correctly tuned."
  ],
  "quickPatterns": [
    "Pattern: Task with Context\n```go\nimport \"context\"\n// In your task definition:\nfunc myTask(ctx context.Context, res *MyResource) (string, error) {\n    select {\n    case <-ctx.Done(): return \"\", ctx.Err()\n    default: /* do work */ return \"OK\", nil\n    }\n}\n// When queuing:\nqueueCtx, queueCancel := context.WithTimeout(context.Background(), 10 * time.Second)\ndefer queueCancel()\nmanager.QueueTask(func(res *MyResource) (string, error) { return myTask(queueCtx, res) })\n```"
  ],
  "diagnosticPaths": [
    "Error `TaskContextCancelled` -> Symptom: Task returns context-related error (e.g., `context deadline exceeded`) -> Check: Timeout settings on `context.WithTimeout` or parent context cancellation -> Fix: Adjust timeout or check if parent context is prematurely cancelled.",
    "Error `PerformanceDegradation` -> Symptom: Tasks backlog despite high worker count, or high CPU/memory usage -> Check: `Config.WorkerCount`, `Config.ResourcePoolSize`, `Config.BurstInterval`. Profile `OnCreate`/`OnDestroy` functions -> Fix: Tune concurrency settings, optimize resource lifecycle methods."
  ]
}
```

---
*Generated using Gemini AI on 6/14/2025, 1:47:53 PM. Review and refine as needed.*