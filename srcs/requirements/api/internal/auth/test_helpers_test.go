package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/models"
	"hypertube/api/internal/userinput"
)

const testJWTSecret = "0123456789abcdef0123456789abcdef"

type memoryUserStore struct {
	nextID          int64
	usersByEmail    map[string]models.User
	usersByID       map[int64]models.User
	usersByUsername map[string]models.User
	oauthAccounts   map[string]int64
	nextOAuthAppID  int64
	oauthAppsByID   map[int64]models.OAuthApplication
	oauthAppSecrets map[int64]string
	oauthAppClient  map[string]int64
	resetTokens     map[string]memoryPasswordResetToken
	userExistsErr   error
}

var _ UserExistenceChecker = (*memoryUserStore)(nil)

type memoryPasswordResetToken struct {
	userID    int64
	expiresAt time.Time
	used      bool
}

func newMemoryUserStore() *memoryUserStore {
	return &memoryUserStore{
		usersByEmail:    make(map[string]models.User),
		usersByID:       make(map[int64]models.User),
		usersByUsername: make(map[string]models.User),
		oauthAccounts:   make(map[string]int64),
		oauthAppsByID:   make(map[int64]models.OAuthApplication),
		oauthAppSecrets: make(map[int64]string),
		oauthAppClient:  make(map[string]int64),
		resetTokens:     make(map[string]memoryPasswordResetToken),
	}
}

func (s *memoryUserStore) UserExists(_ context.Context, userID int64) (bool, error) {
	if s.userExistsErr != nil {
		return false, s.userExistsErr
	}
	_, ok := s.usersByID[userID]
	return ok, nil
}

func (s *memoryUserStore) CreateUser(_ context.Context, params CreateUserParams) (models.User, error) {
	var duplicateFields []string
	if params.PasswordHash != "" {
		if _, ok := s.usersByEmail[params.Email]; ok {
			duplicateFields = append(duplicateFields, "email")
		}
	}
	if _, ok := s.usersByUsername[params.Username]; ok {
		duplicateFields = append(duplicateFields, "username")
	}
	if len(duplicateFields) > 0 {
		return models.User{}, duplicateUserError(duplicateFields...)
	}

	color := strings.TrimSpace(params.Color)
	if color == "" {
		color = models.RandomUserColor()
	}
	if !models.IsValidUserColor(color) {
		return models.User{}, fmt.Errorf("invalid user color: %q", params.Color)
	}

	s.nextID++
	now := time.Now().UTC()
	user := models.User{
		ID:             s.nextID,
		Email:          params.Email,
		Username:       params.Username,
		FirstName:      params.FirstName,
		LastName:       params.LastName,
		ProfilePicture: params.ProfilePicture,
		PasswordHash:   params.PasswordHash,
		Color:          color,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.usersByID[user.ID] = user
	if user.PasswordHash != "" {
		s.usersByEmail[user.Email] = user
	}
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
	if email, ok := userinput.NormalizeEmail(login); ok {
		if user, ok := s.usersByEmail[email]; ok {
			return user, nil
		}
	}
	if user, ok := s.usersByUsername[login]; ok && user.PasswordHash != "" {
		return user, nil
	}
	return models.User{}, ErrUserNotFound
}

func (s *memoryUserStore) FindOrCreateOAuthUser(_ context.Context, params OAuthUserParams) (models.User, error) {
	params = normalizeOAuthUserParams(params)
	key := oauthAccountKey(params.Provider, params.ProviderUserID)
	if userID, ok := s.oauthAccounts[key]; ok {
		user, err := s.findUserByID(userID)
		if err != nil {
			return models.User{}, err
		}
		return s.applyOAuthProfile(user, params), nil
	}

	username := oauthUsernameBase(params.Username, params.Provider, params.ProviderUserID)
	if _, ok := s.usersByUsername[username]; ok {
		username = usernameWithSuffix(username, "_"+params.ProviderUserID)
	}

	user, err := s.CreateUser(context.Background(), CreateUserParams{
		Email:          params.Email,
		Username:       username,
		FirstName:      params.FirstName,
		LastName:       params.LastName,
		ProfilePicture: params.ProfilePicture,
		PasswordHash:   "",
	})
	if err != nil {
		return models.User{}, err
	}
	s.oauthAccounts[key] = user.ID
	return user, nil
}

func (s *memoryUserStore) findUserByID(userID int64) (models.User, error) {
	if user, ok := s.usersByID[userID]; ok {
		return user, nil
	}
	return models.User{}, ErrUserNotFound
}

func (s *memoryUserStore) applyOAuthProfile(user models.User, params OAuthUserParams) models.User {
	if params.FirstName != "" {
		user.FirstName = params.FirstName
	}
	if params.LastName != "" {
		user.LastName = params.LastName
	}
	s.usersByID[user.ID] = user
	s.usersByUsername[user.Username] = user
	return user
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
	s.usersByID[user.ID] = user
	s.usersByEmail[user.Email] = user
	s.usersByUsername[user.Username] = user
	return user, nil
}

func (s *memoryUserStore) CreateOAuthApplication(_ context.Context, params CreateOAuthApplicationParams) (models.OAuthApplication, error) {
	if _, ok := s.oauthAppClient[params.ClientID]; ok {
		return models.OAuthApplication{}, ErrDuplicateOAuthClientID
	}

	s.nextOAuthAppID++
	now := time.Now().UTC()
	app := models.OAuthApplication{
		ID:          s.nextOAuthAppID,
		OwnerID:     params.OwnerUserID,
		Name:        params.Name,
		RedirectURI: params.RedirectURI,
		ClientID:    params.ClientID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.oauthAppsByID[app.ID] = app
	s.oauthAppSecrets[app.ID] = params.ClientSecretHash
	s.oauthAppClient[app.ClientID] = app.ID
	return app, nil
}

func (s *memoryUserStore) ListOAuthApplications(_ context.Context, ownerUserID int64, limit, offset int) ([]models.OAuthApplication, error) {
	apps := []models.OAuthApplication{}
	for _, app := range s.oauthAppsByID {
		if app.OwnerID == ownerUserID {
			apps = append(apps, app)
		}
	}
	sort.Slice(apps, func(i, j int) bool {
		if apps[i].CreatedAt.Equal(apps[j].CreatedAt) {
			return apps[i].ID > apps[j].ID
		}
		return apps[i].CreatedAt.After(apps[j].CreatedAt)
	})
	if offset >= len(apps) {
		return []models.OAuthApplication{}, nil
	}
	end := offset + limit
	if end > len(apps) {
		end = len(apps)
	}
	return apps[offset:end], nil
}

func (s *memoryUserStore) CountOAuthApplications(_ context.Context, ownerUserID int64) (int, error) {
	total := 0
	for _, app := range s.oauthAppsByID {
		if app.OwnerID == ownerUserID {
			total++
		}
	}
	return total, nil
}

func (s *memoryUserStore) UpdateOAuthApplication(_ context.Context, id int64, ownerUserID int64, params UpdateOAuthApplicationParams) (models.OAuthApplication, error) {
	app, ok := s.oauthAppsByID[id]
	if !ok || app.OwnerID != ownerUserID {
		return models.OAuthApplication{}, ErrOAuthApplicationNotFound
	}
	if params.Name != nil {
		app.Name = *params.Name
	}
	if params.RedirectURI != nil {
		app.RedirectURI = *params.RedirectURI
	}
	app.UpdatedAt = time.Now().UTC()
	s.oauthAppsByID[id] = app
	return app, nil
}

func (s *memoryUserStore) DeleteOAuthApplication(_ context.Context, id int64, ownerUserID int64) error {
	app, ok := s.oauthAppsByID[id]
	if !ok || app.OwnerID != ownerUserID {
		return ErrOAuthApplicationNotFound
	}
	delete(s.oauthAppsByID, id)
	delete(s.oauthAppSecrets, id)
	delete(s.oauthAppClient, app.ClientID)
	return nil
}

func (s *memoryUserStore) FindOAuthClientByClientID(_ context.Context, clientID string) (oauthClientCredentials, error) {
	id, ok := s.oauthAppClient[clientID]
	if !ok {
		return oauthClientCredentials{}, ErrOAuthApplicationNotFound
	}
	app := s.oauthAppsByID[id]
	return oauthClientCredentials{
		ID:               app.ID,
		OwnerUserID:      app.OwnerID,
		ClientID:         app.ClientID,
		ClientSecretHash: s.oauthAppSecrets[id],
	}, nil
}

func createPasswordUser(t *testing.T, store *memoryUserStore, email string, username string, password string) models.User {
	t.Helper()

	passwordHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := store.CreateUser(context.Background(), CreateUserParams{
		Email:        email,
		Username:     username,
		FirstName:    "Alice",
		LastName:     "Example",
		PasswordHash: passwordHash,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func newTestTokenManager(t *testing.T) *TokenManager {
	t.Helper()

	tokens, err := NewTokenManager(testJWTSecret, "hypertube-test")
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	return tokens
}

func decodeAuthEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Data authResponse `json:"data"`
} {
	t.Helper()

	var body struct {
		Data authResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func assertNoFrontendNameAliasFields(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	assertAuthUserFieldsAbsent(t, rec, "firstname", "lastname")
}

func assertAuthUserFieldsAbsent(t *testing.T, rec *httptest.ResponseRecorder, fields ...string) {
	t.Helper()

	var body struct {
		Data struct {
			User map[string]json.RawMessage `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response fields: %v", err)
	}
	for _, field := range fields {
		if _, ok := body.Data.User[field]; ok {
			t.Fatalf("response data.user must not include %q: %s", field, rec.Body.String())
		}
	}
}

func assertAuthDataFieldAbsent(t *testing.T, rec *httptest.ResponseRecorder, field string) {
	t.Helper()

	var body struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response fields: %v", err)
	}
	if _, ok := body.Data[field]; ok {
		t.Fatalf("response must not include %q: %s", field, rec.Body.String())
	}
}

func decodeErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Fields  map[string]struct {
			Message string `json:"message"`
		} `json:"fields"`
	} `json:"error"`
} {
	t.Helper()

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Fields  map[string]struct {
				Message string `json:"message"`
			} `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	return body
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

func optionalCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

type fakePasswordResetMailer struct {
	calls     int
	toEmail   string
	toName    string
	resetURL  string
	expiresIn time.Duration
	locale    i18n.Locale
}

func (m *fakePasswordResetMailer) SendPasswordReset(_ context.Context, toEmail string, toName string, resetURL string, expiresIn time.Duration, locale i18n.Locale) error {
	m.calls++
	m.toEmail = toEmail
	m.toName = toName
	m.resetURL = resetURL
	m.expiresIn = expiresIn
	m.locale = locale
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
