package ports

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
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
	UserID   string `json:"userID"`
	Template string `json:"template"`
	Data     any    `json:"data"`
}

func SendEmail(broker Broker, req SendEmailRequest) {
	b, err := json.Marshal(req)
	if err != nil {
		slog.Error("queue send email request failed", "error", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // TOOD: hardcoding
	defer cancel()
	if err := broker.Publish(ctx, EmailQueue, "application/json", b); err != nil {
		slog.Error("queue send email request failed", "error", err)
	}
}
