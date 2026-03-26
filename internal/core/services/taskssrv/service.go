package taskssrv

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
)

type Config struct {
	ContentMaxChars int
}

func DefaultConfig() Config {
	return Config{
		ContentMaxChars: 1024,
	}
}

type service struct {
	Config
	tasksRepo ports.TasksRepository
}

func New(tasksRepo ports.TasksRepository, config Config) ports.TasksService {
	return &service{
		Config:    config,
		tasksRepo: tasksRepo,
	}
}

func (srv *service) Create(ctx context.Context, user domain.User, req ports.CreateTaskRequest) (ports.CreateTaskResponse, error) {
	v := failures.NewValidator()

	v.CheckNotEmpty("list_id", req.ListID)
	v.CheckUTF8("list_id", req.ListID)

	v.CheckUTF8("content", req.Content)
	v.CheckNotEmpty("content", req.Content)
	v.CheckAtMostInc("content", utf8.RuneCountInString(req.Content), srv.ContentMaxChars, "characters long")

	if err := v.Err(); err != nil {
		return ports.CreateTaskResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	task := &domain.Task{
		ListID:      req.ListID,
		Content:     req.Content,
		IsCompleted: false,
	}
	if err := srv.tasksRepo.Create(ctx, task); err != nil {
		return ports.CreateTaskResponse{}, fmt.Errorf("'tasksRepo.Create' failed: %w", err)
	}

	resp := ports.CreateTaskResponse{
		Task: task,
	}
	return resp, nil
}

func (srv *service) Get(ctx context.Context, req ports.GetTaskRequest) (ports.GetTaskResponse, error) {
	v := failures.NewValidator()
	v.CheckUTF8("id", req.ID)
	v.CheckNotEmpty("id", req.ID)

	if err := v.Err(); err != nil {
		return ports.GetTaskResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	task, err := srv.tasksRepo.Get(ctx, req.ID)
	if err != nil {
		return ports.GetTaskResponse{}, fmt.Errorf("'tasksRepo.Get' failed: %w", err)
	}

	resp := ports.GetTaskResponse{
		Task: &task,
	}
	return resp, nil
}

func (srv *service) GetAll(ctx context.Context, req ports.GetTasksRequest) (ports.GetTasksResponse, error) {
	v := failures.NewValidator()
	v.CheckNotEmpty("list_id", req.ListID)
	v.Check(req.Page > 0, "page", "must be positive")
	v.Check(req.PageSize > 0, "page_size", "must be positive")
	v.CheckUTF8(req.Content, "content")
	v.CheckAtMostInc(req.Content, utf8.RuneCountInString(req.Content), srv.ContentMaxChars, "characters long")
	if err := v.Err(); err != nil {
		return ports.GetTasksResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	tasks, total, err := srv.tasksRepo.GetAll(ctx, req.ListID, req.Page, req.PageSize, req.Sort, req.Content, req.IsCompleted)
	if err != nil {
		return ports.GetTasksResponse{}, fmt.Errorf("'tasksRepo.GetAll' failed: %w", err)
	}

	resp := ports.GetTasksResponse{
		Tasks: tasks,
		Total: total,
	}

	return resp, nil
}

func (srv *service) Update(ctx context.Context, req ports.UpdateTaskRequest) (ports.UpdateTaskResponse, error) {
	v := failures.NewValidator()

	v.CheckUTF8("id", req.ID)
	v.CheckNotEmpty("id", req.ID)

	if req.Content != nil {
		v.CheckUTF8("content", *req.Content)
		v.CheckNotEmpty("title", *req.Content)
		v.CheckAtMostInc("title", utf8.RuneCountInString(*req.Content), srv.ContentMaxChars, "characters long")
	}

	if err := v.Err(); err != nil {
		return ports.UpdateTaskResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	task, err := srv.tasksRepo.Get(ctx, req.ID)
	if err != nil {
		return ports.UpdateTaskResponse{}, fmt.Errorf("'tasksRepo.Get' failed: %w", err)
	}

	if req.Content != nil {
		task.Content = *req.Content
	}
	if req.IsCompleted != nil {
		task.IsCompleted = *req.IsCompleted
	}

	if err := srv.tasksRepo.Update(ctx, &task); err != nil {
		return ports.UpdateTaskResponse{}, fmt.Errorf("'tasksRepo.Update' failed: %w", err)
	}

	resp := ports.UpdateTaskResponse{
		Task: &task,
	}
	return resp, nil
}

func (srv *service) Delete(ctx context.Context, req ports.DeleteTaskRequest) (ports.DeleteTaskResponse, error) {
	v := failures.NewValidator()
	v.CheckUTF8("id", req.ID)
	v.CheckNotEmpty("id", req.ID)
	if err := v.Err(); err != nil {
		return ports.DeleteTaskResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	if err := srv.tasksRepo.Delete(ctx, req.ID); err != nil {
		return ports.DeleteTaskResponse{}, fmt.Errorf("'tasksRepo.Delete' failed: %w", err)
	}

	resp := ports.DeleteTaskResponse{
		Message: "task was deleted successfully",
	}

	return resp, nil
}
