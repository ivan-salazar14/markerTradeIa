package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/auth"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type AuthController struct {
	authService *auth.AuthService
}

func NewAuthController(s *auth.AuthService) *AuthController {
	return &AuthController{authService: s}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Mock Authentication for demonstration (replace with actual user validation)
	if req.UID == "" || req.Secret == "" {
		http.Error(w, "missing credentials", http.StatusUnauthorized)
		return
	}

	identity := domain.Identity{
		ID:   req.UID,
		Type: "user",
	}

	tokens, err := c.authService.GenerateTokenPair(identity)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	var tokens domain.TokenPair
	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	identity, err := c.authService.ValidateToken(tokens.RefreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	newTokens, err := c.authService.GenerateTokenPair(*identity)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newTokens)
}
