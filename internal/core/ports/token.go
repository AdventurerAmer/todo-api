package ports

import (
	"context"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

var ErrTokenNotFound = &failures.ResourceNotFoundError{Name: "token"}

type TokensRepository interface {
	Create(ctx context.Context, token *domain.Token) error
	Get(ctx context.Context, id string) (domain.Token, error)
	Delete(ctx context.Context, id string) error
}

type TokensService interface {
	ActivateViaEmail(ctx context.Context, user domain.User) (ActivateViaEmailResponse, error)
	ActivateUser(ctx context.Context, req ActivateUserRequest) (ActivateUserResponse, error)
}

type ActivateUserRequest struct {
	Token string `json:"token"`
}

type ActivateUserResponse struct {
	Message string `json:"message"`
}

type ActivateViaEmailResponse struct {
	Message string `json:"message"`
}
