package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/joshjms/castletown/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

// Handler processes a single message body. Returning an error indicates the message should be requeued.
type Handler func(context.Context, []byte) error

type Consumer interface {
	Run(ctx context.Context, handler Handler)
}

type amqpConsumer struct {
	cfg      config.RabbitMQConfig
	log      zerolog.Logger
	prefetch int
}

// NewConsumer returns a RabbitMQ-backed consumer. If the configuration is incomplete, nil is returned.
func NewConsumer(cfg config.RabbitMQConfig, log zerolog.Logger, prefetch int) Consumer {
	if cfg.URL == "" || cfg.Queue == "" {
		return nil
	}

	if prefetch <= 0 {
		prefetch = 1
	}
	return &amqpConsumer{
		cfg:      cfg,
		log:      log,
		prefetch: prefetch,
	}
}

func (c *amqpConsumer) Run(ctx context.Context, handler Handler) {
	backoff := time.Second
	for {
		if err := c.consumeOnce(ctx, handler); err != nil {
			c.log.Error().Err(err).Msg("rabbitmq consumer exited")
		} else {
			c.log.Info().Msg("rabbitmq consumer stopped")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		if backoff < 30*time.Second {
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}
}

func (c *amqpConsumer) consumeOnce(ctx context.Context, handler Handler) error {
	conn, err := amqp.Dial(c.cfg.URL)
	if err != nil {
		return fmt.Errorf("connect to rabbitmq: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		c.log.Warn().Err(err).Msg("failed to set rabbitmq qos")
	}

	deliveries, err := ch.Consume(c.cfg.Queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("rabbitmq deliveries channel closed")
			}

			itemCtx, cancel := context.WithCancel(ctx)
			err := handler(itemCtx, msg.Body)
			cancel()

			if err != nil {
				c.log.Error().Err(err).Msg("handler failed, requeueing message")
				_ = msg.Nack(false, true)
				continue
			}

			if err := msg.Ack(false); err != nil {
				c.log.Warn().Err(err).Msg("failed to ack rabbitmq message")
			}
		}
	}
}

// Decode is a helper to unmarshal JSON payloads into a destination struct.
func Decode(body []byte, dst interface{}) error {
	return json.Unmarshal(body, dst)
}
