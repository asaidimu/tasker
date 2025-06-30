package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/asaidimu/tasker/v2"
)

// HeavyComputeResource represents a resource for intensive computations.
type HeavyComputeResource struct {
	ID int
}

// onCreate for HeavyComputeResource: simulates allocating CPU/GPU resources.
func createComputeResource() (*HeavyComputeResource, error) {
	id := rand.Intn(1000)
	fmt.Printf("INFO: Creating HeavyComputeResource %d\n", id)
	return &HeavyComputeResource{ID: id}, nil
}

// onDestroy for HeavyComputeResource: simulates releasing CPU/GPU resources.
func destroyComputeResource(r *HeavyComputeResource) error {
	fmt.Printf("INFO: Destroying HeavyComputeResource %d\n", r.ID)
	return nil
}

// checkComputeHealth: All errors are considered healthy (task-specific, not worker-specific)
func checkComputeHealth(err error) bool {
	return true
}

func main() {
	fmt.Println("\n--- Advanced Usage: Dynamic Scaling (Bursting) ---")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the main context is cancelled on exit

	config := tasker.Config[*HeavyComputeResource]{
		OnCreate:        createComputeResource,
		OnDestroy:       destroyComputeResource,
		WorkerCount:     2,                   // Start with 2 base workers
		Ctx:             ctx,
		CheckHealth:     checkComputeHealth,
		BurstInterval:   200 * time.Millisecond, // Check every 200ms
		MaxRetries:      0,                   // No retries for this example
		ResourcePoolSize: 0,                  // Not using RunTask heavily here
	}

	manager, err := tasker.NewTaskManager[*HeavyComputeResource, string](config)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	// Do NOT defer manager.Stop() here. We will call it manually with a delay to observe shutdown.

	var tasksSubmitted atomic.Int32
	var tasksCompleted atomic.Int32

	// --- 1. Aggressively queue a large number of tasks ---
	fmt.Println("\nAggressively queuing 20 heavy computation tasks...")
	for i := range 20 {
		taskID := i
		tasksSubmitted.Add(1)
		go func() {
			result, err := manager.QueueTask(func(ctx context.Context, res *HeavyComputeResource) (string, error) {
				processingTime := time.Duration(100 + rand.Intn(300)) * time.Millisecond // 100-400ms work
				fmt.Printf("Worker %d processing Task %d for %s\n", res.ID, taskID, processingTime)
				time.Sleep(processingTime)
				if rand.Intn(10) == 0 { // 10% chance of failure
					return "", errors.New(fmt.Sprintf("computation_error_task_%d", taskID))
				}
				return fmt.Sprintf("Task %d completed", taskID), nil
			})
			if err != nil {
				fmt.Printf("Task %d failed: %v\n", taskID, err)
			} else {
				fmt.Printf("Task %d result: %s\n", taskID, result)
				tasksCompleted.Add(1)
			}
		}()
	}

	// --- 2. Monitor stats to observe bursting ---
	fmt.Println("\nMonitoring stats (watch for BurstWorkers)...")
	ticker := time.NewTicker(250 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				stats := manager.Stats()
				fmt.Printf("Stats: Active=%d (Base=%d, Burst=%d), Queued=%d (Main=%d, Prio=%d), AvailRes=%d, Comp=%d/%d\n",
					stats.ActiveWorkers, stats.BaseWorkers, stats.BurstWorkers,
					stats.QueuedTasks+stats.PriorityTasks, stats.QueuedTasks, stats.PriorityTasks,
					stats.AvailableResources, tasksCompleted.Load(), tasksSubmitted.Load())
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	// Let the system run for a while to observe scaling up and down
	time.Sleep(5 * time.Second)

	// Trigger a high-priority task during burst
	fmt.Println("\nQueueing a high priority task during active bursting...")
	go func() {
		tasksSubmitted.Add(1)
		result, err := manager.QueueTaskWithPriority(func(ctx context.Context, res *HeavyComputeResource) (string, error) {
			fmt.Printf("Worker %d processing HIGH PRIORITY urgent report!\n", res.ID)
			time.Sleep(50 * time.Millisecond)
			return "Urgent report generated!", nil
		})
		if err != nil {
			fmt.Printf("Urgent task failed: %v\n", err)
		} else {
			fmt.Printf("Urgent task result: %s\n", result)
			tasksCompleted.Add(1)
		}
	}()

	time.Sleep(2 * time.Second) // Allow priority task and more processing

	// --- 3. Graceful Shutdown ---
	fmt.Println("\nInitiating graceful shutdown...")
	close(done) // Stop the stats monitoring goroutine
	err = manager.Stop()
	if err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	} else {
		fmt.Println("Task manager gracefully shut down.")
	}

	stats := manager.Stats()
	fmt.Printf("\n--- Final Stats After Shutdown ---\n")
	fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers) // Should be 0
	fmt.Printf("Burst Workers: %d\n", stats.BurstWorkers)   // Should be 0
	fmt.Printf("Queued Tasks: %d\n", stats.QueuedTasks)     // Should be 0
	fmt.Printf("Priority Tasks: %d\n", stats.PriorityTasks) // Should be 0
	fmt.Printf("Available Resources: %d\n", stats.AvailableResources) // Should be 0
	fmt.Printf("Tasks Completed: %d / %d\n", tasksCompleted.Load(), tasksSubmitted.Load())
	fmt.Println("---------------------------------")

	fmt.Println("Advanced usage example finished.")
}
