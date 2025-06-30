# Advanced Usage

### Custom Logging and Metrics Integration

Tasker provides interfaces to integrate with your preferred logging solution and external metrics systems, giving you full control over observability.

#### Custom Logging with `tasker.Logger`
By default, Tasker uses a no-op logger. To see internal Tasker messages or integrate with your logging framework (e.g., `logrus`, `zap`), implement the `tasker.Logger` interface and pass it to `Config.Logger`.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker"
)

// MyCustomLogger implements tasker.Logger
type MyCustomLogger struct{}

func (l *MyCustomLogger) Debugf(format string, args ...any) { log.Printf("[DEBUG] "+format, args...) }
func (l *MyCustomLogger) Infof(format string, args ...any)  { log.Printf("[INFO] "+format, args...) }
func (l *MyCustomLogger) Warnf(format string, args ...any)  { log.Printf("[WARN] "+format, args...) }
func (l *MyCustomLogger) Errorf(format string, args ...any) { log.Printf("[ERROR] "+format, args...) }

// (Assume CalculatorResource and its onCreate/onDestroy are defined as before)
type CalculatorResource struct{}
func createCalcResource() (*CalculatorResource, error) { /* ... */ return &CalculatorResource{}, nil }
func destroyCalcResource(r *CalculatorResource) error { /* ... */ return nil }

func main() {
	config := tasker.Config[*CalculatorResource]{
		OnCreate:    createCalcResource,
		OnDestroy:   destroyCalcResource,
		WorkerCount: 1,
		Ctx:         context.Background(),
		Logger:      &MyCustomLogger{}, // Inject your custom logger here
	}

	manager, err := tasker.NewTaskManager[*CalculatorResource, int](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }
	defer manager.Stop()

	_, _ = manager.QueueTask(func(r *CalculatorResource) (int, error) { return 1 + 1, nil })
	time.Sleep(100 * time.Millisecond)
}
```

**Outcome**: Tasker's internal messages will now be routed through `MyCustomLogger`, appearing in `log.Printf` output prefixed with `[DEBUG]`, `[INFO]`, etc.

#### Custom Metrics with `tasker.MetricsCollector`
Tasker automatically collects a rich set of performance metrics. You can access these metrics via `manager.Metrics()` (which returns a `TaskMetrics` struct). If you wish to integrate with an external metrics system (e.g., Prometheus), you can provide your own implementation of `tasker.MetricsCollector`.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/asaidimu/tasker"
)

// MyCustomMetricsCollector implements tasker.MetricsCollector
type MyCustomMetricsCollector struct {
	sync.Mutex
	// Store metrics in a way suitable for your external system
	TotalArrived uint64
	TotalCompleted uint64
	LastMetrics tasker.TaskMetrics
}

func (m *MyCustomMetricsCollector) RecordArrival()       { m.Lock(); defer m.Unlock(); m.TotalArrived++ }
func (m *MyCustomMetricsCollector) RecordCompletion(s tasker.TaskLifecycleTimestamps) {
    m.Lock(); defer m.Unlock(); m.TotalCompleted++;
    // In a real collector, you'd process stamps for latency percentiles etc.
}
func (m *MyCustomMetricsCollector) RecordFailure(s tasker.TaskLifecycleTimestamps) { /* ... */ }
func (m *MyCustomMetricsCollector) RecordRetry()       { /* ... */ }
func (m *mYCustomMetricsCollector) Metrics() tasker.TaskMetrics {
    m.Lock(); defer m.Unlock();
    // In a real collector, calculate current metrics from aggregated data
    m.LastMetrics.TotalTasksCompleted = m.TotalCompleted
    m.LastMetrics.TotalTasksArrived = m.TotalArrived // This is simplified, actual rate calc is complex
    return m.LastMetrics
}

// (Assume CalculatorResource and its onCreate/onDestroy are defined as before)
type CalculatorResource struct{}
func createCalcResource() (*CalculatorResource, error) { /* ... */ return &CalculatorResource{}, nil }
func destroyCalcResource(r *CalculatorResource) error { /* ... */ return nil }

func main() {
	customCollector := &MyCustomMetricsCollector{}
	config := tasker.Config[*CalculatorResource]{
		OnCreate:    createCalcResource,
		OnDestroy:   destroyCalcResource,
		WorkerCount: 1,
		Ctx:         context.Background(),
		Collector:   customCollector, // Inject your custom collector here
	}

	manager, err := tasker.NewTaskManager[*CalculatorResource, int](config)
	if err != nil { log.Fatalf("Error creating task manager: %v", err) }
	defer manager.Stop()

	_, _ = manager.QueueTask(func(r *CalculatorResource) (int, error) { return 1 + 1, nil })
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Custom Collector Total Tasks Arrived: %d\n", customCollector.TotalArrived)
	fmt.Printf("Custom Collector Total Tasks Completed: %d\n", customCollector.TotalCompleted)
}
```

**Outcome**: Tasker's internal events (`RecordArrival`, `RecordCompletion`, etc.) will call your `MyCustomMetricsCollector` methods, allowing you to feed this data into your chosen metrics backend. `manager.Metrics()` will also return data from your custom collector.

---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF [internal Tasker logs are required for debugging or monitoring] THEN [implement `tasker.Logger` and assign to `Config.Logger`] ELSE [use default no-op logger]",
    "IF [detailed Tasker performance metrics need to be exposed to an external monitoring system (e.g., Prometheus)] THEN [implement `tasker.MetricsCollector` and assign to `Config.Collector`] ELSE [rely on `manager.Stats()` and `manager.Metrics()` for in-process inspection]"
  ],
  "verificationSteps": [
    "Check: Custom logger methods are called for Tasker internal messages → Expected: Log output matches custom logger format.",
    "Check: Custom metrics collector methods (`RecordArrival`, `RecordCompletion`) are invoked on relevant task events → Expected: Internal counts in custom collector reflect actual task activity.",
    "Check: `manager.Metrics()` returns data populated by custom collector → Expected: Metrics reflect the state managed by the custom collector."
  ],
  "quickPatterns": [
    "Pattern: Minimal custom logger\n```go\nimport \"log\"\n\ntype ConsoleLogger struct{}\nfunc (ConsoleLogger) Debugf(f string, a ...any) { log.Printf(\"DEBUG: \"+f, a...) }\nfunc (ConsoleLogger) Infof(f string, a ...any)  { log.Printf(\"INFO: \"+f, a...) }\nfunc (ConsoleLogger) Warnf(f string, a ...any)  { log.Printf(\"WARN: \"+f, a...) }\nfunc (ConsoleLogger) Errorf(f string, a ...any) { log.Printf(\"ERROR: \"+f, a...) }\n\n// Usage:\nconfig.Logger = ConsoleLogger{}\n```",
    "Pattern: Simple custom metrics collector (for total tasks completed)\n```go\nimport \"sync/atomic\"\n\ntype CounterCollector struct { totalCompleted atomic.Uint64 }\nfunc (c *CounterCollector) RecordArrival() {} // No-op\nfunc (c *CounterCollector) RecordCompletion(s tasker.TaskLifecycleTimestamps) { c.totalCompleted.Add(1) }\nfunc (c *CounterCollector) RecordFailure(s tasker.TaskLifecycleTimestamps) {} // No-op\nfunc (c *CounterCollector) RecordRetry() {} // No-op\nfunc (c *CounterCollector) Metrics() tasker.TaskMetrics { return tasker.TaskMetrics{TotalTasksCompleted: c.totalCompleted.Load()} }\n\n// Usage:\nconfig.Collector = &CounterCollector{}\n```"
  ],
  "diagnosticPaths": [
    "Error `No Tasker logs appearing` -> Symptom: Tasker is running but no `INFO`, `DEBUG` messages are visible -> Check: Ensure `Config.Logger` is assigned a working logger instance and not `nil` or the `noOpLogger` -> Fix: Provide a concrete implementation for `Config.Logger`.",
    "Error `Metrics in external system are incorrect or missing` -> Symptom: `TaskMetrics` values are not propagating to monitoring dashboard -> Check: Verify the custom `MetricsCollector` implementation correctly processes `RecordArrival`, `RecordCompletion`, etc., and exports data to the external system -> Fix: Debug the `MetricsCollector`'s internal logic and export mechanism."
  ]
}
```

---
*Generated using Gemini AI on 6/30/2025, 4:52:27 PM. Review and refine as needed.*