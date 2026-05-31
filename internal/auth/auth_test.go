package auth_test

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/auth"
)

func TestGenerateAndValidateToken(t *testing.T) {
	a := auth.New("test-secret-key")

	token, err := a.GenerateToken("user-1", "test@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := a.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("expected user-1, got %s", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Fatalf("expected test@example.com, got %s", claims.Email)
	}
}

func TestInvalidToken(t *testing.T) {
	a := auth.New("test-secret-key")
	_, err := a.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	plain, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}
	if len(plain) < 10 {
		t.Fatal("api key too short")
	}
	if hash != auth.HashAPIKey(plain) {
		t.Fatal("hash mismatch")
	}
}
