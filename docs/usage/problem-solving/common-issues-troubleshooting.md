# Problem Solving

### Common Issues & Troubleshooting

This section addresses frequently encountered problems and provides guidance for diagnosing and resolving them.

#### "Task manager is shutting down: cannot queue task"
*   **Symptom**: You attempt to `QueueTask`, `RunTask`, or similar methods, and they immediately return an error indicating the manager is shutting down.
*   **Reason**: This error occurs if you try to submit tasks after `manager.Stop()` or `manager.Kill()` has been called, or if the `context.Context` provided in `Config.Ctx` has been cancelled externally.
*   **Diagnosis**: Check the lifecycle of your `TaskManager` instance. Ensure that task submission logic only runs when the manager is expected to be active.
*   **Resolution**: Only submit tasks while the manager is actively running. Implement checks (`if !manager.IsRunning()`) if necessary (though `IsRunning` is internal, the error itself indicates the state). For graceful shutdowns, ensure new tasks are not submitted after `Stop()` is initiated.

#### Resource Creation Failure
*   **Symptom**: `TaskManager` initialization fails, or workers fail to start with errors originating from your `OnCreate` function (e.g., "Failed to create resource for pool: connection refused").
*   **Reason**: Your `OnCreate` function encountered an error while trying to provision a resource.
*   **Diagnosis**: Inspect the error message from `OnCreate`. This typically points to issues with external dependencies (database, network, file system) or misconfigurations.
*   **Resolution**: Debug your `OnCreate` implementation. Ensure all prerequisites for resource creation are met (e.g., correct connection strings, network access, necessary credentials). Make `OnCreate` robust to handle transient setup issues if possible.

#### Workers Not Starting/Stopping as Expected
*   **Symptom**: The number of `ActiveWorkers` in `Stats()` does not match your `WorkerCount` configuration, or workers persist after shutdown calls.
*   **Reason**: 
    *   **Not Starting**: `OnCreate` might be failing, or `WorkerCount` is zero/negative.
    *   **Not Stopping**: Long-running tasks within workers might not be responding to context cancellation during `Stop()`, or `OnDestroy` is blocking.
*   **Diagnosis**: 
    *   Verify `Config.WorkerCount` is positive and `Config.MaxWorkerCount` (if used for bursting) allows for more workers.
    *   Check `OnCreate` for errors.
    *   For stopping issues, review your task functions and `OnDestroy` to ensure they are non-blocking and respect the context cancellation signal where appropriate.
*   **Resolution**: 
    *   Correct `WorkerCount` and debug `OnCreate`.
    *   Instrument your long-running tasks to periodically check `ctx.Done()` (if passing context) and return if cancelled. Ensure `OnDestroy` is quick and non-blocking.

#### Deadlocks/Goroutine Leaks
*   **Symptom**: Your application hangs, CPU usage is high without progress, or `pprof` shows numerous goroutines stuck.
*   **Reason**: While Tasker is designed to prevent these internally, improper usage of `OnCreate`, `OnDestroy`, or your task functions (`func(R) (E, error)`) can lead to deadlocks or goroutine leaks (e.g., blocking indefinitely, not closing channels, or unhandled panics).
*   **Diagnosis**: Use Go's built-in `pprof` tool to inspect goroutine stacks and identify where the deadlock or leak is occurring. Pay close attention to your custom `OnCreate`, `OnDestroy`, and task logic.
*   **Resolution**: Ensure all custom functions are non-blocking and handle their own concurrency and resource management. If a task can panic, wrap its content in a `defer recover()` block to convert panics into errors that Tasker can handle gracefully.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [],
  "verificationSteps": [
    "Check: `manager.Stop()` returns immediately without waiting for tasks to complete -> Expected: The `Kill()` method was invoked instead, or tasks are not blocking graceful shutdown."
  ],
  "quickPatterns": [],
  "diagnosticPaths": [
    "Error `task manager is shutting down` -> Symptom: Task submission fails instantly -> Check: Is `manager.Stop()` or `manager.Kill()` called before task submission? -> Fix: Reorder calls or handle shutdown state appropriately for new task submissions.",
    "Error `Resource creation failed for worker` -> Symptom: Workers are not starting -> Check: `OnCreate` function for external dependency issues (network, credentials, permissions) -> Fix: Resolve external dependency issues or improve `OnCreate` error handling.",
    "Error `Application hangs on manager.Stop()` -> Symptom: Graceful shutdown is not completing -> Check: Long-running tasks or `OnDestroy` functions are blocking -> Fix: Modify tasks to respect context cancellation; ensure `OnDestroy` is non-blocking.",
    "Error `High goroutine count / Application unresponsive` -> Symptom: Resource leaks or deadlocks -> Check: Use `go tool pprof` to analyze goroutine stacks for blocking operations in `OnCreate`, `OnDestroy`, or task functions -> Fix: Identify and unblock or optimize problematic code sections."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*