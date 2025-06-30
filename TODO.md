TODO: Tasker Library Enhancements
This document outlines recommended improvements to the Tasker library to address usability challenges related to its synchronous task submission methods and lack of direct context propagation. These changes aim to improve developer experience, align with Go’s concurrency idioms, and enhance support for asynchronous workflows.
High Priority
1. Introduce Non-Blocking Task Submission Methods

Description: Add asynchronous variants of task submission methods (QueueTaskAsync, QueueTaskWithPriorityAsync, etc.) that return a channel for results, allowing non-blocking task submission.
Justification: The current synchronous methods (QueueTask, RunTask, etc.) block until completion, which is unintuitive for users expecting fire-and-forget semantics. This forces developers to wrap calls in goroutines, adding boilerplate and complexity (e.g., pattern:Queue task in a goroutine). Non-blocking methods would align with Go’s channel-based concurrency and match the ergonomics of libraries like errgroup or Java’s CompletableFuture.
Tasks:
Define new method signatures: QueueTaskAsync(task func(R) (E, error)) <-chan struct{E, error}.
Implement internal buffering to queue results without blocking the caller.
Update documentation with examples (e.g., receiving results via channels).
Ensure compatibility with existing metrics (RecordArrival, RecordCompletion) and shutdown logic.


Acceptance Criteria:
Users can submit tasks without blocking the calling goroutine.
Results are retrievable via a channel, with proper error handling.
Metrics and logging reflect async submissions accurately.


Estimated Effort: 2-3 days (design, implementation, testing).

2. Pass Context to Task Functions

Description: Modify the task function signature to func(ctx context.Context, res R) (E, error), with Tasker providing a context tied to Config.Ctx for cancellation during Stop() or Kill().
Justification: Tasker’s current task functions (func(R) (E, error)) lack a context parameter, requiring users to manually propagate cancellation signals via closures (noted in commonPitfalls). This adds complexity and is error-prone for long-running tasks, especially during shutdown. Passing a context aligns with Go’s context-driven cancellation model (e.g., errgroup) and simplifies task interruption.
Tasks:
Update TaskManager interface and implementation to pass Config.Ctx (or a derived context) to task functions.
Ensure context cancellation propagates during Stop() and Kill().
Update documentation and patterns (e.g., pattern:Queue task in a goroutine) to demonstrate context usage.
Add tests for context cancellation during task execution and shutdown.


Acceptance Criteria:
Tasks receive a context that is cancelled when Config.Ctx is cancelled or during shutdown.
Long-running tasks can check ctx.Done() to exit gracefully.
Existing task submission patterns remain backward-compatible.


Estimated Effort: 2 days (interface changes, testing, documentation).

Medium Priority
3. Add Per-Task Timeout Support

Description: Introduce an optional timeout parameter for task submission methods (e.g., QueueTaskWithTimeout(task, timeout time.Duration)), automatically cancelling tasks after the specified duration.
Justification: Tasker lacks native support for per-task timeouts, requiring users to implement timeout logic via external contexts. This is cumbersome and inconsistent with Go’s context-based timeout patterns. Timeout support would enhance control over long-running tasks, especially for external resource interactions.
Tasks:
Add new methods like QueueTaskWithTimeout that create a context with a timeout.
Integrate timeout logic with the new context-aware task signature (from Task 2).
Update MetricsCollector to track timeout-related failures (RecordFailure).
Document usage with examples for time-sensitive tasks.


Acceptance Criteria:
Tasks exceeding the specified timeout return context.DeadlineExceeded.
Metrics reflect timeout failures accurately.
Timeout functionality works with both synchronous and asynchronous submissions.


Estimated Effort: 1-2 days (implementation, testing).

4. Provide Callback-Based Task Submission

Description: Add methods like QueueTaskWithCallback(task func(R) (E, error), callback func(E, error)) to handle results asynchronously without requiring channels or goroutines.
Justification: While the proposed QueueTaskAsync (Task 1) uses channels, some users may prefer callbacks for simpler result handling, similar to JavaScript’s async patterns. This reduces the need for explicit goroutine management and aligns with use cases where channels are overkill.
Tasks:
Implement callback-based variants for all task submission methods.
Ensure callbacks are executed in a separate goroutine to avoid blocking.
Update documentation with callback-based examples.
Test callback execution under high load and shutdown scenarios.


Acceptance Criteria:
Callbacks are invoked with task results or errors post-execution.
No additional goroutine boilerplate required by users.
Callbacks handle shutdown errors (e.g., task manager is shutting down) correctly.


Estimated Effort: 1 day (implementation, testing).

Low Priority
5. Enhance Documentation for Asynchronous Patterns

Description: Expand documentation to include more robust examples and best practices for asynchronous task submission and context management, reducing reliance on user-implemented workarounds.
Justification: The current documentation mitigates the synchronous API issue with goroutine patterns (e.g., pattern:Queue task in a goroutine) but lacks comprehensive guidance on managing async workflows or context cancellation. Improved examples would lower the learning curve and prevent common pitfalls.
Tasks:
Add a new pattern: “Asynchronous Task Submission with Result Handling” using channels or callbacks.
Include a pattern for context-aware task execution with cancellation handling.
Update commonPitfalls to emphasize best practices for non-blocking workflows.
Provide a utility function example (e.g., queueWithContext) in the documentation.


Acceptance Criteria:
New patterns cover common async use cases (e.g., web servers, batch processing).
Documentation includes clear examples of context cancellation in tasks.
Users can implement async workflows with minimal external code.


Estimated Effort: 1 day (writing, validation).

6. Add Backpressure Mechanism

Description: Implement a backpressure mechanism to throttle task submissions when queues are full, preventing resource exhaustion under high load.
Justification: Tasker’s current design risks overwhelming the system if task arrival rates exceed processing capacity, as there’s no built-in mechanism to reject or delay submissions. A backpressure mechanism would improve robustness, especially for high-throughput systems.
Tasks:
Add a MaxQueueSize field to Config to limit queue capacity.
Return an error (e.g., errors.New("queue full")) when submitting tasks to a full queue.
Update TaskStats to report queue capacity and current usage.
Document backpressure handling with retry or rejection strategies.


Acceptance Criteria:
Task submissions are rejected or delayed when queues reach MaxQueueSize.
Metrics and stats reflect queue pressure accurately.
Applications can handle backpressure errors gracefully.


Estimated Effort: 2 days (implementation, testing).

Implementation Notes

Backward Compatibility: Ensure all changes (especially Tasks 1 and 2) maintain compatibility with existing synchronous methods to avoid breaking current users.
Testing: Add unit tests for async methods, context cancellation, and timeout scenarios, focusing on edge cases like shutdown during high load.
Performance: Benchmark async methods and context propagation to ensure they don’t introduce significant overhead compared to the current synchronous design.
Community Feedback: Consider soliciting feedback from Tasker users (e.g., via GitHub issues) to validate the need for these features and prioritize implementation.

Related References

Go Concurrency Patterns: Leverage errgroup and channel-based patterns for async implementation inspiration.
Java CompletableFuture: Reference for non-blocking, callback-based APIs.
Documentation Patterns: Expand on existing patterns like pattern:High-priority task with result handling to include async and context-aware examples.

