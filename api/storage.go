package main

import (
	"context"
	"fmt"

	"database/sql"

	"github.com/AdventurerAmer/todo-api/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openDB(cfg config.MainDB) (*sql.DB, error) {
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
