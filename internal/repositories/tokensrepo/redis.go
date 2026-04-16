package tokensrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type redisRepo struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) ports.TokensRepository {
	return &redisRepo{
		client: client,
	}
}

func (repo *redisRepo) Create(ctx context.Context, token *domain.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}
	opts := redis.SetArgs{
		Mode:     `NX`,
		ExpireAt: token.ExpiresAt,
	}
	if _, err := repo.client.SetArgs(ctx, token.ID, data, opts).Result(); err != nil {
		return fmt.Errorf("'client.SetArgs' failed: %w", err)
	}
	return nil
}

func (repo *redisRepo) Get(ctx context.Context, id string) (domain.Token, error) {
	data, err := repo.client.Get(ctx, id).Result()
	if err == redis.Nil {
		return domain.Token{}, ports.ErrTokenNotFound
	} else if err != nil {
		return domain.Token{}, fmt.Errorf("'client.Get' failed: %w", err)
	}
	var token domain.Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return domain.Token{}, fmt.Errorf("'json.Unmarshal' failed: %w", err)
	}
	return domain.Token{}, nil
}

func (repo *redisRepo) Delete(ctx context.Context, id string) error {
	if _, err := repo.client.Del(ctx, id).Result(); err != nil {
		return fmt.Errorf("'client.Del' failed: %w", err)
	}
	return nil
}
