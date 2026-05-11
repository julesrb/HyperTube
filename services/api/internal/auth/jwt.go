package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL = 15 * time.Minute

var (
	ErrJWTSecretTooShort = errors.New("jwt secret must be at least 32 bytes")
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("expired token")
)

type TokenManager struct {
	secret []byte
	issuer string
	now    func() time.Time
}

type AccessClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func NewTokenManager(secret, issuer string) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, ErrJWTSecretTooShort
	}
	return &TokenManager{
		secret: []byte(secret),
		issuer: issuer,
		now:    time.Now,
	}, nil
}

func (m *TokenManager) CreateAccessToken(userID int64) (string, time.Time, error) {
	now := m.now().UTC()
	expiresAt := now.Add(AccessTokenTTL)

	claims := AccessClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signedToken, expiresAt, nil
}

func (m *TokenManager) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrInvalidToken
			}
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if !token.Valid || claims.UserID <= 0 {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
