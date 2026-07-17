package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hypertube/api/internal/auth"
	"hypertube/api/internal/comments"
	"hypertube/api/internal/models"
	"hypertube/api/internal/movies"
	"hypertube/api/internal/stream"
	"hypertube/api/internal/users"
)

func TestRouterHealthCheck(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRouterPublicAuthJSONRoutes(t *testing.T) {
	router, _ := newTestRouter(t)

	tests := []struct {
		name string
		path string
		body string
	}{
		{
			name: "login",
			path: "/api/v1/auth/login",
			body: `{"login":`,
		},
		{
			name: "register",
			path: "/api/v1/auth/register",
			body: `{"email":`,
		},
		{
			name: "refresh token",
			path: "/api/v1/auth/refresh-token",
			body: `{"refresh_token":`,
		},
		{
			name: "password reset request",
			path: "/api/v1/auth/password-reset/send-email",
			body: `{"email":`,
		},
		{
			name: "reset password",
			path: "/api/v1/auth/password-reset/set-new-password",
			body: `{"token":`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected public route to return 400 for bad JSON, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeRouterErrorCode(t, rec); got != "BAD_REQUEST" {
				t.Fatalf("expected BAD_REQUEST, got %q", got)
			}
		})
	}
}

func TestRouterOAuthTokenRoutesArePublic(t *testing.T) {
	router, _ := newTestRouter(t)

	tests := []struct {
		path string
		body string
	}{
		{path: "/api/v1/oauth/token", body: "grant_type=authorization_code"},
		{path: "/api/v1/oauth/token", body: "grant_type=password&username=alice_1&password=correct-horse-battery"},
	}

	for _, tt := range tests {
		t.Run(tt.path+" "+tt.body, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected public OAuth token route to return 400 for unsupported grant, got %d: %s", rec.Code, rec.Body.String())
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Error != "unsupported_grant_type" {
				t.Fatalf("expected unsupported_grant_type, got %q", body.Error)
			}
		})
	}
}

func TestRouterOAuthApplicationRoutesRequireBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/api/v1/oauth/applications", body: `{}`},
		{method: http.MethodGet, path: "/api/v1/oauth/applications"},
		{method: http.MethodPatch, path: "/api/v1/oauth/applications/1", body: `{}`},
		{method: http.MethodDelete, path: "/api/v1/oauth/applications/1"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeRouterErrorCode(t, rec); got != "UNAUTHORIZED" {
				t.Fatalf("expected UNAUTHORIZED, got %q", got)
			}
		})
	}
}

func TestRouterOAuthApplicationRoutesWithValidTokenReachHandler(t *testing.T) {
	router, tokens := newTestRouter(t)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	for _, path := range []string{"/api/v1/oauth/applications"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("expected handler response, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeRouterErrorCode(t, rec); got != "INTERNAL_ERROR" {
				t.Fatalf("expected INTERNAL_ERROR from reached handler, got %q", got)
			}
		})
	}
}

func TestRouterPublicOAuthProviderRoutes(t *testing.T) {
	router, _ := newTestRouter(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "42 login", path: "/api/v1/auth/42/login"},
		{name: "42 callback", path: "/api/v1/auth/42/callback?code=x&state=y"},
		{name: "github login", path: "/api/v1/auth/github/login"},
		{name: "github callback", path: "/api/v1/auth/github/callback?code=x&state=y"},
		{name: "gitlab login", path: "/api/v1/auth/gitlab/login"},
		{name: "gitlab callback", path: "/api/v1/auth/gitlab/callback?code=x&state=y"},
		{name: "42 callback alias", path: "/oauth/callback/42?code=x&state=y"},
		{name: "github callback alias", path: "/oauth/callback/github?code=x&state=y"},
		{name: "gitlab callback alias", path: "/oauth/callback/gitlab?code=x&state=y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected public OAuth route to return 503 when unconfigured, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeRouterErrorCode(t, rec); got != "OAUTH_NOT_CONFIGURED" {
				t.Fatalf("expected OAUTH_NOT_CONFIGURED, got %q", got)
			}
		})
	}
}

func TestRouterProtectedRouteRejectsMissingBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/directstream", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeRouterErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestRouterProtectedRouteRejectsInvalidBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/directstream", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeRouterErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestRouterProtectedRouteWithValidTokenReachesHandler(t *testing.T) {
	router, tokens := newTestRouter(t)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/search", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected downstream handler response, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeRouterErrorCode(t, rec); got != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", got)
	}
}

func TestRouterUserRouteRequiresBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing token 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeRouterErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestRouterUserCommentsRouteRequiresBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/7/comments", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing token 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeRouterErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestRouterUserFilmHistoryRouteRequiresBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/7/movie-history", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing token 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeRouterErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestRouterMovieProgressRouteRequiresBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/movies/tt1234567/progress", strings.NewReader(`{"progress":1804,"complete":false}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing token 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeRouterErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestRouterUserRouteRejectsInvalidBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid token 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeRouterErrorCode(t, rec); got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestRouterUserRouteWithValidTokenReachesHandler(t *testing.T) {
	userStore := &routerUserStore{}
	router, tokens := newTestRouterWithUsersStore(t, userStore)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected users handler response, got %d: %s", rec.Code, rec.Body.String())
	}
	if userStore.requestedID != 7 {
		t.Fatalf("expected handler to query id 7, got %d", userStore.requestedID)
	}

	var body struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.ID != 7 {
		t.Fatalf("expected response id 7, got %d", body.Data.ID)
	}
}

func TestRouterUserPatchRouteWithValidTokenReachesHandler(t *testing.T) {
	userStore := &routerUserStore{}
	router, tokens := newTestRouterWithUsersStore(t, userStore)
	token, _, err := tokens.CreateAccessToken(7)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/7", strings.NewReader(`{"color":"green"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected users patch handler response, got %d: %s", rec.Code, rec.Body.String())
	}
	if !userStore.updated || userStore.requestedID != 7 {
		t.Fatalf("expected handler to update id 7, got updated=%v id=%d", userStore.updated, userStore.requestedID)
	}

	var body struct {
		Data struct {
			ID    int64  `json:"id"`
			Color string `json:"color"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.ID != 7 || body.Data.Color != models.UserColorGreen {
		t.Fatalf("unexpected response: %+v", body.Data)
	}
}

func TestRouterSetPasswordRouteRequiresBearerToken(t *testing.T) {
	router, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/new-password", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized || decodeRouterErrorCode(t, rec) != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRouterSetPasswordRouteWithValidTokenReachesHandler(t *testing.T) {
	router, tokens := newTestRouter(t)
	token, _, err := tokens.CreateAccessToken(42)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/new-password", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Error struct {
			Code   string                     `json:"code"`
			Fields map[string]json.RawMessage `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	for _, field := range []string{"current-password", "new-password"} {
		if _, ok := body.Error.Fields[field]; !ok {
			t.Fatalf("expected field %q, got %+v", field, body.Error.Fields)
		}
	}
}

func newTestRouter(t *testing.T) (http.Handler, *auth.TokenManager) {
	return newTestRouterWithUsersStore(t, &routerUserStore{})
}

func newTestRouterWithUsersStore(t *testing.T, userStore *routerUserStore) (http.Handler, *auth.TokenManager) {
	t.Helper()

	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", "hypertube-test")
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	return newRouter(
		movies.NewMoviesHandler(nil, nil, nil, nil),
		comments.NewCommentsHandler(nil),
		auth.NewHandler(nil, tokens),
		users.NewHandler(userStore),
		tokens,
		routerAuthUserChecker{},
		stream.NewStreamHandler(stream.NewStore(nil)),
		"http://localhost:4200",
	), tokens
}

type routerAuthUserChecker struct{}

func (routerAuthUserChecker) UserExists(context.Context, int64) (bool, error) {
	return true, nil
}

type routerUserStore struct {
	requestedID int64
	updated     bool
}

func (s *routerUserStore) ListUsers(_ context.Context, limit, offset int) ([]models.User, error) {
	return nil, nil
}

func (s *routerUserStore) CountUsers(_ context.Context) (int, error) {
	return 0, nil
}

func (s *routerUserStore) FindUserByID(_ context.Context, id int64) (models.User, error) {
	s.requestedID = id
	return models.User{ID: id, Username: "alice"}, nil
}

func (s *routerUserStore) UserHasOAuthAccount(_ context.Context, _ int64) (bool, error) {
	return false, nil
}

func (s *routerUserStore) UpdateUser(_ context.Context, id int64, params users.UpdateUserParams) (models.User, error) {
	s.requestedID = id
	s.updated = true

	user := models.User{ID: id, Username: "alice", Color: models.UserColorPurple}
	if params.Color != nil {
		user.Color = *params.Color
	}
	return user, nil
}

func (s *routerUserStore) UpdatePasswordHash(_ context.Context, id int64, _, _ string) error {
	s.requestedID = id
	return nil
}

func decodeRouterErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body.Error.Code
}
