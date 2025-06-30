// Package tasker provides a robust and flexible task management system
// with dynamic worker scaling, resource pooling, and priority queuing.
// It is designed for concurrent execution of tasks, offering control
// over worker lifecycles, resource allocation, and graceful shutdown.
package tasker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// Tasker defines the core interface for task execution.
// Implementations of Tasker are responsible for running a single task
// with a given resource and returning a result or an error.
type Tasker interface {
	// Execute runs a task with the provided resource.
	// The context can be used for cancellation or passing deadlines to the task.
	// It returns the task's result and any error encountered during execution.
	Execute(ctx context.Context, resource any) (any, error)
}

// TaskManager defines the interface for managing asynchronous and synchronous
// task execution within a pool of workers and resources.
// Generic types R and E represent the Resource type and the Task execution Result type, respectively.
type TaskManager[R any, E any] interface {
	// QueueTask adds a task to the main queue for asynchronous execution.
	// The task will be picked up by an available worker.
	// It returns the result and error of the task once it completes.
	// If the task manager is shutting down, it returns an error immediately.
	QueueTask(task func(context.Context, R) (E, error)) (E, error)

	// QueueTaskWithCallback adds a task to the main queue for asynchronous execution.
	// The callback function will be invoked with the task's result and error once it completes.
	// This method returns immediately and does not block.
	// If the task manager is shutting down, the callback will be invoked with an error.
	QueueTaskWithCallback(task func(context.Context, R) (E, error), callback func(E, error))

	// QueueTaskAsync adds a task to the main queue for asynchronous execution.
	// It returns a channel that will receive the task's result and error once it completes.
	// The caller can then read from this channel to get the result without blocking the submission.
	// If the task manager is shutting down, it returns an error immediately and a closed channel.
	QueueTaskAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)

	// RunTask executes a task immediately, bypassing the main and priority queues.
	// It attempts to acquire a resource from the internal pool. If no resource
	// is immediately available, it temporarily creates a new one for the task.
	// This method is suitable for urgent tasks that should not be delayed by queueing.
	// It returns the result and error of the task.
	// If the task manager is shutting down, it returns an error immediately.
	RunTask(task func(context.Context, R) (E, error)) (E, error)

	// QueueTaskWithPriority adds a high priority task to a dedicated queue.
	// Tasks in the priority queue are processed before tasks in the main queue.
	// It returns the result and error of the task once it completes.
	// If the task manager is shutting down, it returns an error immediately.
	QueueTaskWithPriority(task func(context.Context, R) (E, error)) (E, error)

	// QueueTaskWithPriorityWithCallback adds a high priority task to a dedicated queue for asynchronous execution.
	// The callback function will be invoked with the task's result and error once it completes.
	// This method returns immediately and does not block.
	// If the task manager is shutting down, the callback will be invoked with an error.
	QueueTaskWithPriorityWithCallback(task func(context.Context, R) (E, error), callback func(E, error))

	// QueueTaskWithPriorityAsync adds a high priority task to a dedicated queue for asynchronous execution.
	// It returns a channel that will receive the task's result and error once it completes.
	// The caller can then read from this channel to get the result without blocking the submission.
	// If the task manager is shutting down, it returns an error immediately and a closed channel.
	QueueTaskWithPriorityAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)

	// QueueTaskOnce adds a task to the main queue for asynchronous execution.
	// This task will NOT be re-queued by the task manager's internal retry mechanism
	// if it fails and CheckHealth indicates an unhealthy state. This is suitable
	// for non-idempotent operations where "at-most-once" execution is desired
	// from the task manager's perspective.
	// It returns the result and error of the task once it completes.
	// If the task manager is shutting down, it returns an error immediately.
	QueueTaskOnce(task func(context.Context, R) (E, error)) (E, error)

	// QueueTaskOnceWithCallback adds a task to the main queue for asynchronous execution.
	// This task will NOT be re-queued by the task manager's internal retry mechanism
	// if it fails and CheckHealth indicates an unhealthy state. This is suitable
	// for non-idempotent operations where "at-most-once" execution is desired
	// from the task manager's perspective.
	// The callback function will be invoked with the task's result and error once it completes.
	// This method returns immediately and does not block.
	// If the task manager is shutting down, the callback will be invoked with an error.
	QueueTaskOnceWithCallback(task func(context.Context, R) (E, error), callback func(E, error))

	// QueueTaskOnceAsync adds a task to the main queue for asynchronous execution.
	// This task will NOT be re-queued by the task manager's internal retry mechanism
	// if it fails and CheckHealth indicates an unhealthy state. This is suitable
	// for non-idempotent operations where "at-most-once" execution is desired
	// from the task manager's perspective.
	// It returns a channel that will receive the task's result and error once it completes.
	// The caller can then read from this channel to get the result without blocking the submission.
	// If the task manager is shutting down, it returns an error immediately and a closed channel.
	QueueTaskOnceAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)

	// QueueTaskWithPriorityOnce adds a high priority task to a dedicated queue.
	// This task will NOT be re-queued by the task manager's internal retry mechanism
	// if it fails and CheckHealth indicates an unhealthy state. This is suitable
	// for non-idempotent high-priority operations where "at-most-once" execution
	// is desired from the task manager's perspective.
	// It returns the result and error of the task once it completes.
	// If the task manager is shutting down, it returns an error immediately.
	QueueTaskWithPriorityOnce(task func(context.Context, R) (E, error)) (E, error)

	// QueueTaskWithPriorityOnceWithCallback adds a high priority task to a dedicated queue for asynchronous execution.
	// This task will NOT be re-queued by the task manager's internal retry mechanism
	// if it fails and CheckHealth indicates an unhealthy state. This is suitable
	// for non-idempotent high-priority operations where "at-most-once" execution
	// is desired from the task manager's perspective.
	// The callback function will be invoked with the task's result and error once it completes.
	// This method returns immediately and does not block.
	// If the task manager is shutting down, the callback will be invoked with an error.
	QueueTaskWithPriorityOnceWithCallback(task func(context.Context, R) (E, error), callback func(E, error))

	// QueueTaskWithPriorityOnceAsync adds a high priority task to a dedicated queue for asynchronous execution.
	// This task will NOT be re-queued by the task manager's internal retry mechanism
	// if it fails and CheckHealth indicates an unhealthy state. This is suitable
	// for non-idempotent high-priority operations where "at-most-once" execution
	// is desired from the task manager's perspective.
	// It returns a channel that will receive the task's result and error once it completes.
	// The caller can then read from this channel to get the result without blocking the submission.
	// If the task manager is shutting down, it returns an error immediately and a closed channel.
	QueueTaskWithPriorityOnceAsync(task func(context.Context, R) (E, error)) (<-chan E, <-chan error)

	// Stop gracefully shuts down the task manager.
	// It stops accepting new tasks, waits for all queued tasks to be completed,
	// and then releases all managed resources.
	Stop() error

	// Kill immediately shuts down the task manager.
	// It cancels all running tasks, drops all queued tasks, and releases resources.
	// It does not wait for tasks to complete.
	Kill() error

	// Stats returns current statistics about the task manager's operational state.
	// This includes information on worker counts, queued tasks, and resource availability.
	Stats() TaskStats

	// Metrics returns a snapshot of the currently aggregated performance metrics.
	Metrics() TaskMetrics
}

// TaskStats provides insight into the task manager's current state and performance.
type TaskStats struct {
	BaseWorkers        int32 // Number of permanently active workers.
	ActiveWorkers      int32 // Total number of currently active workers (base + burst).
	BurstWorkers       int32 // Number of dynamically scaled-up workers.
	QueuedTasks        int32 // Number of tasks currently in the main queue.
	PriorityTasks      int32 // Number of tasks currently in the priority queue.
	AvailableResources int32 // Number of resources currently available in the resource pool.
}

// Task represents a unit of work to be executed by a worker.
// It encapsulates the task function, channels for returning results and errors,
// and a retry counter for handling transient failures.
type Task[R any, E any] struct {
	// The actual function representing the task's logic.
	// It now accepts a context.Context for cancellation and a resource of type R.
	run      func(context.Context, R) (E, error)
	result   chan E     // Channel to send the task's successful result.
	err      chan error // Channel to send any error encountered during task execution.
	retries  int        // Current number of retries attempted for this task.
	queuedAt time.Time  // Timestamp when the task was queued.
}

// Config holds configuration parameters for creating a new Runner instance.
// These parameters control worker behavior, resource management, and scaling policies.
// R is the resource type (e.g., *sql.DB, *http.Client).
type Config[R any] struct {
	// OnCreate is a function that creates and initializes a new resource of type R.
	// This function is called when a worker starts or when RunTask needs a temporary resource.
	// It must return a new resource or an error if resource creation fails.
	OnCreate func() (R, error)

	// OnDestroy is a function that performs cleanup or deallocation for a resource of type R.
	// This function is called when a worker shuts down or a temporary resource from RunTask is no longer needed.
	// It should handle any necessary resource finalization and return an error if cleanup fails.
	OnDestroy func(R) error

	// WorkerCount specifies the initial and minimum number of base workers.
	// This many workers will always be running, ready to pick up tasks. Must be greater than 0.
	WorkerCount int

	// Ctx is the parent context for the Runner. All worker contexts will be derived from this.
	// Cancelling this context will initiate a graceful shutdown of the Runner.
	Ctx context.Context

	// CheckHealth is an optional function that determines if a given error
	// indicates an "unhealthy" state for a worker or resource.
	// If a task returns an error and CheckHealth returns false, the worker processing
	// that task might be considered faulty and potentially shut down, and the task re-queued.
	// If nil, all errors are considered healthy (i.e., not cause for worker termination).
	CheckHealth func(error) bool

	// MaxWorkerCount is the maximum total number of workers (base + burst) allowed.
	// If 0, it defaults to WorkerCount * 2.
	MaxWorkerCount int

	// BurstInterval is the frequency at which the burst manager checks queue sizes
	// and adjusts the number of burst workers.
	// If 0, a default of 100 milliseconds is used.
	BurstInterval time.Duration

	// MaxRetries specifies the maximum number of times a task will be re-queued
	// if it fails and CheckHealth indicates an unhealthy state.
	// If 0, a default of 3 retries is used.
	MaxRetries int

	// ResourcePoolSize defines the number of resources to pre-allocate and maintain
	// in the internal pool for `RunTask` operations.
	// If 0, a default of `WorkerCount` is used.
	ResourcePoolSize int

	// Optional: Custom logger interface for internal messages
	Logger Logger

	// Optional: Custom metrics collector
	Collector MetricsCollector

	// Deprecated: The new rate-based scaling automatically adjusts worker
	// counts based on real-time workload, making this field obsolete
	BurstTaskThreshold int

	// Deprecated: The new rate-based scaling automatically adjusts worker
	// counts based on real-time workload, making this field obsolete
	BurstWorkerCount int
}

// Manager implements the TaskManager interface, providing a highly concurrent
// and scalable task execution environment. It manages a pool of workers,
// handles task queuing (main and priority), supports dynamic worker scaling (bursting),
// and includes a resource pool for immediate task execution.
// R is the type of resource used by tasks (e.g., *sql.DB, *http.Client).
// E is the expected result type of the tasks.
type Manager[R any, E any] struct {
	// shutdownState represents the current lifecycle state of the Runner
	shutdownState struct {
		// Constants for shutdown state management - encapsulated within Runner
		running  int32
		stopping int32
		killed   int32
		current  atomic.Int32 // Current state
	}

	// Configuration
	onCreate    func() (R, error) // Function to create a new resource.
	onDestroy   func(R) error     // Function to destroy a resource.
	checkHealth func(error) bool  // Function to check if an error signifies an unhealthy worker/resource.
	maxRetries  int               // Maximum number of retries for a task on unhealthy errors.

	// Worker management
	baseWorkerCount int32         // The fixed number of base workers.
	activeWorkers   atomic.Int32  // Atomically counts all currently active workers (base + burst).
	burstWorkers    atomic.Int32  // Atomically counts currently active burst workers.
	maxWorkerCount  int           // Maximum total workers allowed.
	burstInterval   time.Duration // Interval for burst manager to check queues.

	// Task queues (using channels for efficient communication)
	mainQueue     chan *Task[R, E] // Main queue for standard tasks.
	priorityQueue chan *Task[R, E] // Priority queue for high-priority tasks.

	// Resource pool for immediate execution (`RunTask`)
	resourcePool chan R // Buffered channel holding pre-allocated resources of type R.
	poolSize     int    // Capacity of the resource pool.

	// Synchronization
	ctx    context.Context    // The root context for the Runner, used for graceful shutdown.
	cancel context.CancelFunc // Function to cancel the root context.
	wg     sync.WaitGroup     // Used to wait for all goroutines (workers, managers) to complete.

	// Burst management
	burstMu     sync.Mutex           // Protects access to burst-related fields like burstCtxs.
	burstCtxs   []context.CancelFunc // Stores cancel functions for burst workers to stop them gracefully.
	burstTicker *time.Ticker         // Ticker for the burst manager's periodic checks.
	burstStop   chan struct{}        // Signal channel to stop the burst manager goroutine.

	// Statistics - using individual atomic values for better performance
	queuedTasks        atomic.Int32 // Number of tasks currently in the main queue.
	priorityTasks      atomic.Int32 // Number of tasks currently in the priority queue.
	availableResources atomic.Int32 // Number of resources currently available in the resource pool.

	// Internal utilities - encapsulated within Runner
	utils struct {
		taskResultTimeout time.Duration // Timeout for sending task results back to callers
	}

	// Logger for internal messages
	logger Logger

	// Metrics collector
	collector MetricsCollector
}

// Deprecated: Use NewTaskManager instead
func NewRunner[R any, E any](config Config[R]) (TaskManager[R, E], error) {
	return NewTaskManager[R, E](config)
}

// NewTaskManager creates and initializes a new Runner instance based on the provided configuration.
// It sets up the resource pool, starts the base workers, and kicks off the burst manager.
// Returns a TaskManager interface and an error if initialization fails (e.g., invalid config, resource creation error).
func NewTaskManager[R any, E any](config Config[R]) (TaskManager[R, E], error) {
	if config.WorkerCount <= 0 {
		return nil, errors.New("worker count must be positive")
	}
	if config.OnCreate == nil {
		return nil, errors.New("onCreate function is required")
	}
	if config.OnDestroy == nil {
		return nil, errors.New("onDestroy function is required")
	}
	if config.MaxWorkerCount <= 0 {
		config.MaxWorkerCount = config.WorkerCount * 2 // Default value
	}
	if config.CheckHealth == nil {
		config.CheckHealth = func(error) bool { return true }
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.ResourcePoolSize <= 0 {
		config.ResourcePoolSize = config.WorkerCount // Sensible default
	}
	if config.BurstInterval <= 0 {
		config.BurstInterval = 100 * time.Millisecond
	}

	if config.Logger == nil {
		config.Logger = newNoOpLogger()
	}

	if config.Collector == nil {
		config.Collector = NewCollector()
	}

	ctx, cancel := context.WithCancel(config.Ctx)

	runner := &Manager[R, E]{
		onCreate:        config.OnCreate,
		onDestroy:       config.OnDestroy,
		baseWorkerCount: int32(config.WorkerCount),
		checkHealth:     config.CheckHealth,
		maxRetries:      config.MaxRetries,
		maxWorkerCount:  config.MaxWorkerCount,
		burstInterval:   config.BurstInterval,
		mainQueue:       make(chan *Task[R, E], config.WorkerCount*2),
		priorityQueue:   make(chan *Task[R, E], config.WorkerCount),
		resourcePool:    make(chan R, config.ResourcePoolSize),
		poolSize:        config.ResourcePoolSize,
		ctx:             ctx,
		cancel:          cancel,
		burstStop:       make(chan struct{}),
		burstCtxs:       make([]context.CancelFunc, 0),
		logger:          config.Logger,
		collector:       config.Collector,
	}
	runner.logger.Infof("TaskManager created with %d base workers.", runner.baseWorkerCount)

	// Initialize encapsulated state constants
	runner.shutdownState.running = 0
	runner.shutdownState.stopping = 1
	runner.shutdownState.killed = 2
	runner.shutdownState.current.Store(runner.shutdownState.running)

	// Initialize internal utilities
	runner.utils.taskResultTimeout = 100 * time.Millisecond

	if err := runner.initializeResourcePool(); err != nil {
		runner.logger.Errorf("Failed to initialize resource pool: %v", err)
		cancel()
		return nil, fmt.Errorf("failed to initialize resource pool: %w", err)
	}

	for range config.WorkerCount {
		runner.startWorker(ctx, false)
	}

	runner.startBurstManager()

	return runner, nil
}

// isRunning checks if the runner is in running state
func (r *Manager[R, E]) isRunning() bool {
	return r.shutdownState.current.Load() == r.shutdownState.running
}

// isStopping checks if the runner is in stopping state
func (r *Manager[R, E]) isStopping() bool {
	return r.shutdownState.current.Load() == r.shutdownState.stopping
}

// transitionState atomically transitions from one state to another
func (r *Manager[R, E]) transitionState(from, to int32) bool {
	return r.shutdownState.current.CompareAndSwap(from, to)
}

// minInt returns the minimum of two integers - encapsulated utility
func (r *Manager[R, E]) minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the maximum of two integers - encapsulated utility
func (r *Manager[R, E]) maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// initializeResourcePool pre-allocates and populates the resource pool.
func (r *Manager[R, E]) initializeResourcePool() error {
	r.logger.Debugf("Initializing resource pool with size %d.", r.poolSize)
	for range r.poolSize {
		resource, err := r.onCreate()
		if err != nil {
			r.drainResourcePool()
			return fmt.Errorf("failed to create resource for pool: %w", err)
		}
		select {
		case r.resourcePool <- resource:
			r.availableResources.Add(1)
		case <-r.ctx.Done():
			_ = r.onDestroy(resource)
			r.drainResourcePool()
			return r.ctx.Err()
		}
	}
	r.logger.Infof("Resource pool initialized successfully.")
	return nil
}

// drainResourcePool cleans up all resources currently held within the resource pool.
func (r *Manager[R, E]) drainResourcePool() {
	r.logger.Debugf("Draining resource pool.")
	close(r.resourcePool)
	for resource := range r.resourcePool {
		if err := r.onDestroy(resource); err != nil {
			r.logger.Warnf("Error destroying resource: %v", err)
		}
		r.availableResources.Add(-1)
	}
	r.logger.Infof("Resource pool drained.")
}

// startWorker creates and launches a new worker goroutine.
func (r *Manager[R, E]) startWorker(ctx context.Context, isBurst bool) {
	r.wg.Add(1)
	go r.worker(ctx, isBurst)
	r.logger.Debugf("Started a new worker (isBurst: %t).", isBurst)
}

// worker is the main loop for a worker goroutine.
// It processes tasks until the context is canceled, at which point it either
// drains the queues (for graceful stop) or exits immediately (for kill).
func (r *Manager[R, E]) worker(ctx context.Context, isBurst bool) {
	defer r.wg.Done()
	r.logger.Debugf("Worker started.")

	resource, err := r.onCreate()
	if err != nil {
		r.logger.Errorf("Worker failed to create resource: %v", err)
		return // Cannot start worker without a resource
	}
	defer func() {
		if err := r.onDestroy(resource); err != nil {
			r.logger.Warnf("Error destroying resource on worker exit: %v", err)
		}
	}()

	if isBurst {
		r.burstWorkers.Add(1)
		defer r.burstWorkers.Add(-1)
	}
	r.activeWorkers.Add(1)
	defer r.activeWorkers.Add(-1)

	for {
		select {
		case <-ctx.Done():
			r.logger.Debugf("Worker received context cancellation signal.")
			// Context is canceled. Check if this is a graceful stop.
			if r.isStopping() {
				r.logger.Infof("Worker entering drain mode.")
				// Graceful stop: drain all queues before exiting.
				for r.processQueues(resource) {
				}
				r.logger.Infof("Worker finished draining queues.")
			}
			// For kill, or after draining is complete, exit immediately.
			r.logger.Debugf("Worker exiting.")
			return
		case task := <-r.priorityQueue:
			r.priorityTasks.Add(-1)
			r.logger.Debugf("Worker picked up a priority task.")
			if !r.executeTask(task, resource) {
				r.logger.Warnf("Unhealthy worker is exiting.")
				return // Unhealthy worker exits
			}
		case task := <-r.mainQueue:
			r.queuedTasks.Add(-1)
			r.logger.Debugf("Worker picked up a main queue task.")
			if !r.executeTask(task, resource) {
				r.logger.Warnf("Unhealthy worker is exiting.")
				return // Unhealthy worker exits
			}
		}
	}
}

// processQueues tries to process one task from either queue. Used for draining.
// Returns true if a task was processed, false if queues were empty.
func (r *Manager[R, E]) processQueues(resource R) bool {
	select {
	case task := <-r.priorityQueue:
		r.priorityTasks.Add(-1)
		r.logger.Debugf("Draining: processing priority task.")
		r.executeTask(task, resource)
		return true
	case task := <-r.mainQueue:
		r.queuedTasks.Add(-1)
		r.logger.Debugf("Draining: processing main queue task.")
		r.executeTask(task, resource)
		return true
	default:
		return false // Both queues are empty
	}
}

// executeTask runs a single task and handles its retry logic.
func (r *Manager[R, E]) executeTask(task *Task[R, E], resource R) bool {
	startedAt := time.Now()
	r.logger.Debugf("Executing task (retries: %d).", task.retries)
	// Create a context for the task, derived from the manager's root context.
	// This allows individual tasks to be cancelled via the manager's Stop/Kill methods.
	taskCtx, cancel := context.WithCancel(r.ctx)
	defer cancel()

	result, err := task.run(taskCtx, resource)
	finishedAt := time.Now()

	timestamps := TaskLifecycleTimestamps{
		QueuedAt:   task.queuedAt,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}

	if err != nil {
		r.logger.Warnf("Task execution failed: %v", err)
		if !r.checkHealth(err) {
			r.logger.Warnf("Unhealthy state detected after task failure.")
			if task.retries < r.maxRetries {
				task.retries++
				r.logger.Infof("Re-queuing task (retry %d/%d).", task.retries, r.maxRetries)
				r.collector.RecordRetry()
				select {
				case r.priorityQueue <- task:
					r.priorityTasks.Add(1)
				default:
					r.logger.Errorf("Priority queue full, task requeue failed: %v", err)
					r.collector.RecordFailure(timestamps)
					r.sendTaskResult(task, result, fmt.Errorf("priority queue full, task requeue failed: %w", err))
				}
			} else {
				r.logger.Errorf("Max retries exceeded for task: %v", err)
				r.collector.RecordFailure(timestamps)
				r.sendTaskResult(task, result, fmt.Errorf("max retries exceeded: %w", err))
			}
			return false // Worker is unhealthy
		}
		r.collector.RecordFailure(timestamps)
	} else {
		r.logger.Debugf("Task executed successfully.")
		r.collector.RecordCompletion(timestamps)
	}

	r.sendTaskResult(task, result, err)
	return true // Worker is healthy
}

// sendTaskResult safely sends the task result and error back to the caller.
func (r *Manager[R, E]) sendTaskResult(task *Task[R, E], result E, err error) {
	r.logger.Debugf("Sending task result.")
	// Non-blocking send with a short timeout to prevent being stuck forever
	ctx, cancel := context.WithTimeout(context.Background(), r.utils.taskResultTimeout)
	defer cancel()

	select {
	case task.err <- err:
	case <-ctx.Done():
		r.logger.Warnf("Timeout sending task error result.")
	}

	select {
	case task.result <- result:
	case <-ctx.Done():
		r.logger.Warnf("Timeout sending task success result.")
	}
}

// QueueTask adds a task to the main queue for asynchronous processing.
func (r *Manager[R, E]) QueueTask(taskFunc func(context.Context, R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		r.logger.Warnf("Attempted to queue task while shutting down.")
		return zero, errors.New("task manager is shutting down")
	}

	task := &Task[R, E]{
		run:      taskFunc,
		result:   make(chan E, 1),
		err:      make(chan error, 1),
		queuedAt: time.Now(),
	}

	r.collector.RecordArrival()
	select {
	case r.mainQueue <- task:
		r.queuedTasks.Add(1)
		r.logger.Debugf("Task added to the main queue.")
		return <-task.result, <-task.err
	case <-r.ctx.Done():
		var zero E
		r.logger.Warnf("Attempted to queue task while shutting down (context done).")
		return zero, errors.New("task manager is shutting down")
	}
}

// QueueTaskWithCallback adds a task to the main queue for asynchronous execution.
// The callback function will be invoked with the task's result and error once it completes.
// This method returns immediately and does not block.
// If the task manager is shutting down, the callback will be invoked with an error.
func (r *Manager[R, E]) QueueTaskWithCallback(taskFunc func(context.Context, R) (E, error), callback func(E, error)) {
	go func() {
		var zero E
		if !r.isRunning() {
			callback(zero, errors.New("task manager is shutting down"))
			return
		}
		result, err := r.QueueTask(taskFunc)
		callback(result, err)
	}()
}

// QueueTaskAsync adds a task to the main queue for asynchronous execution.
// It returns a channel that will receive the task's result and error once it completes.
// The caller can then read from this channel to get the result without blocking the submission.
// If the task manager is shutting down, it returns an error immediately and a closed channel.

func (r *Manager[R, E]) QueueTaskAsync(taskFunc func(context.Context, R) (E, error)) (<-chan E, <-chan error) {
	resultChan := make(chan E, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		result, err := r.QueueTask(taskFunc)
		resultChan <- result
		errorChan <- err
	}()

	return resultChan, errorChan
}

// QueTaskOnce adds a task to the main queue that will NOT be retried
// by the task manager's internal retry mechanism if it fails and
// CheckHealth indicates an unhealthy state. This is suitable for
// non-idempotent operations.
// It returns the result and error of the task once it completes.
// If the task manager is shutting down, it returns an error immediately.
func (r *Manager[R, E]) QueueTaskOnce(taskFunc func(context.Context, R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		r.logger.Warnf("Attempted to queue a 'once' task while shutting down.")
		return zero, errors.New("task manager is shutting down")
	}

	task := &Task[R, E]{
		run:      taskFunc,
		result:   make(chan E, 1),
		err:      make(chan error, 1),
		retries:  r.maxRetries, // Set retries to max so it won't retry on health failure
		queuedAt: time.Now(),
	}

	r.collector.RecordArrival()
	select {
	case r.mainQueue <- task:
		r.queuedTasks.Add(1)
		r.logger.Debugf("Task (once) added to the main queue.")
		return <-task.result, <-task.err
	case <-r.ctx.Done():
		var zero E
		r.logger.Warnf("Attempted to queue a 'once' task while shutting down (context done).")
		return zero, errors.New("task manager is shutting down")
	}
}

// QueueTaskOnceWithCallback adds a task to the main queue for asynchronous execution.
// This task will NOT be re-queued by the task manager's internal retry mechanism
// if it fails and CheckHealth indicates an unhealthy state. This is suitable
// for non-idempotent operations where "at-most-once" execution is desired
// from the task manager's perspective.
// The callback function will be invoked with the task's result and error once it completes.
// This method returns immediately and does not block.
// If the task manager is shutting down, the callback will be invoked with an error.
func (r *Manager[R, E]) QueueTaskOnceWithCallback(taskFunc func(context.Context, R) (E, error), callback func(E, error)) {
	go func() {
		var zero E
		if !r.isRunning() {
			callback(zero, errors.New("task manager is shutting down"))
			return
		}
		result, err := r.QueueTaskOnce(taskFunc)
		callback(result, err)
	}()
}

// QueueTaskOnceAsync adds a task to the main queue for asynchronous execution.
// This task will NOT be re-queued by the task manager's internal retry mechanism
// if it fails and CheckHealth indicates an unhealthy state. This is suitable
// for non-idempotent operations where "at-most-once" execution is desired
// from the task manager's perspective.
// It returns a channel that will receive the task's result and error once it completes.
// The caller can then read from this channel to get the result without blocking the submission.
// If the task manager is shutting down, it returns an error immediately and a closed channel.
func (r *Manager[R, E]) QueueTaskOnceAsync(taskFunc func(context.Context, R) (E, error)) (<-chan E, <-chan error) {
	resultChan := make(chan E, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		result, err := r.QueueTaskOnce(taskFunc)
		resultChan <- result
		errorChan <- err
	}()

	return resultChan, errorChan
}

// QueueTaskWithPriority adds a high-priority task to the priority queue.
func (r *Manager[R, E]) QueueTaskWithPriority(taskFunc func(context.Context, R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		r.logger.Warnf("Attempted to queue priority task while shutting down.")
		return zero, errors.New("task manager is shutting down")
	}

	task := &Task[R, E]{
		run:      taskFunc,
		result:   make(chan E, 1),
		err:      make(chan error, 1),
		queuedAt: time.Now(),
	}

	r.collector.RecordArrival()
	select {
	case r.priorityQueue <- task:
		r.priorityTasks.Add(1)
		r.logger.Debugf("Task added to the priority queue.")
		return <-task.result, <-task.err
	case <-r.ctx.Done():
		var zero E
		r.logger.Warnf("Attempted to queue priority task while shutting down (context done).")
		return zero, errors.New("task manager is shutting down")
	}
}

// QueueTaskWithPriorityWithCallback adds a high priority task to a dedicated queue for asynchronous execution.
// The callback function will be invoked with the task's result and error once it completes.
// This method returns immediately and does not block.
// If the task manager is shutting down, the callback will be invoked with an error.
func (r *Manager[R, E]) QueueTaskWithPriorityWithCallback(taskFunc func(context.Context, R) (E, error), callback func(E, error)) {
	go func() {
		var zero E
		if !r.isRunning() {
			callback(zero, errors.New("task manager is shutting down"))
			return
		}
		result, err := r.QueueTaskWithPriority(taskFunc)
		callback(result, err)
	}()
}

// QueueTaskWithPriorityAsync adds a high priority task to a dedicated queue for asynchronous execution.
// It returns a channel that will receive the task's result and error once it completes.
// The caller can then read from this channel to get the result without blocking the submission.
// If the task manager is shutting down, it returns an error immediately and a closed channel.
func (r *Manager[R, E]) QueueTaskWithPriorityAsync(taskFunc func(context.Context, R) (E, error)) (<-chan E, <-chan error) {
	resultChan := make(chan E, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		result, err := r.QueueTaskWithPriority(taskFunc)
		resultChan <- result
		errorChan <- err
	}()

	return resultChan, errorChan
}

// QueueTaskWithPriority adds a high-priority task to the priority queue, but
// the task will not be retried.
func (r *Manager[R, E]) QueueTaskWithPriorityOnce(taskFunc func(context.Context, R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		r.logger.Warnf("Attempted to queue a 'once' priority task while shutting down.")
		return zero, errors.New("task manager is shutting down")
	}

	task := &Task[R, E]{
		run:      taskFunc,
		result:   make(chan E, 1),
		err:      make(chan error, 1),
		retries:  r.maxRetries, // Set retries to max so it won't retry on health failure
		queuedAt: time.Now(),
	}

	r.collector.RecordArrival()
	select {
	case r.priorityQueue <- task:
		r.priorityTasks.Add(1)
		r.logger.Debugf("Task (once) added to the priority queue.")
		return <-task.result, <-task.err
	case <-r.ctx.Done():
		var zero E
		r.logger.Warnf("Attempted to queue a 'once' priority task while shutting down (context done).")
		return zero, errors.New("task manager is shutting down")
	}
}

// QueueTaskWithPriorityOnceWithCallback adds a high priority task to a dedicated queue for asynchronous execution.
// This task will NOT be re-queued by the task manager's internal retry mechanism
// if it fails and CheckHealth indicates an unhealthy state. This is suitable
// for non-idempotent high-priority operations where "at-most-once" execution
// is desired from the task manager's perspective.
// The callback function will be invoked with the task's result and error once it completes.
// This method returns immediately and does not block.
// If the task manager is shutting down, the callback will be invoked with an error.
func (r *Manager[R, E]) QueueTaskWithPriorityOnceWithCallback(taskFunc func(context.Context, R) (E, error), callback func(E, error)) {
	go func() {
		var zero E
		if !r.isRunning() {
			callback(zero, errors.New("task manager is shutting down"))
			return
		}
		result, err := r.QueueTaskWithPriorityOnce(taskFunc)
		callback(result, err)
	}()
}

// QueueTaskWithPriorityOnceAsync adds a high priority task to a dedicated queue for asynchronous execution.
// This task will NOT be re-queued by the task manager's internal retry mechanism
// if it fails and CheckHealth indicates an unhealthy state. This is suitable
// for non-idempotent high-priority operations where "at-most-once" execution
// is desired from the task manager's perspective.
// It returns a channel that will receive the task's result and error once it completes.
// The caller can then read from this channel to get the result without blocking the submission.
// If the task manager is shutting down, it returns an error immediately and a closed channel.
func (r *Manager[R, E]) QueueTaskWithPriorityOnceAsync(taskFunc func(context.Context, R) (E, error)) (<-chan E, <-chan error) {
	resultChan := make(chan E, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		result, err := r.QueueTaskWithPriorityOnce(taskFunc)
		resultChan <- result
		errorChan <- err
	}()

	return resultChan, errorChan
}

// RunTask executes a task immediately using a resource from the pool.
func (r *Manager[R, E]) RunTask(taskFunc func(context.Context, R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		r.logger.Warnf("Attempted to run task while shutting down.")
		return zero, errors.New("task manager is shutting down")
	}

	var resource R
	var err error
	fromPool := false

	// Record metric for arrival
	r.collector.RecordArrival()

	queuedAt := time.Now()

	select {
	case resource = <-r.resourcePool:
		fromPool = true
		r.availableResources.Add(-1)
		r.logger.Debugf("Acquired resource from pool for RunTask.")
	case <-r.ctx.Done():
		var zero E
		r.logger.Warnf("Attempted to run task while shutting down (context done).")
		return zero, errors.New("task manager is shutting down")
	default:
		r.logger.Debugf("Creating temporary resource for RunTask.")
		resource, err = r.onCreate()
		if err != nil {
			var zero E
			r.logger.Errorf("Failed to create temporary resource for RunTask: %v", err)
			return zero, fmt.Errorf("failed to create temporary resource: %w", err)
		}
	}

	startedAt := time.Now()

	defer func() {
		if fromPool {
			if r.isRunning() {
				select {
				case r.resourcePool <- resource:
					r.availableResources.Add(1)
					r.logger.Debugf("Returned resource to pool after RunTask.")
				default:
					r.logger.Warnf("Could not return resource to full pool; destroying it.")
					_ = r.onDestroy(resource) // Pool full, destroy
				}
			} else {
				r.logger.Debugf("Destroying pooled resource on shutdown.")
				_ = r.onDestroy(resource) // Shutting down, destroy
			}
		} else {
			r.logger.Debugf("Destroying temporary resource after RunTask.")
			_ = r.onDestroy(resource) // Always destroy temporary
		}
	}()

	r.logger.Debugf("Executing RunTask.")
	// For RunTask, we use the manager's root context directly.
	result, err := taskFunc(r.ctx, resource)
	finishedAt := time.Now()

	timestamps := TaskLifecycleTimestamps{
		QueuedAt:   queuedAt,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}

	if err != nil {
		r.collector.RecordFailure(timestamps)
	} else {
		r.collector.RecordCompletion(timestamps)
	}

	return result, err
}

// startBurstManager starts the dynamic worker scaling manager.
func (r *Manager[R, E]) startBurstManager() {
	if r.burstInterval <= 0 {
		r.logger.Infof("Burst scaling is disabled.")
		return
	}

	r.burstTicker = time.NewTicker(r.burstInterval)
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer r.burstTicker.Stop()
		r.logger.Infof("Burst manager started.")
		for {
			select {
			case <-r.ctx.Done():
				r.logger.Infof("Burst manager stopping due to context cancellation.")
				return
			case <-r.burstStop:
				r.logger.Infof("Burst manager stopping.")
				return
			case <-r.burstTicker.C:
				r.manageBurst()
			}
		}
	}()
}

// manageBurst handles the logic for scaling workers up or down.
func (r *Manager[R, E]) manageBurst() {
	r.burstMu.Lock()
	defer r.burstMu.Unlock()

	metrics := r.collector.Metrics()
	stats := r.Stats()

	if stats.ActiveWorkers == 0 {
		return // Avoid division by zero
	}

	// Calculate throughput per worker
	throughputPerWorker := metrics.TaskCompletionRate / float64(stats.ActiveWorkers)

	if throughputPerWorker == 0 && metrics.TaskArrivalRate > 0 {
		// If there is no throughput but tasks are arriving, we need to scale up.
		// Add a single worker to get the process started.
		if stats.ActiveWorkers < int32(r.maxWorkerCount) {
			r.logger.Infof("Scaling up: adding a burst worker due to zero throughput.")
			burstCtx, cancel := context.WithCancel(r.ctx)
			r.burstCtxs = append(r.burstCtxs, cancel)
			r.startWorker(burstCtx, true)
		}
		return
	}

	if metrics.TaskArrivalRate > metrics.TaskCompletionRate {
		// System is falling behind
		throughputDeficit := metrics.TaskArrivalRate - metrics.TaskCompletionRate
		workersToAdd := int(math.Ceil(throughputDeficit / throughputPerWorker))

		canAdd := r.maxWorkerCount - int(stats.ActiveWorkers)
		toAdd := r.minInt(workersToAdd, canAdd)

		if toAdd > 0 {
			r.logger.Infof("Scaling up: adding %d burst workers to meet demand.", toAdd)
			for range toAdd {
				burstCtx, cancel := context.WithCancel(r.ctx)
				r.burstCtxs = append(r.burstCtxs, cancel)
				r.startWorker(burstCtx, true)
			}
		}
	} else if metrics.TaskCompletionRate > metrics.TaskArrivalRate {
		// System is over-provisioned
		throughputSurplus := metrics.TaskCompletionRate - metrics.TaskArrivalRate
		workersToRemove := int(math.Floor(throughputSurplus / throughputPerWorker))

		toRemove := r.minInt(workersToRemove, int(stats.BurstWorkers))

		if toRemove > 0 {
			r.logger.Infof("Scaling down: removing %d burst workers due to low demand.", toRemove)
			for i := 0; i < toRemove && len(r.burstCtxs) > 0; i++ {
				lastIdx := len(r.burstCtxs) - 1
				cancel := r.burstCtxs[lastIdx]
				cancel()
				r.burstCtxs = r.burstCtxs[:lastIdx]
			}
		}
	}
}

// Stats returns a snapshot of the runner's current operational state.
func (r *Manager[R, E]) Stats() TaskStats {
	return TaskStats{
		BaseWorkers:        r.baseWorkerCount,
		ActiveWorkers:      r.activeWorkers.Load(),
		BurstWorkers:       r.burstWorkers.Load(),
		QueuedTasks:        r.queuedTasks.Load(),
		PriorityTasks:      r.priorityTasks.Load(),
		AvailableResources: r.availableResources.Load(),
	}
}

// Metrics returns a snapshot of the currently aggregated performance metrics.
func (r *Manager[R, E]) Metrics() TaskMetrics {
	return r.collector.Metrics()
}

// Stop gracefully shuts down the task manager by draining the task queues.
func (r *Manager[R, E]) Stop() error {
	r.logger.Infof("Stop command received. Initiating graceful shutdown.")
	// Set state to 'stopping' to prevent new tasks and guide worker exit.
	if !r.transitionState(r.shutdownState.running, r.shutdownState.stopping) {
		r.logger.Warnf("Stop command ignored: task manager already stopping or killed.")
		return errors.New("task manager already stopping or killed")
	}

	// Stop the burst manager and cancel burst workers.
	r.shutdownBursting()

	// Cancel the main context to signal base workers to start draining.
	r.cancel()

	// Wait for all goroutines to finish.
	r.logger.Infof("Waiting for all workers to finish...")
	r.wg.Wait()
	r.logger.Infof("All workers have finished.")

	// Clean up the resource pool.
	r.drainResourcePool()

	r.logger.Infof("Task manager stopped gracefully.")
	return nil
}

// Kill immediately terminates the task manager without draining queues.
func (r *Manager[R, E]) Kill() error {
	r.logger.Warnf("Kill command received. Initiating immediate shutdown.")
	// Set state to 'killed' to prevent new tasks and guide worker exit.
	if !r.transitionState(r.shutdownState.running, r.shutdownState.killed) {
		// If it's already 'stopping', we can escalate to 'killed'.
		if !r.transitionState(r.shutdownState.stopping, r.shutdownState.killed) {
			r.logger.Warnf("Kill command ignored: task manager already killed.")
			return errors.New("task manager already killed")
		}
	}

	// Stop the burst manager and cancel burst workers.
	r.shutdownBursting()

	// Cancel the main context to signal all workers to exit immediately.
	r.cancel()

	// Wait for all goroutines to finish.
	r.logger.Infof("Waiting for all workers to terminate...")
	r.wg.Wait()
	r.logger.Infof("All workers have terminated.")

	// Clean up the resource pool.
	r.drainResourcePool()

	r.logger.Warnf("Task manager killed.")
	return nil
}

// shutdownBursting stops the burst manager and cancels all burst workers.
func (r *Manager[R, E]) shutdownBursting() {
	r.logger.Debugf("Shutting down burst manager.")
	select {
	case r.burstStop <- struct{}{}:
	default:
	}

	r.burstMu.Lock()
	defer r.burstMu.Unlock()
	if len(r.burstCtxs) > 0 {
		r.logger.Infof("Cancelling %d burst workers.", len(r.burstCtxs))
		for _, cancel := range r.burstCtxs {
			cancel()
		}
		r.burstCtxs = nil
	}
}
