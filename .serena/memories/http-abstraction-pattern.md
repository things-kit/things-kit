# HTTP Abstraction Pattern

## Overview
Things-Kit implements a clean interface abstraction for HTTP servers, following the same pattern as the logger abstraction. This allows applications to be framework-agnostic while still having a sensible default implementation.

## Architecture

### Interface Layer (`module/http`)
- **Location**: `/Users/noxymon/Documents/Code/things-kit/module/http/`
- **Purpose**: Defines the contract that all HTTP implementations must follow
- **Dependencies**: None (interface-only module)
- **Contents**:
  - `Server` interface: Lifecycle methods (Start, Stop, Addr)
  - `Handler` interface: Route registration with `RegisterRoutes(router any)`
  - `Config` struct: Common configuration (Port, Host)

### Default Implementation (`httpgin`)
- **Location**: `/Users/noxymon/Documents/Code/things-kit/httpgin/`
- **Purpose**: Gin-based implementation of the HTTP server interface
- **Dependencies**: gin-gonic/gin, module/http, module/log, viper, fx
- **Contents**:
  - `GinServer`: Implements `http.Server` interface
  - `GinHandler`: Gin-specific handler interface
  - `Module`: Fx module variable
  - `AsGinHandler`: Helper for registering handlers
  - `Config`: Extends `http.Config` with Gin-specific options (Mode)

## Pattern Consistency

This follows the same pattern as the logging abstraction:

| Component | Logger Pattern | HTTP Pattern |
|-----------|---------------|--------------|
| Interface | `module/log` | `module/http` |
| Default Impl | `logging` (Zap) | `httpgin` (Gin) |
| Interface Type | `log.Logger` | `http.Server` |
| Handler Type | N/A | `http.Handler` |
| Usage | `logging.Module` | `httpgin.Module` |
| Helper | N/A | `httpgin.AsGinHandler()` |

## Key Design Decisions

### 1. Interface Accepts `any` for Router
```go
type Handler interface {
    RegisterRoutes(router any)
}
```
- Allows different implementations to pass framework-specific router types
- Gin implementation casts to `*gin.Engine`
- Chi implementation would cast to `*chi.Mux`
- Maintains type safety within each implementation

### 2. Separate Module for Implementation
- `module/http`: Interface only, no dependencies
- `httpgin`: Gin implementation with Gin dependencies
- Clean separation of concerns
- Easy to version independently

### 3. Config Inheritance
```go
type Config struct {
    httpmodule.Config `mapstructure:",squash"`  // Embed common config
    Mode              string `mapstructure:"mode"`  // Gin-specific
}
```
- Implementation configs embed the base `http.Config`
- Add implementation-specific fields as needed
- Consistent configuration structure

### 4. Fx Module Pattern
```go
var Module = fx.Module("httpgin",
    fx.Provide(
        NewConfig,
        NewGinServer,
        fx.Annotate(
            func(s *GinServer) httpmodule.Server { return s },
            fx.As(new(httpmodule.Server)),
        ),
    ),
    fx.Invoke(RunHttpServer),
)
```
- Provides `http.Server` interface, not concrete type
- Uses `fx.As` to bind implementation to interface
- Follows standard Fx module pattern

### 5. Handler Registration
```go
func AsGinHandler(constructor any) fx.Option {
    return fx.Provide(
        fx.Annotate(
            constructor,
            fx.As(new(GinHandler)),
            fx.ResultTags(`group:"http.handlers"`),
        ),
    )
}
```
- Generic helper for any handler constructor
- Groups handlers with `group:"http.handlers"`
- Type-safe through interface casting

## Benefits

1. **Framework Independence**: Applications don't depend on specific HTTP framework
2. **Easy Migration**: Swap Gin for Chi, Echo, or stdlib without changing app code
3. **Testing**: Mock `http.Server` interface for unit tests
4. **Flexibility**: Different services can use different HTTP frameworks
5. **Consistency**: Same pattern as other abstractions (logger, messaging)

## How to Use

### Using the Default (Gin)
```go
import (
    "github.com/things-kit/app"
    "github.com/things-kit/httpgin"
    "github.com/things-kit/logging"
    "github.com/things-kit/module/viperconfig"
)

func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        httpgin.Module,  // Use Gin implementation
        httpgin.AsGinHandler(handler.NewMyHandler),
    ).Run()
}
```

### Creating a Handler
```go
type MyHandler struct {
    logger log.Logger
}

func NewMyHandler(logger log.Logger) *MyHandler {
    return &MyHandler{logger: logger}
}

func (h *MyHandler) RegisterRoutes(engine *gin.Engine) {
    engine.GET("/hello", h.handleHello)
}
```

### Creating a Custom Implementation
1. Implement `http.Server` interface
2. Create handler interface for your framework
3. Create Fx module that provides your implementation
4. Create `AsXxxHandler` helper
5. Use in applications instead of `httpgin`

## Files Modified/Created

### Created
- `/Users/noxymon/Documents/Code/things-kit/httpgin/go.mod`
- `/Users/noxymon/Documents/Code/things-kit/httpgin/module.go`
- `/Users/noxymon/Documents/Code/things-kit/httpgin/README.md`
- `/Users/noxymon/Documents/Code/things-kit/module/http/README.md`

### Modified
- `/Users/noxymon/Documents/Code/things-kit/module/http/module.go`: Converted to interface-only
- `/Users/noxymon/Documents/Code/things-kit/module/http/go.mod`: Removed dependencies
- `/Users/noxymon/Documents/Code/things-kit/example/cmd/server/main.go`: Updated to use httpgin
- `/Users/noxymon/Documents/Code/things-kit/example/go.mod`: Updated dependencies
- `/Users/noxymon/Documents/Code/things-kit/go.work`: Added httpgin module
- `/Users/noxymon/Documents/Code/things-kit/README.md`: Documented abstraction pattern

## Testing Status
✅ All modules compile successfully
✅ Example service builds and runs
✅ Health endpoint: `curl http://localhost:8080/health` → `{"status":"healthy"}`
✅ Greet endpoint: `curl http://localhost:8080/greet/World` → `{"message":"Hello, World!"}`

## Future Possibilities

### Other HTTP Implementations
- `httpchi`: Chi router implementation
- `httpecho`: Echo framework implementation
- `httpfiber`: Fiber framework implementation
- `httpstdlib`: Standard library implementation

### Implementation-Specific Features
- Gin: Router groups, middleware, HTML rendering
- Chi: Middleware chaining, sub-routers, URL params
- Echo: Context-based, middleware, data binding
- Fiber: Express-like, extremely fast, WebSocket support

Each implementation can provide framework-specific features while maintaining compatibility with the core interface.
