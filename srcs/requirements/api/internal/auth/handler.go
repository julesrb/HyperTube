package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/models"
	"hypertube/api/internal/requestjson"
	"hypertube/api/internal/respond"
)

const maxJSONBodyBytes = 1 << 20

type Handler struct {
	store                   userStore
	tokens                  *TokenManager
	fortyTwo                oauthProvider
	github                  oauthProvider
	gitlab                  oauthProvider
	frontendAuthCallbackURL string
	passwordResetMailer     passwordResetMailer
	passwordResetURL        string
	passwordResetTTL        time.Duration
}

type HandlerOption func(*Handler)

func WithFortyTwoOAuth(provider oauthProvider) HandlerOption {
	return func(h *Handler) {
		h.fortyTwo = provider
	}
}

func WithGitHubOAuth(provider oauthProvider) HandlerOption {
	return func(h *Handler) {
		h.github = provider
	}
}

func WithGitLabOAuth(provider oauthProvider) HandlerOption {
	return func(h *Handler) {
		h.gitlab = provider
	}
}

func WithFrontendAuthCallbackURL(callbackURL string) HandlerOption {
	return func(h *Handler) {
		h.frontendAuthCallbackURL = callbackURL
	}
}

func WithPasswordResetMailer(mailer passwordResetMailer) HandlerOption {
	return func(h *Handler) {
		h.passwordResetMailer = mailer
	}
}

func WithPasswordResetURL(resetURL string) HandlerOption {
	return func(h *Handler) {
		h.passwordResetURL = resetURL
	}
}

func WithPasswordResetTTL(ttl time.Duration) HandlerOption {
	return func(h *Handler) {
		if ttl > 0 {
			h.passwordResetTTL = ttl
		}
	}
}

func NewHandler(store userStore, tokens *TokenManager, opts ...HandlerOption) *Handler {
	handler := &Handler{store: store, tokens: tokens, passwordResetTTL: 30 * time.Minute}
	for _, opt := range opts {
		opt(handler)
	}
	return handler
}

type registerRequest struct {
	Email             string `json:"email"`
	Username          string `json:"username"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	FrontendFirstName string `json:"firstname"`
	FrontendLastName  string `json:"lastname"`
	Password          string `json:"password"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         userResponse `json:"user"`
}

type userResponse struct {
	ID             int64   `json:"id"`
	Email          string  `json:"email"`
	Username       string  `json:"username"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	ProfilePicture *string `json:"profile_picture"`
	OAuthMethod    *string `json:"oauth_method"`
	CreatedAt      string  `json:"created_at"`
	JoinedAt       int64   `json:"joined_at"`
	Color          string  `json:"color"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, decodingFields, ok := decodeRegisterRequest(w, r)
	if !ok {
		if len(decodingFields) > 0 {
			writeValidationError(w, r, decodingFields)
		}
		return
	}

	params, validationFields, ok := validateRegisterRequest(req)
	if !ok {
		writeValidationError(w, r, validationFields)
		return
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "password", i18n.MsgPasswordInvalid)
		return
	}
	params.PasswordHash = passwordHash

	user, err := h.store.CreateUser(r.Context(), params)
	if err != nil {
		if errors.Is(err, ErrDuplicateUser) {
			writeDuplicateRegisterError(w, r, duplicateUserFields(err))
			return
		}
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedCreateUser)
		return
	}

	h.writeAuthResponse(w, r, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, decodingFields, ok := decodeLoginRequest(w, r)
	if !ok {
		if len(decodingFields) > 0 {
			writeValidationError(w, r, decodingFields)
		}
		return
	}

	login, validationFields, ok := validateLoginRequest(req)
	if !ok {
		writeValidationError(w, r, validationFields)
		return
	}

	user, err := h.authenticatePassword(r.Context(), login, req.Password)
	if err != nil {
		if errors.Is(err, errInvalidCredentials) {
			respond.LocalizedError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", i18n.MsgInvalidCredentials)
			return
		}
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadUser)
		return
	}

	h.writeLoginResponse(w, r, http.StatusOK, user)
}

func (h *Handler) writeAuthResponse(w http.ResponseWriter, r *http.Request, status int, user models.User) {
	response, err := h.newAuthResponse(user, nil)
	if err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedCreateToken)
		return
	}

	respond.Data(w, status, response)
}

func (h *Handler) writeLoginResponse(w http.ResponseWriter, r *http.Request, status int, user models.User) {
	setTokenResponseHeaders(w)

	response, err := h.newLoginResponse(user)
	if err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedCreateToken)
		return
	}

	respond.Data(w, status, response)
}

func toUserResponse(user models.User, oauthMethod *string) userResponse {
	var profilePicture *string
	if user.ProfilePicture != "" {
		profilePicture = &user.ProfilePicture
	}
	color := user.Color
	if !models.IsValidUserColor(color) {
		color = models.UserColorPurple
	}
	return userResponse{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		ProfilePicture: profilePicture,
		OAuthMethod:    oauthMethod,
		CreatedAt:      user.CreatedAt.Format(time.RFC3339),
		JoinedAt:       user.CreatedAt.UnixMilli(),
		Color:          color,
	}
}

func writeValidationError(w http.ResponseWriter, r *http.Request, fields validationErrors) {
	locale := i18n.FromRequest(r)
	responseFields := make(respond.FieldErrors, len(fields))
	for field, message := range fields {
		responseFields[field] = respond.FieldError{Message: i18n.T(locale, message)}
	}
	respond.ValidationError(w, http.StatusBadRequest, responseFields)
}

func writeDuplicateRegisterError(w http.ResponseWriter, r *http.Request, fields []string) {
	locale := i18n.FromRequest(r)
	responseFields := respond.FieldErrors{}
	if len(fields) == 0 {
		fields = []string{"email", "username"}
	}

	for _, field := range fields {
		switch field {
		case "email":
			responseFields[field] = respond.FieldError{Message: i18n.T(locale, i18n.MsgEmailAlreadyInUse)}
		case "username":
			responseFields[field] = respond.FieldError{Message: i18n.T(locale, i18n.MsgUsernameAlreadyInUse)}
		}
	}
	if len(responseFields) == 0 {
		responseFields["email"] = respond.FieldError{Message: i18n.T(locale, i18n.MsgEmailAlreadyInUse)}
		responseFields["username"] = respond.FieldError{Message: i18n.T(locale, i18n.MsgUsernameAlreadyInUse)}
	}

	respond.ErrorWithFields(w, http.StatusConflict, "ALREADY_EXIST_ERROR", responseFields)
}

func decodeRegisterRequest(w http.ResponseWriter, r *http.Request) (registerRequest, validationErrors, bool) {
	body, ok := requestjson.DecodeJSONObject(w, r, map[string]struct{}{
		"email":      {},
		"username":   {},
		"first_name": {},
		"last_name":  {},
		"firstname":  {},
		"lastname":   {},
		"password":   {},
	})
	if !ok {
		return registerRequest{}, nil, false
	}

	fields := validationErrors{}
	req := registerRequest{
		Email:             decodeStringField(body, "email", fields),
		Username:          decodeStringField(body, "username", fields),
		FirstName:         decodeStringField(body, "first_name", fields),
		LastName:          decodeStringField(body, "last_name", fields),
		FrontendFirstName: decodeStringField(body, "firstname", fields),
		FrontendLastName:  decodeStringField(body, "lastname", fields),
		Password:          decodeStringField(body, "password", fields),
	}
	if len(fields) > 0 {
		return req, fields, false
	}
	return req, nil, true
}

func decodeLoginRequest(w http.ResponseWriter, r *http.Request) (loginRequest, validationErrors, bool) {
	body, ok := requestjson.DecodeJSONObject(w, r, map[string]struct{}{
		"login":    {},
		"password": {},
	})
	if !ok {
		return loginRequest{}, nil, false
	}

	fields := validationErrors{}
	req := loginRequest{
		Login:    decodeStringField(body, "login", fields),
		Password: decodeStringField(body, "password", fields),
	}
	if len(fields) > 0 {
		return req, fields, false
	}
	return req, nil, true
}

func decodeStringField(body map[string]json.RawMessage, field string, fields validationErrors) string {
	raw, ok := body[field]
	if !ok {
		return ""
	}

	value, ok := requestjson.DecodeString(raw)
	if !ok {
		fields[field] = i18n.MsgInvalidRequestBody
		return ""
	}
	return value
}
