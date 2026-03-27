package main

import (
	"context"
	"sync"
	"time"

	"database/sql"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConnections)
	db.SetMaxIdleConns(cfg.db.maxIdelConnections)
	db.SetConnMaxIdleTime(cfg.db.maxIdelTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
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
	useractivationCache *userActivationCache
}

func newStorage() *storage {
	return &storage{
		useractivationCache: newUserActivationCache(),
	}
}
