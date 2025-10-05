# Things-Kit Coding Conventions and Style Guide

## Package Structure
- Each module is a separate Go module with its own `go.mod`
- Package names match directory names (e.g., `package grpc` in `module/grpc/`)
- Main module code typically in `module.go` file

## Naming Conventions

### Interfaces
- Core abstractions are defined as interfaces (e.g., `log.Logger`, `messaging.Handler`)
- Interface names are nouns, not prefixed with "I" (Go idiom)
- Single-method interfaces often end with "-er" suffix (e.g., `Handler`)

### Structs
- Config structs named `Config` within each module's package
- Private implementation structs use descriptive names (e.g., `zapLoggerAdapter`)
- Use struct tags for configuration unmarshaling: `mapstructure:"field_name"`

### Functions and Methods
- Constructors prefixed with `New` (e.g., `NewViper`, `NewUserService`)
- Context-aware methods suffixed with `C` (e.g., `InfoC`, `ErrorC`)
- Helper functions for Fx registration prefixed with `As` (e.g., `AsGrpcService`, `AsHttpHandler`)

### Variables
- Module exports named `Module` (e.g., `var Module = fx.Module(...)`)
- Use camelCase for local variables
- Use PascalCase for exported names

## Code Organization Patterns

### Module Pattern
Every infrastructure module follows this structure:
```go
package modulename

import "go.uber.org/fx"

// Module provides the module to the application
var Module = fx.Module("modulename",
    fx.Provide(NewConfig),
    fx.Provide(NewComponent),
    fx.Invoke(RunComponent),
)

type Config struct {
    Field string `mapstructure:"field"`
}

func NewConfig(v *viper.Viper) *Config {
    cfg := &Config{/* defaults */}
    v.UnmarshalKey("modulename", cfg)
    return cfg
}

// Component implementation...
```

### Dependency Injection Pattern
- All dependencies injected via constructor parameters
- Use Fx `In` and `Out` structs for complex dependency graphs
- Tag group dependencies: `group:"grpc.services"`
- Use `fx.Annotate` for binding implementations to interfaces

### Context Handling
- Always accept `context.Context` as first parameter in I/O operations
- Propagate context through all layers (transport → service → repository → database)
- Framework creates context at transport boundaries (gRPC, HTTP, Kafka)
- Never hide context - it's part of the Go idiom

### Lifecycle Management
- Use `fx.Lifecycle` hooks for startup/shutdown logic
- OnStart: Initialize connections, start servers
- OnStop: Gracefully close connections, stop servers
- Return errors from hooks to signal startup/shutdown failures

## Documentation
- Package-level documentation required for all packages
- Exported types, functions, and methods should have doc comments
- Doc comments start with the name being documented
- Examples provided in plan.md for complex usage patterns

## Error Handling
- Always wrap errors with context: `fmt.Errorf("action failed: %w", err)`
- Return errors, don't panic (except in truly exceptional cases)
- Log errors at the boundary where they're handled, not where they occur

## Configuration
- Each module defines its own `Config` struct
- Use `mapstructure` tags for Viper unmarshaling
- Provide sensible defaults in constructor
- Support environment variable overrides (automatic via Viper)
- Config keys use dot notation: `grpc.port`, `logging.level`

## Testing
- Use `module/testing.RunTest` for integration tests
- Test logger provided automatically in test context
- Mock interfaces, not concrete implementations
- Tests organized alongside source files

## Import Organization
```go
import (
    // Standard library
    "context"
    "fmt"
    
    // Third-party libraries
    "github.com/spf13/viper"
    "go.uber.org/fx"
    
    // Things-Kit modules
    "github.com/things-kit/module/log"
    
    // Local packages
    "myapp/internal/service"
)
```

## Design Patterns

### Interface Segregation
- Small, focused interfaces (e.g., `log.Logger` has one responsibility)
- Clients depend on abstractions, not implementations

### Provider Pattern
- Use Fx's `fx.Provide` to supply components to the dependency graph
- Use `fx.As` to bind implementations to interfaces
- Use `fx.Annotate` for advanced binding scenarios

### Generic Helpers
- `AsGrpcService`, `AsHttpHandler`, `AsStartupFunc` reduce boilerplate
- Type-safe despite being defined as `any` due to Fx reflection

### Pluggable Implementations
- Core components (like logger) can be swapped by providing alternative implementations
- Framework modules depend on interfaces from abstraction modules
