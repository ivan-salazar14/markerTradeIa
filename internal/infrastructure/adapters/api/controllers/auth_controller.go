package controllers

import (
	"encoding/json"
	"errors"
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
	_ = json.NewEncoder(w).Encode(tokens)
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
	_ = json.NewEncoder(w).Encode(newTokens)
}

func (c *AuthController) WalletChallenge(w http.ResponseWriter, r *http.Request) {
	var req domain.WalletChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	challenge, err := c.authService.CreateWalletChallenge(req.Address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(challenge)
}

func (c *AuthController) WalletVerify(w http.ResponseWriter, r *http.Request) {
	var req domain.WalletVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := c.authService.VerifyWalletChallenge(req.Address, req.Nonce, req.Signature)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, auth.ErrInvalidWalletAddress) || errors.Is(err, auth.ErrInvalidWalletChallenge) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokens)
}
