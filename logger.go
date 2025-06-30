// Package tasker provides a robust and flexible task management system.
package tasker

import ()

// Logger defines the interface for logging messages from the TaskManager.
// This allows users to integrate their own preferred logging library.
type Logger interface {
	// Debugf logs a message at the debug level.
	Debugf(format string, args ...any)
	// Infof logs a message at the info level.
	Infof(format string, args ...any)
	// Warnf logs a message at the warning level.
	Warnf(format string, args ...any)
	// Errorf logs a message at the error level.
	Errorf(format string, args ...any)
}

// noOpLogger is a basic implementation of the Logger interface.
// To see logs, users should provide their own logger implementation
type noOpLogger struct {
}

// newDefaultLogger creates a new logger that writes to os.Stderr.
func newNoOpLogger() Logger {
	return &noOpLogger{
	}
}

func (l *noOpLogger) Debugf(format string, args ...any) {
}

func (l *noOpLogger) Infof(format string, args ...any) {
}

func (l *noOpLogger) Warnf(format string, args ...any) {
}

func (l *noOpLogger) Errorf(format string, args ...any) {
}
