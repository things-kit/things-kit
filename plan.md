**Project Plan: Things-Kit**

**A Modular, Opinionated Microservice Framework for Go, Built on Uber Fx.**

  - **Version:** 1.0
  - **Date:** October 5, 2025
  - **Author:** AI Agent

**1. Introduction & Vision**

**Things-Kit** is a microservice framework for Go designed to bring the productivity and developer experience of frameworks like Spring Boot to the Go ecosystem. It is built entirely on the principles of dependency injection, leveraging [Uber's Fx library](https://github.com/uber-go/fx) to manage the application lifecycle and dependency graph.

The vision is to create a "batteries-included, but optional" ecosystem where developers can build robust, production-ready services by simply composing pre-built, independent modules. This avoids the boilerplate of wiring up common components while retaining Go's idiomatic simplicity.

**2. Core Philosophy**

The framework is guided by the following principles:

1.  **Modularity First:** Every piece of infrastructure (gRPC, Kafka, Redis, etc.) is a self-contained, versionable Go module. Services only import and use what they need, keeping binaries lean.
2.  **Dependency Injection is King:** The framework's core is Uber Fx. All components are provided to and consumed from a central dependency graph. This promotes testability and clean, decoupled architecture.
3.  **Program to an Interface:** Core components are defined by abstractions (interfaces), not concrete implementations. This allows developers to easily swap out default components (like the logger) with their own custom implementations.
4.  **Convention over Configuration:** Modules come with sensible defaults. You can run a service with minimal configuration, but every parameter is overridable for production environments.
5.  **Lifecycle Aware:** All modules that manage connections or background processes (e.g., servers, consumers) are integrated with the Fx lifecycle for graceful startup and shutdown.
6.  **Developer Experience:** Reduce boilerplate through generic helpers (AsGrpcService, AsHttpHandler) and a consistent, predictable structure across all modules.

**3. Architecture Overview**

Things-Kit uses a multi-module Go workspace. This allows each component to be developed and versioned independently while enabling easy local development.

The architecture consists of three main layers:

1.  **Application Core:** The `app` package, which provides the main application runner (`app.New(...).Run()`).
2.  **Foundational Modules:** Core, cross-cutting concerns that most services will use.
      - `log`: **(Abstraction)** Defines the core `log.Logger` interface.
      - `logging`: **(Implementation)** Provides the default, Zap-based implementation of the `log.Logger` interface.
      - `viperconfig`: Provides a `*viper.Viper` instance for decentralized configuration.
3.  **Optional Infrastructure Modules:** Pluggable components that a service can choose to use. They depend on the foundational abstractions (like `log.Logger`).

**3.1. Example Service**

The following snippets demonstrate how a developer would assemble a complete gRPC service. Note how the business logic depends on the `log.Logger` interface and uses its new context-aware methods.

**internal/service/user\_service.go**

``` go
package service

import (
 "context"
 "my-awesome-service/pb"

 "[github.com/things-kit/module/log](https://github.com/things-kit/module/log)"
)

// UserService implements the gRPC server logic.
// It depends on the Logger interface, not a concrete logger.
type UserService struct {
 pb.UnimplementedUserServiceServer
 logger log.Logger
}

// NewUserService is the constructor for our service. Fx will automatically
// inject a component that satisfies the log.Logger interface.
func NewUserService(logger log.Logger) *UserService {
 return &UserService{logger: logger}
}

// GetUser is our gRPC method implementation.
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
 // Use the context-aware logger to automatically include trace IDs, etc.
 s.logger.InfoC(ctx, "Handling GetUser request", log.Field{Key: "user_id", Value: req.UserId})
 // In a real app, you would fetch from a database here.
 return &pb.UserResponse{User: &pb.User{Id: req.UserId, Email: "test@example.com"}}, nil
}

```

**cmd/server/main.go**

``` go
package main

import (
 // Framework Imports
 "[github.com/things-kit/app](https://github.com/things-kit/app)"
 "[github.com/things-kit/logging](https://github.com/things-kit/logging)" // Default logger implementation
 grpcmodule "[github.com/things-kit/module/grpc](https://github.com/things-kit/module/grpc)"
 "[github.com/things-kit/module/viperconfig](https://github.com/things-kit/module/viperconfig)"

 // Internal Application Imports
 "my-awesome-service/internal/service"
 "my-awesome-service/pb"
)

// --- Main Application Assembly ---
func main() {
 app.New(
  // 1. Provide the foundational modules.
  viperconfig.Module,
  logging.Module, // Provides the default Zap-based logger

  // 2. Provide the infrastructure modules your service needs.
  grpcmodule.Module,

  // 3. Provide your application's components.
  grpcmodule.AsGrpcService(service.NewUserService, pb.RegisterUserServiceServer),
 ).Run()
}

```

**3.2. Project File Structure**

The framework is designed to be completely separate from your business logic. For a new developer, the framework is simply a third-party dependency managed by Go modules.

``` 
my-awesome-service/       # YOUR APPLICATION ROOT
├── cmd/
│   └── server/
│       └── main.go       # Main assembly (imports and uses the framework)
├── internal/
│   ├── service/          # Business logic implementation (e.g., userService)
│   │   └── user_service.go
│   └── repository/       # Data access logic (e.g., using sqlc)
│       └── user_repo.go
├── pb/                   # Generated Protobuf code
│   └── user.pb.go
├── go.mod                # Your application's go.mod (lists '[github.com/things-kit](https://github.com/things-kit)' as a 'require' dependency)
├── go.sum
└── config.yaml           # Your application's specific configuration

```

**3.3. Handling Synchronous Startup Logic**

To ensure tasks like database migrations run in a guaranteed order before background processes start, the framework provides a helper, `app.AsStartupFunc`.

**internal/repository/migrations.go**

``` go
package repository

import (
 "context"
 "database/sql"
 "[github.com/things-kit/module/log](https://github.com/things-kit/module/log)"
)

// RunMigrations is a synchronous startup function. It should accept a context
// to handle potential timeouts during long-running migrations.
func RunMigrations(ctx context.Context, db *sql.DB, logger log.Logger) error {
 logger.InfoC(ctx, "Running database migrations...", log.Field{Key: "component", Value: "migrations"})
 // ... migration logic using the context ...
 logger.InfoC(ctx, "Database migrations completed.", log.Field{Key: "component", Value: "migrations"})
 return nil
}

```

**cmd/server/main.go**

``` go
package main

import (
 "[github.com/things-kit/app](https://github.com/things-kit/app)"
 sqlcmodule "[github.com/things-kit/module/sqlc](https://github.com/things-kit/module/sqlc)"
 "my-awesome-service/internal/repository"
)

func main() {
 app.New(
  // ... other modules
  sqlcmodule.Module,
  app.AsStartupFunc(repository.RunMigrations),
  // ... your services
 ).Run()
}

```

**3.4. The Role of context.Context**

The framework does **not** hide `context.Context`. It embraces the standard Go idiom of explicit context propagation.

1.  **Framework Manages Context at the Boundary:** The transport modules (like grpc, http, kafka) are responsible for creating the initial `context.Context` for each incoming request or message. This context carries deadlines, cancellation signals, and tracing information.
2.  **Framework Passes the Context to Your Code:** The framework immediately passes this context as the first argument to your handler or service method (e.g., `GetUser(ctx context.Context, ...)`).
3.  **You Propagate the Context Explicitly:** It is the developer's responsibility to pass the received `ctx` down through all subsequent layers of the application (e.g., from service to repository to database driver).

This design ensures that standard Go tooling and patterns for cancellation and distributed tracing work seamlessly.

``` go
// Your service layer receives the context from the framework's transport module.
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
    // You pass the context down to the repository layer.
    user, err := s.userRepo.FindByID(ctx, req.UserId)
    // ...
}

// Your repository layer receives the context from the service layer.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    // You use the context in the final I/O call.
    row := r.db.QueryRowContext(ctx, "SELECT ... FROM users WHERE id=$1", id)
    // ...
}

```

**4. Module Breakdown & Implementation Snippets**
**4.1. Foundational Modules**
**app**

The core application runner.

``` go
// app/app.go
package app

import "go.uber.org/fx"

// ... New() and Run() methods ...

// AsStartupFunc registers a function to be run synchronously on startup.
func AsStartupFunc(constructor any) fx.Option {
 return fx.Invoke(constructor)
}

```

**log (Abstraction)**

Defines the framework's standard `Logger` interface. It includes context-aware methods (`...C`) for distributed tracing support.

``` go
// module/log/interface.go
package log

import "context"

// Field is a generic key-value pair for structured logging.
type Field struct {
 Key   string
 Value any
}

// Logger defines the standard logging interface for the framework.
type Logger interface {
 // Context-unaware logging methods
 Info(msg string, fields ...Field)
 Error(msg string, err error, fields ...Field)
 Debug(msg string, fields ...Field)
 Warn(msg string, fields ...Field)

 // Context-aware logging methods
 InfoC(ctx context.Context, msg string, fields ...Field)
 ErrorC(ctx context.Context, msg string, err error, fields ...Field)
 DebugC(ctx context.Context, msg string, fields ...Field)
 WarnC(ctx context.Context, msg string, err error, fields ...Field)
}

```

**logging (Zap Implementation)**

Provides the default, Zap-based implementation of the `log.Logger` interface.

``` go
// logging/module.go
package logging

import (
 "context"
 "[github.com/things-kit/module/log](https://github.com/things-kit/module/log)"
 "[github.com/spf13/viper](https://github.com/spf13/viper)"
 "go.uber.org/fx"
 "go.uber.org/zap"
)

// Module provides the default, Zap-based implementation of the log.Logger interface.
var Module = fx.Provide(
 NewZapLoggerAdapter,
 fx.As(new(log.Logger)), // Binds the concrete implementation to the interface
)

// zapLoggerAdapter wraps *zap.Logger to implement the log.Logger interface.
type zapLoggerAdapter struct { logger *zap.Logger }

func NewZapLoggerAdapter(v *viper.Viper) (*zapLoggerAdapter, error) { /* ... */ }

// Implement the log.Logger interface methods...
func (a *zapLoggerAdapter) Info(msg string, fields ...log.Field)  { /* ... */ }
func (a *zapLoggerAdapter) Error(msg string, err error, fields ...log.Field) { /* ... */ }
func (a *zapLoggerAdapter) Debug(msg string, fields ...log.Field) { /* ... */ }
func (a *zapLoggerAdapter) Warn(msg string, fields ...log.Field)  { /* ... */ }
func (a *zapLoggerAdapter) InfoC(ctx context.Context, msg string, fields ...log.Field)  { /* ... */ }
func (a *zapLoggerAdapter) ErrorC(ctx context.Context, msg string, err error, fields ...log.Field) { /* ... */ }
func (a *zapLoggerAdapter) DebugC(ctx context.Context, msg string, fields ...log.Field) { /* ... */ }
func (a *zapLoggerAdapter) WarnC(ctx context.Context, msg string, err error, fields ...log.Field)  { /* ... */ }

```

**viperconfig**

Provides the shared `*viper.Viper` instance.

``` go
// module/viperconfig/module.go
package viperconfig

import (
 "strings"
 "[github.com/spf13/viper](https://github.com/spf13/viper)"
 "go.uber.org/fx"
)

var Module = fx.Provide(NewViper)

func NewViper() (*viper.Viper, error) {
 v := viper.New()
 v.SetConfigName("config")
 v.SetConfigType("yaml")
 v.AddConfigPath(".")
 v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
 v.AutomaticEnv()
 _ = v.ReadInConfig()
 return v, nil
}

```

**4.2. Transport Modules**
**grpc**

Provides a lifecycle-managed gRPC server.

``` go
// module/grpc/module.go
package grpc

import (
 "[github.com/things-kit/module/log](https://github.com/things-kit/module/log)"
 "go.uber.org/fx"
 "google.golang.org/grpc"
)

var Module = fx.Module("grpc", ConfigModule, fx.Invoke(RunGrpcServer))

type GrpcServerParams struct {
 fx.In
 Lifecycle  fx.Lifecycle
 Logger     log.Logger
 Config     *Config
 Registrars []ServiceRegistrar `group:"grpc.services"`
}
// ... rest of module implementation ...

```

**http (Gin)**

Provides a lifecycle-managed HTTP/REST server using Gin.

``` go
// module/http/module.go
package http

import (
 "[github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)"
 "[github.com/things-kit/module/log](https://github.com/things-kit/module/log)"
 "go.uber.org/fx"
)

var Module = fx.Module("http", ConfigModule, fx.Invoke(RunHttpServer))

// Handler represents a type that can register routes on a Gin engine.
type Handler interface {
 RegisterRoutes(engine *gin.Engine)
}

// AsHttpHandler is a generic helper to provide an HTTP handler to the Fx graph.
func AsHttpHandler(constructor any) fx.Option {
 return fx.Provide(fx.Annotate(
  constructor,
  fx.As(new(Handler)),
  fx.ResultTags(`group:"http.handlers"`),
 ))
}

type HttpServerParams struct {
 fx.In
 Lifecycle fx.Lifecycle
 Logger    log.Logger
 Config    *Config
 Handlers  []Handler `group:"http.handlers"`
}

// ... rest of module implementation ...

```

**4.3. Data & Messaging Modules**
**sqlc**

Provides a `*sql.DB` connection pool.

``` go
// module/sqlc/module.go
package sqlc

import (
 "context"
 "database/sql"
 "go.uber.org/fx"
)

var Module = fx.Module("sqlc", ConfigModule, fx.Provide(NewDB))

func NewDB(lc fx.Lifecycle, cfg *Config) (*sql.DB, error) {
 db, err := sql.Open("postgres", cfg.DSN)
 if err != nil { return nil, err }
 lc.Append(fx.Hook{OnStop: func(c context.Context) error { return db.Close() }})
 return db, nil
}

```

**redis**

Provides a `*redis.Client` from go-redis.

``` go
// module/redis/module.go
package redis

import (
 "context"
 "[github.com/go-redis/redis/v8](https://github.com/go-redis/redis/v8)"
 "go.uber.org/fx"
)

var Module = fx.Module("redis", ConfigModule, fx.Provide(NewRedisClient))

func NewRedisClient(lc fx.Lifecycle, cfg *Config) (*redis.Client, error) {
 opts, err := redis.ParseURL(cfg.URL)
 if err != nil { return nil, err }

 client := redis.NewClient(opts)
 lc.Append(fx.Hook{
  OnStart: func(ctx context.Context) error {
   return client.Ping(ctx).Err()
  },
  OnStop: func(ctx context.Context) error {
   return client.Close()
  },
 })
 return client, nil
}

```

**messaging (Abstraction)**

A framework-level interface for message consumers.

``` go
// module/messaging/message.go
package messaging

import (
 "context"
 "time"
)

type Message struct {
 Key, Value []byte
 Topic      string
 Timestamp  time.Time
}

type Handler interface {
 Handle(ctx context.Context, msg Message) error
}

```

**kafka**

An implementation of the messaging interface using `segmentio/kafka-go`.

``` go
// module/kafka/consumer.go
package kafka

import (
 "[github.com/things-kit/module/log](https://github.com/things-kit/module/log)"
 "[github.com/things-kit/module/messaging](https://github.com/things-kit/module/messaging)"
 "go.uber.org/fx"
)

var ConsumerModule = fx.Module("kafka-consumer", ConfigModule, fx.Invoke(RunConsumer))

type ConsumerParams struct {
 fx.In
 Lifecycle fx.Lifecycle
 Logger    log.Logger
 Config    *Config
 Handler   messaging.Handler
}
// ... rest of module implementation ...

```

**5. Configuration Strategy**

Configuration is decentralized. Each module defines its own `Config` struct and loads it from the shared `*viper.Viper` instance. For modules wrapping third-party libraries, the framework uses struct embedding with `mapstructure:",squash"` to provide full configuration control.

**Example config.yaml:**

``` yaml
grpc:
  port: 50052

http:
  port: 8080

db:
  dsn: "postgres://user:pass@host:5432/db?sslmode=disable"

redis:
  url: "redis://localhost:6379/0"

kafka:
  brokers: ["kafka-1:9092"]
  topic: "user-events"
  maxWait: "5s"

```

**6. Testing Strategy**

The testing module enables powerful integration testing by building a real Fx application within the test scope.

``` go
// module/testing/module.go
package testing

import (
 "context"
 "testing"
 "[github.com/things-kit/module/log](https://github.com/things-kit/module/log)"
 "go.uber.org/fx"
 "go.uber.org/fx/fxtest"
)

type testLogger struct { t *testing.T }
func (l *testLogger) Info(msg string, fields ...log.Field)  { l.t.Logf("[INFO] %s %v", msg, fields) }
func (l *testLogger) Error(msg string, err error, fields ...log.Field) { /* ... */ }
func (l *testLogger) Debug(msg string, fields ...log.Field) { /* ... */ }
func (l *testLogger) Warn(msg string, fields ...log.Field)  { /* ... */ }
func (l *testLogger) InfoC(ctx context.Context, msg string, fields ...log.Field)  { l.Info(msg, fields...) }
func (l *testLogger) ErrorC(ctx context.Context, msg string, err error, fields ...log.Field) { /* ... */ }
func (l *testLogger) DebugC(ctx context.Context, msg string, fields ...log.Field) { /* ... */ }
func (l *testLogger) WarnC(ctx context.Context, msg string, err error, fields ...log.Field)  { /* ... */ }

func RunTest(t *testing.T, opts ...fx.Option) {
 opts = append(opts, fx.Provide(func() log.Logger { return &testLogger{t: t} }))
 app := fxtest.New(t, opts...)
 app.RequireStart()
 app.RequireStop()
}

```

**7. Pluggable Core Components**

A core design principle is that foundational components are swappable. This is achieved by having framework modules depend on interfaces (like `log.Logger`) rather than concrete types. A developer can write their own module that provides an implementation for a core interface and replace the default module in their `main.go`.

**8. Getting Started: Your First Service**

This tutorial will guide you through creating the `UserService` from scratch.

**Step 1: Set up your project**

``` bash
mkdir my-awesome-service
cd my-awesome-service
go mod init my-awesome-service
# Create directories: cmd/server, internal/service

```

**Step 2: Define your gRPC service (e.g., in user.proto) and generate the Go code.**

**Step 3: Get the framework**

``` bash
go get [github.com/things-kit/app](https://github.com/things-kit/app)
go get [github.com/things-kit/logging](https://github.com/things-kit/logging)
go get [github.com/things-kit/module/grpc](https://github.com/things-kit/module/grpc)
go get [github.com/things-kit/module/viperconfig](https://github.com/things-kit/module/viperconfig)

```

**Step 4: Write your business logic**

Create the file `internal/service/user_service.go` as shown in section 3.1. It should depend on `log.Logger`.

**Step 5: Create your configuration**

Create a `config.yaml` file in your root directory:

``` yaml
grpc:
  port: 50051

```

**Step 6: Assemble your application**

Create the file `cmd/server/main.go` as shown in section 3.1. This file wires together the framework modules and your business logic.

**Step 7: Run your service**

``` bash
go run ./cmd/server

```