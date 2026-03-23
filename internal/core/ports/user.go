package ports

import (
	"context"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

var ErrUserNotFound = &domain.ResourceNotFoundError{Name: "user"}
var ErrUserAlreadyExists = &domain.ResourceAlreadyExistsError{Name: "user"}

type UsersRepository interface {
	Create(context.Context, *domain.User) error
	Get(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Update(context.Context, *domain.User) error
	Delete(ctx context.Context, id string) error
}

type UsersService interface {
	Create(context.Context, CreateUserRequest) (CreateUserResponse, error)
	Get(context.Context, GetUserRequest) (GetUserResponse, error)
	Update(context.Context, *domain.User, UpdateUserRequest) (UpdateUserResponse, error)
	Delete(context.Context, DeleteUserRequest) (DeleteUserResponse, error)
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	User    *domain.User `json:"user"`
	Message string       `json:"message"`
}

type GetUserRequest struct {
	ID string `json:"id"`
}

type GetUserResponse struct {
	User *domain.User `json:"user"`
}

type UpdateUserRequest struct {
	Name *string `json:"name"`
}

type UpdateUserResponse struct {
	User *domain.User `json:"user"`
}

type DeleteUserRequest struct {
	ID string `json:"user"`
}

type DeleteUserResponse struct {
	Message string `json:"message"`
}
