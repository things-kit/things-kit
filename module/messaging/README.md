# module/messaging - Messaging Abstractions

This module defines messaging abstractions for Things-Kit. It contains **only interfaces**, no implementation.

## Purpose

The `module/messaging` package defines contracts for message-driven architectures. It provides interfaces for:
- **Message Handling**: Process incoming messages
- **Consumer Lifecycle**: Start/stop message consumption
- **Producer Operations**: Publish messages to queues/topics

This allows applications to program against stable interfaces while being free to choose any message broker (Kafka, RabbitMQ, NATS, SQS, etc.).

## Interfaces

### Handler

The `Handler` interface defines how to process incoming messages:

```go
type Handler interface {
    Handle(ctx context.Context, msg Message) error
}
```

This is the core abstraction - your application logic implements this interface to process messages.

### Consumer

The `Consumer` interface defines the lifecycle of a message consumer:

```go
type Consumer interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

Implementations manage the connection to the message broker and deliver messages to registered handlers.

### Producer

The `Producer` interface defines operations for publishing messages:

```go
type Producer interface {
    Publish(ctx context.Context, msg Message) error
    PublishBatch(ctx context.Context, messages []Message) error
    Close() error
}
```

### Message

The `Message` struct represents a message flowing through the system:

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

## Available Implementations

### module/kafka (Default)

The [kafka module](../kafka/) provides Kafka-based implementations of Consumer. It's the recommended default for distributed systems.

```go
import (
    "github.com/things-kit/module/messaging"
    "github.com/things-kit/module/kafka"
)

func main() {
    app.New(
        viperconfig.Module,
        logging.Module,
        kafka.Module,  // Provides messaging capabilities
        
        fx.Provide(NewMyHandler),
        fx.Invoke(kafka.RunConsumer),
    ).Run()
}

type MyHandler struct {
    log log.Logger
}

func (h *MyHandler) Handle(ctx context.Context, msg messaging.Message) error {
    h.log.Info("Received message", 
        "topic", msg.Topic,
        "key", string(msg.Key),
        "value", string(msg.Value),
    )
    return nil
}
```

### Custom Implementations

You can create your own messaging implementation using any broker:

- **RabbitMQ**: AMQP-based message broker
- **NATS**: Lightweight, high-performance messaging
- **AWS SQS/SNS**: Cloud-native queuing and pub/sub
- **Redis Streams**: Redis-based message streams
- **Google Pub/Sub**: GCP messaging service
- **Azure Service Bus**: Azure messaging platform

## Creating Your Own Implementation

### Step 1: Implement Handler

Your business logic implements the `Handler` interface:

```go
package myservice

import (
    "context"
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
        return fmt.Errorf("invalid message: %w", err)
    }
    
    h.log.Info("Processing order", "id", order.ID)
    
    if err := h.db.SaveOrder(ctx, &order); err != nil {
        return fmt.Errorf("failed to save order: %w", err)
    }
    
    return nil
}
```

### Step 2: Implement Consumer (Broker Integration)

Create a consumer for your message broker:

```go
package rabbitmq

import (
    "context"
    "github.com/things-kit/module/messaging"
    amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    handler messaging.Handler
    queue   string
}

func NewRabbitMQConsumer(config Config, handler messaging.Handler) (*RabbitMQConsumer, error) {
    conn, err := amqp.Dial(config.URL)
    if err != nil {
        return nil, err
    }
    
    channel, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    
    return &RabbitMQConsumer{
        conn:    conn,
        channel: channel,
        handler: handler,
        queue:   config.Queue,
    }, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
    msgs, err := c.channel.Consume(
        c.queue, // queue
        "",      // consumer
        false,   // auto-ack
        false,   // exclusive
        false,   // no-local
        false,   // no-wait
        nil,     // args
    )
    if err != nil {
        return err
    }
    
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case delivery := <-msgs:
                msg := messaging.Message{
                    Key:   []byte(delivery.MessageId),
                    Value: delivery.Body,
                    Headers: map[string]string{
                        "content-type": delivery.ContentType,
                    },
                    Timestamp: delivery.Timestamp,
                }
                
                if err := c.handler.Handle(ctx, msg); err != nil {
                    delivery.Nack(false, true) // Requeue on error
                } else {
                    delivery.Ack(false)
                }
            }
        }
    }()
    
    return nil
}

func (c *RabbitMQConsumer) Stop(ctx context.Context) error {
    if c.channel != nil {
        c.channel.Close()
    }
    if c.conn != nil {
        c.conn.Close()
    }
    return nil
}
```

### Step 3: Create Fx Module

```go
package rabbitmq

import (
    "github.com/things-kit/module/messaging"
    "go.uber.org/fx"
)

var Module = fx.Module("rabbitmq",
    fx.Provide(
        NewConfig,
        NewRabbitMQConsumer,
        fx.Annotate(
            func(c *RabbitMQConsumer) messaging.Consumer {
                return c
            },
            fx.As(new(messaging.Consumer)),
        ),
    ),
)

func RunConsumer(lc fx.Lifecycle, consumer messaging.Consumer) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            return consumer.Start(ctx)
        },
        OnStop: func(ctx context.Context) error {
            return consumer.Stop(ctx)
        },
    })
}
```

### Step 4: Implement Producer (Optional)

```go
type RabbitMQProducer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func (p *RabbitMQProducer) Publish(ctx context.Context, msg messaging.Message) error {
    return p.channel.PublishWithContext(ctx,
        "",        // exchange
        msg.Topic, // routing key (queue name)
        false,     // mandatory
        false,     // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        msg.Value,
            MessageId:   string(msg.Key),
            Timestamp:   msg.Timestamp,
        },
    )
}

func (p *RabbitMQProducer) PublishBatch(ctx context.Context, messages []messaging.Message) error {
    for _, msg := range messages {
        if err := p.Publish(ctx, msg); err != nil {
            return err
        }
    }
    return nil
}

func (p *RabbitMQProducer) Close() error {
    if p.channel != nil {
        p.channel.Close()
    }
    if p.conn != nil {
        p.conn.Close()
    }
    return nil
}
```

## Usage Patterns

### Simple Handler

```go
type EmailHandler struct {
    emailer *EmailService
}

func (h *EmailHandler) Handle(ctx context.Context, msg messaging.Message) error {
    var email Email
    if err := json.Unmarshal(msg.Value, &email); err != nil {
        return err
    }
    
    return h.emailer.Send(ctx, &email)
}
```

### Handler with Retries

```go
func (h *OrderHandler) Handle(ctx context.Context, msg messaging.Message) error {
    const maxRetries = 3
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err := h.processOrder(ctx, msg)
        if err == nil {
            return nil
        }
        
        if attempt < maxRetries {
            h.log.Warn("Retry processing", 
                "attempt", attempt,
                "error", err,
            )
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }
        
        return err
    }
    
    return nil
}
```

### Handler with Dead Letter Queue

```go
func (h *PaymentHandler) Handle(ctx context.Context, msg messaging.Message) error {
    err := h.processPayment(ctx, msg)
    if err != nil {
        // Send to DLQ for manual investigation
        dlqMsg := messaging.Message{
            Topic: "payments-dlq",
            Value: msg.Value,
            Headers: map[string]string{
                "original-topic": msg.Topic,
                "error":          err.Error(),
                "timestamp":      time.Now().String(),
            },
        }
        
        _ = h.dlqProducer.Publish(ctx, dlqMsg)
        return err
    }
    
    return nil
}
```

### Multiple Handlers

Register multiple handlers for different message types:

```go
type MessageRouter struct {
    handlers map[string]messaging.Handler
}

func NewMessageRouter() *MessageRouter {
    return &MessageRouter{
        handlers: make(map[string]messaging.Handler),
    }
}

func (r *MessageRouter) Register(messageType string, handler messaging.Handler) {
    r.handlers[messageType] = handler
}

func (r *MessageRouter) Handle(ctx context.Context, msg messaging.Message) error {
    msgType := msg.Headers["type"]
    
    handler, ok := r.handlers[msgType]
    if !ok {
        return fmt.Errorf("no handler for message type: %s", msgType)
    }
    
    return handler.Handle(ctx, msg)
}

// Usage
router := NewMessageRouter()
router.Register("order.created", orderHandler)
router.Register("payment.processed", paymentHandler)
router.Register("email.send", emailHandler)
```

## Testing

### Mock Handler

```go
type MockHandler struct {
    mock.Mock
}

func (m *MockHandler) Handle(ctx context.Context, msg messaging.Message) error {
    args := m.Called(ctx, msg)
    return args.Error(0)
}

// Test
func TestConsumer(t *testing.T) {
    handler := new(MockHandler)
    handler.On("Handle", mock.Anything, mock.MatchedBy(func(msg messaging.Message) bool {
        return msg.Topic == "test-topic"
    })).Return(nil)
    
    // Test your consumer with mock handler
}
```

### Mock Consumer

```go
type MockConsumer struct {
    mock.Mock
}

func (m *MockConsumer) Start(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockConsumer) Stop(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}
```

### Integration Tests

```go
func TestKafkaIntegration(t *testing.T) {
    // Start Kafka container
    container := testing.StartKafkaContainer(t)
    defer container.Terminate(context.Background())
    
    // Configure and test with real Kafka
    viper.Set("kafka.brokers", container.Brokers())
    
    // Test consumer and producer
    // ...
}
```

## Design Philosophy

This interface abstraction follows **"Program to an Interface, Not an Implementation"**:

- **Applications** implement `messaging.Handler` for business logic
- **Broker integrations** implement `messaging.Consumer` and `messaging.Producer`
- **Flexibility**: Swap message brokers without changing business logic

Benefits:
1. **Broker Independence**: Business logic doesn't depend on Kafka, RabbitMQ, etc.
2. **Testing**: Easy to mock handlers and consumers
3. **Migration**: Switch brokers without rewriting handlers
4. **Multi-Broker**: Use different brokers for different services

## Best Practices

1. **Idempotency**: Design handlers to safely handle duplicate messages
2. **Error Handling**: Return errors for transient failures, log permanent failures
3. **Timeouts**: Use context deadlines to prevent hanging handlers
4. **Monitoring**: Track message processing times and error rates
5. **Schema Evolution**: Use versioned message schemas
6. **Batching**: Use `PublishBatch` for bulk operations when possible

## Examples

See the [kafka module](../kafka/) for a complete implementation example.

## License

MIT License - see LICENSE file for details
