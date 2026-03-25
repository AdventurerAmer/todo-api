package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

//go:embed templates
var templates embed.FS

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	heathCheck := struct {
		Status      string `json:"status"`
		Environment string `json:"environment"`
		Version     string `json:"version"`
	}{
		Status:      "available",
		Environment: app.config.env,
		Version:     version,
	}
	writeJSON(w, heathCheck, http.StatusOK)
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	resp, err := app.usersService.Create(ctx, req)
	if err != nil {
		// TODO: get status code from error
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp, http.StatusCreated)
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req ports.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	user := getUserFromRequest(r)
	if user == nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := app.usersService.Update(ctx, user, req)
	if err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromRequest(r)
	if user == nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"user": user}, http.StatusOK)
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromRequest(r)
	if user == nil {
		writeError(w, errors.New("user doesn't exist"), http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := ports.DeleteUserRequest{ID: user.ID}
	resp, err := app.usersService.Delete(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp, http.StatusOK)
}

func (app *application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromRequest(r)
	if user == nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	var req ports.CreateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	resp, err := app.tasksService.Create(ctx, *user, req)
	if err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp, http.StatusCreated)
}

func (app *application) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ports.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	req.ID = id

	user := getUserFromRequest(r)
	if user == nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	resp, err := app.tasksService.Update(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}

func (app *application) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user := getUserFromRequest(r)
	if user == nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	req := ports.GetTaskRequest{ID: id}
	resp, err := app.tasksService.Get(ctx, req)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}

// func (app *application) getTasksHandler(w http.ResponseWriter, r *http.Request) {
// 	user := getUserFromRequest(r)
// 	if user == nil {
// 		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
// 		return
// 	}

// 	query := r.URL.Query()
// 	sort := query.Get("sort")
// 	if sort == "" {
// 		sort = "id"
// 	}

// 	page := 1
// 	pageSize := 20

// 	pageStr := query.Get("page")
// 	if pageStr != "" {
// 		p, err := strconv.Atoi(pageStr)
// 		if err != nil || p <= 0 {
// 			writeError(w, errors.New(`invalid query parameter "page": must be a positive integer`), http.StatusBadRequest)
// 			return
// 		}
// 		page = p
// 	}
// 	pageSizeStr := query.Get("page_size")
// 	if pageSizeStr != "" {
// 		size, err := strconv.Atoi(pageSizeStr)
// 		if err != nil || size <= 0 {
// 			writeError(w, errors.New(`invalid query param "page_size": must be a positive integer`), http.StatusBadRequest)
// 			return
// 		}
// 		pageSize = size
// 	}

// 	v := failures.NewValidator()
// 	sortList := []string{"id", "-id", "created_at", "-created_at", "is_completed", "-is_completed"}
// 	v.Check(slices.Index(sortList, sort) != -1, "sort", fmt.Sprintf("must be one of the values %v", sortList))
// 	v.Check(page >= 1 && page <= 10_000_000, "page", "must be between 1 and 10_000_000")
// 	v.Check(pageSize >= 1 && page <= 100, "page_size", "must be between 1 and 100")

// 	content := query.Get("content")

// 	tasks, total, err := app.storage.getTasksForUser(user, sort, page, pageSize, content)
// 	if err != nil {
// 		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
// 		return
// 	}
// 	if tasks == nil {
// 		writeError(w, errors.New("resource doesn't exist"), http.StatusNotFound)
// 		return
// 	}
// 	writeJSON(w, map[string]any{"tasks": tasks, "total": total}, http.StatusOK)
// }

func (app *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user := getUserFromRequest(r)
	if user == nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	req := ports.DeleteTaskRequest{ID: id}
	resp, err := app.tasksService.Delete(ctx, req)
	if err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}

func (app *application) sendActivationCodeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, errors.New("validation error: id is empty"), http.StatusBadRequest)
		return
	}

	// TODO: hardcoding
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	resp, err := app.usersService.Get(ctx, ports.GetUserRequest{ID: id})
	if err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	u := resp.User
	if u.IsActivated {
		writeJSON(w, map[string]any{"message": "user already activated"}, http.StatusConflict)
		return
	}

	if app.storage.useractivationCache.HasExpired(u) {
		tmpl, err := template.ParseFS(templates, "templates/*.gotmpl")
		if err != nil {
			writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
			return
		}
		code := uint16(rand.Uint())
		err = app.mailer.send(u.Email, tmpl, map[string]any{"code": code})
		if err != nil {
			log.Println(err)
			writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
			return
		}
		app.storage.useractivationCache.Set(u, code, time.Minute)
	}
	writeJSON(w, map[string]any{"message": fmt.Sprintf("we have sent an activation code to your email: %s", u.Email)}, http.StatusOK)
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, errors.New("route paramter {id}: must not be empty"), http.StatusBadRequest)
		return
	}

	var input struct {
		Code *int `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if input.Code == nil {
		writeError(w, errors.New("code must be provided in request body"), http.StatusBadRequest)
		return
	}

	// TODO: hardcoding
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	getUserResp, err := app.usersService.Get(ctx, ports.GetUserRequest{ID: id})
	if err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}
	u := getUserResp.User
	if u.IsActivated {
		writeJSON(w, map[string]any{"message": "user already activated"}, http.StatusConflict)
		return
	}
	activationCode, expired := app.storage.useractivationCache.Get(u)
	if expired {
		writeJSON(w, map[string]any{"message": "code has expired"}, http.StatusConflict)
		return
	}
	if activationCode != *input.Code {
		writeJSON(w, map[string]any{"message": "invalid activation code"}, http.StatusConflict)
		return
	}
	u.IsActivated = true
	if err := app.usersRepo.Update(ctx, u); err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"message": "user was updated successfully"}, http.StatusOK)
}

func (app *application) authenticateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	v := failures.NewValidator()
	v.CheckUTF8Email(input.Email)
	v.CheckUTF8Password(input.Password)

	if err := v.Err(); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	// TODO: hardcoding
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	u, err := app.usersRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(input.Password)); err != nil {
		writeError(w, errors.New("email or password are not correct"), http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"user_id":    u.ID,
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC822),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(app.config.jwt.secret))
	if err != nil {
		log.Println(err)
		writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"token": tokenStr}, http.StatusCreated)
}

func composeJSONError(err error) string {
	jsonError := map[string]string{
		"error": err.Error(),
	}
	result, err := json.Marshal(jsonError)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(result)
}

func writeError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, composeJSONError(err))
}

func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(data)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(j)
}
