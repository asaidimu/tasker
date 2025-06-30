package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/asaidimu/tasker"
)

// ImageProcessor represents a resource for image manipulation.
type ImageProcessor struct {
	ID        int
	IsHealthy bool
}

// onCreate for ImageProcessor: simulates creating a connection to an image processing service.
func createImageProcessor() (*ImageProcessor, error) {
	id := rand.Intn(1000)
	fmt.Printf("INFO: Creating ImageProcessor %d\n", id)
	return &ImageProcessor{ID: id, IsHealthy: true}, nil
}

// onDestroy for ImageProcessor: simulates closing the connection.
func destroyImageProcessor(p *ImageProcessor) error {
	fmt.Printf("INFO: Destroying ImageProcessor %d\n", p.ID)
	return nil
}

// checkImageProcessorHealth: Custom health check.
// If the error is "processor_crash", consider the worker/resource unhealthy.
func checkImageProcessorHealth(err error) bool {
	if err != nil && err.Error() == "processor_crash" {
		fmt.Printf("WARN: Detected unhealthy error: %v. Worker will be replaced.\n", err)
		return false // This error indicates an unhealthy state
	}
	return true // Other errors are just task failures, not worker health issues
}

func main() {
	fmt.Println("\n--- Intermediate Usage: Image Processing ---")

	ctx := context.Background()

	config := tasker.Config[*ImageProcessor]{
		OnCreate:         createImageProcessor,
		OnDestroy:        destroyImageProcessor,
		WorkerCount:      2,           // Two base workers
		Ctx:              ctx,
		CheckHealth:      checkImageProcessorHealth, // Use custom health check
		MaxRetries:       1,           // One retry on unhealthy errors
		ResourcePoolSize: 1,           // Small pool for RunTask
	}

	manager, err := tasker.NewTaskManager[*ImageProcessor, string](config) // Tasks return string (e.g., image path)
	if err != nil {
		log.Fatalf("Error creating task manager: %v", err)
	}
	defer manager.Stop()

	// --- 1. Queue a normal image resize task ---
	fmt.Println("\nQueueing a normal image resize task...")
	go func() {
		result, err := manager.QueueTask(func(proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing normal resize for imageA.jpg\n", proc.ID)
			time.Sleep(150 * time.Millisecond)
			if rand.Intn(10) < 3 { // Simulate a healthy but non-fatal processing error 30% of the time
				return "", errors.New("image_corrupted")
			}
			return "imageA_resized.jpg", nil
		})
		if err != nil {
			fmt.Printf("Normal Resize Failed: %v\n", err)
		} else {
			fmt.Printf("Normal Resize Completed: %s\n", result)
		}
	}()

	// --- 2. Queue a high-priority thumbnail generation task ---
	fmt.Println("Queueing a high-priority thumbnail task...")
	go func() {
		result, err := manager.QueueTaskWithPriority(func(proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing HIGH PRIORITY thumbnail for video.mp4\n", proc.ID)
			time.Sleep(50 * time.Millisecond) // Faster processing
			return "video_thumbnail.jpg", nil
		})
		if err != nil {
			fmt.Printf("Priority Thumbnail Failed: %v\n", err)
		} else {
			fmt.Printf("Priority Thumbnail Completed: %s\n", result)
		}
	}()

	// --- 3. Simulate a task that causes an "unhealthy" error ---
	fmt.Println("Queueing a task that might crash a worker (unhealthy error)...")
	go func() {
		result, err := manager.QueueTask(func(proc *ImageProcessor) (string, error) {
			fmt.Printf("Worker %d processing problematic image (might crash)\n", proc.ID)
			time.Sleep(100 * time.Millisecond)
			if rand.Intn(2) == 0 { // 50% chance to simulate a crash
				return "", errors.New("processor_crash") // This triggers CheckHealth to return false
			}
			return "problematic_image_processed.jpg", nil
		})
		if err != nil {
			fmt.Printf("Problematic Image Task Failed: %v\n", err)
		} else {
			fmt.Printf("Problematic Image Task Completed: %s\n", result)
		}
	}()

	// --- 4. Run an immediate task (e.g., generate a preview for a user) ---
	fmt.Println("\nRunning an immediate task (using resource pool or temporary resource)...")
	immediateResult, immediateErr := manager.RunTask(func(proc *ImageProcessor) (string, error) {
		fmt.Printf("IMMEDIATE Task processing fast preview with processor %d\n", proc.ID)
		time.Sleep(20 * time.Millisecond)
		return "fast_preview.jpg", nil
	})
	if immediateErr != nil {
		fmt.Printf("Immediate Task Failed: %v\n", immediateErr)
	} else {
		fmt.Printf("Immediate Task Completed: %s\n", immediateResult)
	}

	// Give time for tasks and potential worker replacements
	time.Sleep(1 * time.Second)

	stats := manager.Stats()
	fmt.Printf("\n--- Current Stats ---\n")
	fmt.Printf("Base Workers: %d\n", stats.BaseWorkers)
	fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers)
	fmt.Printf("Queued Tasks: %d\n", stats.QueuedTasks)
	fmt.Printf("Priority Tasks: %d\n", stats.PriorityTasks)
	fmt.Printf("Available Resources: %d\n", stats.AvailableResources)
	fmt.Println("----------------------")

	fmt.Println("Intermediate usage example finished.")
}
