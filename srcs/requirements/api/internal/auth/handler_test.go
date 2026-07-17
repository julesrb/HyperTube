package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hypertube/api/internal/models"
)

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
	if registerResponse.Data.User.FirstName != "Alice" || registerResponse.Data.User.LastName != "Example" {
		t.Fatalf("expected canonical names, got %q %q", registerResponse.Data.User.FirstName, registerResponse.Data.User.LastName)
	}
	if registerResponse.Data.User.OAuthMethod != nil {
		t.Fatalf("expected register oauth_method to be nil, got %q", *registerResponse.Data.User.OAuthMethod)
	}
	assertNoFrontendNameAliasFields(t, registerRec)
	assertAuthUserFieldsAbsent(t, registerRec, "watch_history")
	if registerResponse.Data.User.JoinedAt == 0 {
		t.Fatal("expected joined_at compatibility field")
	}
	if !models.IsValidUserColor(registerResponse.Data.User.Color) {
		t.Fatal("expected frontend profile color")
	}
	if registerResponse.Data.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if _, err := tokens.ValidateAccessToken(registerResponse.Data.AccessToken); err != nil {
		t.Fatalf("register token should validate: %v", err)
	}
	assertAuthDataFieldAbsent(t, registerRec, "refresh_token")

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

	loginBody := `{"login": "alice@example.com", "password": "correct-horse-battery"}`
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
	if loginResponse.Data.User.Color != registerResponse.Data.User.Color {
		t.Fatalf("expected login color %q, got %q", registerResponse.Data.User.Color, loginResponse.Data.User.Color)
	}
	if loginResponse.Data.User.OAuthMethod != nil {
		t.Fatalf("expected login oauth_method to be nil, got %q", *loginResponse.Data.User.OAuthMethod)
	}
	assertAuthUserFieldsAbsent(t, loginRec, "watch_history")
	if _, err := tokens.ValidateAccessToken(loginResponse.Data.AccessToken); err != nil {
		t.Fatalf("login token should validate: %v", err)
	}
	if loginResponse.Data.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	refreshClaims, err := tokens.ValidateRefreshToken(loginResponse.Data.RefreshToken)
	if err != nil {
		t.Fatalf("login refresh token should validate: %v", err)
	}
	if refreshClaims.UserID != loginResponse.Data.User.ID {
		t.Fatalf("expected refresh token user id %d, got %d", loginResponse.Data.User.ID, refreshClaims.UserID)
	}
	if got := loginRec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected Cache-Control no-store, got %q", got)
	}
	if got := loginRec.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("expected Pragma no-cache, got %q", got)
	}

	usernameLoginBody := `{"login": "alice_1", "password": "correct-horse-battery"}`
	usernameLoginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(usernameLoginBody))
	usernameLoginRec := httptest.NewRecorder()

	handler.Login(usernameLoginRec, usernameLoginReq)

	if usernameLoginRec.Code != http.StatusOK {
		t.Fatalf("expected username login 200, got %d: %s", usernameLoginRec.Code, usernameLoginRec.Body.String())
	}

	usernameLoginResponse := decodeAuthEnvelope(t, usernameLoginRec)
	if usernameLoginResponse.Data.User.ID != registerResponse.Data.User.ID {
		t.Fatalf("expected username login user id %d, got %d", registerResponse.Data.User.ID, usernameLoginResponse.Data.User.ID)
	}
}

func TestLoginReturnsPersistedUserColor(t *testing.T) {
	store := newMemoryUserStore()
	passwordHash, err := HashPassword("right-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, err := store.CreateUser(context.Background(), CreateUserParams{
		Email:        "blue@example.com",
		Username:     "blue_user",
		FirstName:    "Blue",
		LastName:     "User",
		PasswordHash: passwordHash,
		Color:        models.UserColorBlue,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	handler := NewHandler(store, newTestTokenManager(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"login":"blue@example.com","password":"right-password"}`))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	response := decodeAuthEnvelope(t, rec)
	if response.Data.User.Color != models.UserColorBlue {
		t.Fatalf("expected stored color blue, got %q", response.Data.User.Color)
	}
}

func TestRegisterAcceptsFrontendNameAliases(t *testing.T) {
	store := newMemoryUserStore()
	tokens := newTestTokenManager(t)
	handler := NewHandler(store, tokens)

	registerBody := `{
		"email": "Frontend@example.com",
		"username": "frontend_1",
		"firstname": "Front",
		"lastname": "End",
		"password": "correct-horse-battery"
	}`
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(registerBody))
	registerRec := httptest.NewRecorder()

	handler.Register(registerRec, registerReq)

	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", registerRec.Code, registerRec.Body.String())
	}

	registerResponse := decodeAuthEnvelope(t, registerRec)
	if registerResponse.Data.User.FirstName != "Front" || registerResponse.Data.User.LastName != "End" {
		t.Fatalf("expected canonical names from frontend aliases, got %q %q", registerResponse.Data.User.FirstName, registerResponse.Data.User.LastName)
	}
	assertNoFrontendNameAliasFields(t, registerRec)

	loginBody := `{"login": "frontend_1", "password": "correct-horse-battery"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(loginBody))
	loginRec := httptest.NewRecorder()

	handler.Login(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected username login 200, got %d: %s", loginRec.Code, loginRec.Body.String())
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

func TestRegisterValidationErrorReturnsFieldErrors(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
	body := `{"email":"not-an-email","username":"ab","first_name":"","last_name":"","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	errorBody := decodeErrorEnvelope(t, rec).Error
	if errorBody.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", errorBody.Code)
	}
	if errorBody.Message != "" {
		t.Fatalf("expected validation response without top-level message, got %q", errorBody.Message)
	}

	wantFields := map[string]string{
		"email":      "Invalid email",
		"username":   "Username is too short",
		"first_name": "First name is required",
		"last_name":  "Last name is required",
		"password":   "Password is too short",
	}
	for field, wantMessage := range wantFields {
		got, ok := errorBody.Fields[field]
		if !ok {
			t.Fatalf("expected field %q in validation response, got %+v", field, errorBody.Fields)
		}
		if got.Message != wantMessage {
			t.Fatalf("expected %s message %q, got %q", field, wantMessage, got.Message)
		}
	}
}

func TestRegisterMissingRequiredFieldsReturnFieldErrors(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		field       string
		wantMessage string
	}{
		{
			name:        "missing email",
			body:        `{"username":"alice_1","first_name":"Alice","last_name":"Example","password":"correct-horse-battery"}`,
			field:       "email",
			wantMessage: "Email is required",
		},
		{
			name:        "missing username",
			body:        `{"email":"alice@example.com","first_name":"Alice","last_name":"Example","password":"correct-horse-battery"}`,
			field:       "username",
			wantMessage: "Username is required",
		},
		{
			name:        "missing first_name",
			body:        `{"email":"alice@example.com","username":"alice_1","last_name":"Example","password":"correct-horse-battery"}`,
			field:       "first_name",
			wantMessage: "First name is required",
		},
		{
			name:        "missing last_name",
			body:        `{"email":"alice@example.com","username":"alice_1","first_name":"Alice","password":"correct-horse-battery"}`,
			field:       "last_name",
			wantMessage: "Last name is required",
		},
		{
			name:        "missing password",
			body:        `{"email":"alice@example.com","username":"alice_1","first_name":"Alice","last_name":"Example"}`,
			field:       "password",
			wantMessage: "Password is required",
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

			errorBody := decodeErrorEnvelope(t, rec).Error
			if errorBody.Code != "VALIDATION_ERROR" {
				t.Fatalf("expected VALIDATION_ERROR, got %q: %s", errorBody.Code, rec.Body.String())
			}
			if errorBody.Message != "" {
				t.Fatalf("expected validation response without top-level message, got %q", errorBody.Message)
			}
			got, ok := errorBody.Fields[tt.field]
			if !ok {
				t.Fatalf("expected field %q in validation response, got %+v", tt.field, errorBody.Fields)
			}
			if got.Message != tt.wantMessage {
				t.Fatalf("expected %s message %q, got %q", tt.field, tt.wantMessage, got.Message)
			}
		})
	}

	t.Run("missing all required fields", func(t *testing.T) {
		handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		handler.Register(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}

		errorBody := decodeErrorEnvelope(t, rec).Error
		if errorBody.Code != "VALIDATION_ERROR" {
			t.Fatalf("expected VALIDATION_ERROR, got %q: %s", errorBody.Code, rec.Body.String())
		}
		if errorBody.Message != "" {
			t.Fatalf("expected validation response without top-level message, got %q", errorBody.Message)
		}
		for field, wantMessage := range map[string]string{
			"email":      "Email is required",
			"username":   "Username is required",
			"first_name": "First name is required",
			"last_name":  "Last name is required",
			"password":   "Password is required",
		} {
			got, ok := errorBody.Fields[field]
			if !ok {
				t.Fatalf("expected field %q in validation response, got %+v", field, errorBody.Fields)
			}
			if got.Message != wantMessage {
				t.Fatalf("expected %s message %q, got %q", field, wantMessage, got.Message)
			}
		}
	})
}

func TestRegisterInvalidFieldTypesReturnFieldErrors(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
	body := `{"email":6,"username":["test",5,false],"first_name":"Alice","last_name":false,"password":"correct-horse-battery"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	errorBody := decodeErrorEnvelope(t, rec).Error
	if errorBody.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", errorBody.Code)
	}
	for _, field := range []string{"email", "username", "last_name"} {
		if _, ok := errorBody.Fields[field]; !ok {
			t.Fatalf("expected field %q in validation response, got %+v", field, errorBody.Fields)
		}
	}
}

func TestLoginValidationErrorReturnsFieldErrors(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"login":"not-an-email!","password":""}`))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	errorBody := decodeErrorEnvelope(t, rec).Error
	if errorBody.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", errorBody.Code)
	}
	if _, ok := errorBody.Fields["email"]; ok {
		t.Fatalf("did not expect email field validation, got %+v", errorBody.Fields["email"])
	}
	if got := errorBody.Fields["login"].Message; got != "Invalid email or username" {
		t.Fatalf("expected login validation message, got %q", got)
	}
	if got := errorBody.Fields["password"].Message; got != "Password is required" {
		t.Fatalf("expected password validation message, got %q", got)
	}
}

func TestLoginInvalidFieldTypesReturnFieldErrors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantField string
	}{
		{
			name:      "login type",
			body:      `{"login":454,"password":"password_testydoyrE"}`,
			wantField: "login",
		},
		{
			name:      "password type",
			body:      `{"login":"alice@example.com","password":false}`,
			wantField: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			errorBody := decodeErrorEnvelope(t, rec).Error
			if errorBody.Code != "VALIDATION_ERROR" {
				t.Fatalf("expected VALIDATION_ERROR, got %q", errorBody.Code)
			}
			if _, ok := errorBody.Fields[tt.wantField]; !ok {
				t.Fatalf("expected field %q in validation response, got %+v", tt.wantField, errorBody.Fields)
			}
			if _, ok := errorBody.Fields["email"]; ok {
				t.Fatalf("did not expect email validation field, got %+v", errorBody.Fields)
			}
		})
	}
}

func TestLoginRejectsEmailField(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"alice@example.com","password":"correct-horse-battery"}`))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeErrorEnvelope(t, rec).Error.Code; got != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %q", got)
	}
}

func TestRegisterValidationErrorUsesAcceptLanguage(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
	body := `{"email":"not-an-email","username":"ab","first_name":"","last_name":"","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Accept-Language", "fr")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	errorBody := decodeErrorEnvelope(t, rec).Error
	if got := errorBody.Fields["email"].Message; got != "Email invalide" {
		t.Fatalf("expected French email message, got %q", got)
	}
	if got := errorBody.Fields["password"].Message; got != "Mot de passe trop court" {
		t.Fatalf("expected French password message, got %q", got)
	}
}

func TestLoginInvalidCredentialsUsesAcceptLanguage(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"login":"missing@example.com","password":"right-password"}`))
	req.Header.Set("Accept-Language", "de")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	errorBody := decodeErrorEnvelope(t, rec).Error
	if errorBody.Code != "INVALID_CREDENTIALS" {
		t.Fatalf("expected INVALID_CREDENTIALS, got %q", errorBody.Code)
	}
	if got := errorBody.Message; got != "E-Mail, Benutzername oder Passwort ist ungültig" {
		t.Fatalf("expected German invalid credentials message, got %q", got)
	}
}

func TestRegisterDuplicateUserReturnsFieldConflict(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	initialBody := `{
		"email": "alice@example.com",
		"username": "alice_1",
		"first_name": "Alice",
		"last_name": "Example",
		"password": "correct-horse-battery"
	}`

	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(initialBody))
	firstRec := httptest.NewRecorder()
	handler.Register(firstRec, firstReq)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected initial register 201, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	tests := []struct {
		name       string
		body       string
		wantFields map[string]string
	}{
		{
			name: "email",
			body: `{
				"email": "alice@example.com",
				"username": "alice_2",
				"first_name": "Alice",
				"last_name": "Example",
				"password": "correct-horse-battery"
			}`,
			wantFields: map[string]string{
				"email": "Email is already in use",
			},
		},
		{
			name: "username",
			body: `{
				"email": "alice2@example.com",
				"username": "alice_1",
				"first_name": "Alice",
				"last_name": "Example",
				"password": "correct-horse-battery"
			}`,
			wantFields: map[string]string{
				"username": "Username is already in use",
			},
		},
		{
			name: "email and username",
			body: initialBody,
			wantFields: map[string]string{
				"email":    "Email is already in use",
				"username": "Username is already in use",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			handler.Register(rec, req)

			if rec.Code != http.StatusConflict {
				t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
			}

			errorBody := decodeErrorEnvelope(t, rec).Error
			if errorBody.Code != "ALREADY_EXIST_ERROR" {
				t.Fatalf("expected ALREADY_EXIST_ERROR, got %q", errorBody.Code)
			}
			if errorBody.Message != "" {
				t.Fatalf("expected duplicate response without top-level message, got %q", errorBody.Message)
			}
			if len(errorBody.Fields) != len(tt.wantFields) {
				t.Fatalf("expected fields %+v, got %+v", tt.wantFields, errorBody.Fields)
			}
			for field, wantMessage := range tt.wantFields {
				got, ok := errorBody.Fields[field]
				if !ok {
					t.Fatalf("expected field %q in duplicate response, got %+v", field, errorBody.Fields)
				}
				if got.Message != wantMessage {
					t.Fatalf("expected %s message %q, got %q", field, wantMessage, got.Message)
				}
			}
		})
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
			body: `{"login":"missing@example.com","password":"right-password"}`,
		},
		{
			name: "wrong password",
			body: `{"login":"alice@example.com","password":"wrong-password"}`,
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
