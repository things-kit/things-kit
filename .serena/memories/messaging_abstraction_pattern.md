# Messaging Abstraction Pattern

Things-Kit implements comprehensive messaging abstractions to support message-driven architectures with any broker.

## Architecture

### Interface Module: `module/messaging`
- Defines `messaging.Handler` interface for processing messages
- Defines `messaging.Consumer` interface for consumer lifecycle management
- Defines `messaging.Producer` interface for publishing messages
- Defines `messaging.Message` struct as universal message format
- Contains **NO implementation** - only contracts

### Default Implementation: `module/kafka`
- Implements `messaging.Consumer` using segmentio/kafka-go
- Provides KafkaConsumer with consumer group support
- Configured via Viper with `kafka.brokers`, `kafka.group`, `kafka.topics`

## Messaging Interfaces

### Handler (Business Logic)
```go
type Handler interface {
    Handle(ctx context.Context, msg Message) error
}
```

Applications implement this to process messages. Return `nil` to acknowledge, return `error` to reject/retry.

### Consumer (Broker Integration)
```go
type Consumer interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

Broker integrations implement this to manage message consumption lifecycle.

### Producer (Publishing)
```go
type Producer interface {
    Publish(ctx context.Context, msg Message) error
    PublishBatch(ctx context.Context, messages []Message) error
    Close() error
}
```

For sending messages to topics/queues.

### Message (Universal Format)
```go
type Message struct {
    Topic     string
    Key       []byte
    Value     []byte
    Headers   map[string]string
    Partition int
    Offset    int64
    Timestamp time.Time
}
```

Broker-agnostic message representation.

## Usage Pattern

### 1. Implement Handler
```go
type OrderHandler struct {
    log log.Logger
    db  *Database
}

func (h *OrderHandler) Handle(ctx context.Context, msg messaging.Message) error {
    var order Order
    json.Unmarshal(msg.Value, &order)
    
    h.log.Info("Processing order", "id", order.ID, "topic", msg.Topic)
    return h.db.SaveOrder(ctx, &order)
}
```

### 2. Add Kafka Module and Start Consumer
```go
app.New(
    viperconfig.Module,
    logging.Module,
    kafka.Module,
    
    fx.Provide(NewOrderHandler),
    fx.Invoke(kafka.RunConsumer),
).Run()
```

### 3. Configure Topics
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

## Kafka Consumer Features

- **Consumer Groups**: Automatic partition rebalancing and load distribution
- **Multiple Topics**: Subscribe to multiple topics with single consumer
- **Auto-Commit**: Configurable offset commit strategy (auto or manual)
- **Graceful Shutdown**: Waits for in-flight messages before stopping
- **Lifecycle Integration**: Managed by Fx hooks (OnStart/OnStop)

## Advanced Patterns

### Manual Offset Commits
```yaml
kafka:
  auto_commit: false
```

```go
func (h *Handler) Handle(ctx context.Context, msg messaging.Message) error {
    if err := h.process(msg); err != nil {
        return err  // Don't commit - message will be reprocessed
    }
    return nil  // Commit offset
}
```

### Message Router (Multiple Handlers)
```go
type MessageRouter struct {
    handlers map[string]messaging.Handler
}

func (r *MessageRouter) Handle(ctx context.Context, msg messaging.Message) error {
    handler := r.handlers[msg.Topic]
    return handler.Handle(ctx, msg)
}
```

### Retry with Backoff
```go
func (h *Handler) Handle(ctx context.Context, msg messaging.Message) error {
    for attempt := 1; attempt <= 3; attempt++ {
        if err := h.process(msg); err == nil {
            return nil
        }
        time.Sleep(time.Second * time.Duration(attempt))
    }
    return fmt.Errorf("max retries exceeded")
}
```

### Dead Letter Queue
```go
func (h *Handler) Handle(ctx context.Context, msg messaging.Message) error {
    if err := h.process(msg); err != nil {
        h.sendToDLQ(msg, err)
        return nil  // Acknowledge to skip
    }
    return nil
}
```

## Creating Alternative Implementations

To use a different message broker:

1. **Implement `messaging.Consumer` interface** for your broker
2. **Create Fx module** providing the consumer
3. **Implement lifecycle hooks** (Start/Stop)
4. **Use in application** instead of kafka.Module

Examples:
- **RabbitMQ**: AMQP-based messaging
- **NATS**: Lightweight pub/sub
- **AWS SQS/SNS**: Cloud-native queuing
- **Redis Streams**: Stream processing
- **Google Pub/Sub**: GCP messaging
- **Azure Service Bus**: Azure messaging

## Benefits

1. **Broker Independence**: Business logic doesn't depend on Kafka/RabbitMQ/etc.
2. **Easy Testing**: Mock `messaging.Handler` for unit tests
3. **Migration Path**: Switch brokers without rewriting handlers
4. **Flexibility**: Use different brokers for different services
5. **Consistent API**: Same message handling across all brokers

## Configuration

Kafka configuration:
```yaml
kafka:
  brokers: ["kafka-1:9092", "kafka-2:9092"]
  group: "my-service"
  topics: ["events", "commands"]
  auto_commit: true
  start_from_beginning: false
```

Alternative implementations would have their own config keys.

## Files

- `/module/messaging/message.go` - Interface definitions (Handler, Consumer, Producer, Message)
- `/module/messaging/go.mod` - Interface module (no dependencies)
- `/module/messaging/README.md` - Complete documentation with patterns
- `/module/kafka/consumer.go` - Kafka consumer implementation (KafkaConsumer struct)
- `/module/kafka/module.go` - Kafka Fx module and RunConsumer helper
- `/module/kafka/README.md` - Kafka-specific documentation

## Pattern Consistency

Messaging follows the same abstraction pattern:
- **Logger**: `module/log` (interface) → `module/logging` (Zap)
- **HTTP**: `module/http` (interface) → `module/httpgin` (Gin)
- **Cache**: `module/cache` (interface) → `module/redis` (Redis)
- **Messaging**: `module/messaging` (interfaces) → `module/kafka` (Kafka)

## Consumer vs Handler Distinction

- **Handler**: Your business logic (what to do with messages)
- **Consumer**: Broker integration (how to receive messages)

This separation allows handlers to be broker-agnostic. Same handler can work with Kafka, RabbitMQ, or any other consumer implementation.

## Testing

Mock the Handler interface:
```go
type MockHandler struct {
    mock.Mock
}

func (m *MockHandler) Handle(ctx context.Context, msg messaging.Message) error {
    args := m.Called(ctx, msg)
    return args.Error(0)
}
```

Mock the Consumer interface:
```go
type MockConsumer struct {
    mock.Mock
}

func (m *MockConsumer) Start(ctx context.Context) error {
    return m.Called(ctx).Error(0)
}

func (m *MockConsumer) Stop(ctx context.Context) error {
    return m.Called(ctx).Error(0)
}
```

## Producer (Not Yet Implemented)

The `messaging.Producer` interface is defined but Kafka producer is not implemented yet. This is optional functionality - most services only need consumer capabilities for event-driven architectures.
