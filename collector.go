// Package tasker provides a robust and flexible task management system.
package tasker

import (
	"math"
	"sort"
	"sync"
	"time"
)

// collector is a thread-safe implementation of the MetricsCollector interface.
// It aggregates task lifecycle data to produce comprehensive performance metrics.
type collector struct {
	mu                  sync.RWMutex
	startTime           time.Time
	totalTasksArrived   uint64
	totalTasksCompleted uint64
	totalTasksFailed    uint64
	totalTasksRetried   uint64
	totalExecutionTime  time.Duration
	totalWaitTime       time.Duration
	minExecutionTime    time.Duration
	maxExecutionTime    time.Duration

	// executionTimes holds individual execution durations for percentile calculations.
	executionTimes []time.Duration
}

// NewCollector creates and initializes a new collector instance.
func NewCollector() MetricsCollector {
	return &collector{
		startTime:        time.Now(),
		minExecutionTime: time.Duration(math.MaxInt64),
	}
}

// RecordArrival is called by the TaskManager each time a new task is queued.
func (c *collector) RecordArrival() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalTasksArrived++
}

// RecordCompletion is called by the TaskManager whenever a task completes successfully.
func (c *collector) RecordCompletion(stamps TaskLifecycleTimestamps) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalTasksCompleted++

	executionTime := stamps.FinishedAt.Sub(stamps.StartedAt)
	waitTime := stamps.StartedAt.Sub(stamps.QueuedAt)

	c.totalExecutionTime += executionTime
	c.totalWaitTime += waitTime

	if executionTime < c.minExecutionTime {
		c.minExecutionTime = executionTime
	}
	if executionTime > c.maxExecutionTime {
		c.maxExecutionTime = executionTime
	}

	c.executionTimes = append(c.executionTimes, executionTime)
}

// RecordFailure is called by the TaskManager whenever a task fails permanently.
func (c *collector) RecordFailure(stamps TaskLifecycleTimestamps) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalTasksFailed++
}

// RecordRetry is called by the TaskManager each time a task is queued for a retry.
func (c *collector) RecordRetry() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalTasksRetried++
}

// Metrics returns a snapshot of the currently aggregated performance metrics.
func (c *collector) Metrics() TaskMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := TaskMetrics{
		TotalTasksCompleted: c.totalTasksCompleted,
		TotalTasksFailed:    c.totalTasksFailed,
		TotalTasksRetried:   c.totalTasksRetried,
		MaxExecutionTime:    c.maxExecutionTime,
	}

	elapsedSeconds := time.Since(c.startTime).Seconds()
	if elapsedSeconds > 0 {
		metrics.TaskArrivalRate = float64(c.totalTasksArrived) / elapsedSeconds
		metrics.TaskCompletionRate = float64(c.totalTasksCompleted) / elapsedSeconds
	}

	if c.totalTasksCompleted == 0 {
		return metrics
	}

	metrics.MinExecutionTime = c.minExecutionTime
	metrics.AverageExecutionTime = c.totalExecutionTime / time.Duration(c.totalTasksCompleted)
	metrics.AverageWaitTime = c.totalWaitTime / time.Duration(c.totalTasksCompleted)

	totalTerminalTasks := c.totalTasksCompleted + c.totalTasksFailed
	if totalTerminalTasks > 0 {
		metrics.SuccessRate = float64(c.totalTasksCompleted) / float64(totalTerminalTasks)
		metrics.FailureRate = float64(c.totalTasksFailed) / float64(totalTerminalTasks)
	}

	if len(c.executionTimes) > 0 {
		sortedTimes := make([]time.Duration, len(c.executionTimes))
		copy(sortedTimes, c.executionTimes)
		sort.Slice(sortedTimes, func(i, j int) bool { return sortedTimes[i] < sortedTimes[j] })

		p95Index := int(float64(len(sortedTimes))*0.95) - 1
		if p95Index < 0 {
			p95Index = 0
		}
		metrics.P95ExecutionTime = sortedTimes[p95Index]

		p99Index := int(float64(len(sortedTimes))*0.99) - 1
		if p99Index < 0 {
			p99Index = 0
		}
		metrics.P99ExecutionTime = sortedTimes[p99Index]
	}

	return metrics
}

