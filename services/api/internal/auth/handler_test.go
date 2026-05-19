package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"hypertube/api/internal/models"
)

type memoryUserStore struct {
	nextID          int64
	usersByEmail    map[string]models.User
	usersByUsername map[string]models.User
	oauthAccounts   map[string]int64
	resetTokens     map[string]memoryPasswordResetToken
}

type memoryPasswordResetToken struct {
	userID    int64
	expiresAt time.Time
	used      bool
}

func newMemoryUserStore() *memoryUserStore {
	return &memoryUserStore{
		usersByEmail:    make(map[string]models.User),
		usersByUsername: make(map[string]models.User),
		oauthAccounts:   make(map[string]int64),
		resetTokens:     make(map[string]memoryPasswordResetToken),
	}
}

func (s *memoryUserStore) CreateUser(_ context.Context, params CreateUserParams) (models.User, error) {
	if _, ok := s.usersByEmail[params.Email]; ok {
		return models.User{}, ErrDuplicateUser
	}
	if _, ok := s.usersByUsername[params.Username]; ok {
		return models.User{}, ErrDuplicateUser
	}

	s.nextID++
	now := time.Now().UTC()
	user := models.User{
		ID:           s.nextID,
		Email:        params.Email,
		Username:     params.Username,
		FirstName:    params.FirstName,
		LastName:     params.LastName,
		PasswordHash: params.PasswordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.usersByEmail[user.Email] = user
	s.usersByUsername[user.Username] = user
	return user, nil
}

func (s *memoryUserStore) FindUserByEmail(_ context.Context, email string) (models.User, error) {
	user, ok := s.usersByEmail[email]
	if !ok {
		return models.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *memoryUserStore) FindUserByLogin(_ context.Context, login string) (models.User, error) {
	login = strings.TrimSpace(login)
	if email, ok := normalizeEmail(login); ok {
		if user, ok := s.usersByEmail[email]; ok {
			return user, nil
		}
	}
	if user, ok := s.usersByUsername[login]; ok {
		return user, nil
	}
	return models.User{}, ErrUserNotFound
}

func (s *memoryUserStore) FindOrCreateOAuthUser(_ context.Context, params OAuthUserParams) (models.User, error) {
	params = normalizeOAuthUserParams(params)
	key := oauthAccountKey(params.Provider, params.ProviderUserID)
	if userID, ok := s.oauthAccounts[key]; ok {
		return s.findUserByID(userID)
	}

	if user, ok := s.usersByEmail[params.Email]; ok {
		s.oauthAccounts[key] = user.ID
		return user, nil
	}

	username := oauthUsernameBase(params.Username, params.Provider, params.ProviderUserID)
	if _, ok := s.usersByUsername[username]; ok {
		username = usernameWithSuffix(username, "_"+params.ProviderUserID)
	}

	user, err := s.CreateUser(context.Background(), CreateUserParams{
		Email:        params.Email,
		Username:     username,
		FirstName:    params.FirstName,
		LastName:     params.LastName,
		PasswordHash: "",
	})
	if err != nil {
		return models.User{}, err
	}
	s.oauthAccounts[key] = user.ID
	return user, nil
}

func (s *memoryUserStore) findUserByID(userID int64) (models.User, error) {
	for _, user := range s.usersByEmail {
		if user.ID == userID {
			return user, nil
		}
	}
	return models.User{}, ErrUserNotFound
}

func oauthAccountKey(provider string, providerUserID string) string {
	return fmt.Sprintf("%s:%s", provider, providerUserID)
}

func (s *memoryUserStore) CreatePasswordResetToken(_ context.Context, params CreatePasswordResetTokenParams) error {
	s.resetTokens[params.TokenHash] = memoryPasswordResetToken{
		userID:    params.UserID,
		expiresAt: params.ExpiresAt,
	}
	return nil
}

func (s *memoryUserStore) ResetPasswordWithToken(_ context.Context, tokenHash string, passwordHash string) (models.User, error) {
	token, ok := s.resetTokens[tokenHash]
	if !ok || token.used || !token.expiresAt.After(time.Now().UTC()) {
		return models.User{}, ErrInvalidPasswordResetToken
	}

	user, err := s.findUserByID(token.userID)
	if err != nil {
		return models.User{}, err
	}

	token.used = true
	s.resetTokens[tokenHash] = token
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now().UTC()
	s.usersByEmail[user.Email] = user
	s.usersByUsername[user.Username] = user
	return user, nil
}

func TestRegisterAndLoginHappyPath(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	handler := NewHandler(store, tokens)

	registerBody := `{
		"email": "Alice@example.com",
		"username": "alice_1",
		"first_name": "Alice",
		"last_name": "Example",
		"password": "correct-horse-battery"
	}`
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(registerBody))
	registerRec := httptest.NewRecorder()

	handler.Register(registerRec, registerReq)

	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", registerRec.Code, registerRec.Body.String())
	}

	registerResponse := decodeAuthEnvelope(t, registerRec)
	if registerResponse.Data.User.Email != "alice@example.com" {
		t.Fatalf("expected normalized email, got %q", registerResponse.Data.User.Email)
	}
	if registerResponse.Data.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if _, err := tokens.ValidateAccessToken(registerResponse.Data.AccessToken); err != nil {
		t.Fatalf("register token should validate: %v", err)
	}

	storedUser, err := store.FindUserByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("find stored user: %v", err)
	}
	if storedUser.PasswordHash == "correct-horse-battery" {
		t.Fatal("stored password must be hashed")
	}
	if !CheckPassword(storedUser.PasswordHash, "correct-horse-battery") {
		t.Fatal("stored hash must match original password")
	}

	loginBody := `{"email": "alice@example.com", "password": "correct-horse-battery"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(loginBody))
	loginRec := httptest.NewRecorder()

	handler.Login(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", loginRec.Code, loginRec.Body.String())
	}

	loginResponse := decodeAuthEnvelope(t, loginRec)
	if loginResponse.Data.User.ID != registerResponse.Data.User.ID {
		t.Fatalf("expected login user id %d, got %d", registerResponse.Data.User.ID, loginResponse.Data.User.ID)
	}
	if _, err := tokens.ValidateAccessToken(loginResponse.Data.AccessToken); err != nil {
		t.Fatalf("login token should validate: %v", err)
	}
}

func TestRegisterRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name string
		body string
		code string
	}{
		{
			name: "malformed JSON",
			body: `{"email":`,
			code: "BAD_REQUEST",
		},
		{
			name: "unknown field",
			body: `{"email":"alice@example.com","username":"alice_1","first_name":"Alice","last_name":"Example","password":"correct-horse-battery","admin":true}`,
			code: "BAD_REQUEST",
		},
		{
			name: "multiple JSON documents",
			body: `{"email":"alice@example.com","username":"alice_1","first_name":"Alice","last_name":"Example","password":"correct-horse-battery"} {}`,
			code: "BAD_REQUEST",
		},
		{
			name: "invalid email",
			body: `{"email":"not-an-email","username":"alice_1","first_name":"Alice","last_name":"Example","password":"correct-horse-battery"}`,
			code: "VALIDATION_ERROR",
		},
		{
			name: "short password",
			body: `{"email":"alice@example.com","username":"alice_1","first_name":"Alice","last_name":"Example","password":"short"}`,
			code: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeErrorEnvelope(t, rec).Error.Code; got != tt.code {
				t.Fatalf("expected error code %q, got %q", tt.code, got)
			}
		})
	}
}

func TestRegisterDuplicateUserReturnsConflict(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	body := `{
		"email": "alice@example.com",
		"username": "alice_1",
		"first_name": "Alice",
		"last_name": "Example",
		"password": "correct-horse-battery"
	}`

	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	firstRec := httptest.NewRecorder()
	handler.Register(firstRec, firstReq)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected initial register 201, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	secondRec := httptest.NewRecorder()
	handler.Register(secondRec, secondReq)

	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
	if got := decodeErrorEnvelope(t, secondRec).Error.Code; got != "USER_EXISTS" {
		t.Fatalf("expected USER_EXISTS, got %q", got)
	}
}

func TestLoginRejectsUnknownUserAndWrongPassword(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	passwordHash, err := HashPassword("right-password")
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

	tests := []struct {
		name string
		body string
	}{
		{
			name: "unknown user",
			body: `{"email":"missing@example.com","password":"right-password"}`,
		},
		{
			name: "wrong password",
			body: `{"email":"alice@example.com","password":"wrong-password"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeErrorEnvelope(t, rec).Error.Code; got != "INVALID_CREDENTIALS" {
				t.Fatalf("expected INVALID_CREDENTIALS, got %q", got)
			}
		})
	}
}

func TestOAuthTokenPasswordGrantReturnsBearerToken(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	handler := NewHandler(store, tokens)
	passwordHash, err := HashPassword("correct-horse-battery")
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

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", "alice_1")
	form.Set("password", "correct-horse-battery")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.OAuthToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var response oauthTokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
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
	if claims.UserID != user.ID {
		t.Fatalf("expected token user id %d, got %d", user.ID, claims.UserID)
	}
}

func TestOAuthTokenPasswordGrantAcceptsEmailLogin(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	passwordHash, err := HashPassword("correct-horse-battery")
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

	body := `{"grant_type":"password","username":"Alice@Example.COM","password":"correct-horse-battery"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.OAuthToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestOAuthTokenRejectsInvalidGrant(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	passwordHash, err := HashPassword("right-password")
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
				"grant_type": {"client_credentials"},
			},
			wantError: "unsupported_grant_type",
		},
		{
			name: "wrong password",
			form: url.Values{
				"grant_type": {"password"},
				"username":   {"alice_1"},
				"password":   {"wrong-password"},
			},
			wantError: "invalid_grant",
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

func decodeAuthEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Data authResponse `json:"data"`
} {
	t.Helper()

	var body struct {
		Data authResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func decodeErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
} {
	t.Helper()

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	return body
}

func TestMemoryUserStoreFindMissing(t *testing.T) {
	_, err := newMemoryUserStore().FindUserByEmail(context.Background(), "missing@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

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

	response := decodeAuthEnvelope(t, rec)
	if response.Data.User.Email != "ft.user@example.com" {
		t.Fatalf("expected 42 email, got %q", response.Data.User.Email)
	}
	if response.Data.User.Username != "ft_user" {
		t.Fatalf("expected 42 login as username, got %q", response.Data.User.Username)
	}
	if _, err := tokens.ValidateAccessToken(response.Data.AccessToken); err != nil {
		t.Fatalf("42 auth token should validate: %v", err)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/42/callback?code=valid-code&state=second-state", nil)
	secondReq.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "second-state"})
	secondRec := httptest.NewRecorder()

	handler.CallbackFortyTwo(secondRec, secondReq)
	secondResponse := decodeAuthEnvelope(t, secondRec)
	if secondResponse.Data.User.ID != response.Data.User.ID {
		t.Fatalf("expected repeat 42 login to reuse user id %d, got %d", response.Data.User.ID, secondResponse.Data.User.ID)
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

	var user userResponse
	if err := json.Unmarshal([]byte(fragment.Get("user")), &user); err != nil {
		t.Fatalf("decode user fragment: %v", err)
	}
	if user.Email != "ft.user@example.com" || user.Username != "ft_user" {
		t.Fatalf("unexpected redirected user: %+v", user)
	}

	cookie := findCookie(t, rec, oauthStateCookieName)
	if cookie.MaxAge >= 0 {
		t.Fatalf("expected OAuth state cookie to be cleared, got MaxAge=%d", cookie.MaxAge)
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
	if got := location.Query().Get("error_description"); got != "access_denied" {
		t.Fatalf("expected provider error description, got %q", got)
	}

	cookie := findCookie(t, rec, oauthStateCookieName)
	if cookie.MaxAge >= 0 {
		t.Fatalf("expected OAuth state cookie to be cleared, got MaxAge=%d", cookie.MaxAge)
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
	if _, err := tokens.ValidateAccessToken(response.Data.AccessToken); err != nil {
		t.Fatalf("GitHub auth token should validate: %v", err)
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

func findCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("expected cookie %q", name)
	return nil
}

type fakeOAuthProvider struct {
	authURL     string
	authErr     error
	lastState   string
	identity    OAuthIdentity
	exchangeErr error
}

func (p *fakeOAuthProvider) AuthCodeURL(state string) (string, error) {
	if p.authErr != nil {
		return "", p.authErr
	}
	p.lastState = state
	return p.authURL + "?state=" + state, nil
}

func (p *fakeOAuthProvider) Exchange(context.Context, string) (OAuthIdentity, error) {
	if p.exchangeErr != nil {
		return OAuthIdentity{}, p.exchangeErr
	}
	return p.identity, nil
}
