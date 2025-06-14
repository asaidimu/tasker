# Problem Solving

### Troubleshooting Common Issues

*   **Tasks not running / Manager stuck**:
    *   **Verify `manager.Stop()` is called**: If `Stop()` is never called, goroutines (workers, burst manager) might remain active, holding resources and preventing application exit. Always use `defer manager.Stop()` in your `main` function or the controlling goroutine.
    *   **Check for panics**: Panics in your `OnCreate`, `OnDestroy`, or task functions can crash workers or prevent resources from being returned to the pool. Ensure these functions are robust and handle potential errors internally or convert them to `error` returns.
    *   **Premature context cancellation**: Ensure the `context.Context` passed to `tasker.Config` is not cancelled too early. If the root context is cancelled, the manager will shut down.

*   **Resource creation errors (`NewRunner` fails)**:
    *   If your `OnCreate` function returns an error during `NewRunner` initialization or `RunTask`'s temporary resource creation, `tasker` will fail to start or execute the immediate task. Debug your `OnCreate` logic and ensure external dependencies (e.g., database connectivity) are available.

*   **High memory/CPU usage**:
    *   **Aggressive Bursting**: If `Config.BurstTaskThreshold` is too low or `Config.BurstWorkerCount` is too high, `tasker` might spin up too many workers, consuming excessive resources. Adjust these parameters.
    *   **Frequent Burst Checks**: A very low `Config.BurstInterval` can cause the burst manager to check and adjust too frequently, adding minor overhead. Increase it if profiling indicates this is an issue.
    *   **Resource Leaks**: Ensure your `OnDestroy` function properly cleans up resources. If resources are not released, memory usage will climb.

*   **Tasks timing out / Slow processing**:
    *   **Worker Starvation**: If `Config.WorkerCount` is too low for your typical load, tasks will queue up. Increase `WorkerCount` or enable bursting.
    *   **Slow Tasks**: Individual task functions might be too slow. Optimize their logic, or ensure they respect `context.Context` deadlines if provided, allowing them to fail fast rather than block.
    *   **Network Latency/External Dependencies**: If tasks rely on external services, network issues or slow responses from those services can cause slowdowns. Implement timeouts in your task logic.

### Error Reference

Below are common conceptual errors you might encounter or produce when using `tasker`:

*   **`InvalidConfigurationError` (ID: `error:InvalidConfigurationError`)**
*   **`TaskQueuingFailedError` (ID: `error:TaskQueuingFailedError`)**
*   **`ResourceCreationError` (ID: `error:ResourceCreationError`)**
*   **`MaxRetriesExceededError` (ID: `error:MaxRetriesExceededError`)**
*   **`UnhealthyWorkerError` (ID: `error:UnhealthyWorkerError`)**
*   **`TaskProcessingError` (ID: `error:TaskProcessingError`)**

For detailed information on each of these, refer to the `agentReferenceData.errors` section.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF application_exits_without_cleaning_up THEN verify_manager_stop_call",
    "IF resource_creation_fails THEN debug_on_create_function",
    "IF high_resource_usage_occurs THEN review_burst_settings_and_resource_cleanup",
    "IF tasks_are_slow OR timeout THEN analyze_worker_count_and_task_duration"
  ],
  "verificationSteps": [
    "Check: `manager.Stop()` is called on application shutdown -> Expected: All goroutines exit and resources are destroyed.",
    "Check: `NewRunner` initialization is successful -> Expected: No errors from `Config.OnCreate`.",
    "Check: System resources (CPU, Memory) are within acceptable limits during operation -> Expected: Bursting parameters are tuned correctly and `OnDestroy` works.",
    "Check: Task completion times meet requirements -> Expected: Sufficient workers, optimized task logic."
  ],
  "quickPatterns": [],
  "diagnosticPaths": [
    "Error `UncaughtPanic` -> Symptom: Application crashes or worker goroutine exits unexpectedly -> Check: `OnCreate`, `OnDestroy`, and task functions for panic scenarios -> Fix: Add `recover()` or ensure no panics occur in these critical paths.",
    "Error `ResourceLeak` -> Symptom: Memory usage steadily increases over time -> Check: `OnDestroy` function for proper resource cleanup. Also check if resources are returned to `resourcePool` if borrowed by `RunTask` -> Fix: Implement thorough resource cleanup in `OnDestroy`.",
    "Error `QueueBlocking` -> Symptom: `QueueTask` calls block indefinitely or return errors about full queues -> Check: If channels `mainQueue`/`priorityQueue` are full; if workers are processing tasks -> Fix: Increase channel buffer size (though `tasker` default is usually sufficient), increase `WorkerCount` or enable bursting to handle load."
  ]
}
```

---
*Generated using Gemini AI on 6/14/2025, 1:47:53 PM. Review and refine as needed.*