# Things-Kit Module Organization

## Overview
Things-Kit has been refactored to have a clear, consistent module organization where ALL framework-provided components live under the `module/` directory.

## Directory Structure

```
things-kit/
├── app/                      # Application runner (root level - not a module)
├── example/                  # Example service (root level - demonstration)
└── module/                   # ALL framework modules live here
    ├── grpc/                 # gRPC server
    ├── http/                 # HTTP server interface (abstraction)
    ├── httpgin/              # HTTP server implementation (Gin) ⭐
    ├── kafka/                # Kafka consumer
    ├── log/                  # Logger interface (abstraction)
    ├── logging/              # Logger implementation (Zap) ⭐
    ├── messaging/            # Messaging interface
    ├── redis/                # Redis client
    ├── sqlc/                 # SQL database client
    ├── testing/              # Testing utilities
    └── viperconfig/          # Configuration provider

⭐ = Default implementation modules (paired with interface modules)
```

## Organizational Principles

### Clear Rule
**"All framework-provided components go in `module/`, alternative/third-party implementations go at root level"**

### Benefits
1. **Discovery**: New users browse `module/` and find everything the framework provides
2. **Consistency**: All framework components follow the same organizational pattern
3. **Clear Ownership**: `module/*` = framework-provided, `root/*` = alternatives/apps
4. **Logical Grouping**: Related components (interface + default impl) are neighbors
5. **Extensibility**: Clear place for community implementations (root level)

## Module Types

### 1. Interface + Implementation Pattern
Used when multiple implementations are possible:

**Logger:**
- `module/log/` - Interface definition
- `module/logging/` - Default Zap implementation
- `logging-logrus/` (future) - Alternative at root level

**HTTP Server:**
- `module/http/` - Interface definition
- `module/httpgin/` - Default Gin implementation
- `httpchi/` (future) - Alternative at root level

### 2. Direct Implementation Pattern
Used when there's typically only one implementation:

- `module/grpc/` - gRPC server (google.golang.org/grpc)
- `module/redis/` - Redis client (go-redis/v9)
- `module/kafka/` - Kafka consumer (segmentio/kafka-go)
- `module/sqlc/` - SQL database (database/sql)
- `module/viperconfig/` - Viper configuration
- `module/testing/` - Testing utilities

These modules combine interface and implementation because there's no real benefit to separation (the underlying library is standard).

## Import Paths

### Before Refactoring
```go
import (
    "github.com/things-kit/httpgin"        // At root
    "github.com/things-kit/logging"        // At root
    "github.com/things-kit/module/http"    // In module/
    "github.com/things-kit/module/log"     // In module/
)
```

### After Refactoring
```go
import (
    "github.com/things-kit/module/httpgin"  // Now in module/
    "github.com/things-kit/module/logging"  // Now in module/
    "github.com/things-kit/module/http"     // In module/
    "github.com/things-kit/module/log"      // In module/
)
```

**Everything is consistently under `module/`**

## Migration Impact

### Files Moved
1. `httpgin/` → `module/httpgin/`
2. `logging/` → `module/logging/`

### Files Updated
1. `module/httpgin/go.mod` - Updated module path and replace directives
2. `module/logging/go.mod` - Updated module path and replace directives
3. `example/go.mod` - Updated import paths and replace directives
4. `example/cmd/server/main.go` - Updated import statements
5. `go.work` - Updated workspace paths
6. `README.md` - Updated documentation
7. `module/httpgin/README.md` - Updated paths and examples
8. `module/http/README.md` - Updated paths and examples

### No Breaking Changes for Patterns
- Handler interfaces remain the same
- Module usage patterns unchanged
- Configuration unchanged
- Lifecycle management unchanged

## Comparison with Other Frameworks

### Spring Boot
```
spring-boot-starter-web/     (default implementation)
spring-boot-starter-jetty/   (alternative)
spring-boot-starter-undertow/(alternative)
```

### Things-Kit (New)
```
module/httpgin/              (default implementation)
httpchi/                     (alternative - at root if created)
httpecho/                    (alternative - at root if created)
```

**Pattern match!** Framework defaults are "blessed" and in a special location, alternatives are separate.

## Future Extensions

When users create alternative implementations:

### For HTTP Server
```
things-kit/
├── module/
│   ├── http/      (interface - framework)
│   └── httpgin/   (default impl - framework)
└── httpchi/       (alternative - community/root)
```

### For Logger
```
things-kit/
├── module/
│   ├── log/       (interface - framework)
│   └── logging/   (default impl - framework)
└── logging-logrus/(alternative - community/root)
```

Alternative implementations at root level are clearly "not part of the core framework" but are compatible with it.

## Example Usage

### Using Default Implementations
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
        logging.Module,      // Default Zap logger
        httpgin.Module,      // Default Gin HTTP server
    ).Run()
}
```

### Using Alternative Implementation (Future)
```go
package main

import (
    "github.com/things-kit/app"
    "github.com/things-kit/httpchi"          // Alternative at root
    "github.com/things-kit/module/logging"   // Default from module/
    "github.com/things-kit/module/viperconfig"
)

func main() {
    app.New(
        viperconfig.Module,
        logging.Module,      // Framework default
        httpchi.Module,      // Community alternative
    ).Run()
}
```

## Workspace Configuration

`go.work` now lists all modules under `module/`:

```go
use (
    ./app
    ./example
    ./module/grpc
    ./module/http
    ./module/httpgin      ← Moved here
    ./module/kafka
    ./module/log
    ./module/logging      ← Moved here
    ./module/messaging
    ./module/redis
    ./module/sqlc
    ./module/testing
    ./module/viperconfig
)
```

Clean and consistent!

## Key Insights

1. **User Intuition Was Correct**: Moving default implementations to `module/` makes more sense
2. **Consistency Matters**: Having some modules in `module/` and others at root was confusing
3. **Discovery is Important**: New users should browse `module/` to see what's available
4. **Clear Boundaries**: Framework vs. community/alternative is now obvious from directory structure
5. **Follows Established Patterns**: Similar to Spring Boot's "starter" pattern

This refactoring significantly improves the developer experience and makes the framework's architecture more intuitive.
