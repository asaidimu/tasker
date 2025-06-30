---
outline: "deep"
lastUpdated: true
editLink: true
sidebar: true
title: "Interfaces Reference"
description: "Complete reference of all interface definitions"
---
# Interfaces Reference

## `Logger`

### Definition



```go
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}
```


**Purpose**: Defines the interface for logging messages from the TaskManager, allowing users to integrate their own preferred logging library.

**Related Patterns**: `Custom_logging_integration`

### Interface Contract

#### Required Methods

##### `Debugf`

**Signature**: 

```go
Debugf(format string, args ...any)
```


- **Parameters**: 

```go
format: string (printf-style format string); args: ...any (arguments for format string)
```

- **Return Value**: None
- **Side Effects**: Should log a message at debug level.

##### `Infof`

**Signature**: 

```go
Infof(format string, args ...any)
```


- **Parameters**: 

```go
format: string (printf-style format string); args: ...any (arguments for format string)
```

- **Return Value**: None
- **Side Effects**: Should log a message at info level.

##### `Warnf`

**Signature**: 

```go
Warnf(format string, args ...any)
```


- **Parameters**: 

```go
format: string (printf-style format string); args: ...any (arguments for format string)
```

- **Return Value**: None
- **Side Effects**: Should log a message at warning level.

##### `Errorf`

**Signature**: 

```go
Errorf(format string, args ...any)
```


- **Parameters**: 

```go
format: string (printf-style format string); args: ...any (arguments for format string)
```

- **Return Value**: None
- **Side Effects**: Should log a message at error level.

#### Parameter Object Structures


---

## `MetricsCollector`

### Definition



```go
type MetricsCollector interface {
	RecordArrival()
	RecordCompletion(stamps TaskLifecycleTimestamps)
	RecordFailure(stamps TaskLifecycleTimestamps)
	RecordRetry()
	Metrics() TaskMetrics
}
```


**Purpose**: Defines the interface for collecting and calculating performance and reliability metrics for the TaskManager. Implementations process task lifecycle events and aggregate data.

**Related Patterns**: `Custom_metrics_integration`

### Interface Contract

#### Required Methods

##### `RecordArrival`

**Signature**: 

```go
RecordArrival()
```


- **Parameters**: 

```go
None
```

- **Return Value**: None
- **Side Effects**: Increments internal counter for task arrivals.

##### `RecordCompletion`

**Signature**: 

```go
RecordCompletion(stamps TaskLifecycleTimestamps)
```


- **Parameters**: 

```go
stamps: TaskLifecycleTimestamps (timing information for the completed task)
```

- **Return Value**: None
- **Side Effects**: Updates internal metrics for total tasks completed, execution time, wait time, and min/max/percentile execution times.

##### `RecordFailure`

**Signature**: 

```go
RecordFailure(stamps TaskLifecycleTimestamps)
```


- **Parameters**: 

```go
stamps: TaskLifecycleTimestamps (timing information for the failed task)
```

- **Return Value**: None
- **Side Effects**: Increments internal counter for total tasks failed.

##### `RecordRetry`

**Signature**: 

```go
RecordRetry()
```


- **Parameters**: 

```go
None
```

- **Return Value**: None
- **Side Effects**: Increments internal counter for total tasks retried.

##### `Metrics`

**Signature**: 

```go
Metrics() TaskMetrics
```


- **Parameters**: 

```go
None
```

- **Return Value**: TaskMetrics (a snapshot of the currently aggregated performance metrics)
- **Side Effects**: None (reads internal state, does not modify)

#### Parameter Object Structures


---

## `TaskManager`

### Definition



```go
type TaskManager[R any, E any] interface { ... }
```


**Purpose**: Defines the interface for managing asynchronous and synchronous task execution within a pool of workers and resources.

**Related Methods**: `NewTaskManager`

### Interface Contract

#### Required Methods

##### `QueueTask`

**Signature**: 

```go
QueueTask(task func(context.Context, R) (E, error)) (E, error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error) (the function representing the task's logic)
```

- **Return Value**: E (the task's result), error (any error encountered during execution or submission)
- **Side Effects**: Adds task to main queue, increments `queuedTasks` metric, initiates task execution, decrements `queuedTasks` on processing. Blocks caller until task completes.

##### `QueueTaskWithCallback`

**Signature**: 

```go
QueueTaskWithCallback(task func(context.Context, R) (E, error), callback func(E, error))
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error); callback: func(E, error) (function to be invoked with task's result/error)
```

- **Return Value**: None
- **Side Effects**: Adds task to main queue, increments `queuedTasks` metric, initiates task execution. Returns immediately. `callback` is invoked asynchronously.

##### `QueueTaskAsync`

**Signature**: 

```go
QueueTaskAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Return Value**: <-chan E (channel for task result), <-chan error (channel for task error)
- **Side Effects**: Adds task to main queue, increments `queuedTasks` metric, initiates task execution. Returns immediately. Result/error are sent to provided channels.

##### `RunTask`

**Signature**: 

```go
RunTask(task func(context.Context, R) (E, error)) (E, error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Return Value**: E (the task's result), error (any error encountered during execution)
- **Side Effects**: Acquires resource (from pool or temporary), executes task, releases/destroys resource. Blocks caller until task completes.

##### `QueueTaskWithPriority`

**Signature**: 

```go
QueueTaskWithPriority(task func(context.Context, R) (E, error)) (E, error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Return Value**: E (the task's result), error (any error encountered during execution or submission)
- **Side Effects**: Adds task to priority queue, increments `priorityTasks` metric, initiates task execution, decrements `priorityTasks` on processing. Blocks caller until task completes.

##### `QueueTaskWithPriorityWithCallback`

**Signature**: 

```go
QueueTaskWithPriorityWithCallback(task func(context.Context, R) (E, error), callback func(E, error))
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error); callback: func(E, error)
```

- **Return Value**: None
- **Side Effects**: Adds task to priority queue, increments `priorityTasks` metric, initiates task execution. Returns immediately. `callback` is invoked asynchronously.

##### `QueueTaskWithPriorityAsync`

**Signature**: 

```go
QueueTaskWithPriorityAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Return Value**: <-chan E (channel for task result), <-chan error (channel for task error)
- **Side Effects**: Adds task to priority queue, increments `priorityTasks` metric, initiates task execution. Returns immediately. Result/error are sent to provided channels.

##### `QueueTaskOnce`

**Signature**: 

```go
QueueTaskOnce(task func(context.Context, R) (E, error)) (E, error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Return Value**: E (the task's result), error (any error encountered during execution or submission)
- **Side Effects**: Adds task to main queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Blocks caller until task completes.

##### `QueueTaskOnceWithCallback`

**Signature**: 

```go
QueueTaskOnceWithCallback(task func(context.Context, R) (E, error), callback func(E, error))
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error); callback: func(E, error)
```

- **Return Value**: None
- **Side Effects**: Adds task to main queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Returns immediately. `callback` is invoked asynchronously.

##### `QueueTaskOnceAsync`

**Signature**: 

```go
QueueTaskOnceAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Return Value**: <-chan E (channel for task result), <-chan error (channel for task error)
- **Side Effects**: Adds task to main queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Returns immediately. Result/error are sent to provided channels.

##### `QueueTaskWithPriorityOnce`

**Signature**: 

```go
QueueTaskWithPriorityOnce(task func(context.Context, R) (E, error)) (E, error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Return Value**: E (the task's result), error (any error encountered during execution or submission)
- **Side Effects**: Adds task to priority queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Blocks caller until task completes.

##### `QueueTaskWithPriorityOnceWithCallback`

**Signature**: 

```go
QueueTaskWithPriorityOnceWithCallback(task func(context.Context, R) (E, error), callback func(E, error))
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error); callback: func(E, error)
```

- **Return Value**: None
- **Side Effects**: Adds task to priority queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Returns immediately. `callback` is invoked asynchronously.

##### `QueueTaskWithPriorityOnceAsync`

**Signature**: 

```go
QueueTaskWithPriorityOnceAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)
```


- **Parameters**: 

```go
task: func(context.Context, R) (E, error)
```

- **Return Value**: <-chan E (channel for task result), <-chan error (channel for task error)
- **Side Effects**: Adds task to priority queue. Task is configured not to be re-queued if `CheckHealth` indicates unhealthy. Returns immediately. Result/error are sent to provided channels.

##### `Stop`

**Signature**: 

```go
Stop() error
```


- **Parameters**: 

```go
None
```

- **Return Value**: error (nil on successful graceful shutdown, an error if manager is already stopping/killed)
- **Side Effects**: Sets shutdown state to 'stopping', stops accepting new tasks, cancels main context, signals workers to drain queues, waits for all goroutines to complete, drains resource pool.

##### `Kill`

**Signature**: 

```go
Kill() error
```


- **Parameters**: 

```go
None
```

- **Return Value**: error (nil on successful immediate shutdown, an error if manager is already killed)
- **Side Effects**: Sets shutdown state to 'killed', stops accepting new tasks, cancels main context, signals workers to terminate immediately (dropping queued tasks), waits for all goroutines to complete, drains resource pool.

##### `Stats`

**Signature**: 

```go
Stats() TaskStats
```


- **Parameters**: 

```go
None
```

- **Return Value**: TaskStats (current operational state snapshot)
- **Side Effects**: None (reads atomic counters, does not modify state)

##### `Metrics`

**Signature**: 

```go
Metrics() TaskMetrics
```


- **Parameters**: 

```go
None
```

- **Return Value**: TaskMetrics (comprehensive performance metrics snapshot)
- **Side Effects**: None (reads metrics collector state, does not modify)

#### Parameter Object Structures


---

