package ports

import (
	"context"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
)

var ErrTaskNotFound = &failures.ResourceNotFoundError{Name: "task"}

type TasksRepository interface {
	Create(context.Context, *domain.Task) error
	Get(context.Context, string) (domain.Task, error)
	Update(context.Context, *domain.Task) error
	Delete(context.Context, string) error
}

type TasksService interface {
	Create(context.Context, domain.User, CreateTaskRequest) (CreateTaskResponse, error)
	Get(context.Context, GetTaskRequest) (GetTaskResponse, error)
	Update(context.Context, UpdateTaskRequest) (UpdateTaskResponse, error)
	Delete(context.Context, DeleteTaskRequest) (DeleteTaskResponse, error)
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
