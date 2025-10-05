# Things-Kit Codebase Structure

## Project Layout
This is a Go workspace containing multiple independent modules. Each module is versionable and can be used independently.

```
things-kit/
├── go.work                      # Go workspace file
├── plan.md                      # Detailed project plan and documentation
├── README.md                    # Project overview
├── .gitignore                   # Git ignore rules
│
├── app/                         # Core application runner
│   ├── go.mod
│   └── app.go                   # Main App type, New(), Run(), AsStartupFunc()
│
├── logging/                     # Default Zap-based logger implementation
│   ├── go.mod
│   └── module.go                # zapLoggerAdapter implementing log.Logger
│
└── module/                      # Framework modules
    ├── log/                     # Logger interface (abstraction)
    │   ├── go.mod
    │   └── interface.go         # log.Logger interface, log.Field struct
    │
    ├── viperconfig/             # Configuration management
    │   ├── go.mod
    │   └── module.go            # Provides *viper.Viper
    │
    ├── grpc/                    # gRPC server module
    │   ├── go.mod
    │   └── module.go            # gRPC server, AsGrpcService helper
    │
    ├── http/                    # HTTP/REST server (Gin)
    │   ├── go.mod
    │   └── module.go            # Gin server, AsHttpHandler helper
    │
    ├── sqlc/                    # Database connection pool
    │   ├── go.mod
    │   └── module.go            # Provides *sql.DB
    │
    ├── redis/                   # Redis client
    │   ├── go.mod
    │   └── module.go            # Provides *redis.Client
    │
    ├── kafka/                   # Kafka consumer/producer
    │   ├── go.mod
    │   └── consumer.go          # Kafka consumer implementation
    │
    ├── messaging/               # Messaging abstraction
    │   ├── go.mod
    │   └── message.go           # Message struct, Handler interface
    │
    └── testing/                 # Testing utilities
        ├── go.mod
        └── module.go            # RunTest helper, test logger
```

## Module Dependencies
- **module/log**: No dependencies (pure interface)
- **module/messaging**: No dependencies (pure interface)
- **app**: Depends on fx
- **module/viperconfig**: Depends on viper, fx
- **logging**: Depends on log interface, viper, fx, zap
- **module/grpc**: Depends on log interface, fx, grpc
- **module/http**: Depends on log interface, fx, gin
- **module/sqlc**: Depends on fx
- **module/redis**: Depends on fx, go-redis
- **module/kafka**: Depends on log interface, messaging interface, fx, kafka-go
- **module/testing**: Depends on log interface, fx

## Key Files
- `go.work`: Defines the Go workspace with all modules
- `plan.md`: Complete architectural documentation and examples
- Each module has its own `go.mod` for independent versioning
