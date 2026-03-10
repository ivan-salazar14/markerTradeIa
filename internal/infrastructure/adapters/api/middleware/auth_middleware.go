package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/auth"
)

type contextKey string

const (
	IdentityContextKey contextKey = "identity"
)

func AuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Try API Key (X-API-Key)
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" {
				identity, err := authService.ValidateAPIKey(apiKey)
				if err == nil {
					ctx := context.WithValue(r.Context(), IdentityContextKey, identity)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// 2. Try JWT Bearer Token
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				identity, err := authService.ValidateToken(tokenStr)
				if err == nil {
					ctx := context.WithValue(r.Context(), IdentityContextKey, identity)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Unauthorized
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
		})
	}
}
