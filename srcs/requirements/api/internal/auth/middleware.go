package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/respond"
)

type contextKey string

const userIDContextKey contextKey = "auth_user_id"

func RequireAuth(tokens *TokenManager, users UserExistenceChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingBearerToken)
				return
			}

			claims, err := tokens.ValidateAccessToken(tokenString)
			if err != nil {
				if errors.Is(err, ErrExpiredToken) {
					respond.LocalizedError(w, r, http.StatusUnauthorized, "TOKEN_EXPIRED", i18n.MsgTokenExpired)
					return
				}
				respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgInvalidBearerToken)
				return
			}

			if users == nil {
				respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
				return
			}

			exists, err := users.UserExists(r.Context(), claims.UserID)
			if err != nil {
				respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
				return
			}
			if !exists {
				respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgInvalidBearerToken)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func DevAuthenticateAs(userID int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
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
