---
outline: "deep"
lastUpdated: true
editLink: true
sidebar: true
title: "Methods Reference"
description: "Complete reference of all available methods and functions"
---
# Methods Reference

## `NewRunner`

- **Use Case**: Deprecated: Use `NewTaskManager` instead.
- **Signature**: 

```go
NewRunner[R any, E any](config Config[R]) (TaskManager[R, E], error)
```

- **Parameters**: 

```go
config: Config[R] (configuration parameters for the task manager)
```

- **Prerequisites**: None.
- **Side Effects**: Initializes and starts worker goroutines, resource pool, and burst manager.
- **Return Value**: TaskManager[R, E] (an instance of the task manager), error (if initialization fails)
- **Exceptions**: `errors.New("worker count must be positive")`, `errors.New("onCreate function is required")`, `errors.New("onDestroy function is required")`, `fmt.Errorf("failed to initialize resource pool: %w", err)`
- **Availability**: sync
- **Status**: deprecated
- **Related Types**: `Config`

---

## `NewTaskManager`

- **Use Case**: To create and initialize a new `TaskManager` instance for managing concurrent tasks with custom resources.
- **Signature**: 

```go
NewTaskManager[R any, E any](config Config[R]) (TaskManager[R, E], error)
```

- **Parameters**: 

```go
config: Config[R] (configuration parameters for the task manager, including resource lifecycle, worker counts, health checks, and scaling)
```

- **Prerequisites**: The `Config` struct must be valid: `WorkerCount` > 0, `OnCreate` and `OnDestroy` functions must be provided.
- **Side Effects**: Initializes the resource pool, starts the specified number of base worker goroutines, and begins the burst manager's dynamic scaling routine. Resources are created via `OnCreate`.
- **Return Value**: TaskManager[R, E] (an instance of the task manager interface, ready to accept tasks), error (if initialization fails, e.g., invalid config or initial resource creation error)
- **Exceptions**: `errors.New("worker count must be positive")`, `errors.New("onCreate function is required")`, `errors.New("onDestroy function is required")`, `fmt.Errorf("failed to initialize resource pool: %w", err)`
- **Availability**: sync
- **Status**: active
- **Related Types**: `Config`
- **Related Patterns**: `Basic_tasker_setup`

---

## `QueueTask`

- **Use Case**: To submit a standard asynchronous task to the manager's main queue and wait for its completion. This is suitable for background processing where immediate, non-blocking submission is not a strict requirement.
- **Signature**: 

```go
QueueTask(task func(context.Context, R) (E, error)) (E, error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error) (the function defining the task's logic. It receives a `context.Context` for cancellation and a resource `R`.)
```

- **Prerequisites**: The `TaskManager` must be in a running state. The task function should ideally be idempotent if retries are enabled and `CheckHealth` can deem a worker unhealthy.
- **Side Effects**: Adds the task to the main queue. Increments `queuedTasks` and `totalTasksArrived` metrics. An available worker will eventually pick it up, execute it, and potentially retry it if `CheckHealth` indicates an unhealthy worker and `MaxRetries` allows. Blocks the calling goroutine until the task finishes execution or manager shuts down.
- **Return Value**: E (the result of the task's execution), error (nil if successful, or an error if the task fails, is cancelled, or manager shuts down)
- **Exceptions**: `errors.New("task manager is shutting down")`, `fmt.Errorf("priority queue full, task requeue failed: %w", err)`
- **Availability**: sync
- **Status**: active
- **Related Types**: `Task`
- **Related Patterns**: `Queue_and_wait`
- **Related Errors**: `task manager is shutting down`, `priority queue full, task requeue failed`

---

## `QueueTaskWithCallback`

- **Use Case**: To submit a standard asynchronous task to the manager's main queue without blocking the caller. The task's result is delivered via a callback function once completed.
- **Signature**: 

```go
QueueTaskWithCallback(task func(context.Context, R) (E, error), callback func(E, error))
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error) (the function defining the task's logic); callback: func(E, error) (a function that will be invoked with the task's result and error)
```

- **Prerequisites**: The `TaskManager` must be in a running state. The task function should ideally be idempotent if retries are enabled and `CheckHealth` can deem a worker unhealthy.
- **Side Effects**: Adds the task to the main queue. Increments `queuedTasks` and `totalTasksArrived` metrics. An available worker will eventually pick it up. The provided `callback` function is invoked asynchronously upon task completion or failure.
- **Return Value**: None (returns immediately)
- **Availability**: async
- **Status**: active
- **Related Types**: `Task`
- **Related Patterns**: `Queue_with_callback`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskAsync`

- **Use Case**: To submit a standard asynchronous task to the manager's main queue and receive its result/error via channels, allowing for flexible non-blocking consumption of results.
- **Signature**: 

```go
QueueTaskAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error) (the function defining the task's logic)
```

- **Prerequisites**: The `TaskManager` must be in a running state. The task function should ideally be idempotent if retries are enabled and `CheckHealth` can deem a worker unhealthy.
- **Side Effects**: Adds the task to the main queue. Increments `queuedTasks` and `totalTasksArrived` metrics. An available worker will eventually pick it up. Returns immediately. The result and error are sent to the returned channels once the task completes.
- **Return Value**: <-chan E (a receive-only channel for the task's result), <-chan error (a receive-only channel for any error encountered)
- **Availability**: async
- **Status**: active
- **Related Types**: `Task`
- **Related Patterns**: `Queue_asynchronously_with_channels`
- **Related Errors**: `task manager is shutting down`

---

## `RunTask`

- **Use Case**: To execute a task immediately, bypassing queues, typically for urgent or latency-sensitive operations. It tries to use a resource from a pre-allocated pool or creates a temporary one if the pool is empty.
- **Signature**: 

```go
RunTask(task func(context.Context, R) (E, error)) (E, error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error) (the function defining the task's logic. It receives a `context.Context` for cancellation and a resource `R`.)
```

- **Prerequisites**: The `TaskManager` must be in a running state. `ResourcePoolSize` should be configured if aiming to reuse resources.
- **Side Effects**: Acquires a resource (from pool or temporary creation), executes the task, then returns the resource to the pool or destroys it. Increments `totalTasksArrived`, `totalTasksCompleted`, or `totalTasksFailed` metrics. Blocks the calling goroutine until the task finishes execution.
- **Return Value**: E (the result of the task's execution), error (nil if successful, or an error if the task fails, is cancelled, or resource creation fails)
- **Exceptions**: `errors.New("task manager is shutting down")`, `fmt.Errorf("failed to create temporary resource: %w", err)`
- **Availability**: sync
- **Status**: active
- **Related Errors**: `task manager is shutting down`, `failed to create temporary resource`

---

## `QueueTaskWithPriority`

- **Use Case**: To submit a high-priority asynchronous task to a dedicated priority queue and wait for its completion. These tasks are processed before standard tasks.
- **Signature**: 

```go
QueueTaskWithPriority(task func(context.Context, R) (E, error)) (E, error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error) (the function defining the task's logic. It receives a `context.Context` for cancellation and a resource `R`.)
```

- **Prerequisites**: The `TaskManager` must be in a running state. The task function should ideally be idempotent if retries are enabled and `CheckHealth` can deem a worker unhealthy.
- **Side Effects**: Adds the task to the priority queue. Increments `priorityTasks` and `totalTasksArrived` metrics. An available worker will pick it up preferentially, execute it, and potentially retry it. Blocks the calling goroutine until the task finishes execution or manager shuts down.
- **Return Value**: E (the result of the task's execution), error (nil if successful, or an error if the task fails, is cancelled, or manager shuts down)
- **Exceptions**: `errors.New("task manager is shutting down")`
- **Availability**: sync
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskWithPriorityWithCallback`

- **Use Case**: To submit a high-priority asynchronous task to the dedicated priority queue without blocking the caller. The task's result is delivered via a callback function once completed.
- **Signature**: 

```go
QueueTaskWithPriorityWithCallback(task func(context.Context, R) (E, error), callback func(E, error))
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error); callback: func(E, error)
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Adds the task to the priority queue. Increments `priorityTasks` and `totalTasksArrived` metrics. Returns immediately. The `callback` is invoked asynchronously upon task completion or failure.
- **Return Value**: None
- **Availability**: async
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskWithPriorityAsync`

- **Use Case**: To submit a high-priority asynchronous task to the dedicated priority queue and receive its result/error via channels, allowing for flexible non-blocking consumption of results.
- **Signature**: 

```go
QueueTaskWithPriorityAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Adds the task to the priority queue. Increments `priorityTasks` and `totalTasksArrived` metrics. Returns immediately. The result and error are sent to the returned channels once the task completes.
- **Return Value**: <-chan E (a receive-only channel for the task's result), <-chan error (a receive-only channel for any error encountered)
- **Availability**: async
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskOnce`

- **Use Case**: To submit a standard asynchronous task that will *not* be retried by the manager if it fails and `CheckHealth` indicates an unhealthy state. Suitable for non-idempotent operations where "at-most-once" execution by the task manager is desired.
- **Signature**: 

```go
QueueTaskOnce(task func(context.Context, R) (E, error)) (E, error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error) (the task function)
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Adds the task to the main queue. Task is configured with maximum retries from the start, effectively disabling automatic re-queueing by `tasker`'s internal retry mechanism if `CheckHealth` returns false. Blocks the calling goroutine until the task finishes or manager shuts down.
- **Return Value**: E (the result of the task), error (nil if successful, or an error if the task fails or manager shuts down)
- **Exceptions**: `errors.New("task manager is shutting down")`
- **Availability**: sync
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskOnceWithCallback`

- **Use Case**: To submit a standard asynchronous task with "at-most-once" semantics using a callback for non-blocking result delivery.
- **Signature**: 

```go
QueueTaskOnceWithCallback(task func(context.Context, R) (E, error), callback func(E, error))
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error); callback: func(E, error)
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Adds the task to the main queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Returns immediately. `callback` is invoked asynchronously upon task completion or failure.
- **Return Value**: None
- **Availability**: async
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskOnceAsync`

- **Use Case**: To submit a standard asynchronous task with "at-most-once" semantics and receive its result/error via channels for non-blocking consumption.
- **Signature**: 

```go
QueueTaskOnceAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Adds the task to the main queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Returns immediately. Result/error are sent to provided channels.
- **Return Value**: <-chan E (result channel), <-chan error (error channel)
- **Availability**: async
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskWithPriorityOnce`

- **Use Case**: To submit a high-priority asynchronous task that will *not* be retried by the manager if it fails and `CheckHealth` indicates an unhealthy state. Suitable for non-idempotent high-priority operations.
- **Signature**: 

```go
QueueTaskWithPriorityOnce(task func(context.Context, R) (E, error)) (E, error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error) (the task function)
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Adds the task to the priority queue. Task is configured with maximum retries, effectively disabling automatic re-queueing by `tasker`'s internal retry mechanism if `CheckHealth` returns false. Blocks the calling goroutine until the task finishes or manager shuts down.
- **Return Value**: E (the result of the task), error (nil if successful, or an error if the task fails or manager shuts down)
- **Exceptions**: `errors.New("task manager is shutting down")`
- **Availability**: sync
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskWithPriorityOnceWithCallback`

- **Use Case**: To submit a high-priority asynchronous task with "at-most-once" semantics using a callback for non-blocking result delivery.
- **Signature**: 

```go
QueueTaskWithPriorityOnceWithCallback(task func(context.Context, R) (E, error), callback func(E, error))
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error); callback: func(E, error)
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Adds the task to the priority queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Returns immediately. `callback` is invoked asynchronously upon task completion or failure.
- **Return Value**: None
- **Availability**: async
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `QueueTaskWithPriorityOnceAsync`

- **Use Case**: To submit a high-priority asynchronous task with "at-most-once" semantics and receive its result/error via channels for non-blocking consumption.
- **Signature**: 

```go
QueueTaskWithPriorityOnceAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)
```

- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Adds the task to the priority queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Returns immediately. Result/error are sent to provided channels.
- **Return Value**: <-chan E (result channel), <-chan error (error channel)
- **Availability**: async
- **Status**: active
- **Related Types**: `Task`
- **Related Errors**: `task manager is shutting down`

---

## `Stop`

- **Use Case**: To gracefully shut down the `TaskManager`. This involves stopping new task submissions, allowing existing queued tasks to complete, and then releasing all resources.
- **Signature**: 

```go
Stop() error
```

- **Parameters**: 

```go
None
```

- **Prerequisites**: The `TaskManager` must be in a running state.
- **Side Effects**: Transitions manager state to 'stopping'. Cancels the main context (signaling workers to drain queues). Stops the burst manager. Waits for all worker goroutines to complete. Drains and destroys all resources in the pool. Prevents new tasks from being queued.
- **Return Value**: error (nil on successful graceful shutdown; `errors.New("task manager already stopping or killed")` if already in a shutdown state)
- **Availability**: sync
- **Status**: active
- **Related Patterns**: `Graceful_shutdown`
- **Related Errors**: `task manager already stopping or killed`

---

## `Kill`

- **Use Case**: To immediately terminate the `TaskManager` without waiting for queued tasks to complete. All running tasks are cancelled, and resources are released abruptly. Use for urgent shutdowns.
- **Signature**: 

```go
Kill() error
```

- **Parameters**: 

```go
None
```

- **Prerequisites**: The `TaskManager` must be in a running state or stopping state.
- **Side Effects**: Transitions manager state to 'killed'. Cancels the main context (signaling workers to exit immediately). Stops the burst manager. Waits for all worker goroutines to terminate. Drains and destroys all resources. Drops any tasks still in queues. Prevents new tasks from being queued.
- **Return Value**: error (nil on successful immediate shutdown; `errors.New("task manager already killed")` if already in 'killed' state)
- **Availability**: sync
- **Status**: active
- **Related Patterns**: `Immediate_shutdown`
- **Related Errors**: `task manager already killed`

---

## `Stats`

- **Use Case**: To retrieve a real-time snapshot of the `TaskManager`'s current operational state, including worker counts and queue sizes.
- **Signature**: 

```go
Stats() TaskStats
```

- **Parameters**: 

```go
None
```

- **Prerequisites**: None.
- **Side Effects**: None.
- **Return Value**: TaskStats (a struct containing current worker counts, queued tasks, and available resources)
- **Availability**: sync
- **Status**: active
- **Related Types**: `TaskStats`
- **Related Patterns**: `Get_live_stats`

---

## `Metrics`

- **Use Case**: To retrieve comprehensive performance metrics collected by the `TaskManager`, such as task rates, execution times, and success/failure rates.
- **Signature**: 

```go
Metrics() TaskMetrics
```

- **Parameters**: 

```go
None
```

- **Prerequisites**: None.
- **Side Effects**: None.
- **Return Value**: TaskMetrics (a struct containing aggregated performance metrics)
- **Availability**: sync
- **Status**: active
- **Related Types**: `TaskMetrics`
- **Related Patterns**: `Get_performance_metrics`

---

## `NewCollector`

- **Use Case**: To create and initialize a new default `MetricsCollector` instance. Typically used internally or when providing a default collector.
- **Signature**: 

```go
NewCollector() MetricsCollector
```

- **Parameters**: 

```go
None
```

- **Prerequisites**: None.
- **Side Effects**: Initializes internal state for metric aggregation.
- **Return Value**: MetricsCollector (a new instance of the default metrics collector)
- **Availability**: sync
- **Status**: active

---

## `RecordArrival`

- **Use Case**: Records that a new task has arrived in the system (queued). Used internally by `TaskManager`.
- **Signature**: 

```go
RecordArrival()
```

- **Parameters**: 

```go
None
```

- **Prerequisites**: None.
- **Side Effects**: Increments the total tasks arrived counter.
- **Return Value**: None
- **Availability**: sync
- **Status**: active

---

## `RecordCompletion`

- **Use Case**: Records that a task has completed successfully. Used internally by `TaskManager`.
- **Signature**: 

```go
RecordCompletion(stamps TaskLifecycleTimestamps)
```

- **Parameters**: 

```go
stamps: TaskLifecycleTimestamps (timing information for the completed task)
```

- **Prerequisites**: None.
- **Side Effects**: Increments total tasks completed, updates total/min/max/percentile execution times and total wait time.
- **Return Value**: None
- **Availability**: sync
- **Status**: active
- **Related Types**: `TaskLifecycleTimestamps`

---

## `RecordFailure`

- **Use Case**: Records that a task has failed permanently (after retries). Used internally by `TaskManager`.
- **Signature**: 

```go
RecordFailure(stamps TaskLifecycleTimestamps)
```

- **Parameters**: 

```go
stamps: TaskLifecycleTimestamps (timing information for the failed task)
```

- **Prerequisites**: None.
- **Side Effects**: Increments the total tasks failed counter.
- **Return Value**: None
- **Availability**: sync
- **Status**: active
- **Related Types**: `TaskLifecycleTimestamps`

---

## `RecordRetry`

- **Use Case**: Records that a task has been re-queued for a retry. Used internally by `TaskManager`.
- **Signature**: 

```go
RecordRetry()
```

- **Parameters**: 

```go
None
```

- **Prerequisites**: None.
- **Side Effects**: Increments the total tasks retried counter.
- **Return Value**: None
- **Availability**: sync
- **Status**: active

---

## `Debugf`

- **Use Case**: Logs a message at the debug level. Part of the `Logger` interface.
- **Signature**: 

```go
Debugf(format string, args ...any)
```

- **Parameters**: 

```go
format: string (printf-style format string); args: ...any (arguments for format string)
```

- **Prerequisites**: None.
- **Side Effects**: Emits a log message.
- **Return Value**: None
- **Availability**: sync
- **Status**: active

---

## `Infof`

- **Use Case**: Logs a message at the info level. Part of the `Logger` interface.
- **Signature**: 

```go
Infof(format string, args ...any)
```

- **Parameters**: 

```go
format: string (printf-style format string); args: ...any (arguments for format string)
```

- **Prerequisites**: None.
- **Side Effects**: Emits a log message.
- **Return Value**: None
- **Availability**: sync
- **Status**: active

---

## `Warnf`

- **Use Case**: Logs a message at the warning level. Part of the `Logger` interface.
- **Signature**: 

```go
Warnf(format string, args ...any)
```

- **Parameters**: 

```go
format: string (printf-style format string); args: ...any (arguments for format string)
```

- **Prerequisites**: None.
- **Side Effects**: Emits a log message.
- **Return Value**: None
- **Availability**: sync
- **Status**: active

---

## `Errorf`

- **Use Case**: Logs a message at the error level. Part of the `Logger` interface.
- **Signature**: 

```go
Errorf(format string, args ...any)
```

- **Parameters**: 

```go
format: string (printf-style format string); args: ...any (arguments for format string)
```

- **Prerequisites**: None.
- **Side Effects**: Emits a log message.
- **Return Value**: None
- **Availability**: sync
- **Status**: active

---

