package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/models"
	"hypertube/api/internal/requestjson"
	"hypertube/api/internal/respond"

	"github.com/go-chi/chi/v5"
)

const (
	oauthApplicationNameMaxLength        = 100
	oauthApplicationRedirectURIMaxLength = 2048
	oauthApplicationPageLimit            = 12
	oauthCredentialGenerationAttempts    = 3
)

type oauthApplicationCreateRequest struct {
	Name        string
	RedirectURI string
}

type oauthApplicationPatchRequest struct {
	Name        *string
	RedirectURI *string
}

type oauthApplicationResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	RedirectURI string `json:"redirect_uri"`
	ClientID    string `json:"client_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type oauthApplicationCreatedResponse struct {
	oauthApplicationResponse
	ClientSecret string `json:"client_secret"`
}

func (h *Handler) CreateOAuthApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}
	if h.store == nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}

	req, decodingFields, ok := decodeOAuthApplicationCreateRequest(w, r)
	if !ok {
		if len(decodingFields) > 0 {
			writeValidationError(w, r, decodingFields)
		}
		return
	}
	params, validationFields, ok := validateOAuthApplicationCreateRequest(userID, req)
	if !ok {
		writeValidationError(w, r, validationFields)
		return
	}

	app, clientSecret, err := h.createOAuthApplicationWithCredentials(r, params)
	if err != nil {
		if errors.Is(err, errOAuthCredentialGeneration) {
			respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedCreateOAuthCredentials)
			return
		}
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedCreateOAuthApplication)
		return
	}

	respond.Data(w, http.StatusCreated, oauthApplicationCreatedResponse{
		oauthApplicationResponse: toOAuthApplicationResponse(app),
		ClientSecret:             clientSecret,
	})
}

func (h *Handler) ListOAuthApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}
	if h.store == nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}

	page := parsePage(r)
	offset := page * oauthApplicationPageLimit

	total, err := h.store.CountOAuthApplications(r.Context(), userID)
	if err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadOAuthApplications)
		return
	}

	apps, err := h.store.ListOAuthApplications(r.Context(), userID, oauthApplicationPageLimit, offset)
	if err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadOAuthApplications)
		return
	}

	response := make([]oauthApplicationResponse, 0, len(apps))
	for _, app := range apps {
		response = append(response, toOAuthApplicationResponse(app))
	}
	respond.ListPaginated(w, http.StatusOK, response, total, page, oauthApplicationPageLimit)
}

func (h *Handler) UpdateOAuthApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}
	if h.store == nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}

	id, ok := oauthApplicationIDFromRequest(w, r)
	if !ok {
		return
	}

	req, decodingFields, ok := decodeOAuthApplicationPatchRequest(w, r)
	if !ok {
		if len(decodingFields) > 0 {
			writeValidationError(w, r, decodingFields)
		}
		return
	}
	params, validationFields, ok := validateOAuthApplicationPatchRequest(req)
	if !ok {
		writeValidationError(w, r, validationFields)
		return
	}

	app, err := h.store.UpdateOAuthApplication(r.Context(), id, userID, params)
	if err != nil {
		if errors.Is(err, ErrOAuthApplicationNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgOAuthApplicationNotFound)
			return
		}
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedUpdateOAuthApplication)
		return
	}

	respond.Item(w, http.StatusOK, toOAuthApplicationResponse(app))
}

func (h *Handler) DeleteOAuthApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}
	if h.store == nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}

	id, ok := oauthApplicationIDFromRequest(w, r)
	if !ok {
		return
	}

	if err := h.store.DeleteOAuthApplication(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrOAuthApplicationNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgOAuthApplicationNotFound)
			return
		}
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedDeleteOAuthApplication)
		return
	}

	respond.Item(w, http.StatusOK, nil)
}

var errOAuthCredentialGeneration = errors.New("failed to create oauth credentials")

func (h *Handler) createOAuthApplicationWithCredentials(r *http.Request, params CreateOAuthApplicationParams) (models.OAuthApplication, string, error) {
	for attempt := 0; attempt < oauthCredentialGenerationAttempts; attempt++ {
		clientID, err := generateOAuthClientID()
		if err != nil {
			return models.OAuthApplication{}, "", errOAuthCredentialGeneration
		}
		clientSecret, err := generateOAuthClientSecret()
		if err != nil {
			return models.OAuthApplication{}, "", errOAuthCredentialGeneration
		}
		clientSecretHash, err := HashPassword(clientSecret)
		if err != nil {
			return models.OAuthApplication{}, "", errOAuthCredentialGeneration
		}

		params.ClientID = clientID
		params.ClientSecretHash = clientSecretHash
		app, err := h.store.CreateOAuthApplication(r.Context(), params)
		if err == nil {
			return app, clientSecret, nil
		}
		if errors.Is(err, ErrDuplicateOAuthClientID) {
			continue
		}
		return models.OAuthApplication{}, "", err
	}
	return models.OAuthApplication{}, "", errOAuthCredentialGeneration
}

func decodeOAuthApplicationCreateRequest(w http.ResponseWriter, r *http.Request) (oauthApplicationCreateRequest, validationErrors, bool) {
	body, ok := requestjson.DecodeJSONObject(w, r, map[string]struct{}{
		"name":         {},
		"redirect_uri": {},
	})
	if !ok {
		return oauthApplicationCreateRequest{}, nil, false
	}

	fields := validationErrors{}
	req := oauthApplicationCreateRequest{
		Name:        decodeStringField(body, "name", fields),
		RedirectURI: decodeStringField(body, "redirect_uri", fields),
	}
	if len(fields) > 0 {
		return req, fields, false
	}
	return req, nil, true
}

func decodeOAuthApplicationPatchRequest(w http.ResponseWriter, r *http.Request) (oauthApplicationPatchRequest, validationErrors, bool) {
	body, ok := requestjson.DecodeJSONObject(w, r, map[string]struct{}{
		"name":         {},
		"redirect_uri": {},
	})
	if !ok {
		return oauthApplicationPatchRequest{}, nil, false
	}

	fields := validationErrors{}
	var req oauthApplicationPatchRequest
	if raw, ok := body["name"]; ok {
		value, ok := requestjson.DecodeString(raw)
		if !ok {
			fields["name"] = i18n.MsgInvalidRequestBody
		} else {
			req.Name = &value
		}
	}
	if raw, ok := body["redirect_uri"]; ok {
		value, ok := requestjson.DecodeString(raw)
		if !ok {
			fields["redirect_uri"] = i18n.MsgInvalidRequestBody
		} else {
			req.RedirectURI = &value
		}
	}
	if len(fields) > 0 {
		return req, fields, false
	}
	return req, nil, true
}

func validateOAuthApplicationCreateRequest(ownerUserID int64, req oauthApplicationCreateRequest) (CreateOAuthApplicationParams, validationErrors, bool) {
	fields := validationErrors{}
	name, ok := validateOAuthApplicationName(req.Name, fields)
	redirectURI, okRedirectURI := validateOAuthApplicationRedirectURI(req.RedirectURI, fields)
	if !ok || !okRedirectURI {
		return CreateOAuthApplicationParams{}, fields, false
	}

	return CreateOAuthApplicationParams{
		OwnerUserID: ownerUserID,
		Name:        name,
		RedirectURI: redirectURI,
	}, nil, true
}

func validateOAuthApplicationPatchRequest(req oauthApplicationPatchRequest) (UpdateOAuthApplicationParams, validationErrors, bool) {
	fields := validationErrors{}
	params := UpdateOAuthApplicationParams{}
	if req.Name == nil && req.RedirectURI == nil {
		fields["body"] = i18n.MsgInvalidRequestBody
		return UpdateOAuthApplicationParams{}, fields, false
	}

	if req.Name != nil {
		name, ok := validateOAuthApplicationName(*req.Name, fields)
		if ok {
			params.Name = &name
		}
	}
	if req.RedirectURI != nil {
		redirectURI, ok := validateOAuthApplicationRedirectURI(*req.RedirectURI, fields)
		if ok {
			params.RedirectURI = &redirectURI
		}
	}
	if len(fields) > 0 {
		return UpdateOAuthApplicationParams{}, fields, false
	}
	return params, nil, true
}

func validateOAuthApplicationName(name string, fields validationErrors) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		fields["name"] = i18n.MsgOAuthApplicationNameRequired
		return "", false
	}
	if len(name) > oauthApplicationNameMaxLength {
		fields["name"] = i18n.MsgOAuthApplicationNameTooLong
		return "", false
	}
	return name, true
}

func validateOAuthApplicationRedirectURI(raw string, fields validationErrors) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		fields["redirect_uri"] = i18n.MsgOAuthApplicationRedirectURIRequired
		return "", false
	}
	if len(value) > oauthApplicationRedirectURIMaxLength {
		fields["redirect_uri"] = i18n.MsgOAuthApplicationRedirectURITooLong
		return "", false
	}
	if strings.Contains(value, "#") {
		fields["redirect_uri"] = i18n.MsgOAuthApplicationRedirectURIInvalid
		return "", false
	}

	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil {
		fields["redirect_uri"] = i18n.MsgOAuthApplicationRedirectURIInvalid
		return "", false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		fields["redirect_uri"] = i18n.MsgOAuthApplicationRedirectURIInvalid
		return "", false
	}
	return value, true
}

func oauthApplicationIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || id <= 0 {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgOAuthApplicationNotFound)
		return 0, false
	}
	return id, true
}

func parsePage(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		return 0
	}
	return page
}

func toOAuthApplicationResponse(app models.OAuthApplication) oauthApplicationResponse {
	return oauthApplicationResponse{
		ID:          app.ID,
		Name:        app.Name,
		RedirectURI: app.RedirectURI,
		ClientID:    app.ClientID,
		CreatedAt:   app.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   app.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func generateOAuthClientID() (string, error) {
	return randomURLToken("htc_", 24)
}

func generateOAuthClientSecret() (string, error) {
	return randomURLToken("hts_", 32)
}

func randomURLToken(prefix string, byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}
