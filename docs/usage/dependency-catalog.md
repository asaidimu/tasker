# Dependency Catalog

## Peer Dependencies

### context
- **Reason**: Required for managing goroutine cancellation, timeouts, and propagating application-wide signals. Crucial for graceful shutdown and task context management.
- **Version Requirements**: ``

### errors
- **Reason**: Used for creating, wrapping, and inspecting Go errors. Essential for robust error handling and type checking specific error conditions.
- **Version Requirements**: ``

### fmt
- **Reason**: Used for formatted I/O, particularly for constructing descriptive error messages (e.g., `fmt.Errorf`) and logging within examples.
- **Version Requirements**: ``

### log
- **Reason**: Provides basic logging capabilities, used primarily for internal warnings/errors and within examples.
- **Version Requirements**: ``

### sync
- **Reason**: Provides essential synchronization primitives like `sync.WaitGroup` for coordinating goroutine shutdown and `sync.Mutex` for protecting shared state.
- **Version Requirements**: ``

### sync/atomic
- **Reason**: Provides low-level atomic operations for concurrent, lock-free updates to shared counters and pointers, ensuring thread safety and performance for statistics.
- **Version Requirements**: ``

### time
- **Reason**: Used for managing durations, delays, and scheduling periodic checks (e.g., `BurstInterval`).
- **Version Requirements**: ``



---
*Generated using Gemini AI on 6/14/2025, 1:47:53 PM. Review and refine as needed.*