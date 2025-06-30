# Getting Started

### Overview and Core Concepts
Tasker operates around several core concepts that enable its powerful concurrency features:

*   **Resource (`R`)**: This is a generic type representing any dependency or object your tasks require for execution (e.g., a database connection, an HTTP client, a custom compute unit). Tasker manages the creation and destruction of these resources.
*   **Task Result (`E`)**: A generic type for the value returned by a task upon successful completion. This ensures type safety for diverse task outputs.
*   **Task Function**: Your application's specific logic, defined as a `func(resource R) (result E, err error)`. This function is executed by a worker, receiving an `R` resource instance.
*   **`tasker.Config[R]`**: The configuration struct used to initialize the `TaskManager`. It defines resource lifecycle functions (`OnCreate`, `OnDestroy`), worker counts, scaling parameters, health check logic, and optional custom logging/metrics.
*   **`tasker.Manager[R, E]`**: The concrete implementation of the `TaskManager[R, E]` interface. It's the central orchestrator for task management, handling queues, worker synchronization, and resource lifecycles.
*   **`Task[R, E]` (Internal)**: An internal struct encapsulating a task's executable function, channels for its result and errors, a retry counter, and a timestamp for when it was queued.
*   **`TaskStats`**: A snapshot struct providing real-time operational statistics of the `TaskManager`, including worker counts and queued tasks.
*   **`TaskMetrics`**: A comprehensive struct offering aggregated performance metrics such as task arrival/completion rates, execution time percentiles (P95, P99), and success/failure rates.
*   **`Logger` Interface**: An interface allowing integration with custom logging solutions.
*   **`MetricsCollector` Interface**: An interface for integrating with external monitoring and observability systems for detailed performance metrics.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [],
  "verificationSteps": [],
  "quickPatterns": [],
  "diagnosticPaths": []
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*