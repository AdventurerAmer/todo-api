package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/AdventurerAmer/todo-api/failures"
)

func (app *App) Panic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				err := fmt.Errorf("recovered from panic: %+v", err)
				WriteError(w, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *App) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := w.Header().Get("Origin")
		if origin != "" {
			for _, o := range app.TrustedOrigins {
				if origin == o || o == "*" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// preflight request
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

type TokenAuthFunc = func(r *http.Request, token string) (context.Context, error)

func (app *App) TokenAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := &failures.AuthenticationError{Reason: "invalid 'Authorization' header"}
			WriteError(w, err)
			return
		}
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			err := &failures.AuthenticationError{Reason: "invalid 'Authorization' header"}
			WriteError(w, err)
			return
		}
		token := parts[1]

		dctx, err := app.TokenAuthHandler(r, token)
		if err != nil {
			err := fmt.Errorf("'TokenAuthHandler' failed: %w", err)
			WriteError(w, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(dctx))
	}
}

func (app *App) InjectSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; object-src 'none'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "1; mode=block") // Legacy, but harmless
		next.ServeHTTP(w, r)
	})
}
