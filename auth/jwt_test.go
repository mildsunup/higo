package auth

import (
	"context"
	"testing"
	"time"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	cfg := DefaultConfig().WithSecret([]byte("test-secret"))

	jwt, err := NewJWT(cfg)
	if err != nil {
		t.Fatalf("NewJWT failed: %v", err)
	}

	claims := Claims{
		UserID: 123,
		Extra:  map[string]any{"role": "admin"},
	}

	ctx := context.Background()
	token, err := jwt.GenerateToken(ctx, claims)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	parsed, err := jwt.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if parsed.UserID != 123 {
		t.Errorf("Expected UserID 123, got %d", parsed.UserID)
	}

	if parsed.Extra["role"] != "admin" {
		t.Errorf("Expected role 'admin', got '%v'", parsed.Extra["role"])
	}
}

func TestJWT_ValidateInvalidToken(t *testing.T) {
	cfg := DefaultConfig().WithSecret([]byte("test-secret"))
	jwt, _ := NewJWT(cfg)

	_, err := jwt.ValidateToken(context.Background(), "invalid")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	cfg := DefaultConfig().WithSecret([]byte("test-secret"))
	cfg.Expiry = 1 * time.Millisecond

	jwt, _ := NewJWT(cfg)

	claims := Claims{UserID: 123}
	token, _ := jwt.GenerateToken(context.Background(), claims)

	time.Sleep(10 * time.Millisecond)

	_, err := jwt.ValidateToken(context.Background(), token)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}
