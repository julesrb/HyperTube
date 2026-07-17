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

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/requestjson"
	"hypertube/api/internal/respond"
	"hypertube/api/internal/userinput"
)

const passwordResetTokenBytes = 32

type passwordResetMailer interface {
	SendPasswordReset(ctx context.Context, toEmail string, toName string, resetURL string, expiresIn time.Duration, locale i18n.Locale) error
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
	if !requestjson.DecodeJSON(w, r, &req) {
		return
	}

	email, validationMessage, ok := userinput.ValidateEmail(req.Email)
	if !ok {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "email", validationMessage)
		return
	}

	if h.store == nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}
	if h.passwordResetMailer == nil {
		respond.LocalizedError(w, r, http.StatusServiceUnavailable, "EMAIL_NOT_CONFIGURED", i18n.MsgPasswordResetEmailNotConfigured)
		return
	}

	user, err := h.store.FindUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			h.writePasswordResetAccepted(w, r)
			return
		}
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadUser)
		return
	}

	token, tokenHash, err := newPasswordResetToken()
	if err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedCreatePasswordResetToken)
		return
	}

	expiresAt := time.Now().UTC().Add(h.passwordResetTTL)
	if err := h.store.CreatePasswordResetToken(r.Context(), CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedStorePasswordResetToken)
		return
	}

	resetLocale := i18n.FromValue(req.Locale)
	if strings.TrimSpace(req.Locale) == "" {
		resetLocale = i18n.FromRequest(r)
	}
	resetURL, err := h.buildPasswordResetURL(token, resetLocale.String())
	if err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgPasswordResetURLNotConfigured)
		return
	}

	toName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if err := h.passwordResetMailer.SendPasswordReset(r.Context(), user.Email, toName, resetURL, h.passwordResetTTL, resetLocale); err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedSendPasswordResetEmail)
		return
	}

	h.writePasswordResetAccepted(w, r)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if !requestjson.DecodeJSON(w, r, &req) {
		return
	}

	if h.store == nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}

	tokenHash, ok := passwordResetTokenHash(req.Token)
	if !ok {
		respond.LocalizedError(w, r, http.StatusBadRequest, "INVALID_RESET_TOKEN", i18n.MsgInvalidPasswordResetToken)
		return
	}

	if validationMessage, ok := userinput.ValidateRequiredPassword(req.Password); !ok {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "password", validationMessage)
		return
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "password", i18n.MsgPasswordInvalid)
		return
	}

	if _, err := h.store.ResetPasswordWithToken(r.Context(), tokenHash, passwordHash); err != nil {
		if errors.Is(err, ErrInvalidPasswordResetToken) {
			respond.LocalizedError(w, r, http.StatusBadRequest, "INVALID_RESET_TOKEN", i18n.MsgInvalidPasswordResetToken)
			return
		}
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedResetPassword)
		return
	}

	respond.Data(w, http.StatusOK, passwordResetResponse{Message: i18n.T(i18n.FromRequest(r), i18n.MsgPasswordResetSuccess)})
}

func (h *Handler) buildPasswordResetURL(token string, locale string) (string, error) {
	rawURL := strings.TrimSpace(h.passwordResetURL)
	if rawURL == "" {
		rawURL = "http://localhost:4200/{locale}/password-reset/set-new-password"
	}
	locale = strings.TrimSpace(locale)
	if locale == "" {
		locale = "en"
	}
	locale = i18n.FromValue(locale).String()
	if strings.Contains(rawURL, "{locale}") {
		rawURL = strings.ReplaceAll(rawURL, "{locale}", url.PathEscape(locale))
	} else if !strings.Contains(rawURL, "/en/") && !strings.Contains(rawURL, "/de/") && !strings.Contains(rawURL, "/fr/") {
		rawURL = strings.TrimRight(rawURL, "/") + "/" + url.PathEscape(locale) + "/password-reset/set-new-password"
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

func (h *Handler) writePasswordResetAccepted(w http.ResponseWriter, r *http.Request) {
	respond.Data(w, http.StatusAccepted, passwordResetResponse{
		Message: i18n.T(i18n.FromRequest(r), i18n.MsgPasswordResetAccepted),
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
