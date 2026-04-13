package v1

import (
	"net/http"

	"github.com/AdventurerAmer/todo-api/internal/config"
)

func composeRoutes(app *application) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/health", app.health)

	mux.HandleFunc("POST /v1/auth", app.authenticateUser)

	mux.HandleFunc("POST /v1/users", app.createUser)
	mux.HandleFunc("GET /v1/users", app.TokenAuth(requireActivatedUser(app.getUser)))
	mux.HandleFunc("PUT /v1/users", app.TokenAuth(requireActivatedUser(app.updateUser)))
	mux.HandleFunc("DELETE /v1/users", app.TokenAuth(requireActivatedUser(app.deleteUser)))

	mux.HandleFunc("POST /v1/tokens/activation", app.TokenAuth(app.sendActivationCode))
	mux.HandleFunc("PUT /v1/tokens/activation", app.activateUser)

	mux.HandleFunc("POST /v1/lists", app.TokenAuth(requireActivatedUser(app.createList)))
	mux.HandleFunc("GET /v1/lists", app.TokenAuth(requireActivatedUser(app.getLists)))
	mux.HandleFunc("GET /v1/lists/{id}", app.TokenAuth(requireActivatedUser(app.getList)))
	mux.HandleFunc("PUT /v1/lists/{id}", app.TokenAuth(requireActivatedUser(app.updateList)))
	mux.HandleFunc("DELETE /v1/lists/{id}", app.TokenAuth(requireActivatedUser(app.deleteList)))

	mux.HandleFunc("POST /v1/tasks", app.TokenAuth(requireActivatedUser(app.createTask)))
	mux.HandleFunc("GET /v1/lists/{id}/tasks", app.TokenAuth(requireActivatedUser(app.getTasks)))
	mux.HandleFunc("GET /v1/tasks/{id}", app.TokenAuth(requireActivatedUser(app.getTask)))
	mux.HandleFunc("PUT /v1/tasks/{id}", app.TokenAuth(requireActivatedUser(app.updateTask)))
	mux.HandleFunc("DELETE /v1/tasks/{id}", app.TokenAuth(requireActivatedUser(app.deleteTask)))

	if app.config.Env == config.EnvProd {
		return app.InjectSecurityHeaders(app.Panic(app.CORS(mux)))
	}
	return app.Panic(app.CORS(mux))
}
