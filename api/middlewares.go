package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/web"
)

func (app *application) TokenAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := &failures.AuthenticationError{Reason: "invalid 'Authorization' header"}
			web.WriteError(w, err)
			return
		}
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			err := &failures.AuthenticationError{Reason: "invalid 'Authorization' header"}
			web.WriteError(w, err)
			return
		}
		token := parts[1]

		ctx, cancel := context.WithTimeout(r.Context(), app.config.Server.DefaultTimeout)
		defer cancel()

		resp, err := app.tokenAuthService.Check(ctx, ports.CheckAuthTokenRequest{Token: token})
		if err != nil {
			err := fmt.Errorf("'tokenAuthService.Check' failed: %w", err)
			web.WriteError(w, err)
			return
		}

		user := resp.User
		dctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(dctx))
	}
}

func requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustGetUserFromContext(r.Context())
		if !user.IsActivated {
			err := &failures.AuthorizationError{Reason: "your user account must be activated to access this resource"}
			web.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r)
	}
}

type userContext string

const userContextKey userContext = "userContextKey"

func mustGetUserFromContext(ctx context.Context) *domain.User {
	user, ok := ctx.Value(userContextKey).(*domain.User)
	if !ok {
		panic("user doesn't exist")
	}
	return user
}

func getUserFromContext(ctx context.Context) *domain.User {
	user, _ := ctx.Value(userContextKey).(*domain.User)
	return user
}
