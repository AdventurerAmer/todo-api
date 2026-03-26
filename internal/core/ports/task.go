package ports

import (
	"context"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

var ErrTaskNotFound = &failures.ResourceNotFoundError{Name: "task"}

type TasksRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	Get(ctx context.Context, id string) (domain.Task, error)
	GetAll(ctx context.Context, listID string, page, pageSize int, sort, content string, isCompleted *bool) ([]domain.Task, int, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id string) error
}

type TasksService interface {
	Create(ctx context.Context, user domain.User, req CreateTaskRequest) (CreateTaskResponse, error)
	Get(ctx context.Context, req GetTaskRequest) (GetTaskResponse, error)
	GetAll(ctx context.Context, req GetTasksRequest) (GetTasksResponse, error)
	Update(ctx context.Context, req UpdateTaskRequest) (UpdateTaskResponse, error)
	Delete(ctx context.Context, req DeleteTaskRequest) (DeleteTaskResponse, error)
}

type CreateTaskRequest struct {
	ListID  string `json:"list_id"`
	Content string `json:"content"`
}

type CreateTaskResponse struct {
	Task *domain.Task `json:"task"`
}

type GetTaskRequest struct {
	ID string `json:"id"`
}

type GetTaskResponse struct {
	Task *domain.Task `json:"task"`
}

type GetTasksRequest struct {
	ListID      string `json:"list_id"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
	Sort        string `json:"sort"`
	Content     string `json:"conent"`
	IsCompleted *bool  `json:"is_completed"`
}

type GetTasksResponse struct {
	Tasks []domain.Task `json:"tasks"`
	Total int           `json:"total"`
}

type UpdateTaskRequest struct {
	ID          string  `json:"id"`
	Content     *string `json:"content"`
	IsCompleted *bool   `json:"is_completed"`
}

type UpdateTaskResponse struct {
	Task *domain.Task `json:"task"`
}

type DeleteTaskRequest struct {
	ID string `json:"id"`
}

type DeleteTaskResponse struct {
	Message string `json:"message"`
}
