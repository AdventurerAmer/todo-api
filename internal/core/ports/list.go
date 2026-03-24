package ports

import (
	"context"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

var ErrListNotFound = &failures.ResourceNotFoundError{Name: "list"}

type ListsRepository interface {
	Create(context.Context, *domain.List) error
	Get(context.Context, string) (domain.List, error)
	Update(context.Context, *domain.List) error
	Delete(context.Context, string) error
}

type ListsService interface {
	Create(context.Context, domain.User, CreateListRequest) (CreateListResponse, error)
	Get(context.Context, GetListRequest) (GetListResponse, error)
	Update(context.Context, UpdateListRequest) (UpdateListResponse, error)
	Delete(context.Context, DeleteListRequest) (DeleteListResponse, error)
}

type CreateListRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateListResponse struct {
	List *domain.List `json:"list"`
}

type GetListRequest struct {
	ID string `json:"id"`
}

type GetListResponse struct {
	List *domain.List `json:"list"`
}

type UpdateListRequest struct {
	ID          string  `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type UpdateListResponse struct {
	List *domain.List `json:"list"`
}

type DeleteListRequest struct {
	ID string `json:"id"`
}

type DeleteListResponse struct {
	Message string `json:"message"`
}
