# Dependency Catalog

## External Dependencies

### github.com/asaidimu/tasker/v2
- **Purpose**: The core Go module for concurrent task management.
- **Installation**: `go get github.com/asaidimu/tasker/v2`
- **Version Compatibility**: ``

## Peer Dependencies

### context
- **Reason**: Required for `context.Context` propagation to tasks for cancellation and deadlines.
- **Version Requirements**: `Go Standard Library`

### errors
- **Reason**: Required for creating and handling errors.
- **Version Requirements**: `Go Standard Library`

### fmt
- **Reason**: Required for formatted I/O (e.g., printing messages).
- **Version Requirements**: `Go Standard Library`

### log
- **Reason**: Standard logging utility, though `tasker.Logger` allows custom integration.
- **Version Requirements**: `Go Standard Library`

### math
- **Reason**: Used for mathematical operations, e.g., in metrics calculations for ceil/floor.
- **Version Requirements**: `Go Standard Library`

### math/rand
- **Reason**: Used in examples for simulating random outcomes.
- **Version Requirements**: `Go Standard Library`

### sync
- **Reason**: Required for synchronization primitives like `sync.Mutex` and `sync.WaitGroup`.
- **Version Requirements**: `Go Standard Library`

### sync/atomic
- **Reason**: Required for atomic operations on shared counters.
- **Version Requirements**: `Go Standard Library`

### time
- **Reason**: Required for time-related operations, durations, and timestamps.
- **Version Requirements**: `Go Standard Library`



---
*Generated using Gemini AI on 6/30/2025, 9:35:09 PM. Review and refine as needed.*