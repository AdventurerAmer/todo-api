package ports

import (
	"context"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

var ErrListNotFound = &failures.ResourceNotFoundError{Name: "list"}

type ListsRepository interface {
	Create(ctx context.Context, list *domain.List) error
	Get(ctx context.Context, id string) (domain.List, error)
	GetAll(ctx context.Context, userID string, page, pageSize int, sort, title string) ([]domain.List, int, error)
	Update(ctx context.Context, list *domain.List) error
	Delete(ctx context.Context, id string) error
}

type ListsService interface {
	Create(ctx context.Context, user domain.User, req CreateListRequest) (CreateListResponse, error)
	Get(ctx context.Context, req GetListRequest) (GetListResponse, error)
	GetAll(ctx context.Context, user domain.User, req GetListsRequest) (GetListsResponse, error)
	Update(ctx context.Context, req UpdateListRequest) (UpdateListResponse, error)
	Delete(ctx context.Context, req DeleteListRequest) (DeleteListResponse, error)
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

type GetListsRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Title    string `json:"title"`
}

type GetListsResponse struct {
	Lists []domain.List `json:"lists"`
	Total int           `json:"total"`
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
