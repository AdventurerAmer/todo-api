package web

import (
	"fmt"
	"net/http"
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
