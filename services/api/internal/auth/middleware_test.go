package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	nextCalled := false

	handler := RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called without a bearer token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", body.Error.Code)
	}
}

func TestRequireAuthAcceptsValidTokenAndSetsContextUserID(t *testing.T) {
	tokens := newTestTokenManager(t)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	nextCalled := false
	handler := RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected user id in request context")
		}
		if userID != 42 {
			t.Fatalf("expected user id 42, got %d", userID)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/search", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireAuthRejectsInvalidBearerToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	nextCalled := false
	handler := RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/search", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called with an invalid bearer token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeMiddlewareErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestRequireAuthRejectsExpiredBearerToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	tokens.now = func() time.Time { return now.Add(AccessTokenTTL + time.Second) }

	nextCalled := false
	handler := RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/search", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called with an expired bearer token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeMiddlewareErrorCode(t, rec); got != "TOKEN_EXPIRED" {
		t.Fatalf("expected TOKEN_EXPIRED, got %q", got)
	}
}

func TestBearerTokenParsesSchemeCaseInsensitively(t *testing.T) {
	token, ok := bearerToken("bEaReR abc.def.ghi")
	if !ok {
		t.Fatal("expected bearer token to parse")
	}
	if token != "abc.def.ghi" {
		t.Fatalf("expected token body, got %q", token)
	}
}

func decodeMiddlewareErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body.Error.Code
}
