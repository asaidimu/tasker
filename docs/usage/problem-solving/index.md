# Problem Solving

### Troubleshooting Common Issues

*   **"Task manager is shutting down: cannot queue task"**: This error occurs if you attempt to `QueueTask`, `QueueTaskWithPriority`, `QueueTaskOnce`, `QueueTaskWithPriorityOnce`, or `RunTask` after `manager.Stop()` or `manager.Kill()` has been called, or if the `Ctx` provided during `NewTaskManager` initialization has been cancelled. Ensure tasks are only submitted while the manager is actively running.
*   **Resource Creation Failure**: If your `OnCreate` function returns an error, `tasker` will log it, and the associated worker will not start (or a `RunTask` call will fail). Ensure your resource creation logic is robust and handles transient issues gracefully.
*   **Workers Not Starting/Stopping as Expected**: 
    *   Verify your `WorkerCount` and `MaxWorkerCount` settings in `tasker.Config`.
    *   For dynamic scaling (bursting), ensure `BurstInterval` is configured and not set to 0.
    *   Check that your main `context.Context` (passed as `Ctx` in `Config`) and its `cancel` function are managed correctly, especially for graceful shutdown scenarios.
    *   Review your `CheckHealth` logic; an incorrect implementation might lead to workers constantly restarting (thrashing) or not restarting when they should.
*   **Deadlocks/Goroutine Leaks**: While `tasker` is designed to prevent these within its core logic, improper usage (e.g., blocking indefinitely within your `OnCreate`, `OnDestroy`, or task functions, or not using buffered channels for task results outside the library) can lead to such issues. Always ensure your custom functions (`OnCreate`, `OnDestroy`, `taskFunc`) do not block indefinitely.
*   **Task Panics**: It is generally recommended that your task functions (the `func(context.Context, R) (E, error)` you pass to `QueueTask` etc.) internally recover from panics and convert them into errors. If a panic occurs and is *not* recovered within the task function itself, it will crash the specific worker goroutine that was executing it. While `tasker` will detect the worker's exit and attempt to replace it, unhandled panics can lead to unexpected behavior, lost task results, and potential resource leaks if `OnDestroy` is not called due to an abrupt exit.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF error_is_shutdown_related THEN CHECK if_manager_was_stopped_or_killed ELSE IGNORE_SHUTDOWN_ERROR",
    "IF resource_creation_fails THEN INVESTIGATE_OnCreate_function ELSE CONTINUE",
    "IF workers_behave_unexpectedly THEN REVIEW_Config_settings_and_CheckHealth_logic ELSE CONTINUE",
    "IF application_hangs THEN CHECK_for_blocking_operations_in_custom_functions ELSE CONTINUE"
  ],
  "verificationSteps": [
    "Check: Is `manager.Stop()` or `manager.Kill()` called before `QueueTask`? -> Expected: True if shutdown error occurs",
    "Check: Does `OnCreate` return an error? -> Expected: Worker won't start",
    "Check: Are `WorkerCount`, `MaxWorkerCount`, `BurstInterval` configured correctly? -> Expected: Worker scaling as intended",
    "Check: Does `CheckHealth` correctly differentiate between task failure and worker unhealthiness? -> Expected: Correct worker replacement behavior",
    "Check: Are `OnCreate`, `OnDestroy`, and task functions non-blocking? -> Expected: No deadlocks or goroutine leaks"
  ],
  "quickPatterns": [
    "Pattern: Context_cancellation_in_task\n```go\nfunc myTask(ctx context.Context, res *MyResource) (string, error) {\n    select {\n    case <-ctx.Done():\n        return \"\", ctx.Err() // Task cancelled\n    case <-time.After(1 * time.Second):\n        return \"task done\", nil\n    }\n}\n```",
    "Pattern: Recover_from_panic_in_task\n```go\nfunc mySafeTask(ctx context.Context, res *MyResource) (result string, err error) {\n    defer func() {\n        if r := recover(); r != nil {\n            err = fmt.Errorf(\"task panic: %v\", r)\n        }\n    }()\n    // Unsafe operation\n    // panic(\"simulated panic\")\n    return \"success\", nil\n}\n```"
  ],
  "diagnosticPaths": [
    "Error: \"Task manager is shutting down\" -> Symptom: Attempting to queue tasks after shutdown -> Check: Application lifecycle management, ensure tasks are submitted when manager is active -> Fix: Only submit tasks when `manager.isRunning()` is true or before `Stop()`/`Kill()` is invoked.",
    "Error: Resource creation fails repeatedly -> Symptom: Workers fail to start or `RunTask` consistently errors out with resource creation failure -> Check: Review `OnCreate` function for external dependencies, network issues, or configuration problems -> Fix: Ensure external services are available, credentials are correct, and `OnCreate` logic is robust.",
    "Error: Worker thrashing (constant recreation) -> Symptom: High rate of `INFO: Creating [Resource]` and `INFO: Destroying [Resource]` logs, `CheckHealth` returns false frequently -> Check: `CheckHealth` implementation, identify if common task errors are being misidentified as worker health issues -> Fix: Refine `CheckHealth` to return `false` only for errors that truly indicate a faulty worker or resource, not just transient task failures."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 9:35:09 PM. Review and refine as needed.*