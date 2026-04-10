package ports

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

type Queue string

const (
	EmailQueue Queue = "email"
)

var Queues = []Queue{EmailQueue}

type ConsumeHandler = func(contentType string, data []byte) (bool, error)

type Broker interface {
	Publish(ctx context.Context, queue Queue, contentType string, data []byte) error
	Consume(ctx context.Context, queue Queue, handler ConsumeHandler) error
	Close() error
}

type SendEmailRequest struct {
	UserID   string          `json:"userID"`
	Template domain.Template `json:"template"`
	Data     any             `json:"data"`
}

func SendEmail(ctx context.Context, broker Broker, req SendEmailRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}

	if err := broker.Publish(ctx, EmailQueue, "application/json", b); err != nil {
		return fmt.Errorf("'broker.Publish' failed: %w", err)
	}

	return nil
}
