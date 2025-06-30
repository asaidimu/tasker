---
outline: "deep"
lastUpdated: true
editLink: true
prev: true
next: true
title: "Task-Based Guide"
description: "Task-Based Guide documentation and guidance"
---
# Task-Based Guide

### Resource Lifecycle Management
`tasker` manages the lifecycle of your task resources using `OnCreate` and `OnDestroy` functions provided in the `Config`.

*   **`OnCreate func() (R, error)`**: This function is called to create and initialize a new resource. It's invoked when a worker starts or when `RunTask` needs a temporary resource. It must return a new resource or an error.
*   **`OnDestroy func(R) error`**: This function performs cleanup or deallocation for a resource. It's called when a worker shuts down or a temporary resource is no longer needed. It should handle necessary resource finalization.

### Health Checks & Retries
`tasker` offers robust error handling and retries, especially for transient issues. You can define custom health checks and specify retry limits.

*   **`CheckHealth func(error) bool`**: An optional function that determines if an error indicates an "unhealthy" state for a worker or resource. If it returns `false`, the worker might be replaced, and the task re-queued (up to `MaxRetries`). If `nil`, all errors are considered healthy (task failure, not worker issue).
*   **`MaxRetries int`**: Specifies how many times a task will be re-queued if it fails due to an unhealthy worker/resource. Default is 3 retries if not specified.

### Asynchronous Submission Patterns
Beyond the blocking `QueueTask` and `QueueTaskWithPriority` methods, `tasker` provides non-blocking alternatives for more flexible integration into your application's concurrency model.

*   **`QueueTaskWithCallback(task func(context.Context, R) (E, error), callback func(E, error))`**: Submits a task and immediately returns. The provided `callback` function is invoked with the task's result and error once processing completes.
*   **`QueueTaskAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)`**: Submits a task and immediately returns two channels: one for the result (`<-chan E`) and one for the error (`<-chan error`). The caller can read from these channels to obtain the task's outcome without blocking the submission call itself.

These methods are also available for priority and "at-most-once" tasks (e.g., `QueueTaskWithPriorityWithCallback`, `QueueTaskOnceAsync`).

### "At-Most-Once" Execution
For non-idempotent operations where re-execution on a transient worker failure might cause undesirable side effects, `tasker` provides "at-most-once" semantics.

*   **`QueueTaskOnce(task func(context.Context, R) (E, error)) (E, error)`**: Similar to `QueueTask`, but this task will *not* be re-queued by `tasker`'s internal retry mechanism if it fails and `CheckHealth` indicates an unhealthy state. This ensures the task is attempted at most once by the task manager.
*   **`QueueTaskWithPriorityOnce(task func(context.Context, R) (E, error)) (E, error)`**: Combines high priority with "at-most-once" execution semantics. 

Both `QueueTaskOnce` and `QueueTaskWithPriorityOnce` also have `WithCallback` and `Async` variants for non-blocking submission.

### Dynamic Scaling (Bursting)
`tasker` can automatically adjust its worker count based on real-time workload using a rate-based dynamic scaling mechanism. This helps in optimizing resource utilization and responsiveness.

*   **`BurstInterval time.Duration`**: The frequency at which the burst manager checks queue sizes and adjusts the number of burst workers. Default is 100 milliseconds.
*   **`MaxWorkerCount int`**: The maximum total number of workers (base + burst) allowed. If 0, it defaults to `WorkerCount * 2`.

The burst manager monitors the `TaskArrivalRate` and `TaskCompletionRate`. If tasks are arriving faster than they are being completed, it scales up by adding burst workers. Conversely, if the system is over-provisioned, it scales down by gradually removing idle burst workers. This ensures throughput dynamically matches the incoming workload.

### Monitoring & Observability
`tasker` provides built-in mechanisms to monitor its operational state and performance, and allows integration with external logging and metrics systems.

*   **`Stats() TaskStats`**: Returns a snapshot of the manager's current operational state, including active worker counts, queued tasks, and available resources.
*   **`Metrics() TaskMetrics`**: Returns a `TaskMetrics` struct providing comprehensive performance metrics, such as task arrival/completion rates, various execution time percentiles (P95, P99), average wait times, and success/failure rates.
*   **Custom Logging (`tasker.Logger`)**: You can provide your own implementation of the `tasker.Logger` interface to integrate `tasker`'s internal logs with your application's preferred logging framework.
*   **Custom Metrics Collection (`tasker.MetricsCollector`)**: Implement the `tasker.MetricsCollector` interface to integrate `tasker`'s rich performance metrics with your existing monitoring and observability stack (e.g., Prometheus, Datadog).

