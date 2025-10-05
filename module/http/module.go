// Package http defines framework-level HTTP server abstractions.
// This package provides interfaces that HTTP implementations must satisfy,
// allowing users to swap HTTP frameworks (Gin, Chi, Echo, stdlib, etc.) while
// maintaining compatibility with the framework.
//
// For a production-ready implementation, see the httpgin package.
package http

import (
	"context"
)

// Server represents an HTTP server that can be started and stopped.
// Implementations should handle the lifecycle of the HTTP server including
// graceful shutdown and proper error handling.
type Server interface {
	// Start begins listening for HTTP requests.
	// This should be non-blocking and return immediately after the server starts.
	// The implementation should start a goroutine for the actual serving.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the HTTP server.
	// It should wait for in-flight requests to complete within the context deadline.
	// Implementations should respect the context's cancellation/timeout.
	Stop(ctx context.Context) error

	// Addr returns the address the server is listening on (e.g., ":8080").
	Addr() string
}

// Handler represents a component that can register HTTP routes.
// The router parameter type depends on the HTTP implementation being used.
// For example, *gin.Engine for Gin, chi.Router for Chi, etc.
type Handler interface {
	// RegisterRoutes registers this handler's routes with the HTTP router.
	// The router parameter should be cast to the appropriate type by the implementation.
	RegisterRoutes(router any)
}

// Config holds common HTTP server configuration.
// Specific implementations may embed this struct and add framework-specific fields.
type Config struct {
	Port int    `mapstructure:"port"` // Port to listen on
	Host string `mapstructure:"host"` // Host to bind to (empty = all interfaces)
}
