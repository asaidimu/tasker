---
outline: "deep"
lastUpdated: true
editLink: true
prev: true
next: true
title: "Problem Solving"
description: "Problem Solving documentation and guidance"
---
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

