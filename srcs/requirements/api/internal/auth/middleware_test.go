package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeUserExistenceChecker struct {
	exists     bool
	err        error
	calls      int
	seenUserID int64
}

func (f *fakeUserExistenceChecker) UserExists(_ context.Context, userID int64) (bool, error) {
	f.calls++
	f.seenUserID = userID
	if f.err != nil {
		return false, f.err
	}
	return f.exists, nil
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	checker := &fakeUserExistenceChecker{exists: true}
	nextCalled := false

	handler := RequireAuth(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if checker.calls != 0 {
		t.Fatalf("user checker must not be called without a valid token, got %d calls", checker.calls)
	}
}

func TestRequireAuthAcceptsValidTokenAndSetsContextUserID(t *testing.T) {
	tokens := newTestTokenManager(t)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	checker := &fakeUserExistenceChecker{exists: true}

	nextCalled := false
	handler := RequireAuth(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if checker.calls != 1 || checker.seenUserID != 42 {
		t.Fatalf("expected one user lookup for id 42, got calls=%d id=%d", checker.calls, checker.seenUserID)
	}
}

func TestRequireAuthRejectsTokenForMissingUser(t *testing.T) {
	tokens := newTestTokenManager(t)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	checker := &fakeUserExistenceChecker{exists: false}

	nextCalled := false
	handler := RequireAuth(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called for a missing user")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeMiddlewareErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
	if checker.calls != 1 || checker.seenUserID != 42 {
		t.Fatalf("expected one user lookup for id 42, got calls=%d id=%d", checker.calls, checker.seenUserID)
	}
}

func TestRequireAuthReturnsInternalErrorWhenUserLookupFails(t *testing.T) {
	tokens := newTestTokenManager(t)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	checker := &fakeUserExistenceChecker{err: errors.New("db down")}

	nextCalled := false
	handler := RequireAuth(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called when user lookup fails")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeMiddlewareErrorCode(t, rec); got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
	if checker.calls != 1 || checker.seenUserID != 42 {
		t.Fatalf("expected one user lookup for id 42, got calls=%d id=%d", checker.calls, checker.seenUserID)
	}
}

func TestRequireAuthReturnsInternalErrorWithoutUserChecker(t *testing.T) {
	tokens := newTestTokenManager(t)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	nextCalled := false
	handler := RequireAuth(tokens, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called without a user checker")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeMiddlewareErrorCode(t, rec); got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
}

func TestRequireAuthRejectsInvalidBearerToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	checker := &fakeUserExistenceChecker{exists: true}
	nextCalled := false
	handler := RequireAuth(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if checker.calls != 0 {
		t.Fatalf("user checker must not be called for invalid tokens, got %d calls", checker.calls)
	}
}

func TestRequireAuthRejectsRefreshToken(t *testing.T) {
	tokens := newTestTokenManager(t)
	token, _, err := tokens.CreateRefreshToken(42)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}
	checker := &fakeUserExistenceChecker{exists: true}

	nextCalled := false
	handler := RequireAuth(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called with a refresh token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeMiddlewareErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
	if checker.calls != 0 {
		t.Fatalf("user checker must not be called for refresh tokens, got %d calls", checker.calls)
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
	checker := &fakeUserExistenceChecker{exists: true}

	nextCalled := false
	handler := RequireAuth(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if checker.calls != 0 {
		t.Fatalf("user checker must not be called for expired tokens, got %d calls", checker.calls)
	}
}

func TestRequireAuthRejectsWrongIssuer(t *testing.T) {
	tokenIssuer, err := NewTokenManager(testJWTSecret, "issuer-a")
	if err != nil {
		t.Fatalf("new issuer a token manager: %v", err)
	}
	middlewareIssuer, err := NewTokenManager(testJWTSecret, "issuer-b")
	if err != nil {
		t.Fatalf("new issuer b token manager: %v", err)
	}
	token, _, err := tokenIssuer.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	checker := &fakeUserExistenceChecker{exists: true}

	nextCalled := false
	handler := RequireAuth(middlewareIssuer, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called with a wrong issuer token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeMiddlewareErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
	if checker.calls != 0 {
		t.Fatalf("user checker must not be called for wrong issuer tokens, got %d calls", checker.calls)
	}
}

func TestRequireAuthRejectsNonPositiveUserID(t *testing.T) {
	tokens := newTestTokenManager(t)
	token, _, err := tokens.CreateAccessToken(0)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	checker := &fakeUserExistenceChecker{exists: true}

	nextCalled := false
	handler := RequireAuth(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler must not be called with a non-positive user id token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeMiddlewareErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
	if checker.calls != 0 {
		t.Fatalf("user checker must not be called for invalid user ids, got %d calls", checker.calls)
	}
}

func TestDevAuthenticateAsSetsContextUserID(t *testing.T) {
	nextCalled := false
	handler := DevAuthenticateAs(7)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected user id in request context")
		}
		if userID != 7 {
			t.Fatalf("expected user id 7, got %d", userID)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/watched", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
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
