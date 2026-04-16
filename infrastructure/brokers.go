package infrastructure

import (
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	Addr              string
	Username          string
	Password          string
	ConnectionTimeout time.Duration
}

func ConnectToRabbitMQ(cfg RabbitMQConfig) (*amqp.Connection, error) {
	connStr := fmt.Sprintf("amqp://%s:%s@%s/", cfg.Username, cfg.Password, cfg.Addr)
	type connResult struct {
		conn *amqp.Connection
		err  error
	}
	resCh := make(chan connResult)
	go func() {
		conn, err := amqp.Dial(connStr)
		resCh <- connResult{conn: conn, err: err}
	}()
	select {
	case <-time.After(cfg.ConnectionTimeout):
		return nil, errors.New("timeout")
	case res := <-resCh:
		return res.conn, res.err
	}
}
