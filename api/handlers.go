package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/config"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/web"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

//go:embed templates
var templates embed.FS

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
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := web.ReadJSON(r, &input); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	v := failures.NewValidator()
	v.CheckUTF8Email(input.Email)
	v.CheckUTF8Password(input.Password)
	if err := v.Err(); err != nil {
		err := fmt.Errorf("validation failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	u, err := app.usersRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		err := fmt.Errorf("'usersRepo.GetByEmail' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(input.Password)); err != nil {
		err := fmt.Errorf("'bcrypt.CompareHashAndPassword' failed: %w", &failures.AuthenticationError{Reason: "email or password are not correct"})
		web.WriteError(w, err)
		return
	}

	claims := jwt.MapClaims{
		"user_id":    u.ID,
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC822), // TODO: hardcoding
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(app.config.Authentication.JWTSecret))
	if err != nil {
		err := fmt.Errorf("'token.SignedString' failed: %w", err)
		web.WriteError(w, err)
		return
	}
	resp := map[string]any{"token": tokenStr}
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

	user := mustGetUserFromRequestContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.usersService.Update(ctx, user, req)
	if err != nil {
		err := fmt.Errorf("'usersService.Update' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := mustGetUserFromRequestContext(r)
	resp := map[string]any{"user": user}
	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := mustGetUserFromRequestContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

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

	user := mustGetUserFromRequestContext(r)

	resp, err := app.tokensService.ActivateViaEmail(ctx, *user)
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
	user := mustGetUserFromRequestContext(r)

	var req ports.CreateListRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.listsService.Create(ctx, *user, req)
	if err != nil {
		err := fmt.Errorf("'listsService.Create' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusCreated)
}

func (app *application) updateListHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ports.UpdateListRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
	}
	req.ID = id

	_ = mustGetUserFromRequestContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.listsService.Update(ctx, req)
	if err != nil {
		err := fmt.Errorf("'listsService.Update' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getListHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_ = mustGetUserFromRequestContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.GetListRequest{ID: id}
	resp, err := app.listsService.Get(ctx, req)
	if err != nil {
		err := fmt.Errorf("'listsService.Get' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getListsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sort := query.Get("sort")
	if sort == "" {
		sort = "id"
	}

	page := 1
	pageSize := 20

	pageStr := query.Get("page")
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p <= 0 {
			web.WriteError(w, errors.New(`invalid query parameter "page": must be a positive integer`))
			return
		}
		page = p
	}
	pageSizeStr := query.Get("page_size")
	if pageSizeStr != "" {
		size, err := strconv.Atoi(pageSizeStr)
		if err != nil || size <= 0 {
			web.WriteError(w, errors.New(`invalid query param "page_size": must be a positive integer`))
			return
		}
		pageSize = size
	}

	title := query.Get("title")

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.GetListsRequest{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Title:    title,
	}
	user := mustGetUserFromRequestContext(r)
	resp, err := app.listsService.GetAll(ctx, *user, req)
	if err != nil {
		err := fmt.Errorf("'listsService.GetAll' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) deleteListandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_ = mustGetUserFromRequestContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.DeleteListRequest{ID: id}
	resp, err := app.listsService.Delete(ctx, req)
	if err != nil {
		web.WriteError(w, fmt.Errorf("'listsService.Delete' failed: %w", err))
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := mustGetUserFromRequestContext(r)

	var req ports.CreateTaskRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.tasksService.Create(ctx, *user, req)
	if err != nil {
		err := fmt.Errorf("'tasksService.Create' failed: %w", err)
		web.WriteError(w, err)
		return
	}
	web.WriteJSON(w, resp, http.StatusCreated)
}

func (app *application) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ports.UpdateTaskRequest
	if err := web.ReadJSON(r, &req); err != nil {
		err := fmt.Errorf("'web.ReadJSON' failed: %w", err)
		web.WriteError(w, err)
	}
	req.ID = id

	_ = mustGetUserFromRequestContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	resp, err := app.tasksService.Update(ctx, req)
	if err != nil {
		err := fmt.Errorf("'tasksService.Update' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_ = mustGetUserFromRequestContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.GetTaskRequest{ID: id}
	resp, err := app.tasksService.Get(ctx, req)
	if err != nil {
		err := fmt.Errorf("'tasksService.Get' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) getTasksHandler(w http.ResponseWriter, r *http.Request) {
	_ = mustGetUserFromRequestContext(r)

	query := r.URL.Query()
	sort := query.Get("sort")
	if sort == "" {
		sort = "id"
	}

	listID := query.Get("list_id")

	page := 1
	pageSize := 20

	pageStr := query.Get("page")
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p <= 0 {
			web.WriteError(w, errors.New(`invalid query parameter "page": must be a positive integer`))
			return
		}
		page = p
	}
	pageSizeStr := query.Get("page_size")
	if pageSizeStr != "" {
		size, err := strconv.Atoi(pageSizeStr)
		if err != nil || size <= 0 {
			web.WriteError(w, errors.New(`invalid query param "page_size": must be a positive integer`))
			return
		}
		pageSize = size
	}

	content := query.Get("content")
	var isCompleted *bool
	isCompletedQuery := query.Get("is_completed")
	if isCompletedQuery != "" {
		t := false
		if isCompletedQuery == "true" {
			t = true
		}
		isCompleted = &t
	}

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.GetTasksRequest{
		ListID:      listID,
		Page:        page,
		PageSize:    pageSize,
		Sort:        sort,
		Content:     content,
		IsCompleted: isCompleted,
	}
	resp, err := app.tasksService.GetAll(ctx, req)
	if err != nil {
		err := fmt.Errorf("'tasksService.GetAll' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}

func (app *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_ = mustGetUserFromRequestContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
	defer cancel()

	req := ports.DeleteTaskRequest{ID: id}
	resp, err := app.tasksService.Delete(ctx, req)
	if err != nil {
		err := fmt.Errorf("'tasksService.Delete' failed: %w", err)
		web.WriteError(w, err)
		return
	}

	web.WriteJSON(w, resp, http.StatusOK)
}
