package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PostgresConfig struct {
	Username           string
	Password           string
	Host               string
	Port               int
	Name               string
	SSLMode            string
	MaxOpenConnections int
	MaxIdelConnections int
	MaxIdelTime        time.Duration
	StartupTimeout     time.Duration
	PingTimeout        time.Duration
}

func ConnectToPostgres(cfg PostgresConfig) (*sql.DB, error) {
	// TODO: user a better connection string format
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode))
	if err != nil {
		return nil, fmt.Errorf("'sql.Open' failed: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdelConnections)
	db.SetConnMaxIdleTime(cfg.MaxIdelTime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
