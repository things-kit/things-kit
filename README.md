# Things-Kit

**A Modular, Opinionated Microservice Framework for Go, Built on Uber Fx.**

Things-Kit is a microservice framework for Go designed to bring the productivity and developer experience of frameworks like Spring Boot to the Go ecosystem. It is built entirely on the principles of dependency injection, leveraging [Uber's Fx library](https://github.com/uber-go/fx) to manage the application lifecycle and dependency graph.

## Status

✅ **Core Implementation Complete** - All foundational modules and infrastructure modules have been implemented and tested.

## Features

- **Modularity First:** Every piece of infrastructure (gRPC, Kafka, Redis, etc.) is a self-contained, versionable Go module
- **Dependency Injection:** Built on Uber Fx for clean, decoupled architecture
- **Interface-Based Design:** Program to abstractions, not implementations
- **Convention over Configuration:** Sensible defaults with full override capability
- **Lifecycle Aware:** Graceful startup and shutdown for all components
- **Developer Experience:** Minimal boilerplate through generic helpers

## Quick Start

See the [example service](./example/) for a complete working example.

```bash
cd example
cp config.example.yaml config.yaml
go run ./cmd/server
```

Then test it:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/greet/World
```

## Project Structure

This is a Go workspace containing multiple modules:

### Foundational Modules
- `app/` - Core application runner with lifecycle management

### Infrastructure Modules (in `module/`)

All framework-provided modules are organized under `module/`:

- `module/log/` - Logger interface abstraction
- `module/logging/` - Default Zap-based logger implementation ⭐
- `module/http/` - HTTP server interface abstraction (framework-agnostic)
- `module/httpgin/` - Default Gin-based HTTP server implementation ⭐
- `module/cache/` - Cache interface abstraction (key-value operations)
- `module/redis/` - Default Redis-based cache implementation ⭐
- `module/grpc/` - gRPC server with lifecycle management
- `module/sqlc/` - Database connection pool with lifecycle management
- `module/kafka/` - Kafka consumer implementing messaging interfaces
- `module/messaging/` - Message handling interface abstraction (Handler, Consumer, Producer)
- `module/viperconfig/` - Configuration management with Viper
- `module/testing/` - Testing utilities for integration tests

⭐ = Default implementation (interface + impl pattern)

### Example
- `example/` - Complete working example service demonstrating HTTP, logging, and DI

## Getting Started

### 1. Create Your Service

```go
package main

import (
    "github.com/things-kit/app"
    "github.com/things-kit/module/httpgin"
    "github.com/things-kit/module/logging"
    "github.com/things-kit/module/viperconfig"
)

func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        httpgin.Module,  // Use Gin HTTP implementation
        // Add your handlers and services
    ).Run()
}
```

### 2. Configure Your Service

Create `config.yaml`:
```yaml
http:
  port: 8080
  mode: release

logging:
  level: info
  encoding: json
```

### 3. Run Your Service

```bash
go run ./cmd/server
```

## Documentation

- [Detailed Plan](./plan.md) - Complete architectural documentation and implementation details
- [Example Service](./example/README.md) - Working example with usage patterns
- [Configuration Example](./config.example.yaml) - Full configuration reference

## Development

### Building All Modules

```bash
for dir in app logging module/* example; do
  if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
    (cd "$dir" && go build ./...)
  fi
done
```

### Testing

```bash
for dir in app logging module/* example; do
  if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
    (cd "$dir" && go test ./...)
  fi
done
```

### Workspace Management

```bash
go work sync  # Sync all module dependencies
```

## Module Design Philosophy

Each module follows a consistent pattern:

1. **Config Struct** - Define configuration with `mapstructure` tags
2. **NewConfig Function** - Load config from Viper with sensible defaults
3. **Module Variable** - Export `fx.Module` for easy consumption
4. **Lifecycle Integration** - Use `fx.Lifecycle` hooks for startup/shutdown
5. **Helper Functions** - Provide `AsXxx` helpers for easy service registration

See any module's implementation for examples.

## Interface Abstraction Pattern

Things-Kit follows the principle "Program to an Interface, Not an Implementation". This means:

### Module Organization

All framework-provided components are in `module/`:
- **Interface modules** define the contract (e.g., `module/log`, `module/http`, `module/cache`)
- **Implementation modules** provide default implementations (e.g., `module/logging`, `module/httpgin`, `module/redis`)
- **Alternative implementations** would live at the root level (e.g., `httpchi/`, `logging-logrus/`, `cachevalkey/`)

### Logger Abstraction
- **Interface**: `module/log` defines the `Logger` interface
- **Default Implementation**: `module/logging` provides a Zap-based implementation
- **Your Choice**: You can provide your own logger implementation (logrus, zerolog, etc.)

### HTTP Server Abstraction
- **Interface**: `module/http` defines `Server`, `Handler`, and `Config` interfaces
- **Default Implementation**: `module/httpgin` provides a Gin-based implementation
- **Your Choice**: You can use Chi, Echo, standard library, or any HTTP framework

### Cache Abstraction
- **Interface**: `module/cache` defines the `Cache` interface (Get, Set, Delete, etc.)
- **Default Implementation**: `module/redis` provides a Redis-based implementation
- **Your Choice**: You can use Valkey, Memcached, in-memory cache, or any key-value store

### Messaging Abstraction
- **Interfaces**: `module/messaging` defines `Handler`, `Consumer`, and `Producer` interfaces
- **Default Implementation**: `module/kafka` provides Kafka consumer implementation
- **Your Choice**: You can use RabbitMQ, NATS, AWS SQS, or any message broker

### Why This Matters

This design allows you to:

1. **Start Fast**: Use the default implementations (Zap logger, Gin HTTP server, Redis cache, Kafka messaging)
2. **Stay Flexible**: Swap implementations without changing your application code
3. **Test Easily**: Mock interfaces for unit testing
4. **Evolve Gracefully**: Migrate to different frameworks as needs change
5. **Clear Organization**: All framework modules in one place (`module/`)

### Example: Custom HTTP Implementation

```go
// Step 1: Implement the http.Server interface
type MyCustomServer struct {
    // your implementation
}

func (s *MyCustomServer) Start(ctx context.Context) error { /* ... */ }
func (s *MyCustomServer) Stop(ctx context.Context) error { /* ... */ }
func (s *MyCustomServer) Addr() string { /* ... */ }

// Step 2: Create an Fx module
var CustomHttpModule = fx.Module("mycustomhttp",
    fx.Provide(
        NewCustomConfig,
        NewMyCustomServer,
        fx.Annotate(
            func(s *MyCustomServer) http.Server { return s },
            fx.As(new(http.Server)),
        ),
    ),
    fx.Invoke(RunCustomHttpServer),
)

// Step 3: Use it in your application
func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        CustomHttpModule,  // Your custom implementation
        // ...
    ).Run()
}
```

See the [httpgin module](./module/httpgin/README.md) for a complete example of implementing the HTTP interface.

### Example: Custom Cache Implementation

```go
// Step 1: Implement the cache.Cache interface
type MyMemcachedCache struct {
    client *memcache.Client
}

func (c *MyMemcachedCache) Get(ctx context.Context, key string) (string, error) { /* ... */ }
func (c *MyMemcachedCache) Set(ctx context.Context, key, value string, exp time.Duration) error { /* ... */ }
// ... implement all Cache methods

// Step 2: Create an Fx module
var MemcachedModule = fx.Module("memcached",
    fx.Provide(
        NewMemcachedConfig,
        NewMemcachedCache,
        fx.Annotate(
            func(c *MyMemcachedCache) cache.Cache { return c },
            fx.As(new(cache.Cache)),
        ),
    ),
)

// Step 3: Use it in your application
func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        MemcachedModule,  // Your custom cache implementation
        // ...
    ).Run()
}
```

See the [redis module](./module/redis/README.md) and [cache interface](./module/cache/README.md) for complete examples.

## What Makes This Different?

- **True Modularity**: Each module is independently versionable
- **Pluggable Core**: Swap out any component (logger, config, etc.)
- **Context-Aware**: Embraces Go's context idiom for cancellation and tracing
- **Production Ready**: Built-in graceful shutdown, structured logging, configuration management
- **Developer Friendly**: Minimal boilerplate, clear patterns, comprehensive examples

## License

MIT
