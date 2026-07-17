package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestOAuthTokenPasswordGrantIsUnsupported(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))

	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{
			name:        "form",
			contentType: "application/x-www-form-urlencoded",
			body:        "grant_type=password&username=alice_1&password=correct-horse-battery",
		},
		{
			name:        "json",
			contentType: "application/json",
			body:        `{"grant_type":"password","username":"alice_1","password":"correct-horse-battery"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()

			handler.OAuthToken(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			response := decodeOAuthTokenError(t, rec)
			if response.Error != "unsupported_grant_type" {
				t.Fatalf("expected unsupported_grant_type, got %q", response.Error)
			}
			if strings.Contains(response.ErrorDescription, "password") {
				t.Fatalf("error description must not say password is supported: %q", response.ErrorDescription)
			}
		})
	}
}

func TestOAuthTokenClientCredentialsGrantReturnsBearerTokenFromForm(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	handler := NewHandler(store, tokens)
	createStoredOAuthApplication(t, store, 42, "API Client", "hypertube-api", "replace-with-secret")

	form := validOAuthClientCredentialsForm()
	form.Set("client_id", " hypertube-api ")
	rec := postOAuthTokenForm(t, handler, form)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	response := decodeOAuthTokenResponse(t, rec)
	if response.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if response.TokenType != "Bearer" {
		t.Fatalf("expected Bearer token type, got %q", response.TokenType)
	}
	if response.ExpiresIn != int64(AccessTokenTTL.Seconds()) {
		t.Fatalf("expected expires_in %d, got %d", int64(AccessTokenTTL.Seconds()), response.ExpiresIn)
	}
	claims, err := tokens.ValidateAccessToken(response.AccessToken)
	if err != nil {
		t.Fatalf("token should validate: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected token user id 42, got %d", claims.UserID)
	}
	assertJSONFieldsAbsent(t, rec.Body.Bytes(), "scope")
}

func TestOAuthTokenClientCredentialsGrantAcceptsJSON(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	createStoredOAuthApplication(t, store, 42, "API Client", "hypertube-api", "replace-with-secret")

	body := `{"grant_type":"client_credentials","client_id":"hypertube-api","client_secret":"replace-with-secret"}`
	rec := postOAuthTokenJSON(t, handler, body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	response := decodeOAuthTokenResponse(t, rec)
	if response.AccessToken == "" {
		t.Fatal("expected access token")
	}
	assertJSONFieldsAbsent(t, rec.Body.Bytes(), "scope")
}

func TestOAuthTokenClientCredentialsGrantAcceptsJSONWithoutGrantType(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	createStoredOAuthApplication(t, store, 42, "API Client", "hypertube-api", "replace-with-secret")

	body := `{"client_id":"hypertube-api","client_secret":"replace-with-secret"}`
	rec := postOAuthTokenJSON(t, handler, body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	response := decodeOAuthTokenResponse(t, rec)
	if response.AccessToken == "" {
		t.Fatal("expected access token")
	}
	assertJSONFieldsAbsent(t, rec.Body.Bytes(), "scope")
}

func TestOAuthTokenDoesNotInferClientCredentialsWhenPasswordFieldsArePresent(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	createStoredOAuthApplication(t, store, 42, "API Client", "hypertube-api", "replace-with-secret")

	form := url.Values{}
	form.Set("client_id", "hypertube-api")
	form.Set("client_secret", "replace-with-secret")
	form.Set("username", "alice_1")
	form.Set("password", "correct-horse-battery")
	rec := postOAuthTokenForm(t, handler, form)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	response := decodeOAuthTokenError(t, rec)
	if response.Error != "invalid_request" {
		t.Fatalf("expected invalid_request, got %q", response.Error)
	}
}

func TestOAuthTokenClientCredentialsGrantRejectsScope(t *testing.T) {
	tests := []struct {
		name string
		call func(*testing.T, *Handler) *httptest.ResponseRecorder
	}{
		{
			name: "form scope",
			call: func(t *testing.T, handler *Handler) *httptest.ResponseRecorder {
				form := validOAuthClientCredentialsForm()
				form.Set("scope", "read:movies")
				return postOAuthTokenForm(t, handler, form)
			},
		},
		{
			name: "form empty scope",
			call: func(t *testing.T, handler *Handler) *httptest.ResponseRecorder {
				form := validOAuthClientCredentialsForm()
				form.Set("scope", "")
				return postOAuthTokenForm(t, handler, form)
			},
		},
		{
			name: "json scope",
			call: func(t *testing.T, handler *Handler) *httptest.ResponseRecorder {
				body := `{"grant_type":"client_credentials","client_id":"hypertube-api","client_secret":"replace-with-secret","scope":"read:movies"}`
				return postOAuthTokenJSON(t, handler, body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMemoryUserStore()
			handler := NewHandler(store, newTestTokenManager(t))
			createStoredOAuthApplication(t, store, 42, "API Client", "hypertube-api", "replace-with-secret")

			rec := tt.call(t, handler)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			errorResponse := decodeOAuthTokenError(t, rec)
			if errorResponse.Error != "invalid_request" {
				t.Fatalf("expected invalid_request, got %q", errorResponse.Error)
			}
		})
	}
}

func TestOAuthTokenClientCredentialsGrantRejectsInvalidRequest(t *testing.T) {
	tests := []struct {
		name       string
		mutateForm func(url.Values)
		wantStatus int
		wantError  string
	}{
		{
			name: "missing client id",
			mutateForm: func(form url.Values) {
				form.Del("client_id")
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid_request",
		},
		{
			name: "missing client secret",
			mutateForm: func(form url.Values) {
				form.Del("client_secret")
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid_request",
		},
		{
			name: "wrong client id",
			mutateForm: func(form url.Values) {
				form.Set("client_id", "other-client")
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid_client",
		},
		{
			name: "wrong client secret",
			mutateForm: func(form url.Values) {
				form.Set("client_secret", "wrong-secret")
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid_client",
		},
		{
			name: "secret with surrounding whitespace",
			mutateForm: func(form url.Values) {
				form.Set("client_secret", " replace-with-secret ")
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid_client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMemoryUserStore()
			handler := NewHandler(store, newTestTokenManager(t))
			createStoredOAuthApplication(t, store, 42, "API Client", "hypertube-api", "replace-with-secret")
			form := validOAuthClientCredentialsForm()
			tt.mutateForm(form)

			rec := postOAuthTokenForm(t, handler, form)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			response := decodeOAuthTokenError(t, rec)
			if response.Error != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, response.Error)
			}
		})
	}
}

func TestOAuthTokenClientCredentialsGrantRejectsNilStore(t *testing.T) {
	handler := NewHandler(nil, newTestTokenManager(t))

	rec := postOAuthTokenForm(t, handler, validOAuthClientCredentialsForm())

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	response := decodeOAuthTokenError(t, rec)
	if response.Error != "server_error" {
		t.Fatalf("expected server_error, got %q", response.Error)
	}
}

func TestOAuthTokenClientCredentialsGrantRejectsNilTokenManager(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, nil)
	createStoredOAuthApplication(t, store, 42, "API Client", "hypertube-api", "replace-with-secret")

	rec := postOAuthTokenForm(t, handler, validOAuthClientCredentialsForm())

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	response := decodeOAuthTokenError(t, rec)
	if response.Error != "server_error" {
		t.Fatalf("expected server_error, got %q", response.Error)
	}
}

func TestOAuthTokenRejectsInvalidGrant(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))

	tests := []struct {
		name      string
		form      url.Values
		wantError string
	}{
		{
			name: "missing grant type",
			form: url.Values{
				"username": {"alice_1"},
				"password": {"right-password"},
			},
			wantError: "invalid_request",
		},
		{
			name: "unsupported grant type",
			form: url.Values{
				"grant_type": {"authorization_code"},
			},
			wantError: "unsupported_grant_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			handler.OAuthToken(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			var response oauthErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Error != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, response.Error)
			}
		})
	}
}

func TestOAuthTokenErrorUsesAcceptLanguage(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "fr")
	rec := httptest.NewRecorder()

	handler.OAuthToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var response oauthErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error != "invalid_request" {
		t.Fatalf("expected invalid_request, got %q", response.Error)
	}
	if response.ErrorDescription != "Type de grant requis" {
		t.Fatalf("expected French OAuth error description, got %q", response.ErrorDescription)
	}
}

func validOAuthClientCredentialsForm() url.Values {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", "hypertube-api")
	form.Set("client_secret", "replace-with-secret")
	return form
}

func postOAuthTokenForm(t *testing.T, handler *Handler, form url.Values) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.OAuthToken(rec, req)
	return rec
}

func postOAuthTokenJSON(t *testing.T, handler *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.OAuthToken(rec, req)
	return rec
}

func decodeOAuthTokenResponse(t *testing.T, rec *httptest.ResponseRecorder) oauthTokenResponse {
	t.Helper()

	var response oauthTokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}

func decodeOAuthTokenError(t *testing.T, rec *httptest.ResponseRecorder) oauthErrorResponse {
	t.Helper()

	var response oauthErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}
