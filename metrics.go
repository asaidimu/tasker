package tasker

import "time"

// TaskMetrics provides a comprehensive snapshot of performance, throughput, and reliability
// metrics for a TaskManager instance. These metrics offer deep insights into the behavior
// of the task execution system over time.
type TaskMetrics struct {
	//
	// --- Latency & Duration Metrics ---
	// These metrics measure the time taken for various stages of a task's lifecycle.
	// They are crucial for understanding performance bottlenecks and user-perceived speed.
	//

	// AverageExecutionTime is the average time spent executing a task, from the moment
	// a worker picks it up until it completes. (Unit: time.Duration)
	AverageExecutionTime time.Duration

	// MinExecutionTime is the shortest execution time recorded for any single task.
	// Useful for understanding the best-case performance scenario. (Unit: time.Duration)
	MinExecutionTime time.Duration

	// MaxExecutionTime is the longest execution time recorded for any single task.
	// Useful for identifying outliers and potential long-running task issues. (Unit: time.Duration)
	MaxExecutionTime time.Duration

	// P95ExecutionTime is the 95th percentile of task execution time. 95% of tasks
	// completed in this time or less. This is a key indicator of tail latency. (Unit: time.Duration)
	P95ExecutionTime time.Duration

	// P99ExecutionTime is the 99th percentile of task execution time. 99% of tasks
	// completed in this time or less. Helps in understanding the performance for the
	// vast majority of tasks, excluding extreme outliers. (Unit: time.Duration)
	P99ExecutionTime time.Duration

	// AverageWaitTime is the average time a task spends in a queue before being
	// picked up by a worker. High values may indicate that the system is under-provisioned.
	// (Unit: time.Duration)
	AverageWaitTime time.Duration

	//
	// --- Throughput & Volume Metrics ---
	// These metrics measure the overall workload and processing capacity of the system.
	//

	// TaskArrivalRate is the number of new tasks being added to the queues per second,
	// calculated over a recent time window.
	TaskArrivalRate float64

	// TaskCompletionRate is the number of tasks being successfully completed per second,
	// calculated over a recent time window. This is a primary measure of system throughput.
	TaskCompletionRate float64

	// TotalTasksCompleted is the total count of tasks that have completed successfully
	// since the TaskManager started.
	TotalTasksCompleted uint64

	//
	// --- Reliability & Error Metrics -- -
	// These metrics track the success, failure, and retry rates of tasks, which are
	// critical for monitoring system health and diagnosing issues.
	//

	// TotalTasksFailed is the total count of tasks that have failed after all retry
	// attempts have been exhausted.
	TotalTasksFailed uint64

	// TotalTasksRetried is the total number of times any task has been retried due to
	// recoverable errors (e.g., unhealthy resource state). A high value may indicate
	// instability in resources or downstream services.
	TotalTasksRetried uint64

	// SuccessRate is the ratio of successfully completed tasks to the total number of
	// terminal tasks (completed + failed). (Value: 0.0 to 1.0)
	SuccessRate float64

	// FailureRate is the ratio of failed tasks to the total number of terminal tasks
	// (completed + failed). (Value: 0.0 to 1.0)
	FailureRate float64
}

// TaskLifecycleTimestamps holds critical timestamps from a task's journey.
// This data is passed to a MetricsCollector to calculate performance metrics.
type TaskLifecycleTimestamps struct {
	// QueuedAt is the time when the task was first added to a queue.
	QueuedAt time.Time
	// StartedAt is the time when a worker began executing the task.
	StartedAt time.Time
	// FinishedAt is the time when the task execution completed (successfully or not).
	FinishedAt time.Time
}

// MetricsCollector defines the interface for collecting and calculating
// performance and reliability metrics for the TaskManager. Implementations of this
// interface are responsible for processing task lifecycle events and aggregating
// the data into the TaskMetrics struct.
type MetricsCollector interface {
	// RecordArrival is called by the TaskManager each time a new task is queued.
	// This provides the necessary data to calculate the task arrival rate.
	RecordArrival()

	// RecordCompletion is called by the TaskManager whenever a task completes
	// successfully. The collector should use the provided timestamps to update
	// its internal metrics for latency and throughput.
	//
	// The `stamps` parameter contains the timing information for the completed task,
	// allowing the collector to calculate wait and execution times.
	RecordCompletion(stamps TaskLifecycleTimestamps)

	// RecordFailure is called by the TaskManager whenever a task fails permanently
	// (i.e., all retries are exhausted). The collector should update its
	// failure and error rate metrics.
	//
	// The `stamps` parameter provides the timing information for the failed task.
	RecordFailure(stamps TaskLifecycleTimestamps)

	// RecordRetry is called by the TaskManager each time a task is queued for a
	// retry attempt. This allows the collector to track the overall reliability
	// and health of the tasks and their underlying resources.
	RecordRetry()

	// Metrics returns a snapshot of the currently aggregated performance metrics.
	// This method allows users to periodically fetch the latest metrics to
	// monitor the system's health and performance.
	Metrics() TaskMetrics
}

