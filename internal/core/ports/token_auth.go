package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

type TokenAuthService interface {
	Create(ctx context.Context, req CreateAuthTokenRequest) (CreateAuthTokenResponse, error)
	Refresh(ctx context.Context, user domain.User) (RefreshAuthTokenResponse, error)
	Check(ctx context.Context, req CheckAuthTokenRequest) (CheckAuthTokenResponse, error)
}

type CreateAuthTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAuthTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type RefreshAuthTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type CheckAuthTokenRequest struct {
	Token string `json:"token"`
}

type CheckAuthTokenResponse struct {
	User domain.User `json:"user"`
}
