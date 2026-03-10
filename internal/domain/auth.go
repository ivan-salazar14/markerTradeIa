package domain

import "time"

// Identity representa a un actor autenticado (usuario o servicio).
type Identity struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // "user" o "service"
	Scope string `json:"scope"`
}

// TokenPair contiene el par de tokens JWT.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// LoginRequest representa la solicitud de autenticación.
type LoginRequest struct {
	UID    string `json:"uid"`
	Secret string `json:"secret"`
}

// AuthConfig contiene la configuración de seguridad.
type AuthConfig struct {
	JWTSecret      string
	AccessExpiry   time.Duration
	RefreshExpiry  time.Duration
	ServiceAPIKeys map[string]string // M2M API Keys
}
