# Problem Solving

### Error Reference

This section details common error conditions you might encounter when using Tasker, their triggers, and recommended resolutions.

#### 1. `errors.New("task manager is shutting down")`
*   **Type**: Go `error` (returned directly)
*   **Symptoms**: Calls to `QueueTask`, `QueueTaskWithPriority`, `QueueTaskOnce`, `QueueTaskWithPriorityOnce`, or `RunTask` immediately return this error.
*   **Properties**: No specific error properties.
*   **Scenarios**:
    *   **Trigger**: Attempting to queue a task after `manager.Stop()` has been called.
    *   **Example**: 
        ```go
        manager.Stop()
        _, err := manager.QueueTask(func(r *MyResource) (int, error) { return 1, nil })
        // err will be "task manager is shutting down"
        ```
    *   **Reason**: The `TaskManager` transitions to a `stopping` or `killed` state, preventing new tasks from being accepted.
*   **Diagnosis**: Check the call sequence in your application. Is task submission happening after the manager is explicitly told to shut down or after its root context is cancelled?
*   **Resolution**: Ensure tasks are only submitted when the `TaskManager` is in a running state. Refactor task submission logic to respect the application's shutdown signals.
*   **Prevention**: Always check application lifecycle status before attempting to queue tasks, or structure your application so that task producers cease before `TaskManager` shutdown is initiated.
*   **Handling Patterns**: 
    ```go
    result, err := manager.QueueTask(myTaskFunc)
    if err != nil {
        if errors.Is(err, errors.New("task manager is shutting down")) {
            log.Println("Cannot queue task: manager is shutting down.")
            // Gracefully exit or handle the un-queued task
        } else {
            log.Printf("Task failed: %v", err)
        }
    }
    ```
*   **Propagation Behavior**: This error is directly returned to the caller of the task submission method.

#### 2. `errors.New("worker count must be positive")`
*   **Type**: Go `error` (returned directly)
*   **Symptoms**: `tasker.NewTaskManager` returns this error during initialization.
*   **Properties**: No specific error properties.
*   **Scenarios**:
    *   **Trigger**: Providing `Config.WorkerCount` as 0 or a negative number.
    *   **Example**: 
        ```go
        config := tasker.Config{ WorkerCount: 0, /* ... */ }
        _, err := tasker.NewTaskManager(config)
        // err will be "worker count must be positive"
        ```
    *   **Reason**: The `TaskManager` requires at least one worker to operate.
*   **Diagnosis**: Inspect the `WorkerCount` field in your `tasker.Config` struct.
*   **Resolution**: Set `Config.WorkerCount` to a positive integer (e.g., 1 or more).
*   **Prevention**: Validate configuration inputs before passing them to `NewTaskManager`.
*   **Handling Patterns**: Typically handled at application startup; a fatal log is common.
    ```go
    manager, err := tasker.NewTaskManager(config)
    if err != nil {
        log.Fatalf("Failed to initialize TaskManager: %v", err)
    }
    ```
*   **Propagation Behavior**: Returned immediately by `NewTaskManager`.

#### 3. `errors.New("onCreate function is required")`
*   **Type**: Go `error` (returned directly)
*   **Symptoms**: `tasker.NewTaskManager` returns this error during initialization.
*   **Properties**: No specific error properties.
*   **Scenarios**:
    *   **Trigger**: Providing `Config.OnCreate` as `nil`.
    *   **Example**: 
        ```go
        config := tasker.Config{ OnCreate: nil, /* ... */ }
        _, err := tasker.NewTaskManager(config)
        // err will be "onCreate function is required"
        ```
    *   **Reason**: The `TaskManager` needs a function to create new instances of your resource `R`.
*   **Diagnosis**: Inspect the `OnCreate` field in your `tasker.Config` struct.
*   **Resolution**: Provide a non-nil function for `Config.OnCreate` that correctly initializes your resource.
*   **Prevention**: Validate configuration inputs or ensure `OnCreate` is always set.
*   **Handling Patterns**: Same as "worker count must be positive".
*   **Propagation Behavior**: Returned immediately by `NewTaskManager`.

#### 4. `errors.New("onDestroy function is required")`
*   **Type**: Go `error` (returned directly)
*   **Symptoms**: `tasker.NewTaskManager` returns this error during initialization.
*   **Properties**: No specific error properties.
*   **Scenarios**:
    *   **Trigger**: Providing `Config.OnDestroy` as `nil`.
    *   **Example**: 
        ```go
        config := tasker.Config{ OnDestroy: nil, /* ... */ }
        _, err := tasker.NewTaskManager(config)
        // err will be "onDestroy function is required"
        ```
    *   **Reason**: The `TaskManager` needs a function to properly clean up resources when workers shut down.
*   **Diagnosis**: Inspect the `OnDestroy` field in your `tasker.Config` struct.
*   **Resolution**: Provide a non-nil function for `Config.OnDestroy` that correctly cleans up your resource.
*   **Prevention**: Validate configuration inputs or ensure `OnDestroy` is always set.
*   **Handling Patterns**: Same as "worker count must be positive".
*   **Propagation Behavior**: Returned immediately by `NewTaskManager`.

#### 5. `fmt.Errorf("failed to create temporary resource: %w", originalErr)`
*   **Type**: Go `error` (wrapped error)
*   **Symptoms**: `RunTask` returns this error.
*   **Properties**: Wraps the original error from your `Config.OnCreate` function.
*   **Scenarios**:
    *   **Trigger**: `RunTask` attempts to create a temporary resource (because the `resourcePool` is empty or `ResourcePoolSize` is 0) and your `Config.OnCreate` function returns an error.
    *   **Example**: 
        ```go
        // If createImageProcessor returns an error
        _, err := manager.RunTask(func(p *ImageProcessor) (string, error) { /* ... */ })
        // err will contain "failed to create temporary resource: [original error from onCreate]"
        ```
    *   **Reason**: The temporary resource needed for immediate execution could not be provisioned.
*   **Diagnosis**: Inspect the wrapped error (`errors.Unwrap(err)`) to find the root cause from your `OnCreate` function. This is often a transient issue with the external dependency.
*   **Resolution**: Address the underlying cause in your `OnCreate` function. Ensure the resource it tries to create is available and accessible. Consider making `OnCreate` more resilient to transient failures or increasing `ResourcePoolSize` if resource creation is slow.
*   **Prevention**: Implement robust error handling and retries within your `OnCreate` function if it interacts with unreliable external systems.
*   **Handling Patterns**: 
    ```go
    _, err := manager.RunTask(myImmediateTaskFunc)
    if err != nil {
        if errors.Is(err, fmt.Errorf("failed to create temporary resource")) {
            log.Printf("Could not execute immediate task due to resource provisioning error: %v", errors.Unwrap(err))
        } else {
            log.Printf("Immediate task failed: %v", err)
        }
    }
    ```
*   **Propagation Behavior**: Returned directly to the caller of `RunTask`.

#### 6. `fmt.Errorf("max retries exceeded: %w", originalErr)` (Task-specific error)
*   **Type**: Go `error` (wrapped error)
*   **Symptoms**: A `QueueTask` or `QueueTaskWithPriority` call eventually returns this error after internal retries.
*   **Properties**: Wraps the original error that triggered the retry mechanism and failed the `Config.CheckHealth` function.
*   **Scenarios**:
    *   **Trigger**: A task repeatedly fails, and for each failure, your `Config.CheckHealth` function returns `false` (indicating an unhealthy worker/resource). After `Config.MaxRetries` attempts, the task is finally marked as failed.
    *   **Example**: 
        ```go
        // Given CheckHealth marks "processor_crash" as unhealthy
        _, err := manager.QueueTask(func(proc *ImageProcessor) (string, error) {
            return "", errors.New("processor_crash") // This task will retry MaxRetries times
        })
        // After max retries, err will be "max retries exceeded: processor_crash"
        ```
    *   **Reason**: The task continuously leads to an unhealthy worker state, exhausting all configured retry attempts.
*   **Diagnosis**: The root cause lies in the original error (`errors.Unwrap(err)`) that consistently triggers the unhealthy state. Debug the task logic and the resource it uses to understand why it's failing.
*   **Resolution**: Fix the underlying problem causing the task to fail unhealthily. This might involve updating external services, changing task input, or improving resource robustness. Adjust `MaxRetries` if the issue is truly transient and requires more attempts.
*   **Prevention**: Improve task idempotency, make `OnCreate` more robust to unhealthy states, or refine `CheckHealth` if some errors are not truly indicative of a worker/resource health issue.
*   **Handling Patterns**: 
    ```go
    _, err := manager.QueueTask(myProblematicTaskFunc)
    if err != nil {
        if errors.Is(err, errors.New("max retries exceeded")) {
            log.Printf("Task failed permanently after retries, original error: %v", errors.Unwrap(err))
            // Log this as a critical failure, potentially alert
        } else {
            log.Printf("Task failed with other error: %v", err)
        }
    }
    ```
*   **Propagation Behavior**: Returned to the caller of `QueueTask` or `QueueTaskWithPriority`.

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
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*