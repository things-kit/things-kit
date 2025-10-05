# module/kafka - Kafka Consumer Implementation

This module provides a Kafka-based implementation of the `module/messaging` interfaces for Things-Kit.

## Overview

The `module/kafka` package implements the `messaging.Consumer` interface using [segmentio/kafka-go](https://github.com/segmentio/kafka-go). It's the default message broker implementation for Things-Kit applications.

## Features

- ✅ Implements `messaging.Consumer` interface
- ✅ Consumer group support with automatic rebalancing
- ✅ Configurable commit strategies (auto/manual)
- ✅ Multiple topic subscription
- ✅ Lifecycle management via Fx
- ✅ Configuration through Viper (YAML + environment variables)
- ✅ Graceful shutdown with in-flight message handling

## Installation

```bash
go get github.com/things-kit/module/kafka
```

## Basic Usage

### 1. Implement Handler

First, create your message handler implementing `messaging.Handler`:

```go
package myservice

import (
    "context"
    "encoding/json"
    "github.com/things-kit/module/messaging"
    "github.com/things-kit/module/log"
)

type OrderHandler struct {
    log log.Logger
    db  *Database
}

func NewOrderHandler(l log.Logger, db *Database) *OrderHandler {
    return &OrderHandler{log: l, db: db}
}

func (h *OrderHandler) Handle(ctx context.Context, msg messaging.Message) error {
    var order Order
    if err := json.Unmarshal(msg.Value, &order); err != nil {
        h.log.Error("Invalid message format", "error", err)
        return err
    }
    
    h.log.Info("Processing order", 
        "id", order.ID,
        "topic", msg.Topic,
        "partition", msg.Partition,
        "offset", msg.Offset,
    )
    
    if err := h.db.SaveOrder(ctx, &order); err != nil {
        h.log.Error("Failed to save order", "error", err)
        return err
    }
    
    return nil
}
```

### 2. Add Module to Application

```go
package main

import (
    "github.com/things-kit/app"
    "github.com/things-kit/module/kafka"
    "github.com/things-kit/module/viperconfig"
    "github.com/things-kit/module/logging"
    "go.uber.org/fx"
)

func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        kafka.Module,
        
        // Provide your handler
        fx.Provide(NewOrderHandler),
        
        // Start the consumer
        fx.Invoke(kafka.RunConsumer),
    ).Run()
}
```

### 3. Configure

Create `config.yaml`:

```yaml
kafka:
  brokers:
    - "localhost:9092"
  group: "order-service"
  topics:
    - "orders.created"
    - "orders.updated"
  auto_commit: true
```

That's it! Your application will now consume messages from Kafka.

## Configuration

### Complete Configuration Options

```yaml
kafka:
  # Kafka broker addresses (required)
  brokers:
    - "kafka-1:9092"
    - "kafka-2:9092"
    - "kafka-3:9092"
  
  # Consumer group ID (required)
  group: "my-service-group"
  
  # Topics to subscribe to (required)
  topics:
    - "orders"
    - "payments"
    - "notifications"
  
  # Auto-commit offsets (default: true)
  auto_commit: true
  
  # Start from beginning if no offset (default: false)
  # If false, starts from latest
  start_from_beginning: false
```

### Environment Variables

Override configuration with environment variables:

```bash
export KAFKA_BROKERS="kafka-1:9092,kafka-2:9092"
export KAFKA_GROUP="order-service"
export KAFKA_TOPICS="orders.created,orders.updated"
export KAFKA_AUTO_COMMIT="true"
export KAFKA_START_FROM_BEGINNING="false"
```

## Advanced Usage

### Manual Offset Committing

For precise control over when offsets are committed:

```go
kafka:
  auto_commit: false  # Disable auto-commit
```

```go
type CriticalHandler struct {
    log log.Logger
    db  *Database
}

func (h *CriticalHandler) Handle(ctx context.Context, msg messaging.Message) error {
    // Process message
    if err := h.processMessage(ctx, msg); err != nil {
        h.log.Error("Processing failed", "error", err)
        return err // Will NOT commit offset (message will be reprocessed)
    }
    
    h.log.Info("Successfully processed", "offset", msg.Offset)
    return nil // Will commit offset (message acknowledged)
}
```

### Multiple Handlers with Router

Route messages based on topic or content:

```go
type MessageRouter struct {
    handlers map[string]messaging.Handler
    log      log.Logger
}

func NewMessageRouter(
    orderHandler *OrderHandler,
    paymentHandler *PaymentHandler,
    log log.Logger,
) *MessageRouter {
    router := &MessageRouter{
        handlers: make(map[string]messaging.Handler),
        log:      log,
    }
    
    router.handlers["orders.created"] = orderHandler
    router.handlers["orders.updated"] = orderHandler
    router.handlers["payments.processed"] = paymentHandler
    
    return router
}

func (r *MessageRouter) Handle(ctx context.Context, msg messaging.Message) error {
    handler, ok := r.handlers[msg.Topic]
    if !ok {
        r.log.Warn("No handler for topic", "topic", msg.Topic)
        return nil // Skip unknown topics
    }
    
    return handler.Handle(ctx, msg)
}
```

### Error Handling Patterns

#### Retry with Backoff

```go
func (h *Handler) Handle(ctx context.Context, msg messaging.Message) error {
    const maxRetries = 3
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err := h.process(ctx, msg)
        if err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            h.log.Error("Non-retryable error", "error", err)
            return nil // Acknowledge to skip message
        }
        
        if attempt < maxRetries {
            backoff := time.Second * time.Duration(math.Pow(2, float64(attempt)))
            h.log.Warn("Retrying", "attempt", attempt, "backoff", backoff)
            
            select {
            case <-time.After(backoff):
                continue
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }
    
    // Max retries exceeded - send to DLQ or log
    h.sendToDLQ(ctx, msg)
    return nil // Acknowledge to move forward
}
```

#### Dead Letter Queue

```go
type HandlerWithDLQ struct {
    handler  messaging.Handler
    producer messaging.Producer // Implement this for your use case
    log      log.Logger
}

func (h *HandlerWithDLQ) Handle(ctx context.Context, msg messaging.Message) error {
    err := h.handler.Handle(ctx, msg)
    if err != nil {
        // Send to DLQ
        dlqMsg := messaging.Message{
            Topic: msg.Topic + ".dlq",
            Key:   msg.Key,
            Value: msg.Value,
            Headers: map[string]string{
                "original-topic":     msg.Topic,
                "original-partition": fmt.Sprint(msg.Partition),
                "original-offset":    fmt.Sprint(msg.Offset),
                "error":              err.Error(),
                "timestamp":          time.Now().Format(time.RFC3339),
            },
        }
        
        if err := h.producer.Publish(ctx, dlqMsg); err != nil {
            h.log.Error("Failed to send to DLQ", "error", err)
        }
        
        // Acknowledge original message to prevent reprocessing
        return nil
    }
    
    return nil
}
```

### Monitoring and Metrics

Track consumer performance:

```go
type MetricsHandler struct {
    handler messaging.Handler
    metrics *Metrics
}

func (h *MetricsHandler) Handle(ctx context.Context, msg messaging.Message) error {
    start := time.Now()
    
    err := h.handler.Handle(ctx, msg)
    
    duration := time.Since(start)
    h.metrics.RecordMessageProcessing(msg.Topic, duration, err)
    
    return err
}
```

## Consumer Groups

Kafka consumer groups provide:
- **Load Balancing**: Messages distributed across group members
- **Fault Tolerance**: Automatic partition reassignment on failure
- **Scalability**: Add consumers to increase throughput

### Scaling Example

Run multiple instances with the same group ID:

```yaml
# Instance 1
kafka:
  group: "order-processor"
  topics: ["orders"]

# Instance 2 (same group)
kafka:
  group: "order-processor"
  topics: ["orders"]

# Instance 3 (same group)
kafka:
  group: "order-processor"
  topics: ["orders"]
```

Kafka automatically distributes partitions across instances.

## Testing

### Mock Handler

```go
import "github.com/stretchr/testify/mock"

type MockHandler struct {
    mock.Mock
}

func (m *MockHandler) Handle(ctx context.Context, msg messaging.Message) error {
    args := m.Called(ctx, msg)
    return args.Error(0)
}

func TestOrderHandler(t *testing.T) {
    mockDB := new(MockDatabase)
    handler := NewOrderHandler(logger, mockDB)
    
    msg := messaging.Message{
        Topic: "orders.created",
        Value: []byte(`{"id":"123","amount":100}`),
    }
    
    err := handler.Handle(context.Background(), msg)
    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
}
```

### Integration Tests

Use [testcontainers-go](https://github.com/testcontainers/testcontainers-go):

```go
func TestKafkaConsumer(t *testing.T) {
    ctx := context.Background()
    
    // Start Kafka container
    kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "confluentinc/cp-kafka:7.5.0",
            ExposedPorts: []string{"9093/tcp"},
            Env: map[string]string{
                "KAFKA_LISTENERS": "PLAINTEXT://0.0.0.0:9093",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer kafkaContainer.Terminate(ctx)
    
    broker, err := kafkaContainer.Host(ctx)
    require.NoError(t, err)
    
    port, err := kafkaContainer.MappedPort(ctx, "9093")
    require.NoError(t, err)
    
    brokerAddr := fmt.Sprintf("%s:%s", broker, port.Port())
    
    // Configure and test
    viper.Set("kafka.brokers", []string{brokerAddr})
    viper.Set("kafka.group", "test-group")
    viper.Set("kafka.topics", []string{"test-topic"})
    
    // Create consumer and test...
}
```

## Message Format

### Basic Message

```go
msg := messaging.Message{
    Topic:     "orders.created",
    Key:       []byte("order-123"),
    Value:     []byte(`{"id":"123","amount":100}`),
    Headers:   map[string]string{"version": "v1"},
    Partition: 2,
    Offset:    12345,
    Timestamp: time.Now(),
}
```

### JSON Messages

```go
type Order struct {
    ID     string  `json:"id"`
    Amount float64 `json:"amount"`
}

// Producing
orderJSON, _ := json.Marshal(order)
msg := messaging.Message{
    Topic: "orders.created",
    Key:   []byte(order.ID),
    Value: orderJSON,
}

// Consuming
func (h *OrderHandler) Handle(ctx context.Context, msg messaging.Message) error {
    var order Order
    if err := json.Unmarshal(msg.Value, &order); err != nil {
        return err
    }
    // Process order...
}
```

### Avro/Protobuf Messages

Use serialization libraries:

```go
// With Protocol Buffers
import "google.golang.org/protobuf/proto"

// Serialize
data, _ := proto.Marshal(orderProto)
msg.Value = data

// Deserialize
var order orderpb.Order
proto.Unmarshal(msg.Value, &order)
```

## Performance Tips

1. **Batch Processing**: Process multiple messages if possible
2. **Commit Strategy**: Use manual commits for critical data
3. **Partition Key**: Use consistent keys for ordering guarantees
4. **Consumer Instances**: Match partition count for optimal parallelism
5. **Async Processing**: Don't block handler - use worker pools if needed

## Lifecycle

The Kafka consumer integrates with Fx lifecycle:

- **OnStart**: Connects to Kafka and starts consuming messages
- **OnStop**: 
  - Stops accepting new messages
  - Waits for in-flight messages to complete (graceful shutdown)
  - Commits final offsets
  - Closes Kafka connection

No manual lifecycle management needed!

## Troubleshooting

### Connection Refused

```
Error: dial tcp: connect: connection refused
```

**Solution**: Ensure Kafka is running:
```bash
# Docker
docker-compose up kafka

# Or use Confluent Platform
confluent local services start
```

### Consumer Group Rebalancing

```
Error: group rebalancing in progress
```

**Solution**: This is normal when consumers join/leave. Configure:
```yaml
kafka:
  session_timeout_ms: 30000
  heartbeat_interval_ms: 3000
```

### Offset Commit Failed

```
Error: offset commit failed
```

**Solution**: Check broker connectivity and permissions. Enable auto-commit or manually commit after processing.

### Messages Not Being Consumed

1. Check topic exists: `kafka-topics --list --bootstrap-server localhost:9092`
2. Check consumer group: `kafka-consumer-groups --describe --group my-group --bootstrap-server localhost:9092`
3. Verify topic has messages: `kafka-console-consumer --topic my-topic --from-beginning --bootstrap-server localhost:9092`

## Alternatives

Want to use a different message broker? Create your own implementation:

- **RabbitMQ**: Create `module/rabbitmq` implementing `messaging.Consumer`
- **NATS**: Lightweight pub/sub messaging
- **AWS SQS**: Cloud-native queuing
- **Redis Streams**: Simple stream processing

See [module/messaging README](../messaging/README.md) for guidance on creating custom implementations.

## Dependencies

- `github.com/segmentio/kafka-go` - Kafka client
- `github.com/things-kit/module/messaging` - Messaging interfaces
- `github.com/things-kit/module/log` - Logging interface
- `github.com/spf13/viper` - Configuration
- `go.uber.org/fx` - Dependency injection

## Best Practices

1. **Idempotent Handlers**: Design handlers to safely process duplicate messages
2. **Error Handling**: Return errors for retryable failures, nil for permanent failures
3. **Monitoring**: Track lag, processing time, and error rates
4. **Schema Evolution**: Use versioned schemas (Avro Schema Registry, Protobuf)
5. **Backpressure**: Don't overwhelm downstream services

## License

MIT License - see LICENSE file for details
