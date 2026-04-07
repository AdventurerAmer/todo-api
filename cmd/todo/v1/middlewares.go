package v1

import (
	"context"
	"net/http"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/web"
)

type userContext string

const userContextKey userContext = "userContextKey"

func mustGetUserFromCtx(ctx context.Context) domain.User {
	user, ok := ctx.Value(userContextKey).(domain.User)
	if !ok {
		panic("user doesn't exist")
	}
	return user
}

func getUserFromCtx(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(userContextKey).(domain.User)
	return user, ok
}

func requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := getUserFromCtx(r.Context())
		if !ok || !user.IsActivated {
			err := &failures.AuthorizationError{Reason: "your user account must be activated to access this resource"}
			web.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r)
	}
}
