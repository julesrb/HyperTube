package auth

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTValidAccessToken(t *testing.T) {
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
	if claims.TokenUse != accessTokenUse {
		t.Fatalf("expected token_use %q, got %q", accessTokenUse, claims.TokenUse)
	}
}

func TestJWTValidRefreshToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }

	token, expiresAt, err := tokens.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}
	if expiresAt.Sub(now) != RefreshTokenTTL {
		t.Fatalf("expected ttl %s, got %s", RefreshTokenTTL, expiresAt.Sub(now))
	}

	claims, err := tokens.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("validate refresh token: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", claims.UserID)
	}
	if claims.TokenUse != refreshTokenUse {
		t.Fatalf("expected token_use %q, got %q", refreshTokenUse, claims.TokenUse)
	}
}

func TestNewTokenManagerRejectsShortSecret(t *testing.T) {
	_, err := NewTokenManager("too-short", "hypertube-test")
	if !errors.Is(err, ErrJWTSecretTooShort) {
		t.Fatalf("expected ErrJWTSecretTooShort, got %v", err)
	}
}

func TestJWTExpiredTokens(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		create   func(*TokenManager) (string, error)
		validate func(*TokenManager, string) error
	}{
		{
			name: "access",
			ttl:  AccessTokenTTL,
			create: func(tokens *TokenManager) (string, error) {
				token, _, err := tokens.CreateAccessToken(42)
				return token, err
			},
			validate: func(tokens *TokenManager, token string) error {
				_, err := tokens.ValidateAccessToken(token)
				return err
			},
		},
		{
			name: "refresh",
			ttl:  RefreshTokenTTL,
			create: func(tokens *TokenManager) (string, error) {
				token, _, err := tokens.CreateRefreshToken(42)
				return token, err
			},
			validate: func(tokens *TokenManager, token string) error {
				_, err := tokens.ValidateRefreshToken(token)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := newTestTokenManager(t)
			now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
			tokens.now = func() time.Time { return now }

			token, err := tt.create(tokens)
			if err != nil {
				t.Fatalf("create token: %v", err)
			}
			tokens.now = func() time.Time { return now.Add(tt.ttl + time.Second) }
			if err := tt.validate(tokens, token); !errors.Is(err, ErrExpiredToken) {
				t.Fatalf("expected ErrExpiredToken, got %v", err)
			}
		})
	}
}

func TestJWTInvalidTokenText(t *testing.T) {
	tokens := newTestTokenManager(t)

	if _, err := tokens.ValidateAccessToken("not-a-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected access ErrInvalidToken, got %v", err)
	}
	if _, err := tokens.ValidateRefreshToken("not-a-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected refresh ErrInvalidToken, got %v", err)
	}
}

func TestJWTTokensCannotBeUsedInterchangeably(t *testing.T) {
	tokens := newTestTokenManager(t)
	accessToken, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create access token: %v", err)
	}
	refreshToken, _, err := tokens.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	if _, err := tokens.ValidateRefreshToken(accessToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected refresh validation to reject access token, got %v", err)
	}
	if _, err := tokens.ValidateAccessToken(refreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected access validation to reject refresh token, got %v", err)
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

	accessToken, _, err := issuerA.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create access token: %v", err)
	}
	refreshToken, _, err := issuerA.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	if _, err := issuerB.ValidateAccessToken(accessToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected access ErrInvalidToken, got %v", err)
	}
	if _, err := issuerB.ValidateRefreshToken(refreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected refresh ErrInvalidToken, got %v", err)
	}
}

func TestJWTRefreshRejectsWrongSigningAlgorithm(t *testing.T) {
	tokens := newTestTokenManager(t)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }
	claims := RefreshClaims{
		UserID:           42,
		TokenUse:         refreshTokenUse,
		RegisteredClaims: testRegisteredClaims(now, tokens.issuer, 42, RefreshTokenTTL),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString(tokens.secret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := tokens.ValidateRefreshToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTRejectsNonPositiveUserID(t *testing.T) {
	tokens := newTestTokenManager(t)
	accessToken, _, err := tokens.CreateAccessToken(0)
	if err != nil {
		t.Fatalf("create access token: %v", err)
	}
	refreshToken, _, err := tokens.CreateRefreshToken(0)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	if _, err := tokens.ValidateAccessToken(accessToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected access ErrInvalidToken, got %v", err)
	}
	if _, err := tokens.ValidateRefreshToken(refreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected refresh ErrInvalidToken, got %v", err)
	}
}

func TestJWTRejectsMissingTokenUse(t *testing.T) {
	tokens := newTestTokenManager(t)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }
	claims := struct {
		UserID int64 `json:"user_id"`
		jwt.RegisteredClaims
	}{
		UserID:           42,
		RegisteredClaims: testRegisteredClaims(now, tokens.issuer, 42, RefreshTokenTTL),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tokens.secret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := tokens.ValidateAccessToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected access ErrInvalidToken, got %v", err)
	}
	if _, err := tokens.ValidateRefreshToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected refresh ErrInvalidToken, got %v", err)
	}
}

func testRegisteredClaims(now time.Time, issuer string, userID int64, ttl time.Duration) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
}
