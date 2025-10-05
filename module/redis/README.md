# module/redis - Redis Cache Implementation

This module provides a Redis-based implementation of the `module/cache` interface for Things-Kit.

## Overview

The `module/redis` package implements the `cache.Cache` interface using [go-redis/v9](https://github.com/redis/go-redis). It's the default cache implementation for Things-Kit applications.

## Features

- ✅ Implements `cache.Cache` interface completely
- ✅ Supports all Redis data operations (strings, bytes)
- ✅ Connection pooling and lifecycle management via Fx
- ✅ Configuration through Viper (YAML + environment variables)
- ✅ Expiration and TTL management
- ✅ Health checking with Ping
- ⚡ **Power user access**: Direct `*redis.Client` also available

## Installation

```bash
go get github.com/things-kit/module/redis
```

## Basic Usage

### 1. Add Module to Application

```go
package main

import (
    "github.com/things-kit/app"
    "github.com/things-kit/module/cache"
    "github.com/things-kit/module/redis"
    "github.com/things-kit/module/viperconfig"
    "github.com/things-kit/module/logging"
    "go.uber.org/fx"
)

func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        redis.Module,  // Provides both cache.Cache and *redis.Client
        
        fx.Invoke(RunMyService),
    ).Run()
}
```

### 2. Inject Cache Interface

```go
type MyService struct {
    cache cache.Cache
    log   log.Logger
}

func NewMyService(c cache.Cache, l log.Logger) *MyService {
    return &MyService{
        cache: c,
        log:   l,
    }
}

func (s *MyService) GetUser(ctx context.Context, id string) (*User, error) {
    key := fmt.Sprintf("user:%s", id)
    
    // Try cache first
    if data, err := s.cache.Get(ctx, key); err == nil {
        var user User
        if err := json.Unmarshal([]byte(data), &user); err == nil {
            s.log.Info("Cache hit", "key", key)
            return &user, nil
        }
    }
    
    // Cache miss - fetch from database
    user, err := s.fetchUserFromDB(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for 5 minutes
    if data, err := json.Marshal(user); err == nil {
        _ = s.cache.Set(ctx, key, string(data), 5*time.Minute)
    }
    
    return user, nil
}
```

## Configuration

### config.yaml

```yaml
redis:
  url: "redis://localhost:6379/0"  # Redis connection URL
```

### Environment Variables

```bash
export REDIS_URL="redis://username:password@host:6379/0"
```

### Supported URL Formats

```
# Simple
redis://localhost:6379

# With database
redis://localhost:6379/2

# With authentication
redis://:password@localhost:6379

# With username and password
redis://username:password@localhost:6379

# TLS connection
rediss://localhost:6380

# Unix socket
redis:///var/run/redis/redis.sock
```

## Advanced Usage

### Direct Redis Client Access

For advanced Redis operations not covered by the `cache.Cache` interface, you can inject the `*redis.Client` directly:

```go
type AdvancedService struct {
    client *redis.Client  // Full Redis client
}

func NewAdvancedService(client *redis.Client) *AdvancedService {
    return &AdvancedService{client: client}
}

func (s *AdvancedService) UseAdvancedFeatures(ctx context.Context) error {
    // Use Redis-specific features
    pipe := s.client.Pipeline()
    pipe.Incr(ctx, "counter")
    pipe.HSet(ctx, "hash", "field", "value")
    pipe.ZAdd(ctx, "sorted-set", redis.Z{Score: 1.0, Member: "member"})
    
    _, err := pipe.Exec(ctx)
    return err
}

func (s *AdvancedService) UseLua(ctx context.Context) error {
    script := redis.NewScript(`
        redis.call('SET', KEYS[1], ARGV[1])
        return redis.call('GET', KEYS[1])
    `)
    
    result, err := script.Run(ctx, s.client, []string{"mykey"}, "myvalue").Result()
    return err
}
```

### Both Interface and Client

You can inject both for maximum flexibility:

```go
type HybridService struct {
    cache  cache.Cache    // For standard caching
    client *redis.Client  // For advanced operations
}

func NewHybridService(c cache.Cache, client *redis.Client) *HybridService {
    return &HybridService{
        cache:  c,
        client: client,
    }
}
```

## Cache Operations

### String Operations

```go
// Set with expiration
err := cache.Set(ctx, "key", "value", 10*time.Minute)

// Get
value, err := cache.Get(ctx, "key")

// Delete
err := cache.Delete(ctx, "key")

// Check existence
exists, err := cache.Exists(ctx, "key")
```

### Binary Operations

```go
// Store binary data (images, files, etc.)
data := []byte{0x89, 0x50, 0x4E, 0x47, ...}
err := cache.SetBytes(ctx, "image:123", data, 1*time.Hour)

// Retrieve binary data
imageData, err := cache.GetBytes(ctx, "image:123")
```

### Expiration Management

```go
// Set expiration on existing key
success, err := cache.Expire(ctx, "session:abc", 30*time.Minute)

// Get remaining TTL
ttl, err := cache.TTL(ctx, "session:abc")
if ttl > 0 {
    fmt.Printf("Key expires in %v\n", ttl)
}
```

### Health Checks

```go
// Ping Redis
err := cache.Ping(ctx)
if err != nil {
    log.Error("Redis is down!", "error", err)
}
```

## Error Handling

```go
import "github.com/redis/go-redis/v9"

value, err := cache.Get(ctx, "key")
if err != nil {
    if errors.Is(err, redis.Nil) {
        // Key doesn't exist (cache miss)
        return handleCacheMiss(ctx)
    }
    // Other error (connection, timeout, etc.)
    return err
}
```

## Patterns

### Cache-Aside Pattern

```go
func (s *Service) GetData(ctx context.Context, id string) (*Data, error) {
    key := fmt.Sprintf("data:%s", id)
    
    // 1. Try cache
    if cached, err := s.cache.Get(ctx, key); err == nil {
        var data Data
        if err := json.Unmarshal([]byte(cached), &data); err == nil {
            return &data, nil
        }
    }
    
    // 2. Fetch from source
    data, err := s.db.GetData(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 3. Update cache
    if serialized, err := json.Marshal(data); err == nil {
        _ = s.cache.Set(ctx, key, string(serialized), 10*time.Minute)
    }
    
    return data, nil
}
```

### Write-Through Pattern

```go
func (s *Service) UpdateData(ctx context.Context, id string, data *Data) error {
    key := fmt.Sprintf("data:%s", id)
    
    // 1. Update database
    if err := s.db.UpdateData(ctx, id, data); err != nil {
        return err
    }
    
    // 2. Update cache
    if serialized, err := json.Marshal(data); err == nil {
        _ = s.cache.Set(ctx, key, string(serialized), 10*time.Minute)
    }
    
    return nil
}
```

### Cache Invalidation

```go
func (s *Service) DeleteData(ctx context.Context, id string) error {
    key := fmt.Sprintf("data:%s", id)
    
    // Delete from database
    if err := s.db.DeleteData(ctx, id); err != nil {
        return err
    }
    
    // Invalidate cache
    _ = s.cache.Delete(ctx, key)
    
    return nil
}
```

## Testing

### Mock for Unit Tests

```go
import "github.com/stretchr/testify/mock"

type MockCache struct {
    mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
    args := m.Called(ctx, key)
    return args.String(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key, value string, exp time.Duration) error {
    args := m.Called(ctx, key, value, exp)
    return args.Error(0)
}

// Test
func TestService(t *testing.T) {
    mockCache := new(MockCache)
    mockCache.On("Get", mock.Anything, "user:123").
        Return(`{"id":"123","name":"John"}`, nil)
    
    svc := NewService(mockCache)
    user, err := svc.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
    mockCache.AssertExpectations(t)
}
```

### Integration Tests

Use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) or the Things-Kit testing module:

```go
func TestWithRedis(t *testing.T) {
    // Start Redis container
    container := testing.StartRedisContainer(t)
    defer container.Terminate(context.Background())
    
    // Configure Redis module
    viper.Set("redis.url", container.URL())
    
    // Test with real Redis
    // ...
}
```

## Alternatives

Want to use a different cache backend? Create your own implementation:

- **Valkey**: Redis fork - minimal changes needed
- **Memcached**: Create `module/memcached` implementing `cache.Cache`
- **In-Memory**: Simple map-based cache for development
- **DragonflyDB**: Modern Redis-compatible backend

See [module/cache README](../cache/README.md) for guidance on creating custom implementations.

## Performance Tips

1. **Use connection pooling** (handled automatically by go-redis)
2. **Set appropriate expiration times** to prevent memory bloat
3. **Use binary operations** for non-text data (images, files)
4. **Batch operations** when possible (use `*redis.Client` for pipelines)
5. **Monitor memory usage** and set `maxmemory` policy in Redis config

## Lifecycle

The Redis module integrates with Fx lifecycle:

- **OnStart**: Connects to Redis and verifies connectivity with Ping
- **OnStop**: Closes Redis connection gracefully

No manual connection management needed!

## Troubleshooting

### Connection Refused

```
Error: dial tcp [::1]:6379: connect: connection refused
```

**Solution**: Ensure Redis is running:
```bash
# macOS
brew services start redis

# Linux
sudo systemctl start redis

# Docker
docker run -d -p 6379:6379 redis:7-alpine
```

### Authentication Failed

```
Error: NOAUTH Authentication required
```

**Solution**: Update config with password:
```yaml
redis:
  url: "redis://:your-password@localhost:6379"
```

### Timeout Errors

**Solution**: Increase timeout or check network:
```yaml
redis:
  url: "redis://localhost:6379?dial_timeout=5s&read_timeout=3s"
```

## Dependencies

- `github.com/redis/go-redis/v9` - Redis client
- `github.com/things-kit/module/cache` - Cache interface
- `github.com/spf13/viper` - Configuration
- `go.uber.org/fx` - Dependency injection

## License

MIT License - see LICENSE file for details
