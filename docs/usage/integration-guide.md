# Integration Guide

## Environment Requirements

Go 1.22 or higher. The library leverages Go generics and specific atomic types introduced in recent Go versions. No specific operating system or hardware requirements beyond a standard Go development/runtime environment.

## Initialization Patterns

### Standard initialization of a `tasker.Runner` instance. It involves defining resource creation/destruction functions and setting basic worker parameters. Always pair with `defer manager.Stop()` for graceful shutdown.
```[DETECTED_LANGUAGE]
package main

import (
	"context"
	"fmt"
	"log"
	"github.com/asaidimu/tasker"
)

type MyCustomResource struct { ID int }

func createResource() (*MyCustomResource, error) {
	fmt.Println("Creating MyCustomResource...")
	return &MyCustomResource{ID: 123}, nil
}

func destroyResource(res *MyCustomResource) error {
	fmt.Printf("Destroying MyCustomResource %d.\n", res.ID)
	return nil
}

func main() {
	config := tasker.Config[*MyCustomResource]{
		OnCreate: createResource,
		OnDestroy: destroyResource,
		WorkerCount: 3,
		Ctx: context.Background(),
	}
	
	manager, err := tasker.NewRunner[*MyCustomResource, string](config)
	if err != nil {
		log.Fatalf("Failed to initialize tasker: %v", err)
	}
	defer manager.Stop() // Essential for graceful shutdown
	
	fmt.Println("Tasker manager initialized.")
	// ... your application logic ...
}
```

## Common Integration Pitfalls

- **Issue**: Not calling `manager.Stop()`
  - **Solution**: Always ensure `manager.Stop()` is called when your application is shutting down, typically using `defer manager.Stop()` after `NewRunner` in `main` or a top-level goroutine.

- **Issue**: Blocking or slow `OnCreate`/`OnDestroy` functions
  - **Solution**: These functions are critical for worker startup/shutdown. Keep them fast and non-blocking. Offload any heavy initialization or cleanup to the tasks themselves if possible, or ensure external dependencies are highly available.

- **Issue**: Premature `Config.Ctx` cancellation
  - **Solution**: The context provided in `Config.Ctx` controls the entire `tasker` lifecycle. If it's cancelled early, the manager will shut down, rejecting new tasks. Ensure this context has the same lifecycle as your application's `tasker` usage.

## Lifecycle Dependencies

The `tasker.Runner` instance should be initialized during application startup using `NewRunner`. Its operational phase coincides with the application's active processing. During application shutdown, `manager.Stop()` must be called to ensure all worker goroutines complete their current tasks, resources are properly deallocated via `OnDestroy`, and internal channels are closed cleanly. The `Runner`'s lifecycle is directly managed by the `context.Context` provided in its `Config`.



---
*Generated using Gemini AI on 6/14/2025, 1:47:53 PM. Review and refine as needed.*