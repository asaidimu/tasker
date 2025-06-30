// Package tasker provides a robust and flexible task management system
// with dynamic worker scaling, resource pooling, and priority queuing.
// It is designed for concurrent execution of tasks, offering control
// over worker lifecycles, resource allocation, and graceful shutdown.
package tasker

import (
	"context"
	"errors"
	"fmt"
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
	QueueTask(task func(R) (E, error)) (E, error)

	// RunTask executes a task immediately, bypassing the main and priority queues.
	// It attempts to acquire a resource from the internal pool. If no resource
	// is immediately available, it temporarily creates a new one for the task.
	// This method is suitable for urgent tasks that should not be delayed by queueing.
	// It returns the result and error of the task.
	// If the task manager is shutting down, it returns an error immediately.
	RunTask(task func(R) (E, error)) (E, error)

	// QueueTaskWithPriority adds a high priority task to a dedicated queue.
	// Tasks in the priority queue are processed before tasks in the main queue.
	// It returns the result and error of the task once it completes.
	// If the task manager is shutting down, it returns an error immediately.
	QueueTaskWithPriority(task func(R) (E, error)) (E, error)

	// Stop gracefully shuts down the task manager.
	// It stops accepting new tasks, allows currently executing tasks to complete,
	// and releases all managed resources. It waits for all workers to finish
	// before returning.
	// Returns an error if the shutdown process encounters an issue.
	Stop() error

	// Stats returns current statistics about the task manager's operational state.
	// This includes information on worker counts, queued tasks, and resource availability.
	Stats() TaskStats
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
	// R is the resource type itself (e.g., *DatabaseConnection).
	run     func(R) (E, error)
	result  chan E      // Channel to send the task's successful result.
	err     chan error  // Channel to send any error encountered during task execution.
	retries int         // Current number of retries attempted for this task.
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

	// BurstTaskThreshold is the queue size (mainQueue + priorityQueue) that triggers the creation of burst workers.
	// If total queued tasks exceed this value, new burst workers will be spun up.
	// If 0, burst scaling is disabled.
	BurstTaskThreshold int

	// BurstWorkerCount is the number of burst workers to create when the BurstTaskThreshold is exceeded.
	// If 0, burst scaling is disabled. Defaults to 2.
	BurstWorkerCount int

	// MaxWorkerCount is the maximum total number of workers (base + burst) allowed.
	// If 0, it defaults to WorkerCount + BurstWorkerCount.
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
}

// Runner implements the TaskManager interface, providing a highly concurrent
// and scalable task execution environment. It manages a pool of workers,
// handles task queuing (main and priority), supports dynamic worker scaling (bursting),
// and includes a resource pool for immediate task execution.
// R is the type of resource used by tasks (e.g., *sql.DB, *http.Client).
// E is the expected result type of the tasks.
type Runner[R any, E any] struct {
	// Configuration
	onCreate    func() (R, error) // Function to create a new resource.
	onDestroy   func(R) error     // Function to destroy a resource.
	checkHealth func(error) bool  // Function to check if an error signifies an unhealthy worker/resource.
	maxRetries  int               // Maximum number of retries for a task on unhealthy errors.

	// Worker management
	baseWorkerCount     int32           // The fixed number of base workers.
	activeWorkers       atomic.Int32    // Atomically counts all currently active workers (base + burst).
	burstWorkers        atomic.Int32    // Atomically counts currently active burst workers.
	burstTaskThreshold  int
	burstWorkerCount    int
	maxWorkerCount      int             // Maximum total workers allowed.
	burstInterval       time.Duration   // Interval for burst manager to check queues.

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
	burstMu     sync.Mutex             // Protects access to burst-related fields like burstCtxs.
	burstCtxs   []context.CancelFunc   // Stores cancel functions for burst workers to stop them gracefully.
	burstTicker *time.Ticker           // Ticker for the burst manager's periodic checks.
	burstStop   chan struct{}          // Signal channel to stop the burst manager goroutine.

	// Statistics - using individual atomic values for better performance
	queuedTasks        atomic.Int32 // Number of tasks currently in the main queue.
	priorityTasks      atomic.Int32 // Number of tasks currently in the priority queue.
	availableResources atomic.Int32 // Number of resources currently available in the resource pool.
}

// NewRunner creates and initializes a new Runner instance based on the provided configuration.
// It sets up the resource pool, starts the base workers, and kicks off the burst manager.
// Returns a TaskManager interface and an error if initialization fails (e.g., invalid config, resource creation error).
func NewRunner[R any, E any](config Config[R]) (TaskManager[R, E], error) {
	if config.WorkerCount <= 0 {
		return nil, errors.New("worker count must be positive")
	}
	if config.OnCreate == nil {
		return nil, errors.New("onCreate function is required")
	}
	if config.OnDestroy == nil {
		return nil, errors.New("onDestroy function is required")
	}
	// Default BurstWorkerCount if not positive.
	if config.BurstWorkerCount <= 0 {
		config.BurstWorkerCount = 2 // Default value
	}

	// Default MaxWorkerCount.
	if config.MaxWorkerCount <= 0 {
		config.MaxWorkerCount = config.WorkerCount + config.BurstWorkerCount // Default value
	}
	// Default checkHealth to always return true if not provided.
	if config.CheckHealth == nil {
		config.CheckHealth = func(error) bool { return true }
	}
	// Default MaxRetries if not positive.
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	// Default ResourcePoolSize if not positive.
	if config.ResourcePoolSize <= 0 {
		config.ResourcePoolSize = config.WorkerCount // Sensible default: one resource per base worker.
	}
	// Default BurstInterval if not positive.
	if config.BurstInterval <= 0 {
		config.BurstInterval = 100 * time.Millisecond
	}

	ctx, cancel := context.WithCancel(config.Ctx)

	runner := &Runner[R, E]{
		onCreate:            config.OnCreate,
		onDestroy:           config.OnDestroy,
		baseWorkerCount:     int32(config.WorkerCount),
		checkHealth:         config.CheckHealth,
		maxRetries:          config.MaxRetries,
		burstTaskThreshold:  config.BurstTaskThreshold,
		burstWorkerCount:    config.BurstWorkerCount,
		maxWorkerCount:      config.MaxWorkerCount,
		burstInterval:       config.BurstInterval,

		// Buffered channels to prevent blocking on send when queues are not full
		// and to allow for concurrent producers/consumers.
		mainQueue:     make(chan *Task[R, E], config.WorkerCount*2), // Larger buffer for main queue.
		priorityQueue: make(chan *Task[R, E], config.WorkerCount),   // Smaller buffer for priority queue.
		resourcePool:  make(chan R, config.ResourcePoolSize),        // Resource pool for RunTask, type R directly.
		poolSize:      config.ResourcePoolSize,

		ctx:       ctx,
		cancel:    cancel,
		burstStop: make(chan struct{}),
		burstCtxs: make([]context.CancelFunc, 0), // Initialize an empty slice for burst worker contexts.
	}

	// Initialize atomic counters
	runner.availableResources.Store(0)
	runner.queuedTasks.Store(0)
	runner.priorityTasks.Store(0)

	// Initialize resource pool before starting workers to ensure resources are available.
	if err := runner.initializeResourcePool(); err != nil {
		cancel() // Clean up context if pool initialization fails.
		return nil, fmt.Errorf("failed to initialize resource pool: %w", err)
	}

	// Start base workers. These workers run continuously.
	for range config.WorkerCount {
		runner.startWorker(ctx, false) // `false` indicates a base worker.
	}

	// Start burst management if configured.
	runner.startBurstManager()

	return runner, nil
}

// initializeResourcePool pre-allocates and populates the resource pool
// with resources created by the `onCreate` function.
// This pool is used by `RunTask` for immediate execution.
// Returns an error if any resource creation fails.
func (r *Runner[R, E]) initializeResourcePool() error {
	for range r.poolSize {
		resource, err := r.onCreate() // Now onCreate returns R directly
		if err != nil {
			// If resource creation fails, drain any resources already created
			// to prevent leaks and return the error.
			r.drainResourcePool()
			return fmt.Errorf("failed to create resource for pool: %w", err)
		}
		select {
		case r.resourcePool <- resource:
			// Resource added to pool.
			r.availableResources.Add(1)
		case <-r.ctx.Done():
			// Context cancelled during initialization, destroy resource and exit.
			if destroyErr := r.onDestroy(resource); destroyErr != nil {
				// Log error but continue cleanup
			}
			r.drainResourcePool() // Ensure any others are drained
			return r.ctx.Err()
		}
	}
	return nil
}

// drainResourcePool cleans up all resources currently held within the resource pool.
// It is typically called during shutdown or if `initializeResourcePool` fails.
func (r *Runner[R, E]) drainResourcePool() {
	for {
		select {
		case resource := <-r.resourcePool: // resource is of type R
			// Attempt to destroy each resource. Continue draining even on errors.
			if err := r.onDestroy(resource); err != nil {
				// In a production application, you might log this error with a logger.
				// For now, we continue cleanup to prevent resource leaks
			}
			r.availableResources.Add(-1)
		default:
			// No more resources in the pool.
			return
		}
	}
}

// startWorker creates and launches a new worker goroutine.
// The `isBurst` parameter indicates if this is a temporary burst worker (`true`)
// or a permanent base worker (`false`).
func (r *Runner[R, E]) startWorker(ctx context.Context, isBurst bool) {
	r.wg.Add(1) // Increment the WaitGroup counter.
	go r.worker(ctx, isBurst)
}

// worker is the main loop for a worker goroutine.
// Each worker acquires a resource, processes tasks from the priority or main queue,
// and handles task execution, retries, and health checks.
// Workers exit when their context is cancelled or if they encounter an unhealthy error.
func (r *Runner[R, E]) worker(ctx context.Context, isBurst bool) {
	defer r.wg.Done() // Decrement WaitGroup counter when the worker exits.

	// Create a resource for this worker's exclusive use.
	resource, err := r.onCreate() // onCreate now returns R directly
	if err != nil {
		// Log error if resource creation fails and exit worker.
		// In a production app, use a proper logger.
		return
	}
	defer func() {
		if destroyErr := r.onDestroy(resource); destroyErr != nil {
			// Handle cleanup errors appropriately in production
		}
	}()

	// Update worker counters.
	if isBurst {
		r.burstWorkers.Add(1)
		defer r.burstWorkers.Add(-1)
	}
	r.activeWorkers.Add(1)
	defer r.activeWorkers.Add(-1)

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, initiate worker shutdown.
			return
		case task := <-r.priorityQueue:
			// Process high priority tasks first.
			r.priorityTasks.Add(-1) // Decrement count immediately.
			if !r.executeTask(task, resource) {
				// Worker is unhealthy or task failed beyond retries, exit worker.
				return
			}
		case task := <-r.mainQueue:
			// Process regular tasks if no priority tasks are available.
			r.queuedTasks.Add(-1) // Decrement count immediately.
			if !r.executeTask(task, resource) {
				// Worker is unhealthy or task failed beyond retries, exit worker.
				return
			}
		}
	}
}

// executeTask runs a task, handles its result, error, and retry logic.
// It returns `true` if the worker remains healthy and can continue processing tasks,
// and `false` if the worker should exit (e.g., due to an unhealthy resource).
func (r *Runner[R, E]) executeTask(task *Task[R, E], resource R) bool {
	result, err := task.run(resource) // Execute the task function.

	if err != nil && !r.checkHealth(err) {
		// Error occurred AND checkHealth indicates an unhealthy state.
		if task.retries < r.maxRetries {
			// Task has remaining retries; re-queue it to the priority queue.
			task.retries++
			select {
			case r.priorityQueue <- task:
				// Task successfully re-queued.
				r.priorityTasks.Add(1) // Re-increment priority task count.
			default:
				// Priority queue is full, cannot re-queue. Send error back immediately.
				r.sendTaskResult(task, result, fmt.Errorf("priority queue full, task requeue failed: %w", err))
			}
		} else {
			// Max retries exceeded. Send error back to caller.
			r.sendTaskResult(task, result, fmt.Errorf("max retries exceeded for task: %w", err))
		}
		return false // Worker should exit as its resource or itself might be unhealthy.
	}

	// Task completed successfully or with a "healthy" error. Send result back.
	r.sendTaskResult(task, result, err)
	return true // Worker remains healthy and can continue processing.
}

// sendTaskResult safely sends task results back to the caller with proper error handling.
// It ensures that results are delivered reliably without blocking indefinitely.
func (r *Runner[R, E]) sendTaskResult(task *Task[R, E], result E, err error) {
	// Create a context with timeout to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Send error first
	select {
	case task.err <- err:
	case <-ctx.Done():
		// If we can't send the error, there's likely a programming error or shutdown
		// In production, this might be logged
	}

	// Send result
	select {
	case task.result <- result:
	case <-ctx.Done():
		// If we can't send the result, there's likely a programming error or shutdown
		// In production, this might be logged
	}
}

// QueueTask adds a task to the main queue for asynchronous processing.
// It blocks until the task can be added to the queue or the manager shuts down.
// Once the task is processed by a worker, its result and error are returned.
func (r *Runner[R, E]) QueueTask(taskFunc func(R) (E, error)) (E, error) {
	task := &Task[R, E]{
		run: taskFunc,
		// Buffered channels to ensure senders don't block immediately if receiver isn't ready.
		result: make(chan E, 1),
		err:    make(chan error, 1),
	}

	select {
	case r.mainQueue <- task:
		r.queuedTasks.Add(1) // Increment count.
		// Wait for the task to be processed and results to be sent back.
		res := <-task.result
		err := <-task.err
		return res, err
	case <-r.ctx.Done():
		var zero E
		return zero, errors.New("task manager is shutting down: cannot queue task")
	}
}

// QueueTaskWithPriority adds a high priority task to the priority queue.
// Similar to QueueTask, but tasks in this queue are processed first.
func (r *Runner[R, E]) QueueTaskWithPriority(taskFunc func(R) (E, error)) (E, error) {
	task := &Task[R, E]{
		run: taskFunc,
		// Buffered channels for results.
		result: make(chan E, 1),
		err:    make(chan error, 1),
	}

	select {
	case r.priorityQueue <- task:
		r.priorityTasks.Add(1) // Increment count.
		res := <-task.result
		err := <-task.err
		return res, err
	case <-r.ctx.Done():
		var zero E
		return zero, errors.New("task manager is shutting down: cannot queue priority task")
	}
}

// RunTask executes a task immediately, attempting to use a resource from the pool.
// If the pool is empty, a new temporary resource is created and destroyed after use.
// This method is synchronous, blocking until the task completes.
func (r *Runner[R, E]) RunTask(taskFunc func(R) (E, error)) (E, error) {
	var resource R
	var err error
	fromPool := false

	// Try to get a resource from the pool.
	select {
	case resource = <-r.resourcePool:
		fromPool = true
		r.availableResources.Add(-1) // Decrement available count.
	case <-r.ctx.Done():
		var zero E
		return zero, errors.New("task manager is shutting down: cannot run immediate task")
	default:
		// Pool is empty or unavailable, create a temporary resource.
		resource, err = r.onCreate()
		if err != nil {
			var zero E
			return zero, fmt.Errorf("failed to create temporary resource for immediate task: %w", err)
		}
	}

	// Ensure proper cleanup
	defer func() {
		if fromPool {
			// Return to pool
			select {
			case r.resourcePool <- resource:
				r.availableResources.Add(1) // Increment available count.
			case <-r.ctx.Done():
				// If manager is shutting down, destroy the resource instead of returning to pool.
				if destroyErr := r.onDestroy(resource); destroyErr != nil {
					// Handle cleanup error appropriately
				}
			}
		} else {
			// Temporary resource, destroy it
			if destroyErr := r.onDestroy(resource); destroyErr != nil {
				// Handle cleanup error appropriately
			}
		}
	}()

	// Execute the task with the obtained resource.
	return taskFunc(resource)
}

// startBurstManager initializes and runs a goroutine responsible for dynamic
// worker scaling (bursting). It periodically checks queue sizes and adjusts
// the number of burst workers based on `BurstTaskThreshold` and `BurstWorkerCount`.
// This function is a no-op if burst configuration is invalid (BurstTaskThreshold or BurstWorkerCount <= 0).
func (r *Runner[R, E]) startBurstManager() {
	if r.burstTaskThreshold <= 0 || r.burstWorkerCount <= 0 {
		return // Bursting is disabled if thresholds are not set.
	}

	r.burstTicker = time.NewTicker(r.burstInterval) // Create a new ticker for periodic checks.

	r.wg.Add(1) // Increment WaitGroup for the burst manager goroutine.
	go func() {
		defer r.wg.Done()         // Decrement WaitGroup when manager exits.
		defer r.burstTicker.Stop() // Stop the ticker to release resources.

		for {
			select {
			case <-r.ctx.Done():
				// Main context cancelled, shut down burst manager.
				return
			case <-r.burstStop:
				// Explicit stop signal received, shut down burst manager.
				return
			case <-r.burstTicker.C:
				// Time for a burst management check.
				r.manageBurst()
			}
		}
	}()
}

// manageBurst handles the logic for scaling workers up or down dynamically.
// It is called periodically by the burst manager.
// It acquires a mutex to safely modify shared burst-related state.
func (r *Runner[R, E]) manageBurst() {
	r.burstMu.Lock()
	defer r.burstMu.Unlock()

	// Get current queue sizes using atomic operations for accuracy
	mainQueueSize := int(r.queuedTasks.Load())
	priorityQueueSize := int(r.priorityTasks.Load())
	totalQueued := mainQueueSize + priorityQueueSize

	// Get current worker counts
	currentActiveWorkers := int(r.activeWorkers.Load())
	currentBurstWorkers := int(r.burstWorkers.Load())

	if totalQueued > r.burstTaskThreshold {
		// Scale up: If tasks exceed threshold, create more burst workers.
		canAdd := r.maxWorkerCount - currentActiveWorkers
		wantToAdd := r.burstWorkerCount

		// Don't add more than we can or more than needed based on queue size
		// Add at most one worker per queued task above threshold
		maxUseful := totalQueued - r.burstTaskThreshold
		toAdd := min(wantToAdd, min(canAdd, maxUseful))

		if toAdd > 0 {
			for range toAdd {
				// Create a new cancellable context for each burst worker.
				burstCtx, cancel := context.WithCancel(r.ctx)
				r.burstCtxs = append(r.burstCtxs, cancel)
				r.startWorker(burstCtx, true)
			}
		}
	} else if totalQueued < r.burstTaskThreshold/2 && currentBurstWorkers > 0 {
		// Scale down: If queue backlog is significantly reduced and burst workers exist.
		// Remove some burst workers, but not all at once for stability
		toRemove := min(currentBurstWorkers, max(1, currentBurstWorkers/2))

		for i := 0; i < toRemove && len(r.burstCtxs) > 0; i++ {
			// Get the last added burst worker's cancel function and call it.
			lastIdx := len(r.burstCtxs) - 1
			cancel := r.burstCtxs[lastIdx]
			cancel()
			// Remove the cancel function from the slice.
			r.burstCtxs = r.burstCtxs[:lastIdx]
		}
	}
}

// Stats returns current statistics about the task manager's operational state.
// It provides a snapshot of worker counts, queued tasks, and resource availability.
func (r *Runner[R, E]) Stats() TaskStats {
	return TaskStats{
		BaseWorkers:        r.baseWorkerCount,
		ActiveWorkers:      r.activeWorkers.Load(),
		BurstWorkers:       r.burstWorkers.Load(),
		QueuedTasks:        r.queuedTasks.Load(),
		PriorityTasks:      r.priorityTasks.Load(),
		AvailableResources: r.availableResources.Load(),
	}
}

// Stop gracefully shuts down the task manager.
// It signals all workers and managers to stop, waits for their completion,
// and cleans up all allocated resources.
// It should be called to release resources and prevent goroutine leaks.
func (r *Runner[R, E]) Stop() error {
	// Step 1: Signal the burst manager to stop.
	select {
	case r.burstStop <- struct{}{}:
	default:
		// If the burst manager goroutine is already stopped or burstStop channel is not ready,
		// this means the signal was either not needed or it's already in a shutdown state.
	}

	// Step 2: Cancel all burst workers.
	r.burstMu.Lock()
	for _, cancel := range r.burstCtxs {
		cancel() // Call cancel function for each burst worker.
	}
	r.burstCtxs = nil // Clear the slice to free memory.
	r.burstMu.Unlock()

	// Step 3: Cancel the main context.
	// This will signal all base workers and the burst manager (if still running) to exit.
	r.cancel()

	// Step 4: Wait for all goroutines (workers, burst manager) to finish.
	// This ensures that all tasks currently being processed are completed
	// and workers clean up their resources.
	r.wg.Wait()

	// Step 5: Clean up any remaining resources in the resource pool.
	// This happens after all workers are done, so no one tries to put
	// resources back into a draining pool.
	r.drainResourcePool()

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
