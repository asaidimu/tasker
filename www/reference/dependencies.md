---
outline: "deep"
lastUpdated: true
editLink: true
sidebar: true
title: "Dependencies Reference"
description: "External and peer dependencies documentation"
---
# Dependencies

## External Dependencies

### github.com/asaidimu/tasker/v2

- **Purpose**: The core Go module for concurrent task management.
- **Installation**: `go get github.com/asaidimu/tasker/v2`

## Peer Dependencies

### context

- **Reason**: Required for `context.Context` propagation to tasks for cancellation and deadlines.
- **Version**: `Go Standard Library`

### errors

- **Reason**: Required for creating and handling errors.
- **Version**: `Go Standard Library`

### fmt

- **Reason**: Required for formatted I/O (e.g., printing messages).
- **Version**: `Go Standard Library`

### log

- **Reason**: Standard logging utility, though `tasker.Logger` allows custom integration.
- **Version**: `Go Standard Library`

### math

- **Reason**: Used for mathematical operations, e.g., in metrics calculations for ceil/floor.
- **Version**: `Go Standard Library`

### math/rand

- **Reason**: Used in examples for simulating random outcomes.
- **Version**: `Go Standard Library`

### sync

- **Reason**: Required for synchronization primitives like `sync.Mutex` and `sync.WaitGroup`.
- **Version**: `Go Standard Library`

### sync/atomic

- **Reason**: Required for atomic operations on shared counters.
- **Version**: `Go Standard Library`

### time

- **Reason**: Required for time-related operations, durations, and timestamps.
- **Version**: `Go Standard Library`

