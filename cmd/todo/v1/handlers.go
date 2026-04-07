package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/config"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/web"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Status      string             `json:"status"`
		Environment config.Environment `json:"env"`
		Version     string             `json:"version"`
	}{
		Status:      "available",
		Environment: app.config.Env,
		Version:     version,
	}
	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) authenticateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateAuthTokenRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.tokenAuthService.Create(ctx, req)
	if err != nil {
		err := fmt.Errorf("'tokenAuthService.Create' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusCreated)
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateUserRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.usersService.Create(ctx, req)
	if err != nil {
		err := fmt.Errorf("'usersService.Create' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusCreated)
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.UpdateUserRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	user := mustGetUserFromCtx(ctx)

	resp, err := app.usersService.Update(ctx, &user, req)
	if err != nil {
		err := fmt.Errorf("'usersService.Update' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getUserFromCtx(r.Context())
	resp := map[string]any{"user": user}
	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	user := mustGetUserFromCtx(ctx)
	req := ports.DeleteUserRequest{ID: user.ID}
	resp, err := app.usersService.Delete(ctx, req)
	if err != nil {
		err := fmt.Errorf("'usersService.Delete' failed: %w", err)
		web.WriteError(w, err)
		return
	}
	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) sendActivationCodeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	user := mustGetUserFromCtx(ctx)

	resp, err := app.tokensService.ActivateViaEmail(ctx, user)
	if err != nil {
		err := fmt.Errorf("'tokensService.ActivateViaEmail' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.ActivateUserRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.tokensService.ActivateUser(ctx, ports.ActivateUserRequest{Token: req.Token})
	if err != nil {
		err := fmt.Errorf("'tokensService.ActivateUser' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) createListHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateListRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.listsService.Create(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		err := fmt.Errorf("'listsService.Create' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusCreated)
}

func (app *application) updateListHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.UpdateListRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
	}
	req.ID = web.Path(r, "id")

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.listsService.Update(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		err := fmt.Errorf("'listsService.Update' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.GetListRequest{ID: web.Path(r, "id")}
	resp, err := app.listsService.Get(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		err := fmt.Errorf("'listsService.Get' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getListsHandler(w http.ResponseWriter, r *http.Request) {
	v := failures.NewValidator()

	page, err := web.QueryInt(r, "page", 1)
	v.Check(err == nil, "page", "invalid number")

	pageSize, err := web.QueryInt(r, "page_size", 20) // TODO: hardcoding...
	v.Check(err == nil, "page_size", "invalid number")

	if err := v.Err(); err != nil {
		err := fmt.Errorf("validation failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.GetListsRequest{
		Page:     page,
		PageSize: pageSize,
		Sort:     web.Query(r, "sort", "created_at"),
		Title:    web.Query(r, "title", ""),
	}
	resp, err := app.listsService.GetAll(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		err := fmt.Errorf("'listsService.GetAll' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) deleteListandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.DeleteListRequest{ID: web.Path(r, "id")}
	resp, err := app.listsService.Delete(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		web.WriteError(w, fmt.Errorf("'listsService.Delete' failed: %w", err))
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateTaskRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	user := mustGetUserFromCtx(ctx)

	resp, err := app.tasksService.Create(ctx, user, req)
	if err != nil {
		err := fmt.Errorf("'tasksService.Create' failed: %w", err)
		web.WriteError(w, err)
		return
	}
	web.WriteJSON(w, resp, http.StatusCreated)
}

func (app *application) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.UpdateTaskRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
	}
	req.ID = web.Path(r, "id")

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.tasksService.Update(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		err := fmt.Errorf("'tasksService.Update' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.GetTaskRequest{ID: web.Path(r, "id")}
	resp, err := app.tasksService.Get(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		err := fmt.Errorf("'tasksService.Get' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getTasksHandler(w http.ResponseWriter, r *http.Request) {
	v := failures.NewValidator()

	page, err := web.QueryInt(r, "page", 1)
	v.Check(err == nil, "page", "invalid number")

	pageSize, err := web.QueryInt(r, "page_size", 20) // TODO: hardcoding...
	v.Check(err == nil, "page_size", "invalid number")

	isCompleted, err := web.QueryBool(r, "is_completed")
	v.Check(err == nil, "page_size", "invalid bool")

	if err := v.Err(); err != nil {
		err := fmt.Errorf("validation failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.GetTasksRequest{
		ListID:      web.Path(r, "id"),
		Page:        page,
		PageSize:    pageSize,
		Sort:        web.Query(r, "sort", "created_at"),
		Content:     web.Query(r, "content", ""),
		IsCompleted: isCompleted,
	}
	resp, err := app.tasksService.GetAll(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		err := fmt.Errorf("'tasksService.GetAll' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.DeleteTaskRequest{ID: web.Path(r, "id")}
	resp, err := app.tasksService.Delete(ctx, mustGetUserFromCtx(ctx), req)
	if err != nil {
		err := fmt.Errorf("'tasksService.Delete' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}
