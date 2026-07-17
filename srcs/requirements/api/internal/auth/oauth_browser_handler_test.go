package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFortyTwoLoginRedirectsWithStateCookie(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{authURL: "https://api.intra.42.fr/oauth/authorize"}
	handler := NewHandler(store, tokens, WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login", nil)
	rec := httptest.NewRecorder()

	handler.LoginFortyTwo(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	if provider.lastState == "" {
		t.Fatal("expected generated OAuth state")
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "state="+provider.lastState) {
		t.Fatalf("expected redirect location to include state, got %q", location)
	}

	cookie := rec.Result().Cookies()[0]
	if cookie.Name != oauthStateCookieName {
		t.Fatalf("expected state cookie %q, got %q", oauthStateCookieName, cookie.Name)
	}
	if cookie.Value != provider.lastState {
		t.Fatalf("expected cookie state %q, got %q", provider.lastState, cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Fatal("state cookie must be HttpOnly")
	}
}

func TestFortyTwoLoginStoresOAuthLocaleCookie(t *testing.T) {
	provider := &fakeOAuthProvider{authURL: "https://api.intra.42.fr/oauth/authorize"}
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t), WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9")
	rec := httptest.NewRecorder()

	handler.LoginFortyTwo(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	cookie := findCookie(t, rec, oauthLocaleCookieName(oauthStateCookieName))
	if cookie.Value != "de" {
		t.Fatalf("expected stored OAuth locale de, got %q", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Fatal("locale cookie must be HttpOnly")
	}
}

func TestValidOAuthRedirectPath(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "movie path", value: "/movies/123", want: true},
		{name: "path with query", value: "/movies/123?tab=comments&page=2", want: true},
		{name: "user path", value: "/users/42", want: true},
		{name: "external url", value: "https://evil.example", want: false},
		{name: "protocol relative url", value: "//evil.example", want: false},
		{name: "javascript url", value: "javascript:alert(1)", want: false},
		{name: "relative without slash", value: "movies/123", want: false},
		{name: "backslash path", value: `/\evil.example`, want: false},
		{name: "empty", value: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validOAuthRedirectPath(tt.value); got != tt.want {
				t.Fatalf("validOAuthRedirectPath(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestOAuthRedirectCookieRoundTrip(t *testing.T) {
	tests := []string{
		"/movies/123",
		"/movies/123?tab=comments&page=2",
	}

	for _, redirectPath := range tests {
		t.Run(redirectPath, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login", nil)
			cookie := oauthRedirectCookie(req, oauthStateCookieName, redirectPath, int(oauthStateTTL.Seconds()))
			if cookie.Name != oauthRedirectCookieName(oauthStateCookieName) {
				t.Fatalf("expected redirect cookie name %q, got %q", oauthRedirectCookieName(oauthStateCookieName), cookie.Name)
			}
			if !cookie.HttpOnly {
				t.Fatal("redirect cookie must be HttpOnly")
			}
			if cookie.SameSite != http.SameSiteLaxMode {
				t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
			}
			if cookie.Value == redirectPath {
				t.Fatalf("expected encoded redirect cookie value, got raw value %q", cookie.Value)
			}

			callbackReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback", nil)
			callbackReq.AddCookie(cookie)
			if got := oauthRedirectFromCookie(callbackReq, oauthStateCookieName); got != redirectPath {
				t.Fatalf("expected redirect round trip %q, got %q", redirectPath, got)
			}
		})
	}
}

func TestFortyTwoLoginStoresOAuthRedirectCookie(t *testing.T) {
	provider := &fakeOAuthProvider{authURL: "https://api.intra.42.fr/oauth/authorize"}
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t), WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login?redirect=/movies/123", nil)
	rec := httptest.NewRecorder()

	handler.LoginFortyTwo(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	if provider.lastState == "" {
		t.Fatal("expected generated OAuth state")
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "state="+provider.lastState) {
		t.Fatalf("expected redirect location to include state, got %q", location)
	}
	if cookie := findCookie(t, rec, oauthStateCookieName); cookie.Value != provider.lastState {
		t.Fatalf("expected state cookie %q, got %q", provider.lastState, cookie.Value)
	}
	redirectCookie := findCookie(t, rec, oauthRedirectCookieName(oauthStateCookieName))
	redirectValue, err := url.QueryUnescape(redirectCookie.Value)
	if err != nil {
		t.Fatalf("decode redirect cookie: %v", err)
	}
	if redirectValue != "/movies/123" {
		t.Fatalf("expected redirect cookie /movies/123, got %q", redirectValue)
	}
}

func TestFortyTwoLoginStoresOAuthRedirectCookieFromHref(t *testing.T) {
	provider := &fakeOAuthProvider{authURL: "https://api.intra.42.fr/oauth/authorize"}
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t), WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login?href=/movies/123", nil)
	rec := httptest.NewRecorder()

	handler.LoginFortyTwo(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	redirectCookie := findCookie(t, rec, oauthRedirectCookieName(oauthStateCookieName))
	redirectValue, err := url.QueryUnescape(redirectCookie.Value)
	if err != nil {
		t.Fatalf("decode redirect cookie: %v", err)
	}
	if redirectValue != "/movies/123" {
		t.Fatalf("expected redirect cookie /movies/123, got %q", redirectValue)
	}
}

func TestOAuthRedirectQueryTakesPrecedenceOverHref(t *testing.T) {
	provider := &fakeOAuthProvider{authURL: "https://api.intra.42.fr/oauth/authorize"}
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t), WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login?redirect=/movies/123&href=/users/42", nil)
	rec := httptest.NewRecorder()

	handler.LoginFortyTwo(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	redirectCookie := findCookie(t, rec, oauthRedirectCookieName(oauthStateCookieName))
	redirectValue, err := url.QueryUnescape(redirectCookie.Value)
	if err != nil {
		t.Fatalf("decode redirect cookie: %v", err)
	}
	if redirectValue != "/movies/123" {
		t.Fatalf("expected redirect cookie /movies/123, got %q", redirectValue)
	}
}

func TestFortyTwoLoginIgnoresUnsafeOAuthRedirects(t *testing.T) {
	unsafeRedirects := []string{
		"https://evil.example",
		"//evil.example",
		"javascript:alert(1)",
		"movies/123",
		`/\evil.example`,
	}

	for _, redirectPath := range unsafeRedirects {
		t.Run(redirectPath, func(t *testing.T) {
			provider := &fakeOAuthProvider{authURL: "https://api.intra.42.fr/oauth/authorize"}
			handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t), WithFortyTwoOAuth(provider))
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login?redirect="+url.QueryEscape(redirectPath), nil)
			rec := httptest.NewRecorder()

			handler.LoginFortyTwo(rec, req)

			if rec.Code != http.StatusFound {
				t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
			}
			if optionalCookie(rec, oauthRedirectCookieName(oauthStateCookieName)) != nil {
				t.Fatalf("expected no redirect cookie for unsafe redirect %q", redirectPath)
			}
			if cookie := findCookie(t, rec, oauthStateCookieName); cookie.Value != provider.lastState {
				t.Fatalf("expected state cookie %q, got %q", provider.lastState, cookie.Value)
			}
		})
	}
}

func TestUnsafeOAuthRedirectDoesNotFallBackToHref(t *testing.T) {
	provider := &fakeOAuthProvider{authURL: "https://api.intra.42.fr/oauth/authorize"}
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t), WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login?redirect="+url.QueryEscape("https://evil.example")+"&href="+url.QueryEscape("/movies/123"), nil)
	rec := httptest.NewRecorder()

	handler.LoginFortyTwo(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	if optionalCookie(rec, oauthRedirectCookieName(oauthStateCookieName)) != nil {
		t.Fatal("expected no redirect cookie when redirect is unsafe, even with safe href")
	}
	if cookie := findCookie(t, rec, oauthStateCookieName); cookie.Value != provider.lastState {
		t.Fatalf("expected state cookie %q, got %q", provider.lastState, cookie.Value)
	}
}

func TestOAuthLoginStoresProviderRedirectCookies(t *testing.T) {
	tests := []struct {
		name            string
		stateCookieName string
		login           func(http.ResponseWriter, *http.Request)
	}{
		{
			name:            "github",
			stateCookieName: githubOAuthStateCookieName,
			login: NewHandler(
				newMemoryUserStore(),
				newTestTokenManager(t),
				WithGitHubOAuth(&fakeOAuthProvider{authURL: "https://github.com/login/oauth/authorize"}),
			).LoginGitHub,
		},
		{
			name:            "gitlab",
			stateCookieName: gitlabOAuthStateCookieName,
			login: NewHandler(
				newMemoryUserStore(),
				newTestTokenManager(t),
				WithGitLabOAuth(&fakeOAuthProvider{authURL: "https://gitlab.com/oauth/authorize"}),
			).LoginGitLab,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/"+tt.name+"/login?redirect=/movies/123", nil)
			rec := httptest.NewRecorder()

			tt.login(rec, req)

			if rec.Code != http.StatusFound {
				t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
			}
			redirectCookie := findCookie(t, rec, oauthRedirectCookieName(tt.stateCookieName))
			redirectValue, err := url.QueryUnescape(redirectCookie.Value)
			if err != nil {
				t.Fatalf("decode redirect cookie: %v", err)
			}
			if redirectValue != "/movies/123" {
				t.Fatalf("expected redirect cookie /movies/123, got %q", redirectValue)
			}
		})
	}
}

func TestFortyTwoCallbackCreatesUserAndToken(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{
		identity: OAuthIdentity{
			Provider:       fortyTwoProvider,
			ProviderUserID: "12345",
			Email:          "ft.user@example.com",
			Username:       "ft_user",
			FirstName:      "Forty",
			LastName:       "Two",
			ProfilePicture: "https://cdn.intra.42.fr/users/12345/medium_ft_user.jpg",
		},
	}
	handler := NewHandler(store, tokens, WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	assertNoStoreHeaders(t, rec)
	assertAuthUserFieldsAbsent(t, rec, "watch_history")

	response := decodeAuthEnvelope(t, rec)
	if response.Data.User.Email != "ft.user@example.com" {
		t.Fatalf("expected 42 email, got %q", response.Data.User.Email)
	}
	if response.Data.User.Username != "ft_user" {
		t.Fatalf("expected 42 login as username, got %q", response.Data.User.Username)
	}
	if response.Data.User.OAuthMethod == nil || *response.Data.User.OAuthMethod != "42" {
		t.Fatalf("expected oauth_method 42, got %#v", response.Data.User.OAuthMethod)
	}
	if response.Data.User.ProfilePicture == nil || *response.Data.User.ProfilePicture != "https://cdn.intra.42.fr/users/12345/medium_ft_user.jpg" {
		t.Fatalf("expected 42 profile picture, got %+v", response.Data.User.ProfilePicture)
	}
	accessClaims, err := tokens.ValidateAccessToken(response.Data.AccessToken)
	if err != nil {
		t.Fatalf("42 auth token should validate: %v", err)
	}
	if accessClaims.UserID != response.Data.User.ID {
		t.Fatalf("expected access token user id %d, got %d", response.Data.User.ID, accessClaims.UserID)
	}
	if response.Data.RefreshToken == "" {
		t.Fatal("expected 42 OAuth refresh token")
	}
	refreshClaims, err := tokens.ValidateRefreshToken(response.Data.RefreshToken)
	if err != nil {
		t.Fatalf("42 refresh token should validate: %v", err)
	}
	if refreshClaims.UserID != response.Data.User.ID || refreshClaims.UserID != accessClaims.UserID {
		t.Fatalf("expected refresh token user id %d, got %d", response.Data.User.ID, refreshClaims.UserID)
	}

	refreshRec := callRefreshToken(t, handler, response.Data.RefreshToken)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("expected OAuth refresh token exchange 200, got %d: %s", refreshRec.Code, refreshRec.Body.String())
	}
	newAccessToken := decodeRefreshTokenEnvelope(t, refreshRec).Data.AccessToken
	newAccessClaims, err := tokens.ValidateAccessToken(newAccessToken)
	if err != nil {
		t.Fatalf("refreshed access token should validate: %v", err)
	}
	if newAccessClaims.UserID != response.Data.User.ID {
		t.Fatalf("expected refreshed access token user id %d, got %d", response.Data.User.ID, newAccessClaims.UserID)
	}

	storedUser := store.usersByID[response.Data.User.ID]
	storedUser.ProfilePicture = ""
	store.usersByID[storedUser.ID] = storedUser
	store.usersByUsername[storedUser.Username] = storedUser
	provider.identity.ProfilePicture = "https://cdn.intra.42.fr/users/12345/provider-refresh.jpg"

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=second-state", nil)
	secondReq.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "second-state"})
	secondRec := httptest.NewRecorder()

	handler.CallbackFortyTwo(secondRec, secondReq)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected second callback 200, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
	secondResponse := decodeAuthEnvelope(t, secondRec)
	if secondResponse.Data.User.ID != response.Data.User.ID {
		t.Fatalf("expected repeat 42 login to reuse user id %d, got %d", response.Data.User.ID, secondResponse.Data.User.ID)
	}
	if secondResponse.Data.User.ProfilePicture != nil {
		t.Fatalf("expected repeat 42 login to preserve removed profile picture, got %+v", secondResponse.Data.User.ProfilePicture)
	}
}

func TestFortyTwoCallbackRedirectsToFrontendWithTokenFragment(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{
		identity: OAuthIdentity{
			Provider:       fortyTwoProvider,
			ProviderUserID: "12345",
			Email:          "ft.user@example.com",
			Username:       "ft_user",
			FirstName:      "Forty",
			LastName:       "Two",
		},
	}
	handler := NewHandler(
		store,
		tokens,
		WithFortyTwoOAuth(provider),
		WithFrontendAuthCallbackURL("http://frontend.local/auth/callback"),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rec.Code, rec.Body.String())
	}
	assertNoStoreHeaders(t, rec)

	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if location.Scheme != "http" || location.Host != "frontend.local" || location.Path != "/auth/callback" {
		t.Fatalf("unexpected redirect location: %q", location.String())
	}

	fragment, err := url.ParseQuery(location.Fragment)
	if err != nil {
		t.Fatalf("parse redirect fragment: %v", err)
	}
	accessClaims, err := tokens.ValidateAccessToken(fragment.Get("access_token"))
	if err != nil {
		t.Fatalf("redirect access token should validate: %v", err)
	}
	refreshToken := fragment.Get("refresh_token")
	if refreshToken == "" {
		t.Fatal("expected refresh_token fragment")
	}
	refreshClaims, err := tokens.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("redirect refresh token should validate: %v", err)
	}
	if fragment.Get("token_type") != "Bearer" {
		t.Fatalf("expected Bearer token type, got %q", fragment.Get("token_type"))
	}

	var user userResponse
	if err := json.Unmarshal([]byte(fragment.Get("user")), &user); err != nil {
		t.Fatalf("decode user fragment: %v", err)
	}
	if user.Email != "ft.user@example.com" || user.Username != "ft_user" {
		t.Fatalf("unexpected redirected user: %+v", user)
	}
	if user.ID <= 0 {
		t.Fatalf("expected positive redirected user id, got %d", user.ID)
	}
	var userFields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(fragment.Get("user")), &userFields); err != nil {
		t.Fatalf("decode user fragment fields: %v", err)
	}
	if _, ok := userFields["watch_history"]; ok {
		t.Fatalf("redirected user must not include watch_history: %s", fragment.Get("user"))
	}
	if accessClaims.UserID != user.ID || refreshClaims.UserID != user.ID {
		t.Fatalf("expected token user ids to match redirected user id %d, got access=%d refresh=%d", user.ID, accessClaims.UserID, refreshClaims.UserID)
	}

	cookie := findCookie(t, rec, oauthStateCookieName)
	if cookie.MaxAge >= 0 {
		t.Fatalf("expected OAuth state cookie to be cleared, got MaxAge=%d", cookie.MaxAge)
	}
}

func TestFortyTwoCallbackRedirectsToFrontendWithRequestedRedirect(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{
		identity: OAuthIdentity{
			Provider:       fortyTwoProvider,
			ProviderUserID: "12345",
			Email:          "ft.user@example.com",
			Username:       "ft_user",
			FirstName:      "Forty",
			LastName:       "Two",
		},
	}
	handler := NewHandler(
		store,
		tokens,
		WithFortyTwoOAuth(provider),
		WithFrontendAuthCallbackURL("http://frontend.local/auth/callback"),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
	req.AddCookie(oauthRedirectCookie(req, oauthStateCookieName, "/movies/123", int(oauthStateTTL.Seconds())))
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rec.Code, rec.Body.String())
	}
	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if location.Scheme != "http" || location.Host != "frontend.local" || location.Path != "/auth/callback" {
		t.Fatalf("unexpected redirect location: %q", location.String())
	}
	fragment, err := url.ParseQuery(location.Fragment)
	if err != nil {
		t.Fatalf("parse redirect fragment: %v", err)
	}
	if _, err := tokens.ValidateAccessToken(fragment.Get("access_token")); err != nil {
		t.Fatalf("redirect access token should validate: %v", err)
	}
	if fragment.Get("token_type") != "Bearer" {
		t.Fatalf("expected Bearer token type, got %q", fragment.Get("token_type"))
	}
	if fragment.Get("expires_in") == "" {
		t.Fatal("expected expires_in fragment")
	}
	if fragment.Get("user") == "" {
		t.Fatal("expected user fragment")
	}
	if got := fragment.Get("redirect"); got != "/movies/123" {
		t.Fatalf("expected redirect fragment /movies/123, got %q", got)
	}
	if cookie := findCookie(t, rec, oauthRedirectCookieName(oauthStateCookieName)); cookie.MaxAge >= 0 {
		t.Fatalf("expected OAuth redirect cookie to be cleared, got MaxAge=%d", cookie.MaxAge)
	}
}

func TestFortyTwoCallbackPreservesOAuthRedirectQueryString(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{
		identity: OAuthIdentity{
			Provider:       fortyTwoProvider,
			ProviderUserID: "12345",
			Email:          "ft.user@example.com",
			Username:       "ft_user",
			FirstName:      "Forty",
			LastName:       "Two",
		},
	}
	handler := NewHandler(
		store,
		tokens,
		WithFortyTwoOAuth(provider),
		WithFrontendAuthCallbackURL("http://frontend.local/auth/callback"),
	)

	redirectPath := "/movies/123?tab=comments&page=2"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
	req.AddCookie(oauthRedirectCookie(req, oauthStateCookieName, redirectPath, int(oauthStateTTL.Seconds())))
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rec.Code, rec.Body.String())
	}
	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	fragment, err := url.ParseQuery(location.Fragment)
	if err != nil {
		t.Fatalf("parse redirect fragment: %v", err)
	}
	if got := fragment.Get("redirect"); got != redirectPath {
		t.Fatalf("expected redirect fragment %q, got %q", redirectPath, got)
	}
}

func TestFortyTwoCallbackRedirectsProviderErrorToFrontend(t *testing.T) {
	handler := NewHandler(
		newMemoryUserStore(),
		newTestTokenManager(t),
		WithFortyTwoOAuth(&fakeOAuthProvider{}),
		WithFrontendAuthCallbackURL("http://frontend.local/auth/callback"),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?error=access_denied&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
	req.AddCookie(oauthRedirectCookie(req, oauthStateCookieName, "/movies/123", int(oauthStateTTL.Seconds())))
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rec.Code, rec.Body.String())
	}
	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if got := location.Query().Get("error"); got != "OAUTH_DENIED" {
		t.Fatalf("expected OAUTH_DENIED, got %q", got)
	}
	if got := location.Query().Get("error_description"); got != "OAuth authorization was denied for 42" {
		t.Fatalf("expected localized provider error description, got %q", got)
	}
	if got := location.Query().Get("redirect"); got != "" {
		t.Fatalf("expected provider error redirect URL without redirect query, got %q", got)
	}

	cookie := findCookie(t, rec, oauthStateCookieName)
	if cookie.MaxAge >= 0 {
		t.Fatalf("expected OAuth state cookie to be cleared, got MaxAge=%d", cookie.MaxAge)
	}
	redirectCookie := findCookie(t, rec, oauthRedirectCookieName(oauthStateCookieName))
	if redirectCookie.MaxAge >= 0 {
		t.Fatalf("expected OAuth redirect cookie to be cleared, got MaxAge=%d", redirectCookie.MaxAge)
	}
}

func TestFortyTwoCallbackErrorUsesStoredOAuthLocale(t *testing.T) {
	handler := NewHandler(
		newMemoryUserStore(),
		newTestTokenManager(t),
		WithFortyTwoOAuth(&fakeOAuthProvider{}),
		WithFrontendAuthCallbackURL("http://frontend.local/auth/callback"),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=bad-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "good-state"})
	req.AddCookie(&http.Cookie{Name: oauthLocaleCookieName(oauthStateCookieName), Value: "fr"})
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rec.Code, rec.Body.String())
	}
	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if got := location.Query().Get("error"); got != "INVALID_OAUTH_STATE" {
		t.Fatalf("expected INVALID_OAUTH_STATE, got %q", got)
	}
	if got := location.Query().Get("error_description"); got != "État OAuth invalide" {
		t.Fatalf("expected French OAuth error description, got %q", got)
	}
}

func TestFortyTwoCallbackRejectsInvalidState(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{identity: OAuthIdentity{Provider: fortyTwoProvider, ProviderUserID: "1", Username: "ft"}}
	handler := NewHandler(store, tokens, WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=bad-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "good-state"})
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestFortyTwoCallbackRejectsInvalidStateClearsOAuthRedirectCookie(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{identity: OAuthIdentity{Provider: fortyTwoProvider, ProviderUserID: "1", Username: "ft"}}
	handler := NewHandler(store, tokens, WithFortyTwoOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=bad-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "good-state"})
	req.AddCookie(oauthRedirectCookie(req, oauthStateCookieName, "/movies/123", int(oauthStateTTL.Seconds())))
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "INVALID_OAUTH_STATE" {
		t.Fatalf("expected INVALID_OAUTH_STATE, got %q", got)
	}
	redirectCookie := findCookie(t, rec, oauthRedirectCookieName(oauthStateCookieName))
	if redirectCookie.MaxAge >= 0 {
		t.Fatalf("expected OAuth redirect cookie to be cleared, got MaxAge=%d", redirectCookie.MaxAge)
	}
}

func TestFortyTwoCallbackRejectsExchangeError(t *testing.T) {
	handler := NewHandler(
		newMemoryUserStore(),
		newTestTokenManager(t),
		WithFortyTwoOAuth(&fakeOAuthProvider{exchangeErr: errors.New("provider down")}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
	rec := httptest.NewRecorder()

	handler.CallbackFortyTwo(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "OAUTH_EXCHANGE_FAILED" {
		t.Fatalf("expected OAUTH_EXCHANGE_FAILED, got %q", got)
	}
}

func TestFortyTwoLoginRequiresConfiguredProvider(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/login", nil)
	rec := httptest.NewRecorder()

	handler.LoginFortyTwo(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "OAUTH_NOT_CONFIGURED" {
		t.Fatalf("expected OAUTH_NOT_CONFIGURED, got %q", got)
	}
}

func TestGitHubLoginRedirectsWithStateCookie(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{authURL: "https://github.com/login/oauth/authorize"}
	handler := NewHandler(store, tokens, WithGitHubOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/login", nil)
	rec := httptest.NewRecorder()

	handler.LoginGitHub(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	if provider.lastState == "" {
		t.Fatal("expected generated OAuth state")
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "state="+provider.lastState) {
		t.Fatalf("expected redirect location to include state, got %q", location)
	}

	cookie := rec.Result().Cookies()[0]
	if cookie.Name != githubOAuthStateCookieName {
		t.Fatalf("expected state cookie %q, got %q", githubOAuthStateCookieName, cookie.Name)
	}
	if cookie.Value != provider.lastState {
		t.Fatalf("expected cookie state %q, got %q", provider.lastState, cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Fatal("state cookie must be HttpOnly")
	}
}

func TestGitHubCallbackCreatesUserAndToken(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{
		identity: OAuthIdentity{
			Provider:       githubProvider,
			ProviderUserID: "98765",
			Email:          "gh.user@example.com",
			Username:       "gh_user",
			FirstName:      "Git",
			LastName:       "Hub",
		},
	}
	handler := NewHandler(store, tokens, WithGitHubOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/callback?code=valid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: githubOAuthStateCookieName, Value: "test-state"})
	rec := httptest.NewRecorder()

	handler.CallbackGitHub(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	response := decodeAuthEnvelope(t, rec)
	if response.Data.User.Email != "gh.user@example.com" {
		t.Fatalf("expected GitHub email, got %q", response.Data.User.Email)
	}
	if response.Data.User.Username != "gh_user" {
		t.Fatalf("expected GitHub login as username, got %q", response.Data.User.Username)
	}
	accessClaims, err := tokens.ValidateAccessToken(response.Data.AccessToken)
	if err != nil {
		t.Fatalf("GitHub auth token should validate: %v", err)
	}
	if response.Data.RefreshToken == "" {
		t.Fatal("expected GitHub OAuth refresh token")
	}
	refreshClaims, err := tokens.ValidateRefreshToken(response.Data.RefreshToken)
	if err != nil {
		t.Fatalf("GitHub refresh token should validate: %v", err)
	}
	if accessClaims.UserID != response.Data.User.ID || refreshClaims.UserID != response.Data.User.ID {
		t.Fatalf("expected token user ids to match user id %d, got access=%d refresh=%d", response.Data.User.ID, accessClaims.UserID, refreshClaims.UserID)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/callback?code=valid-code&state=second-state", nil)
	secondReq.AddCookie(&http.Cookie{Name: githubOAuthStateCookieName, Value: "second-state"})
	secondRec := httptest.NewRecorder()

	handler.CallbackGitHub(secondRec, secondReq)
	secondResponse := decodeAuthEnvelope(t, secondRec)
	if secondResponse.Data.User.ID != response.Data.User.ID {
		t.Fatalf("expected repeat GitHub login to reuse user id %d, got %d", response.Data.User.ID, secondResponse.Data.User.ID)
	}
}

func TestGitHubCallbackRejectsInvalidState(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{identity: OAuthIdentity{Provider: githubProvider, ProviderUserID: "1", Username: "gh"}}
	handler := NewHandler(store, tokens, WithGitHubOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/callback?code=valid-code&state=bad-state", nil)
	req.AddCookie(&http.Cookie{Name: githubOAuthStateCookieName, Value: "good-state"})
	rec := httptest.NewRecorder()

	handler.CallbackGitHub(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGitHubLoginRequiresConfiguredProvider(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/login", nil)
	rec := httptest.NewRecorder()

	handler.LoginGitHub(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "OAUTH_NOT_CONFIGURED" {
		t.Fatalf("expected OAUTH_NOT_CONFIGURED, got %q", got)
	}
}

func TestGitLabLoginRedirectsWithStateCookie(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{authURL: "https://gitlab.com/oauth/authorize"}
	handler := NewHandler(store, tokens, WithGitLabOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/gitlab/login", nil)
	rec := httptest.NewRecorder()

	handler.LoginGitLab(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", rec.Code, rec.Body.String())
	}
	if provider.lastState == "" {
		t.Fatal("expected generated OAuth state")
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "state="+provider.lastState) {
		t.Fatalf("expected redirect location to include state, got %q", location)
	}

	cookie := rec.Result().Cookies()[0]
	if cookie.Name != gitlabOAuthStateCookieName {
		t.Fatalf("expected state cookie %q, got %q", gitlabOAuthStateCookieName, cookie.Name)
	}
	if cookie.Value != provider.lastState {
		t.Fatalf("expected cookie state %q, got %q", provider.lastState, cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Fatal("state cookie must be HttpOnly")
	}
}

func TestGitLabCallbackCreatesUserAndToken(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{
		identity: OAuthIdentity{
			Provider:       gitlabProvider,
			ProviderUserID: "13579",
			Email:          "gl.user@example.com",
			Username:       "gl_user",
			FirstName:      "Git",
			LastName:       "Lab",
			ProfilePicture: "https://gitlab.com/uploads/-/system/user/avatar/13579/avatar.png",
		},
	}
	handler := NewHandler(store, tokens, WithGitLabOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/gitlab/callback?code=valid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: gitlabOAuthStateCookieName, Value: "test-state"})
	rec := httptest.NewRecorder()

	handler.CallbackGitLab(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	response := decodeAuthEnvelope(t, rec)
	if response.Data.User.Email != "gl.user@example.com" {
		t.Fatalf("expected GitLab email, got %q", response.Data.User.Email)
	}
	if response.Data.User.Username != "gl_user" {
		t.Fatalf("expected GitLab username, got %q", response.Data.User.Username)
	}
	if response.Data.User.ProfilePicture == nil || *response.Data.User.ProfilePicture != "https://gitlab.com/uploads/-/system/user/avatar/13579/avatar.png" {
		t.Fatalf("expected GitLab profile picture, got %+v", response.Data.User.ProfilePicture)
	}
	accessClaims, err := tokens.ValidateAccessToken(response.Data.AccessToken)
	if err != nil {
		t.Fatalf("GitLab auth token should validate: %v", err)
	}
	if response.Data.RefreshToken == "" {
		t.Fatal("expected GitLab OAuth refresh token")
	}
	refreshClaims, err := tokens.ValidateRefreshToken(response.Data.RefreshToken)
	if err != nil {
		t.Fatalf("GitLab refresh token should validate: %v", err)
	}
	if accessClaims.UserID != response.Data.User.ID || refreshClaims.UserID != response.Data.User.ID {
		t.Fatalf("expected token user ids to match user id %d, got access=%d refresh=%d", response.Data.User.ID, accessClaims.UserID, refreshClaims.UserID)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/gitlab/callback?code=valid-code&state=second-state", nil)
	secondReq.AddCookie(&http.Cookie{Name: gitlabOAuthStateCookieName, Value: "second-state"})
	secondRec := httptest.NewRecorder()

	handler.CallbackGitLab(secondRec, secondReq)
	secondResponse := decodeAuthEnvelope(t, secondRec)
	if secondResponse.Data.User.ID != response.Data.User.ID {
		t.Fatalf("expected repeat GitLab login to reuse user id %d, got %d", response.Data.User.ID, secondResponse.Data.User.ID)
	}
}

func TestGitLabCallbackRejectsInvalidState(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	provider := &fakeOAuthProvider{identity: OAuthIdentity{Provider: gitlabProvider, ProviderUserID: "1", Username: "gl"}}
	handler := NewHandler(store, tokens, WithGitLabOAuth(provider))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/gitlab/callback?code=valid-code&state=bad-state", nil)
	req.AddCookie(&http.Cookie{Name: gitlabOAuthStateCookieName, Value: "good-state"})
	rec := httptest.NewRecorder()

	handler.CallbackGitLab(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGitLabLoginRequiresConfiguredProvider(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/gitlab/login", nil)
	rec := httptest.NewRecorder()

	handler.LoginGitLab(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "OAUTH_NOT_CONFIGURED" {
		t.Fatalf("expected OAUTH_NOT_CONFIGURED, got %q", got)
	}
}
