package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"hypertube/api/internal/models"

	"github.com/go-chi/chi/v5"
)

const testOAuthApplicationRedirectURI = "http://localhost:4200/auth/callback"

type oauthApplicationsListResponse struct {
	Data []oauthApplicationListItem `json:"data"`
	Meta struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"meta"`
}

type oauthApplicationListItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	RedirectURI string `json:"redirect_uri"`
	ClientID    string `json:"client_id"`
}

func TestOAuthApplicationCredentialGeneration(t *testing.T) {
	clientID1, err := generateOAuthClientID()
	if err != nil {
		t.Fatalf("generate client id: %v", err)
	}
	clientID2, err := generateOAuthClientID()
	if err != nil {
		t.Fatalf("generate second client id: %v", err)
	}
	if !strings.HasPrefix(clientID1, "htc_") || clientID1 == "htc_" {
		t.Fatalf("unexpected client id %q", clientID1)
	}
	if clientID1 == clientID2 {
		t.Fatal("expected generated client ids to differ")
	}

	secret1, err := generateOAuthClientSecret()
	if err != nil {
		t.Fatalf("generate client secret: %v", err)
	}
	secret2, err := generateOAuthClientSecret()
	if err != nil {
		t.Fatalf("generate second client secret: %v", err)
	}
	if !strings.HasPrefix(secret1, "hts_") || secret1 == "hts_" {
		t.Fatalf("unexpected client secret %q", secret1)
	}
	if secret1 == secret2 {
		t.Fatal("expected generated client secrets to differ")
	}

	hash, err := HashPassword(secret1)
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}
	if hash == secret1 {
		t.Fatal("secret hash must not equal plaintext secret")
	}
	if !CheckPassword(hash, secret1) {
		t.Fatal("expected hashed secret to verify")
	}
}

func TestCreateOAuthApplicationRequiresAuthContext(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/applications", strings.NewReader(`{"name":"My App","redirect_uri":"http://localhost:4200/auth/callback"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateOAuthApplication(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOAuthApplicationValidation(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantStatus    int
		wantField     string
		wantErrorCode string
	}{
		{name: "invalid JSON", body: `{"name":`, wantStatus: http.StatusBadRequest, wantErrorCode: "BAD_REQUEST"},
		{name: "missing name", body: `{"redirect_uri":"http://localhost:4200/auth/callback"}`, wantStatus: http.StatusBadRequest, wantField: "name"},
		{name: "blank name", body: `{"name":"   ","redirect_uri":"http://localhost:4200/auth/callback"}`, wantStatus: http.StatusBadRequest, wantField: "name"},
		{name: "unknown scope", body: `{"name":"My App","scope":"read:movies","redirect_uri":"http://localhost:4200/auth/callback"}`, wantStatus: http.StatusBadRequest, wantErrorCode: "BAD_REQUEST"},
		{name: "missing redirect_uri", body: `{"name":"My App"}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "blank redirect_uri", body: `{"name":"My App","redirect_uri":"   "}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "null redirect_uri", body: `{"name":"My App","redirect_uri":null}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "non-string redirect_uri", body: `{"name":"My App","redirect_uri":42}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "relative redirect_uri", body: `{"name":"My App","redirect_uri":"/auth/callback"}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "unsupported redirect_uri scheme", body: `{"name":"My App","redirect_uri":"javascript:alert(1)"}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "redirect_uri missing host", body: `{"name":"My App","redirect_uri":"https:///callback"}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "redirect_uri userinfo", body: `{"name":"My App","redirect_uri":"https://user:pass@example.com/callback"}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "redirect_uri non-empty fragment", body: `{"name":"My App","redirect_uri":"https://example.com/callback#token"}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "redirect_uri empty fragment delimiter", body: `{"name":"My App","redirect_uri":"https://example.com/callback#"}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
		{name: "redirect_uri too long", body: `{"name":"My App","redirect_uri":"https://example.com/` + strings.Repeat("a", oauthApplicationRedirectURIMaxLength) + `"}`, wantStatus: http.StatusBadRequest, wantField: "redirect_uri"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))

			rec := callOAuthApplicationRoute(t, handler, 42, http.MethodPost, "/oauth/applications", tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if tt.wantField != "" {
				assertValidationErrorField(t, rec, tt.wantField)
			}
			if tt.wantErrorCode != "" {
				assertErrorCode(t, rec, tt.wantErrorCode)
			}
		})
	}
}

func TestCreateOAuthApplicationReturnsOneTimeSecret(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))

	rec := callOAuthApplicationRoute(t, handler, 42, http.MethodPost, "/oauth/applications", `{
		"name": "  My App  ",
		"redirect_uri": "  http://localhost:4200/auth/callback?next=/movies  "
	}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			ID           int64  `json:"id"`
			Name         string `json:"name"`
			RedirectURI  string `json:"redirect_uri"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			CreatedAt    string `json:"created_at"`
			UpdatedAt    string `json:"updated_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.ID == 0 || body.Data.Name != "My App" {
		t.Fatalf("unexpected app response: %+v", body.Data)
	}
	if body.Data.RedirectURI != "http://localhost:4200/auth/callback?next=/movies" {
		t.Fatalf("expected trimmed redirect URI, got %q", body.Data.RedirectURI)
	}
	if !strings.HasPrefix(body.Data.ClientID, "htc_") {
		t.Fatalf("expected generated client id, got %q", body.Data.ClientID)
	}
	if !strings.HasPrefix(body.Data.ClientSecret, "hts_") {
		t.Fatalf("expected generated client secret, got %q", body.Data.ClientSecret)
	}
	if body.Data.CreatedAt == "" || body.Data.UpdatedAt == "" {
		t.Fatalf("expected timestamps, got %+v", body.Data)
	}

	assertJSONFieldsAbsent(t, rec.Body.Bytes(), "client_secret_hash", "owner_id", "scope")

	client, err := store.FindOAuthClientByClientID(context.Background(), body.Data.ClientID)
	if err != nil {
		t.Fatalf("find created client: %v", err)
	}
	if client.ClientSecretHash == body.Data.ClientSecret {
		t.Fatal("store must not keep plaintext client secret")
	}
	if !CheckPassword(client.ClientSecretHash, body.Data.ClientSecret) {
		t.Fatal("stored client secret hash should verify one-time secret")
	}
}

func TestListOAuthApplicationsFiltersByAuthenticatedUser(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	createStoredOAuthApplication(t, store, 42, "First", "client-one", "secret-one")
	createStoredOAuthApplication(t, store, 99, "Other", "client-other", "secret-other")
	createStoredOAuthApplication(t, store, 42, "Second", "client-two", "secret-two")

	rec := callOAuthApplicationRoute(t, handler, 42, http.MethodGet, "/oauth/applications", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeOAuthApplicationsListResponse(t, rec)
	if len(body.Data) != 2 || body.Meta.Total != 2 {
		t.Fatalf("expected two owned apps, got %+v", body)
	}
	if body.Meta.Page != 0 {
		t.Fatalf("expected page 0, got %d", body.Meta.Page)
	}
	if body.Meta.PerPage != oauthApplicationPageLimit {
		t.Fatalf("expected per_page %d, got %d", oauthApplicationPageLimit, body.Meta.PerPage)
	}
	for _, app := range body.Data {
		if app.Name == "Other" || app.ClientID == "client-other" {
			t.Fatalf("listed app owned by another user: %+v", app)
		}
		if app.RedirectURI != testOAuthApplicationRedirectURI {
			t.Fatalf("expected listed redirect URI, got %+v", app)
		}
	}
	assertJSONFieldsAbsent(t, rec.Body.Bytes(), "client_secret", "client_secret_hash", "owner_id", "scope")
}

func TestListOAuthApplicationsReturnsEmptyList(t *testing.T) {
	handler := NewHandler(newMemoryUserStore(), newTestTokenManager(t))

	rec := callOAuthApplicationRoute(t, handler, 42, http.MethodGet, "/oauth/applications", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeOAuthApplicationsListResponse(t, rec)
	if body.Data == nil || len(body.Data) != 0 {
		t.Fatalf("expected empty list data array, got %+v", body.Data)
	}
	if body.Meta.Total != 0 || body.Meta.Page != 0 || body.Meta.PerPage != oauthApplicationPageLimit {
		t.Fatalf("unexpected empty list metadata: %+v", body.Meta)
	}
}

func TestListOAuthApplicationsPaginates(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	for i := 0; i < 14; i++ {
		suffix := strconv.Itoa(i)
		createStoredOAuthApplication(t, store, 42, "App "+suffix, "client-"+suffix, "secret-"+suffix)
	}
	createStoredOAuthApplication(t, store, 99, "Other", "client-other", "secret-other")

	rec := callOAuthApplicationRoute(t, handler, 42, http.MethodGet, "/oauth/applications?page=1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeOAuthApplicationsListResponse(t, rec)
	if len(body.Data) != 2 {
		t.Fatalf("expected second page to contain 2 apps, got %+v", body.Data)
	}
	if body.Meta.Total != 14 || body.Meta.Page != 1 || body.Meta.PerPage != oauthApplicationPageLimit {
		t.Fatalf("unexpected pagination metadata: %+v", body.Meta)
	}

	wantClientIDs := map[string]bool{"client-1": true, "client-0": true}
	for _, app := range body.Data {
		if app.ClientID == "client-other" || app.Name == "Other" {
			t.Fatalf("listed app owned by another user: %+v", app)
		}
		if !wantClientIDs[app.ClientID] {
			t.Fatalf("expected page 1 to contain oldest owned apps, got %+v", body.Data)
		}
		delete(wantClientIDs, app.ClientID)
	}
	if len(wantClientIDs) != 0 {
		t.Fatalf("missing expected page 1 apps: %+v", wantClientIDs)
	}
	assertJSONFieldsAbsent(t, rec.Body.Bytes(), "client_secret", "client_secret_hash", "owner_id", "scope")
}

func TestListOAuthApplicationsInvalidPageFallsBackToFirstPage(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "missing page", path: "/oauth/applications"},
		{name: "empty page", path: "/oauth/applications?page="},
		{name: "text page", path: "/oauth/applications?page=abc"},
		{name: "negative page", path: "/oauth/applications?page=-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMemoryUserStore()
			handler := NewHandler(store, newTestTokenManager(t))
			for i := 0; i < 13; i++ {
				suffix := strconv.Itoa(i)
				createStoredOAuthApplication(t, store, 42, "App "+suffix, "client-"+suffix, "secret-"+suffix)
			}

			rec := callOAuthApplicationRoute(t, handler, 42, http.MethodGet, tt.path, "")

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			body := decodeOAuthApplicationsListResponse(t, rec)
			if body.Meta.Total != 13 || body.Meta.Page != 0 || body.Meta.PerPage != oauthApplicationPageLimit {
				t.Fatalf("unexpected pagination metadata: %+v", body.Meta)
			}
			if len(body.Data) != oauthApplicationPageLimit {
				t.Fatalf("expected first page to contain %d apps, got %+v", oauthApplicationPageLimit, body.Data)
			}
			if body.Data[0].ClientID != "client-12" {
				t.Fatalf("expected invalid page to return first page, got %+v", body.Data)
			}
		})
	}
}

func TestUpdateOAuthApplicationValidationAndOwnership(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	owned := createStoredOAuthApplication(t, store, 42, "Old", "owned-client", "owned-secret")
	other := createStoredOAuthApplication(t, store, 99, "Other", "other-client", "other-secret")

	rec := callOAuthApplicationRoute(t, handler, 42, http.MethodPatch, "/oauth/applications/not-a-number", `{"name":"New"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected invalid id 404, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = callOAuthApplicationRoute(t, handler, 42, http.MethodPatch, "/oauth/applications/"+strconvInt64(owned.ID), `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected empty patch 400, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = callOAuthApplicationRoute(t, handler, 42, http.MethodPatch, "/oauth/applications/"+strconvInt64(owned.ID), `{"scope":"read:movies"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected scope patch 400, got %d: %s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "BAD_REQUEST")

	rec = callOAuthApplicationRoute(t, handler, 42, http.MethodPatch, "/oauth/applications/"+strconvInt64(other.ID), `{"name":"Nope"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected foreign app 404, got %d: %s", rec.Code, rec.Body.String())
	}

	invalidRedirectURIPatches := []struct {
		name string
		body string
	}{
		{name: "blank redirect_uri", body: `{"redirect_uri":"   "}`},
		{name: "relative redirect_uri", body: `{"redirect_uri":"/auth/callback"}`},
		{name: "null redirect_uri", body: `{"redirect_uri":null}`},
		{name: "fragment redirect_uri", body: `{"redirect_uri":"https://example.com/callback#token"}`},
	}
	for _, tt := range invalidRedirectURIPatches {
		t.Run(tt.name, func(t *testing.T) {
			rec := callOAuthApplicationRoute(t, handler, 42, http.MethodPatch, "/oauth/applications/"+strconvInt64(owned.ID), tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected redirect URI patch 400, got %d: %s", rec.Code, rec.Body.String())
			}
			assertValidationErrorField(t, rec, "redirect_uri")
		})
	}

	rec = callOAuthApplicationRoute(t, handler, 42, http.MethodPatch, "/oauth/applications/"+strconvInt64(owned.ID), `{
		"redirect_uri": "https://example.com/oauth/callback"
	}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected redirect URI-only patch 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var redirectOnlyBody struct {
		Data struct {
			RedirectURI string `json:"redirect_uri"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &redirectOnlyBody); err != nil {
		t.Fatalf("decode redirect-only patch response: %v", err)
	}
	if redirectOnlyBody.Data.RedirectURI != "https://example.com/oauth/callback" {
		t.Fatalf("unexpected redirect-only patch response: %+v", redirectOnlyBody.Data)
	}

	rec = callOAuthApplicationRoute(t, handler, 42, http.MethodPatch, "/oauth/applications/"+strconvInt64(owned.ID), `{
		"name": "  New Name  ",
		"redirect_uri": "  http://localhost:4200/auth/callback?next=/profile  "
	}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected patch 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			Name        string `json:"name"`
			RedirectURI string `json:"redirect_uri"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Name != "New Name" || body.Data.RedirectURI != "http://localhost:4200/auth/callback?next=/profile" {
		t.Fatalf("unexpected patch response: %+v", body.Data)
	}
	assertJSONFieldsAbsent(t, rec.Body.Bytes(), "client_secret", "client_secret_hash", "owner_id", "scope")
}

func TestDeleteOAuthApplicationDeletesOnlyOwnedApp(t *testing.T) {
	store := newMemoryUserStore()
	handler := NewHandler(store, newTestTokenManager(t))
	owned := createStoredOAuthApplication(t, store, 42, "Owned", "owned-client", "owned-secret")
	other := createStoredOAuthApplication(t, store, 99, "Other", "other-client", "other-secret")

	rec := callOAuthApplicationRoute(t, handler, 42, http.MethodDelete, "/oauth/applications/"+strconvInt64(other.ID), "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected foreign delete 404, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = callOAuthApplicationRoute(t, handler, 42, http.MethodDelete, "/oauth/applications/"+strconvInt64(owned.ID), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected delete 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"data":null`)) {
		t.Fatalf("expected data:null, got %s", rec.Body.String())
	}

	if _, err := store.FindOAuthClientByClientID(context.Background(), "owned-client"); !errors.Is(err, ErrOAuthApplicationNotFound) {
		t.Fatalf("expected deleted app to be missing, got %v", err)
	}
	if _, err := store.FindOAuthClientByClientID(context.Background(), "other-client"); err != nil {
		t.Fatalf("foreign app should remain, got %v", err)
	}
}

func callOAuthApplicationRoute(t *testing.T, handler *Handler, userID int64, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	router := chi.NewRouter()
	router.With(DevAuthenticateAs(userID)).Post("/oauth/applications", handler.CreateOAuthApplication)
	router.With(DevAuthenticateAs(userID)).Get("/oauth/applications", handler.ListOAuthApplications)
	router.With(DevAuthenticateAs(userID)).Patch("/oauth/applications/{id}", handler.UpdateOAuthApplication)
	router.With(DevAuthenticateAs(userID)).Delete("/oauth/applications/{id}", handler.DeleteOAuthApplication)

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if method == http.MethodPost || method == http.MethodPatch {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func createStoredOAuthApplication(t *testing.T, store *memoryUserStore, ownerID int64, name string, clientID string, secret string) models.OAuthApplication {
	t.Helper()

	hash, err := HashPassword(secret)
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}
	app, err := store.CreateOAuthApplication(context.Background(), CreateOAuthApplicationParams{
		OwnerUserID:      ownerID,
		Name:             name,
		RedirectURI:      testOAuthApplicationRedirectURI,
		ClientID:         clientID,
		ClientSecretHash: hash,
	})
	if err != nil {
		t.Fatalf("create oauth application: %v", err)
	}
	return app
}

func decodeOAuthApplicationsListResponse(t *testing.T, rec *httptest.ResponseRecorder) oauthApplicationsListResponse {
	t.Helper()

	var body oauthApplicationsListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func assertJSONFieldsAbsent(t *testing.T, body []byte, fields ...string) {
	t.Helper()

	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	for _, field := range fields {
		if jsonFieldExists(value, field) {
			t.Fatalf("expected field %q to be absent in %s", field, string(body))
		}
	}
}

func assertValidationErrorField(t *testing.T, rec *httptest.ResponseRecorder, field string) {
	t.Helper()

	errorBody := decodeErrorEnvelope(t, rec).Error
	if errorBody.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q: %s", errorBody.Code, rec.Body.String())
	}
	if _, ok := errorBody.Fields[field]; !ok {
		t.Fatalf("expected validation field %q, got %+v", field, errorBody.Fields)
	}
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, code string) {
	t.Helper()

	errorBody := decodeErrorEnvelope(t, rec).Error
	if errorBody.Code != code {
		t.Fatalf("expected error code %q, got %q: %s", code, errorBody.Code, rec.Body.String())
	}
}

func jsonFieldExists(value any, field string) bool {
	switch typed := value.(type) {
	case map[string]any:
		if _, ok := typed[field]; ok {
			return true
		}
		for _, child := range typed {
			if jsonFieldExists(child, field) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if jsonFieldExists(child, field) {
				return true
			}
		}
	}
	return false
}

func strconvInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}
