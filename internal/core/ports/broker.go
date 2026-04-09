package ports

import "context"

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
