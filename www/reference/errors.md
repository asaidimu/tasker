---
outline: "deep"
lastUpdated: true
editLink: true
sidebar: true
title: "Error Reference"
description: "Complete error reference with scenarios, diagnosis, and resolution"
---
# Error Reference

## `task manager is shutting down`

- **Type**: 

Go `error` (specific string match)
- **Symptoms**: 

Calling `QueueTask`, `RunTask`, or other submission methods returns an error with the message "task manager is shutting down" or "task manager is shutting down (context done)".
- **Properties**: 

None (standard `error` interface, typically `errors.New` or wrapped by `fmt.Errorf`)

### Scenarios

###### Attempting to submit a new task after `manager.Stop()` or `manager.Kill()` has been called.

**Example**:



```go
manager.Stop() // Initiates shutdown
_, err := manager.QueueTask(...) // This will likely return 'task manager is shutting down'
```


**Reason**:

 The TaskManager has transitioned to a non-running state (stopping or killed) and no longer accepts new tasks.

---

###### Attempting to submit a new task after the context provided to `NewTaskManager` (`config.Ctx`) has been cancelled externally.

**Example**:



```go
ctx, cancel := context.WithCancel(context.Background())
config := tasker.Config{Ctx: ctx, ...}
manager := tasker.NewTaskManager(config)
cancel() // Cancels the main context
_, err := manager.QueueTask(...) // This will likely return 'task manager is shutting down (context done)'
```


**Reason**:

 The primary control context for the manager has been cancelled, signaling it to begin shutdown.

---

- **Diagnosis**: 

Check application lifecycle management. Verify if `Stop()` or `Kill()` are being called prematurely. Ensure task submissions are coordinated with the manager's active state.
- **Resolution**: 

Only submit tasks when the `TaskManager` is in a running state. Add checks around submission calls (e.g., `if manager.isRunning() { ... }`). Ensure graceful shutdown routines are triggered only when no more new tasks are expected.
- **Prevention**: 

Implement robust application lifecycle hooks. Use `defer manager.Stop()` in `main` or main goroutine. Integrate `manager.Stop()` or `manager.Kill()` with OS signals (e.g., `SIGINT`, `SIGTERM`).
- **Handling Patterns**: 



```
`if errors.Is(err, errors.New("task manager is shutting down"))` (or contains substring) then discard/log task rather than retrying.
```

- **Propagation Behavior**: 

This error is returned directly to the caller of submission methods (`QueueTask`, `RunTask`, etc.) and propagated through result/error channels or callbacks for async methods.

---

## `processor_crash`

- **Type**: 

Go `error` (example custom error string)
- **Symptoms**: 

A task function explicitly returns an error indicating an underlying resource or worker failure (e.g., a connection drop, service outage). Logs may show `WARN: Detected unhealthy error: processor_crash. Worker will be replaced.`
- **Properties**: 

None (typically a simple `errors.New` or custom error type from the application's domain)

### Scenarios

###### A task encounters a critical, unrecoverable error during its execution that indicates the worker's associated resource is faulty.

**Example**:



```go
func myProblematicTask(ctx context.Context, proc *ImageProcessor) (string, error) {
    if rand.Intn(2) == 0 { // 50% chance to simulate a crash
        return "", errors.New("processor_crash") // This triggers CheckHealth to return false
    }
    return "processed", nil
}
// In Config: CheckHealth: func(err error) bool { return err.Error() != "processor_crash" }
```


**Reason**:

 The `CheckHealth` function (provided in `tasker.Config`) detected this specific error string and returned `false`, signaling `tasker` that the worker or its resource is unhealthy.

---

- **Diagnosis**: 

Review the `CheckHealth` implementation in `tasker.Config`. Analyze task logs for the specific error returned by the task function. Check the health of external services that the resource (`R`) interacts with.
- **Resolution**: 

If the error genuinely means the resource is unusable, `tasker`'s retry mechanism (if `MaxRetries > 0`) will attempt to re-process the task on a new worker. Ensure `OnCreate` for the resource is robust to create new, healthy instances. If the external service is down, that requires external intervention.
- **Prevention**: 

Implement robust error handling within your task functions. Use circuit breakers or exponential backoff for external service calls. Monitor resource health externally to proactively replace faulty instances.
- **Handling Patterns**: 



```
`tasker` automatically handles this by replacing the worker and potentially retrying the task. For the task caller, it's treated as a normal task failure, but with the possibility of being retried internally.
```

- **Propagation Behavior**: 

This error is returned by the task function. It is then passed to the `CheckHealth` function. If `CheckHealth` returns `false`, `tasker` handles the worker replacement and task retry internally. Eventually, if retries are exhausted, the error is returned to the original task submitter.

---

## `max retries exceeded`

- **Type**: 

Go `error` (specific string match, often wrapped)
- **Symptoms**: 

A task that previously triggered an unhealthy worker condition (i.e., `CheckHealth` returned `false`) is reported as failed with a message indicating retry exhaustion.
- **Properties**: 

None (standard `error` interface)

### Scenarios

###### A task repeatedly fails with an error that `CheckHealth` identifies as unhealthy, and the number of retry attempts reaches `MaxRetries`.

**Example**:



```go
// Config with MaxRetries: 1
// task returns errors.New("processor_crash") (unhealthy)
// first attempt fails, worker replaced, task re-queued (retry 1/1)
// second attempt fails with "processor_crash"
// -> Result: Task Failed: max retries exceeded: computation_error_task_X
```


**Reason**:

 The task manager exhausted all allowed retries for a task that consistently caused unhealthy worker conditions or was placed on an unhealthy worker, and could not complete successfully.

---

- **Diagnosis**: 

This error indicates a persistent problem with the task or the resources it's trying to use. Examine the underlying error that caused `CheckHealth` to return `false` repeatedly.
- **Resolution**: 

Investigate the root cause of the persistent unhealthy condition. This might involve debugging the task logic, checking external service health, or reviewing resource creation/destruction (`OnCreate`/`OnDestroy`). Consider increasing `MaxRetries` if the errors are truly transient but require more attempts.
- **Prevention**: 

Improve task robustness to handle transient errors internally without relying solely on `tasker`'s retry. Ensure external dependencies are stable. Set `MaxRetries` appropriately for the expected transience of errors.
- **Handling Patterns**: 



```
Catch this error to log it as a critical task failure. Potentially escalate to an alert if many tasks hit this state. These tasks cannot be recovered by `tasker`'s internal mechanism.
```

- **Propagation Behavior**: 

This error is returned directly to the caller of the submission methods (`QueueTask`, `QueueTaskWithPriority`, etc.) or delivered via callbacks/channels after all retries are exhausted.

---

## `failed to create temporary resource`

- **Type**: 

Go `error` (specific string match, wrapped with `%w`)
- **Symptoms**: 

A call to `RunTask` returns an error with this message, often wrapping another underlying error from `OnCreate`.
- **Properties**: 

None (standard `error` interface, contains underlying error via wrapping)

### Scenarios

###### When `RunTask` is called and the `resourcePool` is empty (or full, causing it to bypass the pool), it attempts to create a new, temporary resource via `OnCreate`, but `OnCreate` returns an error.

**Example**:



```go
// In Config: OnCreate: func() (*BadResource, error) { return nil, errors.New("connection_refused") }
_, err := manager.RunTask(...) // Will return 'failed to create temporary resource: connection_refused'
```


**Reason**:

 The `OnCreate` function, responsible for providing a resource to the task, failed during a `RunTask` call, preventing the task from executing.

---

- **Diagnosis**: 

Inspect the wrapped error within the returned `failed to create temporary resource` error. This wrapped error (`errors.Unwrap(err)`) will contain the specific reason `OnCreate` failed.
- **Resolution**: 

Debug the `OnCreate` function. Common causes include incorrect connection strings, unavailable external services, insufficient permissions, or resource exhaustion. Ensure `OnCreate` is robust and handles its own potential errors gracefully.
- **Prevention**: 

Pre-flight checks for external dependencies. Robust error handling within `OnCreate`. Consider implementing retries within `OnCreate` itself for transient resource creation issues.
- **Handling Patterns**: 



```
Log the error, especially unwrapping the root cause. This typically indicates an environment or configuration issue that needs attention.
```

- **Propagation Behavior**: 

This error is returned directly to the caller of `RunTask`.

---

## `priority queue full, task requeue failed`

- **Type**: 

Go `error` (specific string match, wrapped with `%w`)
- **Symptoms**: 

A task that was intended to be re-queued (due to `CheckHealth` returning `false` and retries remaining) fails with this error, indicating it could not be re-added to the priority queue.
- **Properties**: 

None (standard `error` interface, contains underlying error via wrapping)

### Scenarios

###### A task fails, `CheckHealth` indicates an unhealthy worker, retries are allowed, and `tasker` attempts to re-queue it to the priority queue, but the priority queue's buffer is full.

**Example**:



```go
// Config: MaxRetries: 1, PriorityQueue buffer is small
// Task fails with unhealthy error, manager tries to re-queue
// Priority queue is full at that exact moment
// -> Result: Task Failed: priority queue full, task requeue failed: [original_error]
```


**Reason**:

 The internal priority queue did not have capacity to accept a task that needed to be re-queued after an unhealthy worker condition. This usually happens under extreme load or if the priority queue's buffer size is too small relative to `MaxRetries` and task failure rate.

---

- **Diagnosis**: 

Check the `priorityQueue` buffer size (not directly configurable by user, but tied to `WorkerCount` in default implementation). Review overall system load and task failure rates. This is an internal `tasker` queue, so it usually points to overloaded conditions.
- **Resolution**: 

This indicates that even the priority retry mechanism is overloaded. Consider increasing `WorkerCount` or `MaxWorkerCount` to increase processing capacity, or reduce the rate of unhealthy errors. The task is permanently failed from `tasker`'s perspective.
- **Prevention**: 

Ensure adequate worker capacity (`WorkerCount`, `MaxWorkerCount`). Minimize unhealthy worker conditions through robust resources. For very high load, consider custom `MetricsCollector` to monitor queue depths and pre-scale.
- **Handling Patterns**: 



```
Log this error. This task is definitively lost from `tasker`'s management. Manual intervention or external retry logic may be needed.
```

- **Propagation Behavior**: 

This error is returned to the original caller of the submission methods or delivered via callbacks/channels.

---

