# module/httpgin - Gin HTTP Server Implementation

This module provides a Gin-based implementation of the Things-Kit HTTP server interface. It's the default HTTP server implementation included in the framework, but users can swap it with their own implementations using different frameworks (Chi, Echo, standard library, etc.).

## Features

- Full implementation of the `http.Server` interface
- Gin framework with sensible defaults
- Graceful shutdown with configurable timeout
- Automatic route registration for handlers
- Context-aware logging integration
- Configuration via Viper
- Lifecycle management via Fx

## Installation

```bash
go get github.com/things-kit/module/httpgin
```

## Quick Start

### Basic Usage

```go
package main

import (
    "go.uber.org/fx"
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
    ).Run()
}
```

### Creating Handlers

Handlers must implement the `GinHandler` interface:

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/things-kit/module/log"
)

type MyHandler struct {
    logger log.Logger
}

func NewMyHandler(logger log.Logger) *MyHandler {
    return &MyHandler{logger: logger}
}

// RegisterRoutes implements GinHandler interface
func (h *MyHandler) RegisterRoutes(engine *gin.Engine) {
    engine.GET("/hello", h.handleHello)
    engine.POST("/data", h.handleData)
}

func (h *MyHandler) handleHello(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "Hello, World!",
    })
}

func (h *MyHandler) handleData(c *gin.Context) {
    var data map[string]interface{}
    if err := c.ShouldBindJSON(&data); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, data)
}
```

### Registering Handlers

Use the `AsGinHandler` helper to register handlers with the Fx graph:

```go
func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        httpgin.Module,
        
        // Register handlers
        httpgin.AsGinHandler(handler.NewMyHandler),
        httpgin.AsGinHandler(handler.NewAnotherHandler),
    ).Run()
}
```

## Configuration

Configure the server via `config.yaml`:

```yaml
http:
  port: 8080
  host: ""  # Empty for all interfaces, or specify like "localhost"
  mode: "release"  # "debug", "release", or "test"
```

Or via environment variables:

```bash
export HTTP_PORT=8080
export HTTP_HOST=localhost
export HTTP_MODE=release
```

### Configuration Options

| Field | Type   | Default     | Description                                |
|-------|--------|-------------|--------------------------------------------|
| port  | int    | 8080        | The port the server listens on             |
| host  | string | ""          | The host interface to bind to              |
| mode  | string | "release"   | Gin mode: "debug", "release", or "test"    |

## Advanced Usage

### Accessing the Gin Engine

If you need direct access to the Gin engine (for example, to add global middleware), you can inject the `*GinServer`:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/things-kit/httpgin"
    "go.uber.org/fx"
)

func setupMiddleware(server *httpgin.GinServer) {
    engine := server.Engine()
    
    // Add custom middleware
    engine.Use(func(c *gin.Context) {
        c.Header("X-Custom-Header", "MyValue")
        c.Next()
    })
}

func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        httpgin.Module,
        
        fx.Invoke(setupMiddleware),
        
        httpgin.AsGinHandler(handler.NewMyHandler),
    ).Run()
}
```

### Handler Groups

You can organize routes using Gin's router groups:

```go
func (h *MyHandler) RegisterRoutes(engine *gin.Engine) {
    // Public API
    public := engine.Group("/api/v1")
    {
        public.GET("/health", h.handleHealth)
        public.GET("/version", h.handleVersion)
    }
    
    // Admin API
    admin := engine.Group("/admin")
    admin.Use(h.authMiddleware())
    {
        admin.GET("/users", h.handleListUsers)
        admin.POST("/users", h.handleCreateUser)
    }
}
```

## Creating Custom HTTP Implementations

The beauty of Things-Kit's HTTP abstraction is that you can easily swap Gin for another framework. To create a custom implementation:

1. Implement the `http.Server` interface from `module/http`:

```go
type Server interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Addr() string
}
```

2. Implement the `http.Handler` interface:

```go
type Handler interface {
    RegisterRoutes(router any)
}
```

3. Create an Fx module that provides your implementation:

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

See the `httpgin` implementation for a complete example.

## Architecture

The `httpgin` module follows Things-Kit's interface abstraction pattern:

- **`module/http`**: Defines interfaces (`Server`, `Handler`, `Config`)
- **`httpgin`**: Provides Gin-based implementations
- **Your code**: Programs against the interface, can swap implementations

This design allows you to:
- Start with Gin for rapid development
- Switch to Chi, Echo, or standard library later
- Use different implementations for different services
- Test with mock implementations

## Testing

The module integrates with Things-Kit's testing utilities:

```go
package handler_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/things-kit/httpgin"
    "github.com/things-kit/module/testing"
)

func TestMyHandler(t *testing.T) {
    logger := testing.NewTestLogger(t)
    handler := NewMyHandler(logger)
    
    engine := gin.New()
    handler.RegisterRoutes(engine)
    
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/hello", nil)
    
    engine.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}
```

## Best Practices

1. **Keep handlers focused**: Each handler should be responsible for a single domain or resource
2. **Use dependency injection**: Inject services and dependencies via constructor
3. **Handle errors properly**: Always check and handle errors, return appropriate HTTP status codes
4. **Use middleware wisely**: Apply middleware at the appropriate level (global, group, or route)
5. **Validate input**: Always validate and sanitize user input
6. **Log appropriately**: Use the injected logger for consistent logging
7. **Return structured responses**: Use consistent JSON structures for all responses

## Examples

See the [example service](../example) for a complete working example of using httpgin with Things-Kit.

## License

MIT License - see LICENSE file for details
