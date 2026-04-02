package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AdventurerAmer/todo-api/failures"
	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
	"github.com/AdventurerAmer/todo-api/web"
	"github.com/golang-jwt/jwt"
)

func (app *application) Auth(next http.HandlerFunc) http.HandlerFunc {
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
		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(app.config.Authentication.JWTSecret), nil
		})
		if err != nil {
			err := &failures.AuthenticationError{Reason: "invalid token"}
			web.WriteError(w, err)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			err := &failures.AuthenticationError{Reason: "invalid token"}
			web.WriteError(w, err)
			return
		}
		userID := claims["user_id"].(string)
		expiresAtStr := claims["expires_at"].(string)
		expiresAt, err := time.Parse(time.RFC822, expiresAtStr)
		if err != nil {
			err := &failures.AuthenticationError{Reason: "invalid token"}
			web.WriteError(w, err)
			return
		}

		if time.Now().After(expiresAt) {
			err := &failures.AuthenticationError{Reason: "invalid token"}
			web.WriteError(w, err)
			return
		}

		// TOOD: hardcoding duration
		dctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		resp, err := app.usersService.Get(dctx, ports.GetUserRequest{ID: userID})
		if err != nil {
			err := fmt.Errorf("'usersService.Get' failed: %w", err)
			web.WriteError(w, err)
			return
		}
		user := resp.User
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mustGetUserFromRequestContext(r)
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

func mustGetUserFromRequestContext(r *http.Request) *domain.User {
	user, ok := r.Context().Value(userContextKey).(*domain.User)
	if !ok {
		panic("user doesn't exist")
	}
	return user
}
