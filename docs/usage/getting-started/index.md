# Getting Started

## Overview
Tasker is a powerful and flexible Go library designed for efficient management of concurrent tasks. It provides a highly customizable worker pool, dynamic scaling (bursting), priority queuing, and robust resource lifecycle management, making it ideal for processing background jobs, handling I/O-bound operations, or managing CPU-intensive computations with controlled concurrency. It abstracts away the complexities of goroutine management, worker pools, and resource lifecycles, allowing developers to define tasks that operate on specific resources and then queue these tasks for execution, letting `tasker` handle the underlying concurrency, scaling, and error recovery.

## Core Concepts
*   **Resource (`R`)**: This is the generic type of resource your tasks will operate on. It could be anything: a database connection, an HTTP client, a custom processing struct, a CPU/GPU compute unit, or any other external dependency or expensive object that needs managed lifecycle. `tasker` manages the creation, use, and destruction of these resources.
*   **Task Result (`E`)**: This is the generic type of value your tasks will return upon successful completion. This allows `tasker` to be type-safe for various task outputs.
*   **Task Function**: Your actual work logic is defined as a `func(ctx context.Context, resource R) (result E, err error)`. This function receives a `context.Context` for cancellation and an instance of your defined `R` resource type. It's expected to return the result `E` or an `error`.
*   **`tasker.Config[R]`**: A struct used to configure the `TaskManager` with essential parameters such as resource lifecycle functions (`OnCreate`, `OnDestroy`), initial worker counts, dynamic scaling parameters, optional health check logic, and custom logging/metrics.
*   **`tasker.Manager[R, E]`**: The concrete implementation of the `TaskManager[R, E]` interface, providing the core task management capabilities. You instantiate this via `tasker.NewTaskManager`.

## Quick Setup Guide

### Prerequisites
*   Go **1.24.3** or higher

### Installation Steps
To add `tasker` to your Go project, use `go get`:

```bash
go get github.com/asaidimu/tasker/v2
```

### Verification
You can verify the installation and see `tasker` in action by building and running the provided examples:

```bash
# Navigate to the examples directory within your Go module path
# This assumes your GOPATH is set correctly, typically within your user home directory.
cd "$(go env GOPATH)/src/github.com/asaidimu/tasker/examples/basic"
go run main.go

cd "$(go env GOPATH)/src/github.com/asaidimu/tasker/examples/intermediate"
go run main.go

cd "$(go env GOPATH)/src/github.com/asaidimu/tasker/examples/advanced"
go run main.go
```


---
### 🤖 AI Agent Guidance

```json
{
  "decisionPoints": [
    "IF project_type IS Go_application THEN install_command IS \"go get github.com/asaidimu/tasker/v2\" ELSE \"Not_applicable\""
  ],
  "verificationSteps": [
    "Check: `go list github.com/asaidimu/tasker/v2/...` -> Expected: package information is displayed",
    "Check: Run `go run examples/basic/main.go` -> Expected: output from example indicates successful task execution and manager shutdown."
  ],
  "quickPatterns": [
    "Pattern: Go_installation_check\n```go\npackage main\n\nimport (\n\t\"fmt\"\n\t\"runtime\"\n)\n\nfunc main() {\n\tfmt.Printf(\"Go Version: %s\\n\", runtime.Version())\n\tfmt.Printf(\"GOOS: %s\\n\", runtime.GOOS)\n\tfmt.Printf(\"GOARCH: %s\\n\", runtime.GOARCH)\n}\n```",
    "Pattern: Basic_tasker_setup\n```go\npackage main\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\t\"log\"\n\t\"time\"\n\n\t\"github.com/asaidimu/tasker/v2\"\n)\n\ntype MyResource struct{}\n\nfunc createMyResource() (*MyResource, error) {\n\tfmt.Println(\"Creating MyResource\")\n\treturn &MyResource{}, nil\n}\n\nfunc destroyMyResource(r *MyResource) error {\n\tfmt.Println(\"Destroying MyResource\")\n\treturn nil\n}\n\nfunc main() {\n\tconfig := tasker.Config[*MyResource]{\n\t\tOnCreate:    createMyResource,\n\t\tOnDestroy:   destroyMyResource,\n\t\tWorkerCount: 1,\n\t\tCtx:         context.Background(),\n\t}\n\tmanager, err := tasker.NewTaskManager[*MyResource, string](config)\n\tif err != nil {\n\t\tlog.Fatalf(\"Failed to create manager: %v\", err)\n\t}\n\tdefer manager.Stop()\n\tfmt.Println(\"Task manager initialized.\")\n}\n```"
  ],
  "diagnosticPaths": []
}
```

---
*Generated using Gemini AI on 6/30/2025, 9:35:09 PM. Review and refine as needed.*