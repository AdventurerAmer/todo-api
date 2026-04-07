package v1

import (
	"net/http"
)

func composeRoutes(app *application) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/health", app.healthCheckHandler)

	mux.HandleFunc("POST /v1/tokens/activation", app.TokenAuth(app.sendActivationCodeHandler))
	mux.HandleFunc("PUT /v1/tokens/activation", app.activateUserHandler)

	mux.HandleFunc("POST /v1/auth", app.authenticateUserHandler)

	mux.HandleFunc("POST /v1/users", app.createUserHandler)
	mux.HandleFunc("GET /v1/users", app.TokenAuth(requireActivatedUser(app.getUserHandler)))
	mux.HandleFunc("PUT /v1/users", app.TokenAuth(requireActivatedUser(app.updateUserHandler)))
	mux.HandleFunc("DELETE /v1/users", app.TokenAuth(requireActivatedUser(app.deleteUserHandler)))

	mux.HandleFunc("POST /v1/lists", app.TokenAuth(requireActivatedUser(app.createListHandler)))
	mux.HandleFunc("GET /v1/lists", app.TokenAuth(requireActivatedUser(app.getListsHandler)))
	mux.HandleFunc("GET /v1/lists/{id}", app.TokenAuth(requireActivatedUser(app.getListHandler)))
	mux.HandleFunc("PUT /v1/lists/{id}", app.TokenAuth(requireActivatedUser(app.updateListHandler)))
	mux.HandleFunc("DELETE /v1/lists/{id}", app.TokenAuth(requireActivatedUser(app.deleteListandler)))

	mux.HandleFunc("POST /v1/tasks", app.TokenAuth(requireActivatedUser(app.createTaskHandler)))
	mux.HandleFunc("GET /v1/lists/{id}/tasks", app.TokenAuth(requireActivatedUser(app.getTasksHandler)))
	mux.HandleFunc("GET /v1/tasks/{id}", app.TokenAuth(requireActivatedUser(app.getTaskHandler)))
	mux.HandleFunc("PUT /v1/tasks/{id}", app.TokenAuth(requireActivatedUser(app.updateTaskHandler)))
	mux.HandleFunc("DELETE /v1/tasks/{id}", app.TokenAuth(requireActivatedUser(app.deleteTaskHandler)))

	return app.Panic(app.CORS(mux))
}
