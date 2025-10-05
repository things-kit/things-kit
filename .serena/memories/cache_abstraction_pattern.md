# Cache Abstraction Pattern

Things-Kit implements a cache abstraction pattern to allow applications to use any caching backend without code changes.

## Architecture

### Interface Module: `module/cache`
- Defines `cache.Cache` interface with standard key-value operations
- Defines `cache.BatchCache` interface for bulk operations
- Contains **NO implementation** - only contracts

### Default Implementation: `module/redis`
- Implements `cache.Cache` using go-redis/v9
- Provides both the interface (`cache.Cache`) AND direct client access (`*redis.Client`)
- Configured via Viper with `redis.url` setting

## Cache Interface Operations

```go
type Cache interface {
    // Basic operations
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, expiration time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    
    // Binary data operations
    GetBytes(ctx context.Context, key string) ([]byte, error)
    SetBytes(ctx context.Context, key string, value []byte, expiration time.Duration) error
    
    // Expiration management
    Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
    TTL(ctx context.Context, key string) (time.Duration, error)
    
    // Connection management
    Ping(ctx context.Context) error
    Close() error
}
```

## Usage Pattern

### 1. Add Redis Module (Default)
```go
app.New(
    viperconfig.Module,
    logging.Module,
    redis.Module,  // Provides cache.Cache
    
    fx.Provide(NewMyService),
).Run()
```

### 2. Inject Cache Interface
```go
type MyService struct {
    cache cache.Cache  // Program to interface
}

func NewMyService(c cache.Cache) *MyService {
    return &MyService{cache: c}
}

func (s *MyService) GetUser(ctx context.Context, id string) (*User, error) {
    key := fmt.Sprintf("user:%s", id)
    
    // Try cache first
    if data, err := s.cache.Get(ctx, key); err == nil {
        var user User
        json.Unmarshal([]byte(data), &user)
        return &user, nil
    }
    
    // Cache miss - fetch from DB and cache
    user := s.fetchFromDB(ctx, id)
    if data, _ := json.Marshal(user); data != nil {
        s.cache.Set(ctx, key, string(data), 5*time.Minute)
    }
    
    return user, nil
}
```

## Power User Access

For advanced Redis features, inject `*redis.Client` directly:

```go
type AdvancedService struct {
    cache  cache.Cache    // Standard interface
    client *redis.Client  // Redis-specific features
}

func (s *AdvancedService) UseRedisFeatures(ctx context.Context) {
    // Use pipelines, Lua scripts, pub/sub, etc.
    pipe := s.client.Pipeline()
    pipe.Incr(ctx, "counter")
    pipe.HSet(ctx, "hash", "field", "value")
    pipe.Exec(ctx)
}
```

## Creating Alternative Implementations

To use a different cache backend:

1. **Implement `cache.Cache` interface**
2. **Create Fx module** providing the implementation
3. **Use in application** instead of redis.Module

Examples:
- **Valkey**: Redis fork with enhanced features
- **Memcached**: Distributed memory caching
- **In-Memory**: Simple map-based cache for testing
- **DragonflyDB**: Modern Redis-compatible cache
- **Multi-tier**: L1 (local) + L2 (distributed) caching

## Benefits

1. **Backend Independence**: Code doesn't depend on specific cache technology
2. **Easy Testing**: Mock `cache.Cache` interface for unit tests
3. **Migration Path**: Switch from Redis to Valkey without code changes
4. **Flexibility**: Use different caches for different services
5. **Consistent API**: Same operations across all cache backends

## Configuration

Redis cache configuration:
```yaml
redis:
  url: "redis://localhost:6379/0"
```

Alternative implementations would have their own config keys.

## Files

- `/module/cache/interface.go` - Cache interface definition
- `/module/cache/go.mod` - Interface module (no dependencies)
- `/module/cache/README.md` - Complete documentation and examples
- `/module/redis/module.go` - Redis implementation with RedisCache struct
- `/module/redis/README.md` - Redis-specific documentation
- `/go.work` - Includes module/cache in workspace

## Pattern Consistency

This follows the same pattern as other abstractions:
- **Logger**: `module/log` (interface) → `module/logging` (Zap)
- **HTTP**: `module/http` (interface) → `module/httpgin` (Gin)
- **Cache**: `module/cache` (interface) → `module/redis` (Redis)
- **Messaging**: `module/messaging` (interfaces) → `module/kafka` (Kafka)
