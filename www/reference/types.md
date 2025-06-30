---
outline: "deep"
lastUpdated: true
editLink: true
sidebar: true
title: "Types Reference"
description: "Complete reference of all type definitions"
---
# Types Reference

## `Config`

### Definition



```go
type Config[R any] struct {
	OnCreate func() (R, error)
	OnDestroy func(R) error
	WorkerCount int
	Ctx context.Context
	CheckHealth func(error) bool
	MaxWorkerCount int
	BurstInterval time.Duration
	MaxRetries int
	ResourcePoolSize int
	Logger Logger
	Collector MetricsCollector
	DeprecatedBurstTaskThreshold int
	DeprecatedBurstWorkerCount int
}
```


**Purpose**: Holds configuration parameters for creating a new TaskManager instance, controlling worker behavior, resource management, and scaling policies.

**Related Methods**: `NewTaskManager`

### Interface Contract

#### Parameter Object Structures

- **`Config`**: 

```go
type Config[R any] struct {
	OnCreate func() (R, error) // Required: Function to create a new resource.
	OnDestroy func(R) error     // Required: Function to destroy a resource.
	WorkerCount int             // Required: Initial and minimum number of base workers (must be > 0).
	Ctx context.Context         // Required: Parent context for the TaskManager.
	CheckHealth func(error) bool // Optional: Custom health check function. Defaults to always true.
	MaxWorkerCount int          // Optional: Maximum total workers (base + burst). Defaults to WorkerCount * 2.
	BurstInterval time.Duration  // Optional: Frequency for burst manager checks. Defaults to 100ms. Set to 0 to disable bursting.
	MaxRetries int               // Optional: Max retries for tasks on unhealthy errors. Defaults to 3.
	ResourcePoolSize int         // Optional: Number of resources to pre-allocate for RunTask. Defaults to WorkerCount.
	Logger Logger                 // Optional: Custom logger interface.
	Collector MetricsCollector   // Optional: Custom metrics collector.
	BurstTaskThreshold int       // Deprecated: No longer used for rate-based scaling.
	BurstWorkerCount int         // Deprecated: No longer used for rate-based scaling.
}
```


---

## `TaskStats`

### Definition



```go
type TaskStats struct {
	BaseWorkers        int32
	ActiveWorkers      int32
	BurstWorkers       int32
	QueuedTasks        int32
	PriorityTasks      int32
	AvailableResources int32
}
```


**Purpose**: Provides insight into the task manager's current operational state.

**Related Methods**: `Stats`

**Related Patterns**: `Get_live_stats`

### Interface Contract

#### Parameter Object Structures

- **`TaskStats`**: 

```go
type TaskStats struct {
	BaseWorkers        int32 // Number of permanently active workers.
	ActiveWorkers      int32 // Total number of currently active workers (base + burst).
	BurstWorkers       int32 // Number of dynamically scaled-up workers.
	QueuedTasks        int32 // Number of tasks currently in the main queue.
	PriorityTasks      int32 // Number of tasks currently in the priority queue.
	AvailableResources int32 // Number of resources currently available in the resource pool for RunTask.
}
```


---

## `TaskMetrics`

### Definition



```go
type TaskMetrics struct {
	AverageExecutionTime time.Duration
	MinExecutionTime time.Duration
	MaxExecutionTime time.Duration
	P95ExecutionTime time.Duration
	P99ExecutionTime time.Duration
	AverageWaitTime time.Duration
	TaskArrivalRate float64
	TaskCompletionRate float64
	TotalTasksCompleted uint64
	TotalTasksFailed uint64
	TotalTasksRetried uint64
	SuccessRate float64
	FailureRate float64
}
```


**Purpose**: Provides a comprehensive snapshot of performance, throughput, and reliability metrics for a TaskManager instance.

**Related Methods**: `Metrics`

**Related Patterns**: `Get_performance_metrics`

### Interface Contract

#### Parameter Object Structures

- **`TaskMetrics`**: 

```go
type TaskMetrics struct {
	AverageExecutionTime time.Duration // Average time spent executing a task.
	MinExecutionTime time.Duration    // Shortest execution time recorded.
	MaxExecutionTime time.Duration    // Longest execution time recorded.
	P95ExecutionTime time.Duration    // 95th percentile of task execution time.
	P99ExecutionTime time.Duration    // 99th percentile of task execution time.
	AverageWaitTime time.Duration     // Average time a task spends in a queue.
	TaskArrivalRate float64           // New tasks added to queues per second.
	TaskCompletionRate float64        // Tasks successfully completed per second.
	TotalTasksCompleted uint64       // Total tasks completed successfully since manager started.
	TotalTasksFailed uint64          // Total tasks failed after all retries.
	TotalTasksRetried uint64         // Total times any task has been retried.
	SuccessRate float64               // Ratio of completed to total terminal tasks.
	FailureRate float64               // Ratio of failed to total terminal tasks.
}
```


---

## `TaskLifecycleTimestamps`

### Definition



```go
type TaskLifecycleTimestamps struct {
	QueuedAt time.Time
	StartedAt time.Time
	FinishedAt time.Time
}
```


**Purpose**: Holds critical timestamps from a task's journey, used by `MetricsCollector` to calculate performance metrics.

**Related Methods**: `RecordCompletion`, `RecordFailure`

### Interface Contract

#### Parameter Object Structures

- **`TaskLifecycleTimestamps`**: 

```go
type TaskLifecycleTimestamps struct {
	QueuedAt time.Time   // Time when the task was first added to a queue.
	StartedAt time.Time  // Time when a worker began executing the task.
	FinishedAt time.Time // Time when the task execution completed (successfully or not).
}
```


---

