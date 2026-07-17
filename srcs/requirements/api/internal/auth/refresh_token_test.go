package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRefreshTokenReturnsNewAccessToken(t *testing.T) {
	store := newMemoryUserStore()
	user := createPasswordUser(t, store, "alice@example.com", "alice_1", "right-password")
	tokens := newTestTokenManager(t)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }
	refreshToken, _, err := tokens.CreateRefreshToken(user.ID)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}
	handler := NewHandler(store, tokens)

	rec := callRefreshToken(t, handler, refreshToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	assertNoStoreHeaders(t, rec)

	response := decodeRefreshTokenEnvelope(t, rec)
	if response.Data.TokenType != "Bearer" {
		t.Fatalf("expected Bearer token type, got %q", response.Data.TokenType)
	}
	if response.Data.ExpiresIn != int64(AccessTokenTTL.Seconds()) {
		t.Fatalf("expected expires_in %d, got %d", int64(AccessTokenTTL.Seconds()), response.Data.ExpiresIn)
	}
	claims, err := tokens.ValidateAccessToken(response.Data.AccessToken)
	if err != nil {
		t.Fatalf("new access token should validate: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, claims.UserID)
	}

	var raw struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw response: %v", err)
	}
	for _, field := range []string{"user", "refresh_token"} {
		if _, ok := raw.Data[field]; ok {
			t.Fatalf("refresh response must not include %q: %s", field, rec.Body.String())
		}
	}

	if _, err := tokens.ValidateRefreshToken(refreshToken); err != nil {
		t.Fatalf("refresh token should remain valid: %v", err)
	}
	secondRec := callRefreshToken(t, handler, refreshToken)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected refresh token reuse to return 200, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
}

func TestRefreshTokenRejectsMissingUser(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	refreshToken, _, err := tokens.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	rec := callRefreshToken(t, NewHandler(store, tokens), refreshToken)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	assertNoStoreHeaders(t, rec)
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "INVALID_REFRESH_TOKEN" {
		t.Fatalf("expected INVALID_REFRESH_TOKEN, got %q", got)
	}
}

func TestRefreshTokenReturnsInternalErrorWhenUserLookupFails(t *testing.T) {
	store := newMemoryUserStore()
	store.userExistsErr = errors.New("db down")
	tokens := newTestTokenManager(t)
	refreshToken, _, err := tokens.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	rec := callRefreshToken(t, NewHandler(store, tokens), refreshToken)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	assertNoStoreHeaders(t, rec)
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
}

func TestRefreshTokenWithoutStoreReturnsInternalError(t *testing.T) {
	tokens := newTestTokenManager(t)
	refreshToken, _, err := tokens.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	rec := callRefreshToken(t, NewHandler(nil, tokens), refreshToken)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	assertNoStoreHeaders(t, rec)
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
}

func TestRefreshTokenRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantCode  string
		wantField bool
	}{
		{name: "missing field", body: `{}`, wantCode: "VALIDATION_ERROR", wantField: true},
		{name: "null", body: `{"refresh_token":null}`, wantCode: "VALIDATION_ERROR", wantField: true},
		{name: "empty", body: `{"refresh_token":""}`, wantCode: "VALIDATION_ERROR", wantField: true},
		{name: "whitespace", body: `{"refresh_token":"  \t "}`, wantCode: "VALIDATION_ERROR", wantField: true},
		{name: "wrong field type", body: `{"refresh_token":42}`, wantCode: "BAD_REQUEST"},
		{name: "malformed JSON", body: `{"refresh_token":`, wantCode: "BAD_REQUEST"},
		{name: "unknown field", body: `{"refresh_token":"token","extra":true}`, wantCode: "BAD_REQUEST"},
		{name: "multiple documents", body: `{"refresh_token":"token"} {}`, wantCode: "BAD_REQUEST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(nil, newTestTokenManager(t))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			handler.RefreshToken(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			assertNoStoreHeaders(t, rec)
			errorBody := decodeErrorEnvelope(t, rec).Error
			if errorBody.Code != tt.wantCode {
				t.Fatalf("expected %s, got %s", tt.wantCode, errorBody.Code)
			}
			if tt.wantField {
				if _, ok := errorBody.Fields["refresh_token"]; !ok {
					t.Fatalf("expected refresh_token field error, got %+v", errorBody.Fields)
				}
			}
		})
	}
}

func TestRefreshTokenRejectsInvalidTokens(t *testing.T) {
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens := newTestTokenManager(t)
	tokens.now = func() time.Time { return now }

	accessToken, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create access token: %v", err)
	}
	expiredRefreshToken, _, err := tokens.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create expired refresh token: %v", err)
	}
	wrongIssuer, err := NewTokenManager(testJWTSecret, "other-issuer")
	if err != nil {
		t.Fatalf("new wrong-issuer manager: %v", err)
	}
	wrongIssuer.now = func() time.Time { return now }
	wrongIssuerToken, _, err := wrongIssuer.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create wrong-issuer token: %v", err)
	}
	wrongSignature, err := NewTokenManager("abcdef0123456789abcdef0123456789", tokens.issuer)
	if err != nil {
		t.Fatalf("new wrong-signature manager: %v", err)
	}
	wrongSignature.now = func() time.Time { return now }
	wrongSignatureToken, _, err := wrongSignature.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create wrong-signature token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		expired bool
	}{
		{name: "random text", token: "not-a-token"},
		{name: "access token", token: accessToken},
		{name: "expired refresh token", token: expiredRefreshToken, expired: true},
		{name: "wrong issuer", token: wrongIssuerToken},
		{name: "wrong signature", token: wrongSignatureToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expired {
				tokens.now = func() time.Time { return now.Add(RefreshTokenTTL + time.Second) }
			} else {
				tokens.now = func() time.Time { return now }
			}
			rec := callRefreshToken(t, NewHandler(nil, tokens), tt.token)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
			}
			assertNoStoreHeaders(t, rec)
			if got := decodeErrorEnvelope(t, rec).Error.Code; got != "INVALID_REFRESH_TOKEN" {
				t.Fatalf("expected INVALID_REFRESH_TOKEN, got %q", got)
			}
		})
	}
}

func TestRefreshTokenWithoutTokenManagerReturnsInternalError(t *testing.T) {
	handler := NewHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	handler.RefreshToken(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	assertNoStoreHeaders(t, rec)
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
}

func TestLoginRefreshesExpiredAccessToken(t *testing.T) {
	store := newMemoryUserStore()
	user := createPasswordUser(t, store, "alice@example.com", "alice_1", "right-password")
	tokens := newTestTokenManager(t)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }
	handler := NewHandler(store, tokens)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"login":"alice@example.com","password":"right-password"}`))
	loginRec := httptest.NewRecorder()
	handler.Login(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d: %s", loginRec.Code, loginRec.Body.String())
	}
	loginResponse := decodeAuthEnvelope(t, loginRec).Data

	tokens.now = func() time.Time { return now.Add(AccessTokenTTL + time.Second) }
	if _, err := tokens.ValidateAccessToken(loginResponse.AccessToken); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected old access token to be expired, got %v", err)
	}

	refreshRec := callRefreshToken(t, handler, loginResponse.RefreshToken)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("expected refresh 200, got %d: %s", refreshRec.Code, refreshRec.Body.String())
	}
	newAccessToken := decodeRefreshTokenEnvelope(t, refreshRec).Data.AccessToken
	claims, err := tokens.ValidateAccessToken(newAccessToken)
	if err != nil {
		t.Fatalf("new access token should validate: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, claims.UserID)
	}
}

func callRefreshToken(t *testing.T, handler *Handler, token string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(refreshTokenRequest{RefreshToken: token})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", strings.NewReader(string(body)))
	rec := httptest.NewRecorder()
	handler.RefreshToken(rec, req)
	return rec
}

func decodeRefreshTokenEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Data refreshTokenResponse `json:"data"`
} {
	t.Helper()
	var body struct {
		Data refreshTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func assertNoStoreHeaders(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected Cache-Control no-store, got %q", got)
	}
	if got := rec.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("expected Pragma no-cache, got %q", got)
	}
}
