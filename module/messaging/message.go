// Package messaging defines framework-level interfaces for message consumers.
// This abstraction allows different messaging implementations (Kafka, RabbitMQ, etc.)
// to be used interchangeably.
package messaging

import (
	"context"
	"time"
)

// Message represents a generic message from a messaging system.
type Message struct {
	Key       []byte
	Value     []byte
	Topic     string
	Timestamp time.Time
}

// Handler defines the interface for handling incoming messages.
// Implementations should process the message and return an error if processing fails.
type Handler interface {
	Handle(ctx context.Context, msg Message) error
}

// Consumer defines the interface for message consumers.
// Implementations manage the lifecycle of consuming messages from a message queue.
// The consumer is responsible for fetching messages and delegating to a Handler.
type Consumer interface {
	// Start begins consuming messages from the message queue.
	// This should be non-blocking and return immediately after the consumer starts.
	// The implementation should start a goroutine for actual message consumption.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the message consumer.
	// It should wait for in-flight message processing to complete within the context deadline.
	Stop(ctx context.Context) error
}

// Producer defines the interface for message producers.
// Implementations publish messages to a message queue or topic.
type Producer interface {
	// Publish sends a message to the specified topic.
	// The key can be used for partitioning in systems like Kafka.
	Publish(ctx context.Context, topic string, key []byte, value []byte) error

	// PublishBatch sends multiple messages to the specified topic efficiently.
	// Returns an error if any message fails to publish.
	PublishBatch(ctx context.Context, topic string, messages []Message) error

	// Close closes the producer and releases resources.
	Close() error
}
