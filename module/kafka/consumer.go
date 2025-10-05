// Package kafka provides Kafka consumer implementation for Things-Kit applications.
// It implements the messaging.Consumer interface.
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"github.com/things-kit/module/log"
	"github.com/things-kit/module/messaging"
	"go.uber.org/fx"
)

// ConsumerModule provides the Kafka consumer module to the application.
var ConsumerModule = fx.Module("kafka-consumer",
	fx.Provide(
		NewConfig,
		NewKafkaConsumer,
		// Provide as messaging.Consumer interface
		fx.Annotate(
			func(c *KafkaConsumer) messaging.Consumer { return c },
			fx.As(new(messaging.Consumer)),
		),
	),
	fx.Invoke(RunConsumer),
)

// Config holds the Kafka consumer configuration.
type Config struct {
	Brokers  []string      `mapstructure:"brokers"`
	Topic    string        `mapstructure:"topic"`
	GroupID  string        `mapstructure:"group_id"`
	MaxWait  time.Duration `mapstructure:"max_wait"`
	MinBytes int           `mapstructure:"min_bytes"`
	MaxBytes int           `mapstructure:"max_bytes"`
}

// ConsumerParams contains all dependencies needed to run the Kafka consumer.
type ConsumerParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    log.Logger
	Config    *Config
	Handler   messaging.Handler
}

// KafkaConsumer implements the messaging.Consumer interface using Kafka.
type KafkaConsumer struct {
	reader  *kafka.Reader
	handler messaging.Handler
	logger  log.Logger
	cancel  context.CancelFunc
	ctx     context.Context
}

// NewKafkaConsumer creates a new Kafka consumer.
func NewKafkaConsumer(cfg *Config, handler messaging.Handler, logger log.Logger) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MaxWait:  cfg.MaxWait,
		MinBytes: cfg.MinBytes,
		MaxBytes: cfg.MaxBytes,
	})

	return &KafkaConsumer{
		reader:  reader,
		handler: handler,
		logger:  logger,
	}
}

// Start begins consuming messages from Kafka.
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting Kafka consumer",
		log.Field{Key: "topic", Value: c.reader.Config().Topic},
		log.Field{Key: "group_id", Value: c.reader.Config().GroupID},
	)

	// Create a context for the consumer goroutine
	c.ctx, c.cancel = context.WithCancel(context.Background())

	// Start consuming in a goroutine
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				// Read message with context
				msg, err := c.reader.FetchMessage(c.ctx)
				if err != nil {
					if err == context.Canceled {
						return
					}
					c.logger.Error("Failed to fetch Kafka message", err)
					continue
				}

				// Convert to framework message
				frameworkMsg := messaging.Message{
					Key:       msg.Key,
					Value:     msg.Value,
					Topic:     msg.Topic,
					Timestamp: msg.Time,
				}

				// Handle message
				if err := c.handler.Handle(c.ctx, frameworkMsg); err != nil {
					c.logger.ErrorC(c.ctx, "Failed to handle message", err,
						log.Field{Key: "topic", Value: msg.Topic},
						log.Field{Key: "partition", Value: msg.Partition},
						log.Field{Key: "offset", Value: msg.Offset},
					)
					// Continue processing other messages even if one fails
					continue
				}

				// Commit message after successful processing
				if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
					c.logger.ErrorC(c.ctx, "Failed to commit Kafka message", err,
						log.Field{Key: "topic", Value: msg.Topic},
						log.Field{Key: "partition", Value: msg.Partition},
						log.Field{Key: "offset", Value: msg.Offset},
					)
				}
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the Kafka consumer.
func (c *KafkaConsumer) Stop(ctx context.Context) error {
	c.logger.Info("Stopping Kafka consumer")

	// Cancel the consumer context
	if c.cancel != nil {
		c.cancel()
	}

	// Close the reader
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka reader: %w", err)
	}

	return nil
}

// NewConfig creates a new Kafka configuration from Viper.
func NewConfig(v *viper.Viper) *Config {
	cfg := &Config{
		Brokers:  []string{"localhost:9092"},
		Topic:    "events",
		GroupID:  "things-kit-consumer",
		MaxWait:  5 * time.Second,
		MinBytes: 1,
		MaxBytes: 10e6, // 10MB
	}

	// Load configuration from viper
	if v != nil {
		_ = v.UnmarshalKey("kafka", cfg)
	}

	return cfg
}

// RunConsumer starts the Kafka consumer with lifecycle management.
func RunConsumer(p ConsumerParams, consumer *KafkaConsumer) {
	p.Lifecycle.Append(fx.Hook{
		OnStart: consumer.Start,
		OnStop:  consumer.Stop,
	})
}
