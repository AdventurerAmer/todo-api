package ports

import (
	"context"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

type TokenAuthService interface {
	Create(ctx context.Context, req CreateAuthTokenRequest) (CreateAuthTokenResponse, error)
	Check(ctx context.Context, req CheckAuthTokenRequest) (CheckAuthTokenResponse, error)
}

type CreateAuthTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAuthTokenResponse struct {
	Token string `json:"token"`
}

type CheckAuthTokenRequest struct {
	Token string `json:"token"`
}

type CheckAuthTokenResponse struct {
	User domain.User `json:"user"`
}
