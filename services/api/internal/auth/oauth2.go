package auth

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

type oauthTokenRequest struct {
	GrantType string `json:"grant_type"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Scope     string `json:"scope"`
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func (h *Handler) OAuthToken(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeOAuthTokenRequest(w, r)
	if !ok {
		return
	}

	switch strings.TrimSpace(req.GrantType) {
	case "password":
		h.oauthPasswordGrant(w, r, req)
	case "":
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "grant_type is required")
	default:
		writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "only password grant_type is supported")
	}
}

func (h *Handler) oauthPasswordGrant(w http.ResponseWriter, r *http.Request, req oauthTokenRequest) {
	if h.store == nil || h.tokens == nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "authentication service is unavailable")
		return
	}

	login := strings.TrimSpace(req.Username)
	if login == "" || req.Password == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "username and password are required")
		return
	}
	if len(req.Password) > maxPasswordBytes {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "password is too long")
		return
	}

	user, err := h.store.FindUserByLogin(r.Context(), login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "invalid username or password")
			return
		}
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to load user")
		return
	}

	if user.PasswordHash == "" || !CheckPassword(user.PasswordHash, req.Password) {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "invalid username or password")
		return
	}

	token, _, err := h.tokens.CreateAccessToken(user.ID)
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to create access token")
		return
	}

	response := oauthTokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(AccessTokenTTL.Seconds()),
		Scope:       normalizeOAuthScope(req.Scope),
	}
	writeOAuthJSON(w, http.StatusOK, response)
}

func decodeOAuthTokenRequest(w http.ResponseWriter, r *http.Request) (oauthTokenRequest, bool) {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil && r.Header.Get("Content-Type") != "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid Content-Type")
		return oauthTokenRequest{}, false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	if contentType == "" || contentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid form body")
			return oauthTokenRequest{}, false
		}
		return oauthTokenRequest{
			GrantType: r.PostForm.Get("grant_type"),
			Username:  r.PostForm.Get("username"),
			Password:  r.PostForm.Get("password"),
			Scope:     r.PostForm.Get("scope"),
		}, true
	}

	if contentType == "application/json" {
		var req oauthTokenRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return oauthTokenRequest{}, false
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return oauthTokenRequest{}, false
		}
		return req, true
	}

	writeOAuthError(w, http.StatusUnsupportedMediaType, "invalid_request", "request body must be form encoded or JSON")
	return oauthTokenRequest{}, false
}

func normalizeOAuthScope(scope string) string {
	return strings.Join(strings.Fields(scope), " ")
}

func writeOAuthError(w http.ResponseWriter, status int, code string, description string) {
	writeOAuthJSON(w, status, oauthErrorResponse{
		Error:            code,
		ErrorDescription: description,
	})
}

func writeOAuthJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
