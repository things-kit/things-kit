# module/http - HTTP Server Interface

This module defines the HTTP server interface abstraction for Things-Kit. It contains **only interfaces**, no implementation.

## Purpose

The `module/http` package defines the contract that all HTTP server implementations must follow. This allows applications to program against a stable interface while being free to choose any HTTP framework implementation.

## Interfaces

### Server

The `Server` interface defines the lifecycle methods for an HTTP server:

```go
type Server interface {
    Start(ctx context.Context) error  // Start the HTTP server
    Stop(ctx context.Context) error   // Stop the server gracefully
    Addr() string                     // Get the server's listening address
}
```

### Handler

The `Handler` interface defines how HTTP handlers register routes:

```go
type Handler interface {
    RegisterRoutes(router any)  // Register routes with the framework's router
}
```

The `router` parameter is typed as `any` to allow different HTTP frameworks to pass their own router types (e.g., `*gin.Engine`, `*chi.Mux`, `*echo.Echo`, etc.).

### Config

The `Config` struct defines common HTTP server configuration:

```go
type Config struct {
    Port int    `mapstructure:"port"`  // Server port (default: 8080)
    Host string `mapstructure:"host"`  // Server host (default: "" for all interfaces)
}
```

## Available Implementations

### module/httpgin (Default)

The [httpgin module](../httpgin/) provides a Gin-based implementation of the HTTP server interface. It's the recommended default for most applications.

```go
import "github.com/things-kit/module/httpgin"

app.New(
    viperconfig.Module,
    logging.Module,
    httpgin.Module,  // Use Gin implementation
    // ...
)
```

### Custom Implementations

You can create your own HTTP server implementation using any framework:

- **Chi**: Lightweight router with standard library compatibility
- **Echo**: High-performance framework with middleware
- **Fiber**: Express-inspired framework for speed
- **Standard Library**: Full control with `net/http`

## Creating Your Own Implementation

To create a custom HTTP server implementation:

1. **Implement the `Server` interface**:
   ```go
   type MyServer struct {
       config Config
       // ... your fields
   }
   
   func (s *MyServer) Start(ctx context.Context) error { /* ... */ }
   func (s *MyServer) Stop(ctx context.Context) error { /* ... */ }
   func (s *MyServer) Addr() string { /* ... */ }
   ```

2. **Define your handler interface** (optional):
   ```go
   type MyHandler interface {
       RegisterRoutes(router *mypkg.Router)
   }
   ```

3. **Create an Fx module**:
   ```go
   var Module = fx.Module("myhttp",
       fx.Provide(
           NewConfig,
           NewMyServer,
           fx.Annotate(
               func(s *MyServer) http.Server { return s },
               fx.As(new(http.Server)),
           ),
       ),
       fx.Invoke(RunHttpServer),
   )
   ```

4. **Register handlers with Fx**:
   ```go
   func AsMyHandler(constructor any) fx.Option {
       return fx.Provide(
           fx.Annotate(
               constructor,
               fx.As(new(MyHandler)),
               fx.ResultTags(`group:"http.handlers"`),
           ),
       )
   }
   ```

See the [httpgin implementation](../../httpgin/) as a reference example.

## Design Philosophy

This interface abstraction follows the principle **"Program to an Interface, Not an Implementation"**:

- **Applications** depend on `module/http` (interface)
- **Implementations** provide `http.Server` (httpgin, httpchi, etc.)
- **Flexibility**: Swap implementations without changing application code

This is the same pattern used for the logger abstraction:
- Interface: `module/log` → Implementation: `logging` (Zap)
- Interface: `module/http` → Implementation: `httpgin` (Gin)

## Benefits

1. **Framework Independence**: Your application code doesn't depend on a specific HTTP framework
2. **Testing**: Easy to mock the `Server` interface for unit tests
3. **Migration**: Switch frameworks without rewriting handler logic
4. **Experimentation**: Try different frameworks for performance or features
5. **Consistency**: All HTTP implementations follow the same contract

## Usage in Applications

Applications should:

1. Import the interface package for types:
   ```go
   import httpmodule "github.com/things-kit/module/http"
   ```

2. Import an implementation package for the actual server:
   ```go
   import "github.com/things-kit/module/httpgin"
   ```

3. Use the implementation's module and helpers:
   ```go
   app.New(
       httpgin.Module,
       httpgin.AsGinHandler(NewMyHandler),
       // ...
   )
   ```

## Examples

See the [example service](../../../example/) for a complete working example using the httpgin implementation.

## License

MIT License - see LICENSE file for details
