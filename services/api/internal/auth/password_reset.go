package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hypertube/api/internal/respond"
)

const passwordResetTokenBytes = 32

type passwordResetMailer interface {
	SendPasswordReset(ctx context.Context, toEmail string, toName string, resetURL string, expiresIn time.Duration) error
}

type passwordResetRequest struct {
	Email  string `json:"email"`
	Locale string `json:"locale,omitempty"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type passwordResetResponse struct {
	Message string `json:"message"`
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req passwordResetRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	email, ok := normalizeEmail(req.Email)
	if !ok {
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "valid email is required")
		return
	}

	if h.store == nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "authentication service is unavailable")
		return
	}
	if h.passwordResetMailer == nil {
		respond.Error(w, http.StatusServiceUnavailable, "EMAIL_NOT_CONFIGURED", "password reset email is not configured")
		return
	}

	user, err := h.store.FindUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			h.writePasswordResetAccepted(w)
			return
		}
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load user")
		return
	}

	token, tokenHash, err := newPasswordResetToken()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create password reset token")
		return
	}

	expiresAt := time.Now().UTC().Add(h.passwordResetTTL)
	if err := h.store.CreatePasswordResetToken(r.Context(), CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store password reset token")
		return
	}

	resetURL, err := h.buildPasswordResetURL(token, req.Locale)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "password reset URL is not configured")
		return
	}

	toName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if err := h.passwordResetMailer.SendPasswordReset(r.Context(), user.Email, toName, resetURL, h.passwordResetTTL); err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to send password reset email")
		return
	}

	h.writePasswordResetAccepted(w)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if h.store == nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "authentication service is unavailable")
		return
	}

	tokenHash, ok := passwordResetTokenHash(req.Token)
	if !ok {
		respond.Error(w, http.StatusBadRequest, "INVALID_RESET_TOKEN", "password reset link is invalid or expired")
		return
	}

	if validationMessage, ok := validatePassword(req.Password); !ok {
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", validationMessage)
		return
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "password is invalid")
		return
	}

	if _, err := h.store.ResetPasswordWithToken(r.Context(), tokenHash, passwordHash); err != nil {
		if errors.Is(err, ErrInvalidPasswordResetToken) {
			respond.Error(w, http.StatusBadRequest, "INVALID_RESET_TOKEN", "password reset link is invalid or expired")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reset password")
		return
	}

	respond.Data(w, http.StatusOK, passwordResetResponse{Message: "password has been reset"})
}

func (h *Handler) buildPasswordResetURL(token string, locale string) (string, error) {
	rawURL := strings.TrimSpace(h.passwordResetURL)
	if rawURL == "" {
		rawURL = "http://localhost:4200/{locale}/reset-password"
	}
	locale = strings.TrimSpace(locale)
	if locale == "" {
		locale = "en"
	}
	if strings.Contains(rawURL, "{locale}") {
		rawURL = strings.ReplaceAll(rawURL, "{locale}", url.PathEscape(locale))
	} else if !strings.Contains(rawURL, "/en/") && !strings.Contains(rawURL, "/de/") && !strings.Contains(rawURL, "/fr/") {
		rawURL = strings.TrimRight(rawURL, "/") + "/" + url.PathEscape(locale) + "/reset-password"
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := parsed.Query()
	q.Set("token", token)
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

func (h *Handler) writePasswordResetAccepted(w http.ResponseWriter) {
	respond.Data(w, http.StatusAccepted, passwordResetResponse{
		Message: "if the email exists, a password reset link has been sent",
	})
}

func newPasswordResetToken() (string, string, error) {
	raw := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, mustPasswordResetTokenHash(token), nil
}

func passwordResetTokenHash(token string) (string, bool) {
	token = strings.TrimSpace(token)
	if len(token) < 32 || len(token) > 256 {
		return "", false
	}
	return mustPasswordResetTokenHash(token), true
}

func mustPasswordResetTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
