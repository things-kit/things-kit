# Things-Kit Architecture Patterns and Design Decisions

## Architectural Principles

### 1. Dependency Injection via Uber Fx
- **Why**: Provides clean separation of concerns, testability, and lifecycle management
- **How**: All components are registered with `fx.Provide` and consumed via constructor injection
- **Pattern**: Constructor functions accept dependencies as parameters, Fx resolves the dependency graph

### 2. Interface-Based Design
- **Core Principle**: Framework components depend on abstractions, not implementations
- **Benefits**:
  - Easy to swap implementations (e.g., replace Zap logger with custom logger)
  - Testability through mocking
  - Clear contracts between components
- **Key Interfaces**:
  - `log.Logger`: Logging abstraction
  - `messaging.Handler`: Message processing abstraction
  - Transport interfaces: gRPC ServiceRegistrar, HTTP Handler

### 3. Module Independence
- **Structure**: Each module is a separate Go module with its own `go.mod`
- **Benefits**:
  - Independent versioning
  - Minimal dependencies
  - Services only import what they need
  - Lean binaries
- **Tradeoff**: More complex workspace setup, but better long-term maintainability

### 4. Convention over Configuration
- **Philosophy**: Sensible defaults allow services to run with minimal configuration
- **Implementation**: Each module provides default Config values
- **Override Mechanism**: Viper enables file-based and environment variable configuration
- **Example**: gRPC server defaults to port 50051 but can be overridden

## Key Design Patterns

### Provider Pattern (Fx)
```go
var Module = fx.Module("name",
    fx.Provide(NewConfig),      // Provides *Config
    fx.Provide(NewComponent),   // Provides component
    fx.Invoke(StartComponent),  // Runs on startup
)
```
**Purpose**: Declarative component registration and lifecycle management

### Adapter Pattern
```go
type zapLoggerAdapter struct {
    logger *zap.Logger
}
// Implements log.Logger interface
```
**Purpose**: Wrap third-party libraries to conform to framework interfaces

### Registry Pattern
```go
fx.Provide(
    fx.Annotate(
        constructor,
        fx.ResultTags(`group:"grpc.services"`),
    ),
)
```
**Purpose**: Collect multiple implementations of a type (e.g., multiple gRPC services)

### Builder Pattern (Application Assembly)
```go
app.New(
    module1.Module,
    module2.Module,
    app.AsStartupFunc(migration),
).Run()
```
**Purpose**: Fluent API for assembling applications from modules

## Context Propagation Strategy

### Philosophy
- **Not Hidden**: Framework embraces Go's standard context.Context idiom
- **Created at Boundaries**: Transport modules (gRPC, HTTP, Kafka) create context
- **Explicit Propagation**: Developer responsible for passing context through layers

### Implementation
```
Transport Layer (grpc/http/kafka) → creates context.Context
    ↓
Service Layer → receives and propagates ctx
    ↓
Repository Layer → receives and propagates ctx
    ↓
Database/External Service → uses ctx
```

### Context-Aware Logging
- Logger interface has both context-aware (`InfoC`) and regular (`Info`) methods
- Context-aware methods extract trace IDs, request IDs for distributed tracing
- Developers choose when to use context-aware methods

## Configuration Architecture

### Decentralized Configuration
- **Pattern**: Each module loads its own config from shared Viper instance
- **Structure**: Nested YAML keys match module names (e.g., `grpc.port`)
- **Implementation**:
  ```go
  cfg := &Config{/* defaults */}
  v.UnmarshalKey("modulename", cfg)
  ```

### Environment Variable Override
- Automatic via Viper's `AutomaticEnv()`
- Key transformation: dots → underscores (e.g., `grpc.port` → `GRPC_PORT`)
- Precedence: Env vars > Config file > Defaults

### Third-Party Library Configuration
- Use struct embedding with `mapstructure:",squash"`
- Exposes full third-party config in YAML
- Example:
  ```go
  type Config struct {
      CustomField string
      kafka.ReaderConfig `mapstructure:",squash"`
  }
  ```

## Lifecycle Management

### Graceful Startup
1. Fx builds dependency graph
2. `OnStart` hooks run in dependency order
3. Synchronous startup functions run (via `AsStartupFunc`)
4. Background services start (servers, consumers)
5. Application signals ready

### Graceful Shutdown
1. Shutdown signal received (SIGINT, SIGTERM)
2. `OnStop` hooks run in reverse dependency order
3. Servers stop accepting new requests
4. In-flight requests complete (with timeout)
5. Connections closed
6. Application exits

### Hook Implementation
```go
lc.Append(fx.Hook{
    OnStart: func(ctx context.Context) error {
        // Initialize resources
        return nil
    },
    OnStop: func(ctx context.Context) error {
        // Cleanup resources
        return nil
    },
})
```

## Testing Strategy

### Unit Testing
- Mock interfaces (e.g., mock logger, mock database)
- Test business logic in isolation
- No Fx required for pure unit tests

### Integration Testing
- Use `testing.RunTest` to create test Fx application
- Real dependencies where practical
- Test logger logs to testing.T

### Example
```go
func TestService(t *testing.T) {
    testing.RunTest(t,
        logging.Module,
        fx.Provide(NewMockDB),
        fx.Provide(service.NewUserService),
        fx.Invoke(func(svc *service.UserService) {
            // Test the service
        }),
    )
}
```

## Helper Functions Design

### Generic Helpers (AsGrpcService, AsHttpHandler)
- **Purpose**: Reduce boilerplate in application assembly
- **Implementation**: Use `fx.Annotate` for type-safe generic behavior
- **Type**: Defined as `any` but type-safe due to Fx's reflection
- **Benefit**: Clean, declarative service registration

### Startup Function Helper (AsStartupFunc)
- **Purpose**: Run synchronous initialization logic before services start
- **Use Cases**: Database migrations, cache warming, validation
- **Implementation**: Wraps `fx.Invoke` for clarity

## Error Handling Philosophy

### Explicit Over Implicit
- Return errors, don't panic
- Let caller decide error handling strategy
- Framework doesn't hide errors

### Error Wrapping
- Always provide context: `fmt.Errorf("operation failed: %w", err)`
- Preserve error chain for debugging

### Logging Errors
- Log errors at boundaries (where they're handled)
- Don't log intermediate errors (avoid log spam)
- Use structured logging with context

## Extension Points

### Custom Logger
Implement `log.Logger` interface, provide via `fx.Provide`

### Custom Transport
Create module with lifecycle hooks, register handlers via group pattern

### Custom Configuration Source
Replace `viperconfig.Module` with custom provider

### Middleware/Interceptors
Framework modules expose configuration for gRPC interceptors, Gin middleware
