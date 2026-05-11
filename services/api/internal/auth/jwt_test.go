package auth

import (
	"errors"
	"testing"
	"time"
)

const testJWTSecret = "0123456789abcdef0123456789abcdef"

func newTestTokenManager(t *testing.T) *TokenManager {
	t.Helper()

	tokens, err := NewTokenManager(testJWTSecret, "hypertube-test")
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	return tokens
}

func TestJWTValidToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }

	token, expiresAt, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if expiresAt.Sub(now) != AccessTokenTTL {
		t.Fatalf("expected ttl %s, got %s", AccessTokenTTL, expiresAt.Sub(now))
	}

	claims, err := tokens.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", claims.UserID)
	}
}

func TestNewTokenManagerRejectsShortSecret(t *testing.T) {
	_, err := NewTokenManager("too-short", "hypertube-test")
	if !errors.Is(err, ErrJWTSecretTooShort) {
		t.Fatalf("expected ErrJWTSecretTooShort, got %v", err)
	}
}

func TestJWTExpiredToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }

	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	tokens.now = func() time.Time { return now.Add(AccessTokenTTL + time.Second) }
	_, err = tokens.ValidateAccessToken(token)
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
}

func TestJWTInvalidToken(t *testing.T) {
	tokens := newTestTokenManager(t)

	_, err := tokens.ValidateAccessToken("not-a-token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTRejectsWrongIssuer(t *testing.T) {
	issuerA, err := NewTokenManager(testJWTSecret, "issuer-a")
	if err != nil {
		t.Fatalf("new issuer a token manager: %v", err)
	}
	issuerB, err := NewTokenManager(testJWTSecret, "issuer-b")
	if err != nil {
		t.Fatalf("new issuer b token manager: %v", err)
	}

	token, _, err := issuerA.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	_, err = issuerB.ValidateAccessToken(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTRejectsNonPositiveUserID(t *testing.T) {
	tokens := newTestTokenManager(t)
	token, _, err := tokens.CreateAccessToken(0)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	_, err = tokens.ValidateAccessToken(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
