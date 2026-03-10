package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type AuthService struct {
	config domain.AuthConfig
}

func NewAuthService(cfg domain.AuthConfig) *AuthService {
	return &AuthService{config: cfg}
}

type Claims struct {
	Identity domain.Identity `json:"identity"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateTokenPair(identity domain.Identity) (domain.TokenPair, error) {
	now := time.Now()

	// Access Token
	accessClaims := &Claims{
		Identity: identity,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	at, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return domain.TokenPair{}, err
	}

	// Refresh Token
	refreshClaims := &Claims{
		Identity: identity,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rt, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{
		AccessToken:  at,
		RefreshToken: rt,
		ExpiresAt:    now.Add(s.config.AccessExpiry).Unix(),
	}, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (*domain.Identity, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &claims.Identity, nil
	}

	return nil, ErrInvalidToken
}

func (s *AuthService) ValidateAPIKey(key string) (*domain.Identity, error) {
	for serviceName, apiKey := range s.config.ServiceAPIKeys {
		if apiKey == key {
			return &domain.Identity{
				ID:   serviceName,
				Type: "service",
			}, nil
		}
	}
	return nil, errors.New("invalid api key")
}
