package auth

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

func TestWalletChallengeVerification(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	service := NewAuthService(domain.AuthConfig{
		JWTSecret:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: time.Hour,
	})

	challenge, err := service.CreateWalletChallenge(address)
	if err != nil {
		t.Fatalf("failed to create challenge: %v", err)
	}

	hash := accounts.TextHash([]byte(challenge.Message))
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		t.Fatalf("failed to sign challenge: %v", err)
	}

	tokens, err := service.VerifyWalletChallenge(address, challenge.Nonce, "0x"+hex.EncodeToString(signature))
	if err != nil {
		t.Fatalf("failed to verify challenge: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected JWT tokens")
	}

	identity, err := service.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if identity.Type != "wallet" {
		t.Fatalf("expected wallet identity, got %s", identity.Type)
	}
	if identity.ID != address {
		t.Fatalf("expected %s, got %s", address, identity.ID)
	}
}

func TestWalletChallengeRejectsWrongSignature(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	otherKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate other key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	service := NewAuthService(domain.AuthConfig{
		JWTSecret:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: time.Hour,
	})

	challenge, err := service.CreateWalletChallenge(address)
	if err != nil {
		t.Fatalf("failed to create challenge: %v", err)
	}

	hash := accounts.TextHash([]byte(challenge.Message))
	signature, err := crypto.Sign(hash, otherKey)
	if err != nil {
		t.Fatalf("failed to sign challenge: %v", err)
	}

	if _, err := service.VerifyWalletChallenge(address, challenge.Nonce, "0x"+hex.EncodeToString(signature)); err == nil {
		t.Fatalf("expected verification to fail")
	}
}
