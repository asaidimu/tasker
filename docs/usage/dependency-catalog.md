# Dependency Catalog

## External Dependencies

### github.com/asaidimu/tasker
- **Purpose**: The core Tasker library itself; used as a dependency in example applications. Provides the TaskManager interface and implementation.
  - **Required Interfaces**:
    - `TaskManager`: The primary interface for managing asynchronous and synchronous task execution within a pool of workers and resources.
      - **Methods**:
        - `QueueTask`
          - **Signature**: `QueueTask(task func(R) (E, error)) (E, error)`
          - **Parameters**: task: A function representing the task's logic, accepting a resource of type `R` and returning a result `E` or an error.
          - **Returns**: Returns the task's result of type `E` and any error encountered during execution. The call blocks until completion.
        - `RunTask`
          - **Signature**: `RunTask(task func(R) (E, error)) (E, error)`
          - **Parameters**: task: A function representing the task's logic, accepting a resource of type `R` and returning a result `E` or an error.
          - **Returns**: Returns the task's result of type `E` and any error encountered during execution. The call blocks until completion.
        - `QueueTaskWithPriority`
          - **Signature**: `QueueTaskWithPriority(task func(R) (E, error)) (E, error)`
          - **Parameters**: task: A function representing the task's high-priority logic, accepting a resource of type `R` and returning a result `E` or an error.
          - **Returns**: Returns the task's result of type `E` and any error encountered during execution. The call blocks until completion.
        - `QueueTaskOnce`
          - **Signature**: `QueueTaskOnce(task func(R) (E, error)) (E, error)`
          - **Parameters**: task: A function representing the task's logic, accepting a resource of type `R` and returning a result `E` or an error. This task will not be re-queued if `CheckHealth` indicates an unhealthy state.
          - **Returns**: Returns the task's result of type `E` and any error encountered during execution. The call blocks until completion.
        - `QueueTaskWithPriorityOnce`
          - **Signature**: `QueueTaskWithPriorityOnce(task func(R) (E, error)) (E, error)`
          - **Parameters**: task: A function representing the task's high-priority logic, accepting a resource of type `R` and returning a result `E` or an error. This task will not be re-queued if `CheckHealth` indicates an unhealthy state.
          - **Returns**: Returns the task's result of type `E` and any error encountered during execution. The call blocks until completion.
        - `Stop`
          - **Signature**: `Stop() error`
          - **Parameters**: None.
          - **Returns**: Returns an error if the manager is already stopping or killed, otherwise nil.
        - `Kill`
          - **Signature**: `Kill() error`
          - **Parameters**: None.
          - **Returns**: Returns an error if the manager is already killed, otherwise nil.
        - `Stats`
          - **Signature**: `Stats() TaskStats`
          - **Parameters**: None.
          - **Returns**: Returns a `TaskStats` struct containing current operational statistics.
        - `Metrics`
          - **Signature**: `Metrics() TaskMetrics`
          - **Parameters**: None.
          - **Returns**: Returns a `TaskMetrics` struct containing aggregated performance metrics.
    - `Logger`: Interface for custom logging within Tasker. Users can provide their own implementation.
      - **Methods**:
        - `Debugf`
          - **Signature**: `Debugf(format string, args ...any)`
          - **Parameters**: format: Format string; args: Arguments for formatting.
          - **Returns**: None.
        - `Infof`
          - **Signature**: `Infof(format string, args ...any)`
          - **Parameters**: format: Format string; args: Arguments for formatting.
          - **Returns**: None.
        - `Warnf`
          - **Signature**: `Warnf(format string, args ...any)`
          - **Parameters**: format: Format string; args: Arguments for formatting.
          - **Returns**: None.
        - `Errorf`
          - **Signature**: `Errorf(format string, args ...any)`
          - **Parameters**: format: Format string; args: Arguments for formatting.
          - **Returns**: None.
    - `MetricsCollector`: Interface for collecting and calculating performance and reliability metrics for the TaskManager. Users can provide their own implementation.
      - **Methods**:
        - `RecordArrival`
          - **Signature**: `RecordArrival()`
          - **Parameters**: None.
          - **Returns**: None.
        - `RecordCompletion`
          - **Signature**: `RecordCompletion(stamps TaskLifecycleTimestamps)`
          - **Parameters**: stamps: `TaskLifecycleTimestamps` containing queued, started, and finished times for a completed task.
          - **Returns**: None.
        - `RecordFailure`
          - **Signature**: `RecordFailure(stamps TaskLifecycleTimestamps)`
          - **Parameters**: stamps: `TaskLifecycleTimestamps` containing queued, started, and finished times for a failed task.
          - **Returns**: None.
        - `RecordRetry`
          - **Signature**: `RecordRetry()`
          - **Parameters**: None.
          - **Returns**: None.
        - `Metrics`
          - **Signature**: `Metrics() TaskMetrics`
          - **Parameters**: None.
          - **Returns**: Returns a `TaskMetrics` struct containing a snapshot of aggregated performance metrics.
- **Installation**: `go get github.com/asaidimu/tasker`
- **Version Compatibility**: `>=1.0.0`

### context
- **Purpose**: Go standard library package for carrying deadlines, cancellation signals, and other request-scoped values across API boundaries and between goroutines.
  - **Required Interfaces**:
    - `context.Context`: The fundamental interface for context propagation in Go.
      - **Methods**:
        - `Done`
          - **Signature**: `Done() <-chan struct{}`
          - **Parameters**: None.
          - **Returns**: Returns a channel that is closed when the context is cancelled or times out.
        - `Err`
          - **Signature**: `Err() error`
          - **Parameters**: None.
          - **Returns**: Returns a non-nil error if Done is closed, specifying why the context was cancelled (e.g., `Canceled` or `DeadlineExceeded`).
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### errors
- **Purpose**: Go standard library package for error handling, including creating new errors and unwrapping them.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### fmt
- **Purpose**: Go standard library package for formatted I/O.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### log
- **Purpose**: Go standard library package for simple logging.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### math
- **Purpose**: Go standard library package for common mathematical functions.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### math/rand
- **Purpose**: Go standard library package for pseudo-random number generation.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### sort
- **Purpose**: Go standard library package for sorting slices and user-defined collections.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### sync
- **Purpose**: Go standard library package for basic synchronization primitives like mutexes and wait groups.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### sync/atomic
- **Purpose**: Go standard library package for low-level atomic memory primitives.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

### time
- **Purpose**: Go standard library package for measuring and displaying time.
- **Installation**: `Built-in to Go standard library.`
- **Version Compatibility**: `>=1.24.3 (Go version)`

## Peer Dependencies

### Go Runtime
- **Reason**: Required for compiling and running Tasker applications. Tasker leverages Go's concurrency primitives (goroutines, channels) directly.
- **Version Requirements**: `>=1.24.3`



---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*