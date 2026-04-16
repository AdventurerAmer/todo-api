package brokers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rmq struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQ(conn *amqp.Connection) (ports.Broker, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("'conn.Channel' failed: %w", err)
	}
	for _, queue := range ports.Queues {
		if _, err := channel.QueueDeclare(string(queue), true, false, false, false, nil); err != nil {
			return nil, fmt.Errorf("'channel.QueueDeclare' failed: %w", err)
		}
	}
	return &rmq{conn: conn, channel: channel}, nil
}

func (r *rmq) Publish(ctx context.Context, queue ports.Queue, contentType string, data []byte) error {
	pub := amqp.Publishing{
		ContentType: contentType,
		Body:        data,
	}
	if err := r.channel.PublishWithContext(ctx, "", string(queue), false, false, pub); err != nil {
		return fmt.Errorf("'channel.PublishWithContext' failed: %w", err)
	}
	return nil
}

func (r *rmq) Consume(ctx context.Context, queue ports.Queue, handler ports.ConsumeHandler) error {
	if err := r.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("'channel.Qos' failed: %w", err)
	}
	msgs, err := r.channel.Consume(string(queue), "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("'channel.Consume' failed: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			requeue, err := handler(ctx, msg.ContentType, msg.Body)
			if err != nil {
				slog.Error("request failed", "err", err)
				if err := msg.Nack(false, requeue); err != nil {
					slog.Error("nack failed", "err", err)
				}
			} else {
				if err := msg.Ack(false); err != nil {
					slog.Error("ack failed", "err", err)
				}
			}
		}
	}
}

func (r *rmq) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("'channel.Close' failed: %w", err)
	}

	return nil
}
