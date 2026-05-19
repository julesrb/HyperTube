package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fakePasswordResetMailer struct {
	calls     int
	toEmail   string
	toName    string
	resetURL  string
	expiresIn time.Duration
}

func (m *fakePasswordResetMailer) SendPasswordReset(_ context.Context, toEmail string, toName string, resetURL string, expiresIn time.Duration) error {
	m.calls++
	m.toEmail = toEmail
	m.toName = toName
	m.resetURL = resetURL
	m.expiresIn = expiresIn
	return nil
}

func TestRequestPasswordResetSendsResetLinkForExistingUser(t *testing.T) {
	store := newMemoryUserStore()
	passwordHash, err := HashPassword("old-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := store.CreateUser(context.Background(), CreateUserParams{
		Email:        "alice@example.com",
		Username:     "alice_1",
		FirstName:    "Alice",
		LastName:     "Example",
		PasswordHash: passwordHash,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	mailer := &fakePasswordResetMailer{}
	ttl := 15 * time.Minute
	handler := NewHandler(
		store,
		newTestTokenManager(t),
		WithPasswordResetMailer(mailer),
		WithPasswordResetURL("https://frontend.test/{locale}/reset-password?source=email"),
		WithPasswordResetTTL(ttl),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset", strings.NewReader(`{"email":" Alice@Example.COM ","locale":"de"}`))
	rec := httptest.NewRecorder()

	handler.RequestPasswordReset(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodePasswordResetEnvelope(t, rec).Data.Message; got != "if the email exists, a password reset link has been sent" {
		t.Fatalf("unexpected response message: %q", got)
	}
	if mailer.calls != 1 {
		t.Fatalf("expected one reset email, got %d", mailer.calls)
	}
	if mailer.toEmail != "alice@example.com" {
		t.Fatalf("expected normalized recipient email, got %q", mailer.toEmail)
	}
	if mailer.toName != "Alice Example" {
		t.Fatalf("expected full recipient name, got %q", mailer.toName)
	}
	if mailer.expiresIn != ttl {
		t.Fatalf("expected ttl %s, got %s", ttl, mailer.expiresIn)
	}

	resetURL, err := url.Parse(mailer.resetURL)
	if err != nil {
		t.Fatalf("parse reset URL: %v", err)
	}
	if resetURL.Scheme != "https" || resetURL.Host != "frontend.test" || resetURL.Path != "/de/reset-password" {
		t.Fatalf("unexpected reset URL: %q", mailer.resetURL)
	}
	if got := resetURL.Query().Get("source"); got != "email" {
		t.Fatalf("expected existing query param to be preserved, got %q", got)
	}
	token := resetURL.Query().Get("token")
	if token == "" {
		t.Fatal("expected reset URL token")
	}
	tokenHash, ok := passwordResetTokenHash(token)
	if !ok {
		t.Fatalf("generated token should be accepted, got %q", token)
	}
	resetToken, ok := store.resetTokens[tokenHash]
	if !ok {
		t.Fatalf("expected token hash to be stored")
	}
	if resetToken.userID != user.ID {
		t.Fatalf("expected token user id %d, got %d", user.ID, resetToken.userID)
	}
	if resetToken.expiresAt.Before(time.Now().UTC().Add(ttl - time.Minute)) {
		t.Fatalf("expected token expiry near ttl, got %s", resetToken.expiresAt)
	}
	if _, rawTokenStored := store.resetTokens[token]; rawTokenStored {
		t.Fatal("raw reset token must not be stored")
	}
}

func TestRequestPasswordResetDoesNotRevealUnknownEmail(t *testing.T) {
	store := newMemoryUserStore()
	mailer := &fakePasswordResetMailer{}
	handler := NewHandler(store, newTestTokenManager(t), WithPasswordResetMailer(mailer))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset", strings.NewReader(`{"email":"missing@example.com"}`))
	rec := httptest.NewRecorder()

	handler.RequestPasswordReset(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	if mailer.calls != 0 {
		t.Fatalf("expected no email for unknown user, got %d calls", mailer.calls)
	}
	if len(store.resetTokens) != 0 {
		t.Fatalf("expected no token for unknown user, got %d", len(store.resetTokens))
	}
}

func TestRequestPasswordResetRejectsInvalidInputAndMissingMailer(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		withMailer bool
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid email",
			body:       `{"email":"not-an-email"}`,
			withMailer: false,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "unknown field",
			body:       `{"email":"alice@example.com","admin":true}`,
			withMailer: true,
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "missing mailer",
			body:       `{"email":"alice@example.com"}`,
			withMailer: false,
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "EMAIL_NOT_CONFIGURED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMemoryUserStore()
			passwordHash, err := HashPassword("old-password")
			if err != nil {
				t.Fatalf("hash password: %v", err)
			}
			if _, err := store.CreateUser(context.Background(), CreateUserParams{
				Email:        "alice@example.com",
				Username:     "alice_1",
				FirstName:    "Alice",
				LastName:     "Example",
				PasswordHash: passwordHash,
			}); err != nil {
				t.Fatalf("create user: %v", err)
			}

			opts := []HandlerOption{}
			if tt.withMailer {
				opts = append(opts, WithPasswordResetMailer(&fakePasswordResetMailer{}))
			}
			handler := NewHandler(store, newTestTokenManager(t), opts...)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handler.RequestPasswordReset(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if got := decodeErrorEnvelope(t, rec).Error.Code; got != tt.wantCode {
				t.Fatalf("expected error code %q, got %q", tt.wantCode, got)
			}
		})
	}
}

func TestResetPasswordConsumesTokenAndUpdatesPassword(t *testing.T) {
	store := newMemoryUserStore()
	oldHash, err := HashPassword("old-password")
	if err != nil {
		t.Fatalf("hash old password: %v", err)
	}
	user, err := store.CreateUser(context.Background(), CreateUserParams{
		Email:        "alice@example.com",
		Username:     "alice_1",
		FirstName:    "Alice",
		LastName:     "Example",
		PasswordHash: oldHash,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	token := "valid-reset-token-with-enough-length-123"
	tokenHash := mustPasswordResetTokenHash(token)
	if err := store.CreatePasswordResetToken(context.Background(), CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute),
	}); err != nil {
		t.Fatalf("create reset token: %v", err)
	}
	handler := NewHandler(store, newTestTokenManager(t))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", strings.NewReader(`{"token":"`+token+`","password":"new-password"}`))
	rec := httptest.NewRecorder()

	handler.ResetPassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodePasswordResetEnvelope(t, rec).Data.Message; got != "password has been reset" {
		t.Fatalf("unexpected response message: %q", got)
	}

	updatedUser, err := store.FindUserByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("find updated user: %v", err)
	}
	if CheckPassword(updatedUser.PasswordHash, "old-password") {
		t.Fatal("old password should not match after reset")
	}
	if !CheckPassword(updatedUser.PasswordHash, "new-password") {
		t.Fatal("new password should match after reset")
	}

	reuseReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", strings.NewReader(`{"token":"`+token+`","password":"another-password"}`))
	reuseRec := httptest.NewRecorder()

	handler.ResetPassword(reuseRec, reuseReq)

	if reuseRec.Code != http.StatusBadRequest {
		t.Fatalf("expected token reuse to return 400, got %d: %s", reuseRec.Code, reuseRec.Body.String())
	}
	if got := decodeErrorEnvelope(t, reuseRec).Error.Code; got != "INVALID_RESET_TOKEN" {
		t.Fatalf("expected INVALID_RESET_TOKEN, got %q", got)
	}
}

func TestResetPasswordRejectsInvalidPasswordAndTokens(t *testing.T) {
	store := newMemoryUserStore()
	passwordHash, err := HashPassword("old-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := store.CreateUser(context.Background(), CreateUserParams{
		Email:        "alice@example.com",
		Username:     "alice_1",
		FirstName:    "Alice",
		LastName:     "Example",
		PasswordHash: passwordHash,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	expiredToken := "expired-reset-token-with-enough-length-123"
	if err := store.CreatePasswordResetToken(context.Background(), CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: mustPasswordResetTokenHash(expiredToken),
		ExpiresAt: time.Now().UTC().Add(-time.Minute),
	}); err != nil {
		t.Fatalf("create expired reset token: %v", err)
	}

	handler := NewHandler(store, newTestTokenManager(t))
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "short token",
			body:       `{"token":"short","password":"new-password"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_RESET_TOKEN",
		},
		{
			name:       "short password",
			body:       `{"token":"valid-reset-token-with-enough-length-123","password":"short"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "unknown token",
			body:       `{"token":"unknown-reset-token-with-enough-length-123","password":"new-password"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_RESET_TOKEN",
		},
		{
			name:       "expired token",
			body:       `{"token":"` + expiredToken + `","password":"new-password"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_RESET_TOKEN",
		},
		{
			name:       "unknown field",
			body:       `{"token":"valid-reset-token-with-enough-length-123","password":"new-password","extra":true}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handler.ResetPassword(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if got := decodeErrorEnvelope(t, rec).Error.Code; got != tt.wantCode {
				t.Fatalf("expected error code %q, got %q", tt.wantCode, got)
			}
		})
	}
}

func decodePasswordResetEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Data passwordResetResponse `json:"data"`
} {
	t.Helper()

	var body struct {
		Data passwordResetResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}
