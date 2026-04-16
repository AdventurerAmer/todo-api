package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/utils"
)

type Queue string

const (
	EmailQueue Queue = "email"
)

var Queues = []Queue{EmailQueue}

type ConsumeHandler = func(ctx context.Context, contentType string, data []byte) (bool, error)

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

type EmailSender struct {
	Context      context.Context
	Broker       Broker
	Timeout      time.Duration
	MaxRetries   int
	MaxRetryTime time.Duration
}

func (s EmailSender) SendAsync(req SendEmailRequest) {
	go func() {
		data, err := json.Marshal(req)
		if err != nil {
			slog.Error("send email failed", "userID", req.UserID, "template", req.Template, "error", err)
			return
		}
		handler := func(ctx context.Context) (bool, error) {
			err := s.Broker.Publish(s.Context, EmailQueue, "application/json", data)
			if err != nil {
				return true, fmt.Errorf("'Publish' failed: %w", err)
			}
			return false, nil
		}
		if err := utils.Retry(s.Context, handler, s.Timeout, s.MaxRetries, s.MaxRetryTime); err != nil {
			slog.Error("send email failed", "userID", req.UserID, "template", req.Template, "error", err)
		}
	}()
}
