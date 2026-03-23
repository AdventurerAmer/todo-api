package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	_ "github.com/lib/pq"
)

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConnections)
	db.SetMaxIdleConns(cfg.db.maxIdelConnections)
	db.SetConnMaxIdleTime(cfg.db.maxIdelTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type userActivationCacheEntry struct {
	code      uint16
	expiresAt time.Time
}

type userActivationCache struct {
	mu      sync.RWMutex
	entries map[string]userActivationCacheEntry
}

func newUserActivationCache() *userActivationCache {
	c := &userActivationCache{
		entries: make(map[string]userActivationCacheEntry),
	}
	go func(c *userActivationCache) {
		ticker := time.NewTicker(time.Minute)
		for {
			<-ticker.C
			func() {
				c.mu.Lock()
				defer c.mu.Unlock()
				for k, v := range c.entries {
					if time.Now().After(v.expiresAt) {
						delete(c.entries, k)
					}
				}
			}()
		}
	}(c)
	return c
}

func (c *userActivationCache) Set(u *domain.User, code uint16, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[u.ID] = userActivationCacheEntry{
		code:      code,
		expiresAt: time.Now().Add(d),
	}
}

func (c *userActivationCache) Get(u *domain.User) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[u.ID]
	if !ok {
		return 0, true
	}
	return int(e.code), time.Now().After(e.expiresAt)
}

func (c *userActivationCache) Clear(u *domain.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, u.ID)
}

func (c *userActivationCache) HasExpired(u *domain.User) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[u.ID]
	return !ok || time.Now().After(e.expiresAt)
}

type storage struct {
	db                  *sql.DB
	useractivationCache *userActivationCache
}

func newStorage(db *sql.DB) *storage {
	return &storage{
		db:                  db,
		useractivationCache: newUserActivationCache(),
	}
}

func (s *storage) insertTask(u *domain.User, t *task) error {
	query := `INSERT INTO tasks (user_id, content, is_completed)
			  VALUES ($1, $2, $3)
			  RETURNING id, created_at, version`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, u.ID, t.Content, t.IsCompleted).Scan(&t.ID, &t.CreatedAt, &t.Version)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) getTaskByID(id int) (*task, error) {
	query := `SELECT created_at, user_id, content, is_completed, version
			  FROM tasks
			  WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t := &task{
		ID: id,
	}
	err := s.db.QueryRowContext(ctx, query, id).Scan(&t.CreatedAt, &t.UserID, &t.Content, &t.IsCompleted, &t.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil
		default:
			return nil, err
		}
	}
	return t, err
}

func (s *storage) getTasksForUser(u *domain.User, sort string, page, pageSize int, content string) ([]task, int, error) {
	order := "ASC"
	if strings.HasPrefix(sort, "-") {
		order = "DESC"
		sort, _ = strings.CutPrefix(sort, "-")
	}
	sortStr := fmt.Sprintf("%s %s", sort, order)
	if sort != "id" {
		sortStr = fmt.Sprintf("%s %s, id ASC", sort, order)
	}
	limit := pageSize
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT count(*) OVER(), id, created_at, content, is_completed, version
			  FROM tasks
			  WHERE user_id = $1 AND ($2 = '' OR to_tsvector('simple', content) @@ plainto_tsquery('simple', $2))
			  ORDER BY %s
			  LIMIT $3 OFFSET $4`, sortStr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks := make([]task, 0)
	rows, err := s.db.QueryContext(ctx, query, u.ID, content, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	total := 0
	for rows.Next() {
		t := task{
			UserID: u.ID,
		}
		err = rows.Scan(&total, &t.ID, &t.CreatedAt, &t.Content, &t.IsCompleted, &t.Version)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (s *storage) updateTask(t *task) error {
	query := `UPDATE tasks
	          SET content = $1, is_completed = $2, version = version + 1
			  WHERE id = $3 AND version = $4
			  RETURNING version`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, t.Content, t.IsCompleted, t.ID, t.Version).Scan(&t.Version)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) deleteTask(t *task) error {
	query := `DELETE FROM tasks
	          WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, t.ID)
	if err != nil {
		return err
	}
	return nil
}
