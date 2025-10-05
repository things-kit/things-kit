// Package log defines the framework's standard Logger interface.
// It provides both context-aware and context-unaware logging methods
// for distributed tracing support.
package log

import "context"

// Field is a generic key-value pair for structured logging.
type Field struct {
	Key   string
	Value any
}

// Logger defines the standard logging interface for the framework.
// All framework modules and user code should depend on this interface,
// not on concrete implementations.
type Logger interface {
	// Context-unaware logging methods
	Info(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)

	// Context-aware logging methods for distributed tracing
	InfoC(ctx context.Context, msg string, fields ...Field)
	ErrorC(ctx context.Context, msg string, err error, fields ...Field)
	DebugC(ctx context.Context, msg string, fields ...Field)
	WarnC(ctx context.Context, msg string, err error, fields ...Field)
}
