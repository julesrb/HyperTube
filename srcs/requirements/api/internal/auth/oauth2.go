package auth

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"hypertube/api/internal/i18n"
)

type oauthTokenRequest struct {
	GrantType    string `json:"grant_type"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func (h *Handler) OAuthToken(w http.ResponseWriter, r *http.Request) {
	locale := i18n.FromRequest(r)
	req, ok := decodeOAuthTokenRequest(w, r, locale)
	if !ok {
		return
	}

	grantType := strings.TrimSpace(req.GrantType)
	if grantType == "" && strings.TrimSpace(req.ClientID) != "" && req.ClientSecret != "" && req.Username == "" && req.Password == "" {
		grantType = "client_credentials"
	}

	switch grantType {
	case "client_credentials":
		h.oauthClientCredentialsGrant(w, r, req, locale)
	case "":
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", i18n.T(locale, i18n.MsgGrantTypeRequired))
	default:
		writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", i18n.T(locale, i18n.MsgUnsupportedGrantType))
	}
}

func (h *Handler) oauthClientCredentialsGrant(w http.ResponseWriter, r *http.Request, req oauthTokenRequest, locale i18n.Locale) {
	if h.store == nil || h.tokens == nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", i18n.T(locale, i18n.MsgAuthServiceUnavailable))
		return
	}

	clientID := strings.TrimSpace(req.ClientID)
	clientSecret := req.ClientSecret
	if clientID == "" || clientSecret == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", i18n.T(locale, i18n.MsgClientCredentialsRequired))
		return
	}

	client, err := h.store.FindOAuthClientByClientID(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, ErrOAuthApplicationNotFound) {
			writeOAuthError(w, http.StatusUnauthorized, "invalid_client", i18n.T(locale, i18n.MsgInvalidClientCredentials))
			return
		}
		writeOAuthError(w, http.StatusInternalServerError, "server_error", i18n.T(locale, i18n.MsgAuthServiceUnavailable))
		return
	}

	if !CheckPassword(client.ClientSecretHash, clientSecret) {
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", i18n.T(locale, i18n.MsgInvalidClientCredentials))
		return
	}

	token, expiresIn, err := h.issueAccessToken(client.OwnerUserID)
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", i18n.T(locale, i18n.MsgFailedCreateAccessToken))
		return
	}

	response := oauthTokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}
	writeOAuthJSON(w, http.StatusOK, response)
}

func decodeOAuthTokenRequest(w http.ResponseWriter, r *http.Request, locale i18n.Locale) (oauthTokenRequest, bool) {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil && r.Header.Get("Content-Type") != "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", i18n.T(locale, i18n.MsgInvalidContentType))
		return oauthTokenRequest{}, false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	if contentType == "" || contentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", i18n.T(locale, i18n.MsgInvalidFormBody))
			return oauthTokenRequest{}, false
		}
		if _, ok := r.PostForm["scope"]; ok {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", i18n.T(locale, i18n.MsgInvalidFormBody))
			return oauthTokenRequest{}, false
		}
		return oauthTokenRequest{
			GrantType:    r.PostForm.Get("grant_type"),
			Username:     r.PostForm.Get("username"),
			Password:     r.PostForm.Get("password"),
			ClientID:     r.PostForm.Get("client_id"),
			ClientSecret: r.PostForm.Get("client_secret"),
		}, true
	}

	if contentType == "application/json" {
		var req oauthTokenRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", i18n.T(locale, i18n.MsgInvalidJSONBody))
			return oauthTokenRequest{}, false
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", i18n.T(locale, i18n.MsgInvalidJSONBody))
			return oauthTokenRequest{}, false
		}
		return req, true
	}

	writeOAuthError(w, http.StatusUnsupportedMediaType, "invalid_request", i18n.T(locale, i18n.MsgRequestBodyFormOrJSON))
	return oauthTokenRequest{}, false
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
