package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"hypertube/api/internal/respond"
)

type contextKey string

const userIDContextKey contextKey = "auth_user_id"

func RequireAuth(tokens *TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				respond.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token")
				return
			}

			claims, err := tokens.ValidateAccessToken(tokenString)
			if err != nil {
				if errors.Is(err, ErrExpiredToken) {
					respond.Error(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "token expired")
					return
				}
				respond.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid bearer token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], true
}
