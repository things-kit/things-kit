# Things-Kit Example Service

This is a simple example service demonstrating how to use the Things-Kit framework.

## Features

- HTTP REST API using Gin
- Structured logging with Zap
- Configuration management with Viper
- Dependency injection with Uber Fx
- Graceful shutdown

## Running the Example

### 1. Create a configuration file

```bash
cp config.example.yaml config.yaml
```

### 2. Run the service

```bash
go run ./cmd/server
```

The service will start on port 8080 (configurable in config.yaml).

### 3. Test the endpoints

**Health check:**
```bash
curl http://localhost:8080/health
```

**Greeting:**
```bash
curl http://localhost:8080/greet/World
```

Response:
```json
{
  "message": "Hello, World!"
}
```

## Project Structure

```
example/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── service/
│   │   └── greeting_service.go  # Business logic
│   └── handler/
│       └── greeting_handler.go  # HTTP handlers
├── config.example.yaml          # Configuration example
└── go.mod                       # Module dependencies
```

## How It Works

### 1. Application Assembly (main.go)

The `main.go` file assembles the application using the Things-Kit framework:

```go
app.New(
    // Foundational modules
    viperconfig.Module,  // Configuration
    logging.Module,       // Logging

    // Infrastructure
    httpmodule.Module,    // HTTP server

    // Application components
    service.NewGreetingService,
    httpmodule.AsHttpHandler(handler.NewGreetingHandler),
).Run()
```

### 2. Service Layer (greeting_service.go)

The service contains business logic and depends on abstractions:

```go
type GreetingService struct {
    logger log.Logger  // Depends on interface, not implementation
}

func NewGreetingService(logger log.Logger) *GreetingService {
    return &GreetingService{logger: logger}
}
```

### 3. Handler Layer (greeting_handler.go)

The handler implements the HTTP Handler interface:

```go
type GreetingHandler struct {
    service *service.GreetingService
    logger  log.Logger
}

func (h *GreetingHandler) RegisterRoutes(engine *gin.Engine) {
    engine.GET("/greet/:name", h.handleGreet)
}
```

## Adding More Features

### Adding a Database

1. Add the sqlc module to `main.go`:
```go
import sqlcmodule "github.com/things-kit/module/sqlc"

app.New(
    // ... other modules
    sqlcmodule.Module,
    // ... your services
)
```

2. Inject `*sql.DB` into your service:
```go
func NewGreetingService(logger log.Logger, db *sql.DB) *GreetingService {
    return &GreetingService{logger: logger, db: db}
}
```

3. Configure in `config.yaml`:
```yaml
db:
  dsn: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

### Adding gRPC

1. Define your protobuf service
2. Generate Go code: `protoc --go_out=. --go-grpc_out=. proto/service.proto`
3. Add the grpc module and register your service:

```go
import grpcmodule "github.com/things-kit/module/grpc"

app.New(
    // ... other modules
    grpcmodule.Module,
    grpcmodule.AsGrpcService(service.NewUserService, pb.RegisterUserServiceServer),
)
```

### Adding Redis

```go
import redismodule "github.com/things-kit/module/redis"

app.New(
    // ... other modules
    redismodule.Module,
    // Inject *redis.Client into your services
)
```

### Adding Kafka Consumer

```go
import kafkamodule "github.com/things-kit/module/kafka"
import "github.com/things-kit/module/messaging"

// Implement messaging.Handler
type MyMessageHandler struct {
    logger log.Logger
}

func (h *MyMessageHandler) Handle(ctx context.Context, msg messaging.Message) error {
    // Process message
    return nil
}

app.New(
    // ... other modules
    kafkamodule.ConsumerModule,
    fx.Provide(NewMyMessageHandler),
)
```

## Configuration

All modules support configuration through:
1. **config.yaml** - Primary configuration file
2. **Environment variables** - Override config values (e.g., `HTTP_PORT=9090`)

Environment variables use uppercase with underscores replacing dots:
- `http.port` → `HTTP_PORT`
- `db.dsn` → `DB_DSN`
- `logging.level` → `LOGGING_LEVEL`

## Testing

Run tests:
```bash
go test ./...
```

## Learn More

See the [main project plan](../plan.md) for detailed documentation on:
- Architecture principles
- Module design patterns
- Advanced configuration
- Testing strategies
