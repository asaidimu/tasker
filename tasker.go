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

	// QueueTaskOnce adds a task to the main queue for asynchronous execution.
	// This task will NOT be re-queued by the task manager's internal retry mechanism
	// if it fails and CheckHealth indicates an unhealthy state. This is suitable
	// for non-idempotent operations where "at-most-once" execution is desired
	// from the task manager's perspective.
	// It returns the result and error of the task once it completes.
	// If the task manager is shutting down, it returns an error immediately.
	QueueTaskOnce(task func(R) (E, error)) (E, error)

	// QueueTaskWithPriorityOnce adds a high priority task to a dedicated queue.
	// This task will NOT be re-queued by the task manager's internal retry mechanism
	// if it fails and CheckHealth indicates an unhealthy state. This is suitable
	// for non-idempotent high-priority operations where "at-most-once" execution
	// is desired from the task manager's perspective.
	// It returns the result and error of the task once it completes.
	// If the task manager is shutting down, it returns an error immediately.
	QueueTaskWithPriorityOnce(task func(R) (E, error)) (E, error)

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

	// Internal utilities - encapsulated within Runner
	utils struct {
		taskResultTimeout time.Duration // Timeout for sending task results back to callers
	}
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
	if config.BurstWorkerCount <= 0 {
		config.BurstWorkerCount = 2 // Default value
	}
	if config.MaxWorkerCount <= 0 {
		config.MaxWorkerCount = config.WorkerCount + config.BurstWorkerCount // Default value
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
		mainQueue:           make(chan *Task[R, E], config.WorkerCount*2),
		priorityQueue:       make(chan *Task[R, E], config.WorkerCount),
		resourcePool:        make(chan R, config.ResourcePoolSize),
		poolSize:            config.ResourcePoolSize,
		ctx:                 ctx,
		cancel:              cancel,
		burstStop:           make(chan struct{}),
		burstCtxs:           make([]context.CancelFunc, 0),
	}

	// Initialize encapsulated state constants
	runner.shutdownState.running = 0
	runner.shutdownState.stopping = 1
	runner.shutdownState.killed = 2
	runner.shutdownState.current.Store(runner.shutdownState.running)

	// Initialize internal utilities
	runner.utils.taskResultTimeout = 100 * time.Millisecond

	if err := runner.initializeResourcePool(); err != nil {
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
func (r *Runner[R, E]) isRunning() bool {
	return r.shutdownState.current.Load() == r.shutdownState.running
}

// isStopping checks if the runner is in stopping state
func (r *Runner[R, E]) isStopping() bool {
	return r.shutdownState.current.Load() == r.shutdownState.stopping
}

// isKilled checks if the runner is in killed state
func (r *Runner[R, E]) isKilled() bool {
	return r.shutdownState.current.Load() == r.shutdownState.killed
}

// transitionState atomically transitions from one state to another
func (r *Runner[R, E]) transitionState(from, to int32) bool {
	return r.shutdownState.current.CompareAndSwap(from, to)
}

// getCurrentState returns the current shutdown state
func (r *Runner[R, E]) getCurrentState() int32 {
	return r.shutdownState.current.Load()
}

// minInt returns the minimum of two integers - encapsulated utility
func (r *Runner[R, E]) minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the maximum of two integers - encapsulated utility
func (r *Runner[R, E]) maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// initializeResourcePool pre-allocates and populates the resource pool.
func (r *Runner[R, E]) initializeResourcePool() error {
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
	return nil
}

// drainResourcePool cleans up all resources currently held within the resource pool.
func (r *Runner[R, E]) drainResourcePool() {
	close(r.resourcePool)
	for resource := range r.resourcePool {
		_ = r.onDestroy(resource) // Log errors in a real app
		r.availableResources.Add(-1)
	}
}

// startWorker creates and launches a new worker goroutine.
func (r *Runner[R, E]) startWorker(ctx context.Context, isBurst bool) {
	r.wg.Add(1)
	go r.worker(ctx, isBurst)
}

// worker is the main loop for a worker goroutine.
// It processes tasks until the context is canceled, at which point it either
// drains the queues (for graceful stop) or exits immediately (for kill).
func (r *Runner[R, E]) worker(ctx context.Context, isBurst bool) {
	defer r.wg.Done()

	resource, err := r.onCreate()
	if err != nil {
		return // Cannot start worker without a resource
	}
	defer func() { _ = r.onDestroy(resource) }()

	if isBurst {
		r.burstWorkers.Add(1)
		defer r.burstWorkers.Add(-1)
	}
	r.activeWorkers.Add(1)
	defer r.activeWorkers.Add(-1)

	for {
		select {
		case <-ctx.Done():
			// Context is canceled. Check if this is a graceful stop.
			if r.isStopping() {
				// Graceful stop: drain all queues before exiting.
				for r.processQueues(resource) {
				}
			}
			// For kill, or after draining is complete, exit immediately.
			return
		case task := <-r.priorityQueue:
			r.priorityTasks.Add(-1)
			if !r.executeTask(task, resource) {
				return // Unhealthy worker exits
			}
		case task := <-r.mainQueue:
			r.queuedTasks.Add(-1)
			if !r.executeTask(task, resource) {
				return // Unhealthy worker exits
			}
		}
	}
}

// processQueues tries to process one task from either queue. Used for draining.
// Returns true if a task was processed, false if queues were empty.
func (r *Runner[R, E]) processQueues(resource R) bool {
	select {
	case task := <-r.priorityQueue:
		r.priorityTasks.Add(-1)
		r.executeTask(task, resource)
		return true
	case task := <-r.mainQueue:
		r.queuedTasks.Add(-1)
		r.executeTask(task, resource)
		return true
	default:
		return false // Both queues are empty
	}
}

// executeTask runs a single task and handles its retry logic.
func (r *Runner[R, E]) executeTask(task *Task[R, E], resource R) bool {
	result, err := task.run(resource)

	if err != nil && !r.checkHealth(err) {
		if task.retries < r.maxRetries {
			task.retries++
			select {
			case r.priorityQueue <- task:
				r.priorityTasks.Add(1)
			default:
				r.sendTaskResult(task, result, fmt.Errorf("priority queue full, task requeue failed: %w", err))
			}
		} else {
			r.sendTaskResult(task, result, fmt.Errorf("max retries exceeded: %w", err))
		}
		return false // Worker is unhealthy
	}

	r.sendTaskResult(task, result, err)
	return true // Worker is healthy
}

// sendTaskResult safely sends the task result and error back to the caller.
func (r *Runner[R, E]) sendTaskResult(task *Task[R, E], result E, err error) {
	// Non-blocking send with a short timeout to prevent being stuck forever
	ctx, cancel := context.WithTimeout(context.Background(), r.utils.taskResultTimeout)
	defer cancel()

	select {
	case task.err <- err:
	case <-ctx.Done():
	}

	select {
	case task.result <- result:
	case <-ctx.Done():
	}
}

// QueueTask adds a task to the main queue for asynchronous processing.
func (r *Runner[R, E]) QueueTask(taskFunc func(R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		return zero, errors.New("task manager is shutting down")
	}

	task := &Task[R, E]{
		run:    taskFunc,
		result: make(chan E, 1),
		err:    make(chan error, 1),
	}

	select {
	case r.mainQueue <- task:
		r.queuedTasks.Add(1)
		return <-task.result, <-task.err
	case <-r.ctx.Done():
		var zero E
		return zero, errors.New("task manager is shutting down")
	}
}

// QueTaskOnce adds a task to the main queue that will NOT be retried
// by the task manager's internal retry mechanism if it fails and
// CheckHealth indicates an unhealthy state. This is suitable for
// non-idempotent operations.
// It returns the result and error of the task once it completes.
// If the task manager is shutting down, it returns an error immediately.
func (r *Runner[R, E]) QueueTaskOnce(taskFunc func(R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		return zero, errors.New("task manager is shutting down")
	}

	task := &Task[R, E]{
		run:    taskFunc,
		result: make(chan E, 1),
		err:    make(chan error, 1),
		retries: r.maxRetries, // Set retries to max so it won't retry on health failure
	}

	select {
	case r.mainQueue <- task:
		r.queuedTasks.Add(1)
		return <-task.result, <-task.err
	case <-r.ctx.Done():
		var zero E
		return zero, errors.New("task manager is shutting down")
	}
}


// QueueTaskWithPriority adds a high-priority task to the priority queue.
func (r *Runner[R, E]) QueueTaskWithPriority(taskFunc func(R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		return zero, errors.New("task manager is shutting down")
	}

	task := &Task[R, E]{
		run:    taskFunc,
		result: make(chan E, 1),
		err:    make(chan error, 1),
	}

	select {
	case r.priorityQueue <- task:
		r.priorityTasks.Add(1)
		return <-task.result, <-task.err
	case <-r.ctx.Done():
		var zero E
		return zero, errors.New("task manager is shutting down")
	}
}

// QueueTaskWithPriority adds a high-priority task to the priority queue, but
// the task will not be retried.
func (r *Runner[R, E]) QueueTaskWithPriorityOnce(taskFunc func(R) (E, error)) (E, error) {
 	if !r.isRunning() {
 		var zero E
 		return zero, errors.New("task manager is shutting down")
 	}

 	task := &Task[R, E]{
 		run:    taskFunc,
 		result: make(chan E, 1),
 		err:    make(chan error, 1),
 		retries: r.maxRetries, // Set retries to max so it won't retry on health failure
 	}

 	select {
 	case r.priorityQueue <- task:
 		r.priorityTasks.Add(1)
 		return <-task.result, <-task.err
 	case <-r.ctx.Done():
 		var zero E
 		return zero, errors.New("task manager is shutting down")
 	}
 }

// RunTask executes a task immediately using a resource from the pool.
func (r *Runner[R, E]) RunTask(taskFunc func(R) (E, error)) (E, error) {
	if !r.isRunning() {
		var zero E
		return zero, errors.New("task manager is shutting down")
	}

	var resource R
	var err error
	fromPool := false

	select {
	case resource = <-r.resourcePool:
		fromPool = true
		r.availableResources.Add(-1)
	case <-r.ctx.Done():
		var zero E
		return zero, errors.New("task manager is shutting down")
	default:
		resource, err = r.onCreate()
		if err != nil {
			var zero E
			return zero, fmt.Errorf("failed to create temporary resource: %w", err)
		}
	}

	defer func() {
		if fromPool {
			if r.isRunning() {
				select {
				case r.resourcePool <- resource:
					r.availableResources.Add(1)
				default:
					_ = r.onDestroy(resource) // Pool full, destroy
				}
			} else {
				_ = r.onDestroy(resource) // Shutting down, destroy
			}
		} else {
			_ = r.onDestroy(resource) // Always destroy temporary
		}
	}()

	return taskFunc(resource)
}

// startBurstManager starts the dynamic worker scaling manager.
func (r *Runner[R, E]) startBurstManager() {
	if r.burstTaskThreshold <= 0 || r.burstWorkerCount <= 0 {
		return
	}

	r.burstTicker = time.NewTicker(r.burstInterval)
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer r.burstTicker.Stop()
		for {
			select {
			case <-r.ctx.Done():
				return
			case <-r.burstStop:
				return
			case <-r.burstTicker.C:
				r.manageBurst()
			}
		}
	}()
}

// manageBurst handles the logic for scaling workers up or down.
func (r *Runner[R, E]) manageBurst() {
	r.burstMu.Lock()
	defer r.burstMu.Unlock()

	totalQueued := int(r.queuedTasks.Load() + r.priorityTasks.Load())
	currentActiveWorkers := int(r.activeWorkers.Load())
	currentBurstWorkers := int(r.burstWorkers.Load())

	if totalQueued > r.burstTaskThreshold {
		canAdd := r.maxWorkerCount - currentActiveWorkers
		wantToAdd := r.burstWorkerCount
		maxUseful := totalQueued - r.burstTaskThreshold
		toAdd := r.minInt(wantToAdd, r.minInt(canAdd, maxUseful))

		for range toAdd {
			burstCtx, cancel := context.WithCancel(r.ctx)
			r.burstCtxs = append(r.burstCtxs, cancel)
			r.startWorker(burstCtx, true)
		}
	} else if totalQueued < r.burstTaskThreshold/2 && currentBurstWorkers > 0 {
		toRemove := r.minInt(currentBurstWorkers, r.maxInt(1, currentBurstWorkers/2))
		for i := 0; i < toRemove && len(r.burstCtxs) > 0; i++ {
			lastIdx := len(r.burstCtxs) - 1
			cancel := r.burstCtxs[lastIdx]
			cancel()
			r.burstCtxs = r.burstCtxs[:lastIdx]
		}
	}
}

// Stats returns a snapshot of the runner's current operational state.
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

// Stop gracefully shuts down the task manager by draining the task queues.
func (r *Runner[R, E]) Stop() error {
	// Set state to 'stopping' to prevent new tasks and guide worker exit.
	if !r.transitionState(r.shutdownState.running, r.shutdownState.stopping) {
		return errors.New("task manager already stopping or killed")
	}

	// Stop the burst manager and cancel burst workers.
	r.shutdownBursting()

	// Cancel the main context to signal base workers to start draining.
	r.cancel()

	// Wait for all goroutines to finish.
	r.wg.Wait()

	// Clean up the resource pool.
	r.drainResourcePool()

	return nil
}

// Kill immediately terminates the task manager without draining queues.
func (r *Runner[R, E]) Kill() error {
	// Set state to 'killed' to prevent new tasks and guide worker exit.
	if !r.transitionState(r.shutdownState.running, r.shutdownState.killed) {
		// If it's already 'stopping', we can escalate to 'killed'.
		if !r.transitionState(r.shutdownState.stopping, r.shutdownState.killed) {
			return errors.New("task manager already killed")
		}
	}

	// Stop the burst manager and cancel burst workers.
	r.shutdownBursting()

	// Cancel the main context to signal all workers to exit immediately.
	r.cancel()

	// Wait for all goroutines to finish.
	r.wg.Wait()

	// Clean up the resource pool.
	r.drainResourcePool()

	return nil
}

// shutdownBursting stops the burst manager and cancels all burst workers.
func (r *Runner[R, E]) shutdownBursting() {
	select {
	case r.burstStop <- struct{}{}:
	default:
	}

	r.burstMu.Lock()
	for _, cancel := range r.burstCtxs {
		cancel()
	}
	r.burstCtxs = nil
	r.burstMu.Unlock()
}
