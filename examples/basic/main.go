package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asaidimu/tasker"
)

// CalculatorResource represents a simple resource,
// in this case, just a placeholder.
type CalculatorResource struct{}

// onCreate for CalculatorResource - no actual setup needed
func createCalcResource() (*CalculatorResource, error) {
	fmt.Println("INFO: Creating CalculatorResource")
	return &CalculatorResource{}, nil
}

// onDestroy for CalculatorResource - no actual cleanup needed
func destroyCalcResource(r *CalculatorResource) error {
	fmt.Println("INFO: Destroying CalculatorResource")
	return nil
}

func main() {
	fmt.Println("--- Basic Usage: Simple Calculator ---")

	ctx := context.Background()

	// Configure the tasker for our CalculatorResource
	config := tasker.Config[*CalculatorResource]{
		OnCreate:    createCalcResource,
		OnDestroy:   destroyCalcResource,
		WorkerCount: 2, // Two base workers
		Ctx:         ctx,
		// No specific health check or burst settings for this basic example
	}

	// Create a new task manager
	manager, err := tasker.NewRunner[*CalculatorResource, int](config) // Tasks will return an int result
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop() // Ensure the manager is stopped gracefully

	fmt.Println("Queuing a simple addition task...")
	task1Start := time.Now()
	// Queue a task to perform addition
	go func() {
		sum, err := manager.QueueTask(func(r *CalculatorResource) (int, error) {
			// In a real scenario, 'r' could be a connection to a math service
			time.Sleep(50 * time.Millisecond) // Simulate some work
			a, b := 10, 25
			fmt.Printf("Worker processing: %d + %d\n", a, b)
			return a + b, nil
		})

		if err != nil {
			fmt.Printf("Task 1 failed: %v\n", err)
		} else {
			fmt.Printf("Task 1 (Addition) Result: %d (took %s)\n", sum, time.Since(task1Start))
		}
	}()

	fmt.Println("Queuing another subtraction task...")
	task2Start := time.Now()
	// Queue another task for subtraction
	go func() {
		difference, err := manager.QueueTask(func(r *CalculatorResource) (int, error) {
			time.Sleep(70 * time.Millisecond) // Simulate some work
			a, b := 100, 40
			fmt.Printf("Worker processing: %d - %d\n", a, b)
			return a - b, nil
		})

		if err != nil {
			fmt.Printf("Task 2 failed: %v\n", err)
		} else {
			fmt.Printf("Task 2 (Subtraction) Result: %d (took %s)\n", difference, time.Since(task2Start))
		}
	}()

	// Allow some time for tasks to complete
	time.Sleep(500 * time.Millisecond)

	stats := manager.Stats()
	fmt.Printf("\n--- Current Stats ---\n")
	fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers)
	fmt.Printf("Queued Tasks: %d\n", stats.QueuedTasks)
	fmt.Printf("Available Resources: %d\n", stats.AvailableResources)
	fmt.Println("----------------------")

	fmt.Println("Basic usage example finished.")
}
