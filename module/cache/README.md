# module/cache - Cache Interface

This module defines the cache abstraction for Things-Kit. It contains **only interfaces**, no implementation.

## Purpose

The `module/cache` package defines the contract that all cache implementations must follow. This allows applications to program against a stable interface while being free to choose any cache backend (Redis, Valkey, Memcached, in-memory, etc.).

## Interfaces

### Cache

The `Cache` interface defines operations for a distributed key-value cache:

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

### BatchCache

The `BatchCache` interface extends `Cache` with batch operations for improved performance:

```go
type BatchCache interface {
    Cache
    
    MGet(ctx context.Context, keys ...string) (map[string]string, error)
    MSet(ctx context.Context, pairs map[string]string, expiration time.Duration) error
    MDelete(ctx context.Context, keys ...string) error
}
```

## Available Implementations

### module/redis (Default)

The [redis module](../redis/) provides a Redis-based implementation of the Cache interface. It's the recommended default for most applications.

```go
import (
    "github.com/things-kit/module/cache"
    "github.com/things-kit/module/redis"
)

func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        redis.Module,  // Provides cache.Cache
        
        // Inject cache.Cache in your services
        fx.Provide(NewMyService),
    ).Run()
}

type MyService struct {
    cache cache.Cache
}

func NewMyService(c cache.Cache) *MyService {
    return &MyService{cache: c}
}
```

### Custom Implementations

You can create your own cache implementation using any backend:

- **Valkey**: Redis fork with improved features
- **Memcached**: Distributed memory caching system
- **In-Memory**: Simple map-based cache for testing
- **DragonflyDB**: Modern Redis-compatible cache
- **Multi-tier**: Combine L1 (local) + L2 (distributed) caching

## Creating Your Own Implementation

To create a custom cache implementation:

1. **Implement the `Cache` interface**:
   ```go
   type MyCache struct {
       // your fields
   }
   
   func (c *MyCache) Get(ctx context.Context, key string) (string, error) { /* ... */ }
   func (c *MyCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error { /* ... */ }
   // ... implement all interface methods
   ```

2. **Create an Fx module**:
   ```go
   var Module = fx.Module("mycache",
       fx.Provide(
           NewConfig,
           NewMyCache,
           fx.Annotate(
               func(c *MyCache) cache.Cache { return c },
               fx.As(new(cache.Cache)),
           ),
       ),
   )
   ```

3. **Use in your application**:
   ```go
   app.New(
       viperconfig.Module,
       logging.Module,
       mycache.Module,  // Your custom cache
       // ...
   )
   ```

See the [redis implementation](../redis/) as a reference example.

## Design Philosophy

This interface abstraction follows the principle **"Program to an Interface, Not an Implementation"**:

- **Applications** depend on `module/cache` (interface)
- **Implementations** provide `cache.Cache` (redis, memcached, etc.)
- **Flexibility**: Swap implementations without changing application code

This is the same pattern used for other abstractions:
- Interface: `module/log` → Implementation: `module/logging` (Zap)
- Interface: `module/http` → Implementation: `module/httpgin` (Gin)
- Interface: `module/cache` → Implementation: `module/redis` (Redis)

## Benefits

1. **Backend Independence**: Your application code doesn't depend on a specific cache backend
2. **Testing**: Easy to mock the `Cache` interface for unit tests
3. **Migration**: Switch backends without rewriting cache logic
4. **Experimentation**: Try different backends for performance or features
5. **Consistency**: All cache implementations follow the same contract

## Usage Patterns

### Simple Caching

```go
func (s *MyService) GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("user:%s", id)
    if data, err := s.cache.Get(ctx, cacheKey); err == nil {
        var user User
        if err := json.Unmarshal([]byte(data), &user); err == nil {
            return &user, nil
        }
    }
    
    // Cache miss - fetch from database
    user, err := s.db.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    if data, err := json.Marshal(user); err == nil {
        _ = s.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
    }
    
    return user, nil
}
```

### Binary Data Caching

```go
func (s *MyService) GetImage(ctx context.Context, id string) ([]byte, error) {
    cacheKey := fmt.Sprintf("image:%s", id)
    
    // Check cache
    if data, err := s.cache.GetBytes(ctx, cacheKey); err == nil {
        return data, nil
    }
    
    // Fetch and cache
    data, err := s.storage.GetImage(ctx, id)
    if err != nil {
        return nil, err
    }
    
    _ = s.cache.SetBytes(ctx, cacheKey, data, 1*time.Hour)
    return data, nil
}
```

### Cache with TTL Management

```go
func (s *MyService) ExtendSession(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("session:%s", sessionID)
    
    // Check if session exists
    exists, err := s.cache.Exists(ctx, key)
    if err != nil || !exists {
        return ErrSessionNotFound
    }
    
    // Extend TTL
    _, err = s.cache.Expire(ctx, key, 30*time.Minute)
    return err
}
```

## Testing

Mock the `Cache` interface for unit tests:

```go
type MockCache struct {
    mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
    args := m.Called(ctx, key)
    return args.String(0), args.Error(1)
}

// ... implement other methods

func TestMyService(t *testing.T) {
    mockCache := new(MockCache)
    mockCache.On("Get", mock.Anything, "user:123").Return(`{"id":"123","name":"John"}`, nil)
    
    svc := NewMyService(mockCache)
    user, err := svc.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
    mockCache.AssertExpectations(t)
}
```

## Configuration

Each cache implementation has its own configuration. For Redis:

```yaml
redis:
  url: "redis://localhost:6379/0"
```

See the specific implementation's documentation for configuration details.

## License

MIT License - see LICENSE file for details
