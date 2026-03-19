package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

var (
	ErrInvalidToken           = errors.New("invalid token")
	ErrExpiredToken           = errors.New("token expired")
	ErrInvalidWalletAddress   = errors.New("invalid wallet address")
	ErrInvalidWalletChallenge = errors.New("invalid wallet challenge")
	ErrWalletChallengeExpired = errors.New("wallet challenge expired")
	ErrInvalidWalletSignature = errors.New("invalid wallet signature")
)

type walletChallenge struct {
	Address   string
	Nonce     string
	Message   string
	ExpiresAt time.Time
}

type AuthService struct {
	config     domain.AuthConfig
	mu         sync.Mutex
	challenges map[string]walletChallenge
}

func NewAuthService(cfg domain.AuthConfig) *AuthService {
	return &AuthService{
		config:     cfg,
		challenges: make(map[string]walletChallenge),
	}
}

type Claims struct {
	Identity domain.Identity `json:"identity"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateTokenPair(identity domain.Identity) (domain.TokenPair, error) {
	now := time.Now()

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

func (s *AuthService) CreateWalletChallenge(address string) (domain.WalletChallengeResponse, error) {
	normalized, err := normalizeWalletAddress(address)
	if err != nil {
		return domain.WalletChallengeResponse{}, err
	}

	nonce, err := randomNonce(16)
	if err != nil {
		return domain.WalletChallengeResponse{}, err
	}

	expiresAt := time.Now().Add(5 * time.Minute)
	message := fmt.Sprintf(
		"MarkerTradeIa wallet verification\nAddress: %s\nNonce: %s\nExpiresAt: %s",
		normalized,
		nonce,
		expiresAt.UTC().Format(time.RFC3339),
	)

	s.mu.Lock()
	s.challenges[strings.ToLower(normalized)] = walletChallenge{
		Address:   normalized,
		Nonce:     nonce,
		Message:   message,
		ExpiresAt: expiresAt,
	}
	s.mu.Unlock()

	return domain.WalletChallengeResponse{
		Address:   normalized,
		Nonce:     nonce,
		Message:   message,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

func (s *AuthService) VerifyWalletChallenge(address string, nonce string, signature string) (domain.TokenPair, error) {
	normalized, err := normalizeWalletAddress(address)
	if err != nil {
		return domain.TokenPair{}, err
	}

	s.mu.Lock()
	challenge, ok := s.challenges[strings.ToLower(normalized)]
	if ok && time.Now().After(challenge.ExpiresAt) {
		delete(s.challenges, strings.ToLower(normalized))
		s.mu.Unlock()
		return domain.TokenPair{}, ErrWalletChallengeExpired
	}
	if !ok {
		s.mu.Unlock()
		return domain.TokenPair{}, ErrInvalidWalletChallenge
	}
	if challenge.Nonce != strings.TrimSpace(nonce) {
		s.mu.Unlock()
		return domain.TokenPair{}, ErrInvalidWalletChallenge
	}
	delete(s.challenges, strings.ToLower(normalized))
	s.mu.Unlock()

	recovered, err := recoverAddress(challenge.Message, signature)
	if err != nil {
		return domain.TokenPair{}, err
	}
	if !strings.EqualFold(recovered.Hex(), normalized) {
		return domain.TokenPair{}, ErrInvalidWalletSignature
	}

	identity := domain.Identity{
		ID:    normalized,
		Type:  "wallet",
		Scope: "wallet:manage",
	}
	return s.GenerateTokenPair(identity)
}

func normalizeWalletAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	if !common.IsHexAddress(address) {
		return "", ErrInvalidWalletAddress
	}
	return common.HexToAddress(address).Hex(), nil
}

func randomNonce(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func recoverAddress(message string, signature string) (common.Address, error) {
	signature = strings.TrimSpace(signature)
	signature = strings.TrimPrefix(signature, "0x")
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return common.Address{}, ErrInvalidWalletSignature
	}
	if len(sigBytes) != crypto.SignatureLength {
		return common.Address{}, ErrInvalidWalletSignature
	}
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}
	if sigBytes[64] > 1 {
		return common.Address{}, ErrInvalidWalletSignature
	}

	hash := accounts.TextHash([]byte(message))
	pubKey, err := crypto.SigToPub(hash, sigBytes)
	if err != nil {
		return common.Address{}, ErrInvalidWalletSignature
	}
	return crypto.PubkeyToAddress(*pubKey), nil
}
