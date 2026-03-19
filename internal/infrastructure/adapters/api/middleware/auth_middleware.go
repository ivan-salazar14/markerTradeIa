package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/auth"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type contextKey string

const (
	IdentityContextKey contextKey = "identity"
)

// isLocalhost checks if the request is coming from localhost
func isLocalhost(r *http.Request) bool {
	// Check the Host header first
	host := r.Host
	if strings.HasPrefix(host, "localhost:") ||
		host == "localhost" ||
		host == "127.0.0.1" ||
		strings.HasPrefix(host, "127.0.0.1:") ||
		host == "0.0.0.0" ||
		strings.HasPrefix(host, "0.0.0.0:") ||
		host == "[::1]" ||
		strings.HasPrefix(host, "[::1]:") {
		return true
	}

	// Also check the remote address (client IP)
	remoteAddr := r.RemoteAddr
	if remoteAddr == "" {
		return false
	}

	// Remove port from address
	ip := remoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		ip = remoteAddr[:idx]
	}

	// Check for IPv4 localhost
	if ip == "127.0.0.1" {
		return true
	}

	// Check for IPv6 localhost
	if ip == "[::1]" || ip == "::1" {
		return true
	}

	return false
}

func AuthMiddleware(authService *auth.AuthService, disableAuthLocalhost bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Skip authentication for localhost if disabled
			if disableAuthLocalhost && isLocalhost(r) {
				// Create a default identity for localhost requests
				identity := &domain.Identity{
					ID:   "localhost",
					Type: "service",
				}
				ctx := context.WithValue(r.Context(), IdentityContextKey, identity)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

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
