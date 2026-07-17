package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"hypertube/api/internal/auth"
	"hypertube/api/internal/models"
)

func TestListUsersReturnsUserSmallList(t *testing.T) {
	createdAt := time.Date(2026, time.June, 20, 12, 0, 0, 0, time.UTC)
	store := &fakeUserStore{list: []models.User{
		{ID: 1, Username: "alice", Email: "alice@example.com", FirstName: "alice", LastName: "gu", PasswordHash: "secret", Color: models.UserColorGreen, CreatedAt: createdAt},
		{ID: 2, Username: "bob", Email: "bob@example.com", FirstName: "alice", LastName: "gu", PasswordHash: "hunter2", Color: models.UserColorBlue},
	}}
	handler := NewHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []models.UserSmall `json:"data"`
		Meta struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 2 {
		t.Fatalf("expected 2 users, got %d", len(body.Data))
	}
	if body.Meta.Total != 2 {
		t.Fatalf("expected total 2, got %d", body.Meta.Total)
	}
	if body.Meta.Page != 0 {
		t.Fatalf("expected page 0, got %d", body.Meta.Page)
	}
	if body.Meta.PerPage != userPageLimit {
		t.Fatalf("expected per_page %d, got %d", userPageLimit, body.Meta.PerPage)
	}
	if store.gotLimit != userPageLimit {
		t.Fatalf("expected limit %d, got %d", userPageLimit, store.gotLimit)
	}
	if store.gotOffset != 0 {
		t.Fatalf("expected offset 0, got %d", store.gotOffset)
	}
	if body.Data[0].Username != "alice" || body.Data[1].Username != "bob" {
		t.Fatalf("unexpected users: %+v", body.Data)
	}
	if body.Data[0].Color != models.UserColorGreen {
		t.Fatalf("expected first user color green, got %q", body.Data[0].Color)
	}
	if raw := rec.Body.String(); strings.Contains(raw, "secret") || strings.Contains(raw, "alice@example.com") {
		t.Fatalf("response leaked sensitive data: %s", raw)
	}
	if strings.Contains(rec.Body.String(), "created_at") {
		t.Fatalf("list response unexpectedly contains created_at: %s", rec.Body.String())
	}
}

func TestListUsersIncludesNullProfilePicture(t *testing.T) {
	store := &fakeUserStore{list: []models.User{
		{
			ID:        1,
			Username:  "alice",
			FirstName: "Alice",
			LastName:  "Liddell",
			Color:     models.UserColorGreen,
		},
	}}
	handler := NewHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(body.Data))
	}
	assertRawJSONField(t, body.Data[0], "profile_picture", "null")
}

func TestListUsersUsesPageOneQueryForSecondPagePagination(t *testing.T) {
	store := &fakeUserStore{
		list: []models.User{
			{ID: 13, Username: "page_two_user", FirstName: "Page", LastName: "Two", Color: models.UserColorBlue},
		},
		total: 25,
	}
	handler := NewHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []models.UserSmall `json:"data"`
		Meta struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if store.gotLimit != userPageLimit {
		t.Fatalf("expected limit %d, got %d", userPageLimit, store.gotLimit)
	}
	if store.gotOffset != userPageLimit {
		t.Fatalf("expected offset %d, got %d", userPageLimit, store.gotOffset)
	}
	if body.Meta.Total != 25 {
		t.Fatalf("expected total 25, got %d", body.Meta.Total)
	}
	if body.Meta.Page != 1 {
		t.Fatalf("expected page 1, got %d", body.Meta.Page)
	}
	if body.Meta.PerPage != userPageLimit {
		t.Fatalf("expected per_page %d, got %d", userPageLimit, body.Meta.PerPage)
	}
	if len(body.Data) != 1 || body.Data[0].Username != "page_two_user" {
		t.Fatalf("unexpected users: %+v", body.Data)
	}
}

func TestListUsersInvalidPageFallsBackToZero(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "missing page", path: "/api/v1/users"},
		{name: "empty page", path: "/api/v1/users?page="},
		{name: "text page", path: "/api/v1/users?page=abc"},
		{name: "negative page", path: "/api/v1/users?page=-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeUserStore{
				list: []models.User{
					{ID: 1, Username: "alice", FirstName: "Alice", LastName: "Example", Color: models.UserColorGreen},
				},
				total: 1,
			}
			handler := NewHandler(store)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ListUsers(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if store.gotOffset != 0 {
				t.Fatalf("expected offset 0 for invalid page, got %d", store.gotOffset)
			}

			var body struct {
				Meta struct {
					Page int `json:"page"`
				} `json:"meta"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Meta.Page != 0 {
				t.Fatalf("expected page 0 for invalid page, got %d", body.Meta.Page)
			}
		})
	}
}

func TestListUsersAcceptsZeroPageQuery(t *testing.T) {
	store := &fakeUserStore{
		list: []models.User{
			{ID: 1, Username: "alice", FirstName: "Alice", LastName: "Example", Color: models.UserColorGreen},
		},
		total: 1,
	}
	handler := NewHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=0", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.gotOffset != 0 {
		t.Fatalf("expected offset 0, got %d", store.gotOffset)
	}

	var body struct {
		Meta struct {
			Page int `json:"page"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Meta.Page != 0 {
		t.Fatalf("expected page 0, got %d", body.Meta.Page)
	}
}

func TestListUsersEmpty(t *testing.T) {
	handler := NewHandler(&fakeUserStore{list: []models.User{}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []models.UserSmall `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 0 {
		t.Fatalf("expected empty list, got %+v", body.Data)
	}
}

func TestListUsersStoreErrorReturnsInternalError(t *testing.T) {
	handler := NewHandler(&fakeUserStore{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeUsersErrorEnvelope(t, rec).Error.Code; got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
}

func TestListUsersCountErrorReturnsInternalError(t *testing.T) {
	handler := NewHandler(&fakeUserStore{countErr: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeUsersErrorEnvelope(t, rec).Error.Code; got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
}

func TestGetUserReturnsUserProfile(t *testing.T) {
	createdAt := time.Date(2026, time.June, 20, 12, 0, 0, 0, time.UTC)
	store := &fakeUserStore{users: map[int64]models.User{
		42: {
			ID:           42,
			Email:        "alice@example.com",
			Username:     "alice",
			FirstName:    "Alice",
			LastName:     "Liddell",
			Color:        models.UserColorGreen,
			PasswordHash: "secret",
			CreatedAt:    createdAt,
			UpdatedAt:    createdAt.Add(time.Hour),
		},
	}}
	handler := NewHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42", nil)
	req.SetPathValue("id", "42")
	rec := httptest.NewRecorder()

	handler.GetUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	responseBytes := rec.Body.Bytes()
	var body struct {
		Data models.UserProfile `json:"data"`
	}
	if err := json.Unmarshal(responseBytes, &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.ID != 42 {
		t.Fatalf("expected id 42, got %d", body.Data.ID)
	}
	if body.Data.Username != "alice" {
		t.Fatalf("expected username alice, got %q", body.Data.Username)
	}
	if body.Data.Color != models.UserColorGreen {
		t.Fatalf("expected color green, got %q", body.Data.Color)
	}
	if body.Data.CreatedAt != createdAt {
		t.Fatalf("expected created_at %v, got %v", createdAt, body.Data.CreatedAt)
	}

	var raw struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(responseBytes, &raw); err != nil {
		t.Fatalf("decode raw response: %v", err)
	}
	if _, ok := raw.Data["created_at"]; !ok {
		t.Fatalf("expected exact created_at key, got %+v", raw.Data)
	}
	for _, forbidden := range []string{"CreatedAt", "email", "password", "password_hash", "updated_at"} {
		if _, ok := raw.Data[forbidden]; ok {
			t.Fatalf("response leaked forbidden key %q", forbidden)
		}
	}
	if rawJSON := string(responseBytes); strings.Contains(rawJSON, "secret") || strings.Contains(rawJSON, "alice@example.com") {
		t.Fatalf("response leaked sensitive data: %s", rawJSON)
	}
}

func TestGetUserIncludesNullProfilePicture(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {
			ID:        42,
			Username:  "alice",
			FirstName: "Alice",
			LastName:  "Liddell",
			Color:     models.UserColorGreen,
		},
	}}
	handler := NewHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42", nil)
	req.SetPathValue("id", "42")
	rec := httptest.NewRecorder()

	handler.GetUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assertRawJSONField(t, body.Data, "profile_picture", "null")
}

func TestGetUserInvalidIDReturnsNotFound(t *testing.T) {
	handler := NewHandler(&fakeUserStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	handler.GetUser(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeUsersErrorEnvelope(t, rec).Error.Code; got != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %q", got)
	}
}

func TestGetUserMapsStoreErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "missing user", err: ErrUserNotFound, wantStatus: http.StatusNotFound, wantCode: "NOT_FOUND"},
		{name: "internal", err: errors.New("db down"), wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&fakeUserStore{err: tt.err})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42", nil)
			req.SetPathValue("id", "42")
			rec := httptest.NewRecorder()

			handler.GetUser(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if got := decodeUsersErrorEnvelope(t, rec).Error.Code; got != tt.wantCode {
				t.Fatalf("expected %s, got %q", tt.wantCode, got)
			}
		})
	}
}

func TestUpdateUserAppliesPartialProfileUpdate(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {
			ID:             42,
			Email:          "old@example.com",
			Username:       "old_username",
			FirstName:      "Old",
			LastName:       "Name",
			ProfilePicture: "https://example.com/old.png",
			Color:          models.UserColorPurple,
		},
	}}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{
		"email":"  New@Example.COM  ",
		"username":"  new_username  ",
		"first_name":"  Alice  ",
		"last_name":"  Liddell  ",
		"color":"green"
	}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !store.updated || store.updatedID != 42 {
		t.Fatalf("expected update for user 42, got updated=%v id=%d", store.updated, store.updatedID)
	}
	assertStringPointer(t, "email", store.updatedParams.Email, "new@example.com")
	assertStringPointer(t, "username", store.updatedParams.Username, "new_username")
	assertStringPointer(t, "first_name", store.updatedParams.FirstName, "Alice")
	assertStringPointer(t, "last_name", store.updatedParams.LastName, "Liddell")
	assertStringPointer(t, "color", store.updatedParams.Color, models.UserColorGreen)

	var body struct {
		Data models.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Email != "new@example.com" || body.Data.Username != "new_username" {
		t.Fatalf("unexpected response user: %+v", body.Data)
	}
	if body.Data.Color != models.UserColorGreen {
		t.Fatalf("expected response color green, got %q", body.Data.Color)
	}
}

func TestUpdateUserCanRemoveProfilePicture(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {ID: 42, Email: "alice@example.com", Username: "alice", ProfilePicture: "https://example.com/avatar.png"},
	}}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{"profile_picture":null}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !store.updatedParams.ClearProfilePicture {
		t.Fatalf("expected profile picture to be marked for removal")
	}

	var body struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assertRawJSONField(t, body.Data, "profile_picture", "null")
}

func TestUpdateUserEmptyProfilePictureReturnsNull(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty string", body: `{"profile_picture":""}`},
		{name: "whitespace string", body: `{"profile_picture":"   "}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeUserStore{users: map[int64]models.User{
				42: {ID: 42, Email: "alice@example.com", Username: "alice", ProfilePicture: "https://example.com/avatar.png"},
			}}
			handler := NewHandler(store)

			rec := serveUpdateUser(t, handler, 42, "42", tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if !store.updatedParams.ClearProfilePicture {
				t.Fatalf("expected profile picture to be marked for removal")
			}

			var body struct {
				Data map[string]json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			assertRawJSONField(t, body.Data, "profile_picture", "null")
		})
	}
}

func TestUpdateUserRejectsProfilePictureURL(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {ID: 42, Email: "alice@example.com", Username: "alice", ProfilePicture: "https://example.com/old.png"},
	}}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{"profile_picture":"https://example.com/avatar.png"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if _, ok := body.Error.Fields["profile_picture"]; !ok {
		t.Fatalf("expected profile_picture field error, got %+v", body.Error.Fields)
	}
	if store.updated {
		t.Fatalf("store update should not be called for blocked profile picture URL")
	}
}

func TestUpdateUserRejectsProfilePictureURLWithoutPartialUpdate(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {ID: 42, Email: "alice@example.com", Username: "alice", Color: models.UserColorPurple},
	}}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{
		"profile_picture":"https://example.com/avatar.png",
		"color":"green"
	}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if _, ok := body.Error.Fields["profile_picture"]; !ok {
		t.Fatalf("expected profile_picture field error, got %+v", body.Error.Fields)
	}
	if store.updated {
		t.Fatalf("store update should not be called when profile_picture is invalid")
	}
}

func TestUpdateUserRejectsPasswordField(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {ID: 42, Email: "alice@example.com", Username: "alice"},
	}}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{"password":"new-secret-password"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeUsersErrorEnvelope(t, rec).Error.Code; got != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %q", got)
	}
	if store.updated {
		t.Fatalf("store update should not be called for rejected password field")
	}
}

func TestUpdateUserRejectsOAuthEmailUpdate(t *testing.T) {
	store := &fakeUserStore{
		oauthUsers: map[int64]bool{42: true},
	}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{"email":"new@example.com"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if _, ok := body.Error.Fields["email"]; !ok {
		t.Fatalf("expected email field error, got %+v", body.Error.Fields)
	}
	if _, ok := body.Error.Fields["password"]; ok {
		t.Fatalf("did not expect password field error, got %+v", body.Error.Fields)
	}
	if store.updated {
		t.Fatalf("store update should not be called for blocked OAuth email update")
	}
	if !store.oauthChecked || store.oauthCheckedID != 42 {
		t.Fatalf("expected OAuth check for user 42, got checked=%v id=%d", store.oauthChecked, store.oauthCheckedID)
	}
}

func TestUpdateUserRejectsOAuthEmailAndUsernameUpdate(t *testing.T) {
	store := &fakeUserStore{
		oauthUsers: map[int64]bool{42: true},
	}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{
		"email":"new@example.com",
		"username":"new_username"
	}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if _, ok := body.Error.Fields["email"]; !ok {
		t.Fatalf("expected email field error, got %+v", body.Error.Fields)
	}
	if _, ok := body.Error.Fields["username"]; !ok {
		t.Fatalf("expected username field error, got %+v", body.Error.Fields)
	}
	if store.updated {
		t.Fatalf("store update should not be called for blocked OAuth credential update")
	}
	if !store.oauthChecked || store.oauthCheckedID != 42 {
		t.Fatalf("expected OAuth check for user 42, got checked=%v id=%d", store.oauthChecked, store.oauthCheckedID)
	}
}

func TestUpdateUserAllowsOAuthAppearanceUpdates(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		assert func(t *testing.T, store *fakeUserStore)
	}{
		{
			name: "color",
			body: `{"color":"green"}`,
			assert: func(t *testing.T, store *fakeUserStore) {
				assertStringPointer(t, "color", store.updatedParams.Color, models.UserColorGreen)
				if store.updatedParams.ClearProfilePicture {
					t.Fatalf("did not expect profile picture update")
				}
			},
		},
		{
			name: "profile picture null",
			body: `{"profile_picture":null}`,
			assert: func(t *testing.T, store *fakeUserStore) {
				if !store.updatedParams.ClearProfilePicture {
					t.Fatalf("expected profile picture removal")
				}
			},
		},
		{
			name: "profile picture empty string",
			body: `{"profile_picture":""}`,
			assert: func(t *testing.T, store *fakeUserStore) {
				if !store.updatedParams.ClearProfilePicture {
					t.Fatalf("expected profile picture removal")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeUserStore{
				users: map[int64]models.User{
					42: {
						ID:             42,
						Email:          "oauth@example.com",
						Username:       "oauth_user",
						ProfilePicture: "https://example.com/old.png",
						Color:          models.UserColorPurple,
					},
				},
				oauthUsers: map[int64]bool{42: true},
			}
			handler := NewHandler(store)

			rec := serveUpdateUser(t, handler, 42, "42", tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if !store.updated || store.updatedID != 42 {
				t.Fatalf("expected update for user 42, got updated=%v id=%d", store.updated, store.updatedID)
			}
			tt.assert(t, store)
			if store.oauthChecked {
				t.Fatalf("OAuth check should not run for avatar-only updates")
			}
		})
	}
}

func TestUpdateUserRejectsOAuthProviderManagedFields(t *testing.T) {
	store := &fakeUserStore{
		oauthUsers: map[int64]bool{42: true},
	}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{
		"username":"new_username",
		"first_name":"New",
		"last_name":"Name"
	}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	for _, field := range []string{"username", "first_name", "last_name"} {
		if _, ok := body.Error.Fields[field]; !ok {
			t.Fatalf("expected %s field error, got %+v", field, body.Error.Fields)
		}
	}
	if store.updated {
		t.Fatalf("store update should not be called for blocked OAuth provider-managed field update")
	}
	if !store.oauthChecked || store.oauthCheckedID != 42 {
		t.Fatalf("expected OAuth check for user 42, got checked=%v id=%d", store.oauthChecked, store.oauthCheckedID)
	}
}

func TestUpdateUserRejectsOAuthMixedAllowedAndRestrictedFields(t *testing.T) {
	store := &fakeUserStore{
		oauthUsers: map[int64]bool{42: true},
	}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{
		"color":"green",
		"username":"new_username"
	}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if _, ok := body.Error.Fields["username"]; !ok {
		t.Fatalf("expected username field error, got %+v", body.Error.Fields)
	}
	if store.updated {
		t.Fatalf("store update should not be called when a restricted OAuth field is present")
	}
	if !store.oauthChecked || store.oauthCheckedID != 42 {
		t.Fatalf("expected OAuth check for user 42, got checked=%v id=%d", store.oauthChecked, store.oauthCheckedID)
	}
}

func TestUpdateUserAllowsEmailForPasswordUser(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {ID: 42, Email: "old@example.com", Username: "alice"},
	}}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{"email":"new@example.com"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	assertStringPointer(t, "email", store.updatedParams.Email, "new@example.com")
	if !store.oauthChecked || store.oauthCheckedID != 42 {
		t.Fatalf("expected OAuth check for credential update, got checked=%v id=%d", store.oauthChecked, store.oauthCheckedID)
	}
}

func TestUpdateUserAllowsProviderManagedFieldsForPasswordUser(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {ID: 42, Email: "alice@example.com", Username: "old_username", FirstName: "Old", LastName: "Name"},
	}}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{
		"username":"new_username",
		"first_name":"New",
		"last_name":"Name"
	}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	assertStringPointer(t, "username", store.updatedParams.Username, "new_username")
	assertStringPointer(t, "first_name", store.updatedParams.FirstName, "New")
	assertStringPointer(t, "last_name", store.updatedParams.LastName, "Name")
	if !store.oauthChecked || store.oauthCheckedID != 42 {
		t.Fatalf("expected OAuth check for provider-managed field update, got checked=%v id=%d", store.oauthChecked, store.oauthCheckedID)
	}
}

func TestUpdateUserOAuthStatusErrorReturnsInternalError(t *testing.T) {
	store := &fakeUserStore{oauthErr: errors.New("db down")}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{"email":"new@example.com"}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", body.Error.Code)
	}
	if store.updated {
		t.Fatalf("store update should not be called when OAuth status lookup fails")
	}
	if !store.oauthChecked || store.oauthCheckedID != 42 {
		t.Fatalf("expected OAuth check for user 42, got checked=%v id=%d", store.oauthChecked, store.oauthCheckedID)
	}
}

func TestUpdateUserValidatesEmailBeforeOAuthPolicy(t *testing.T) {
	store := &fakeUserStore{
		oauthUsers: map[int64]bool{42: true},
	}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "42", `{"email":"not-an-email"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if _, ok := body.Error.Fields["email"]; !ok {
		t.Fatalf("expected email field error, got %+v", body.Error.Fields)
	}
	if store.oauthChecked {
		t.Fatalf("OAuth check should not run before body validation succeeds")
	}
	if store.updated {
		t.Fatalf("store update should not be called on validation error")
	}
}

func TestUpdateUserRejectsDifferentAuthenticatedUser(t *testing.T) {
	store := &fakeUserStore{}
	handler := NewHandler(store)

	rec := serveUpdateUser(t, handler, 42, "7", `{"color":"green"}`)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.updated {
		t.Fatalf("store update should not be called for another user's profile")
	}
	if got := decodeUsersErrorEnvelope(t, rec).Error.Code; got != "FORBIDDEN" {
		t.Fatalf("expected FORBIDDEN, got %q", got)
	}
}

func TestUpdateUserRejectsMissingAuthContext(t *testing.T) {
	handler := NewHandler(&fakeUserStore{})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/42", strings.NewReader(`{"color":"green"}`))
	req.SetPathValue("id", "42")
	rec := httptest.NewRecorder()

	handler.UpdateUser(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeUsersErrorEnvelope(t, rec).Error.Code; got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestUpdateUserValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		pathID     string
		body       string
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{name: "invalid id", pathID: "abc", body: `{"color":"green"}`, wantStatus: http.StatusNotFound, wantCode: "NOT_FOUND"},
		{name: "malformed JSON", pathID: "42", body: `{"color":`, wantStatus: http.StatusBadRequest, wantCode: "BAD_REQUEST"},
		{name: "unknown field", pathID: "42", body: `{"role":"admin"}`, wantStatus: http.StatusBadRequest, wantCode: "BAD_REQUEST"},
		{name: "empty object", pathID: "42", body: `{}`, wantStatus: http.StatusBadRequest, wantCode: "VALIDATION_ERROR", wantField: "body"},
		{name: "invalid color", pathID: "42", body: `{"color":"orange"}`, wantStatus: http.StatusBadRequest, wantCode: "VALIDATION_ERROR", wantField: "color"},
		{name: "invalid email", pathID: "42", body: `{"email":"not-an-email"}`, wantStatus: http.StatusBadRequest, wantCode: "VALIDATION_ERROR", wantField: "email"},
		{name: "password field is rejected", pathID: "42", body: `{"password":"short"}`, wantStatus: http.StatusBadRequest, wantCode: "BAD_REQUEST"},
		{name: "numeric profile picture", pathID: "42", body: `{"profile_picture":123}`, wantStatus: http.StatusBadRequest, wantCode: "VALIDATION_ERROR", wantField: "profile_picture"},
		{name: "object profile picture", pathID: "42", body: `{"profile_picture":{"url":"https://example.com/avatar.png"}}`, wantStatus: http.StatusBadRequest, wantCode: "VALIDATION_ERROR", wantField: "profile_picture"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeUserStore{}
			handler := NewHandler(store)

			rec := serveUpdateUser(t, handler, 42, tt.pathID, tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			body := decodeUsersErrorEnvelope(t, rec)
			if body.Error.Code != tt.wantCode {
				t.Fatalf("expected %s, got %q", tt.wantCode, body.Error.Code)
			}
			if tt.wantField != "" {
				if _, ok := body.Error.Fields[tt.wantField]; !ok {
					t.Fatalf("expected field error for %q, got %+v", tt.wantField, body.Error.Fields)
				}
			}
			if store.updated {
				t.Fatalf("store update should not be called on validation error")
			}
		})
	}
}

func TestUpdateUserMapsStoreErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{name: "missing user", err: ErrUserNotFound, wantStatus: http.StatusNotFound, wantCode: "NOT_FOUND"},
		{name: "duplicate email", err: duplicateUserError("email"), wantStatus: http.StatusConflict, wantCode: "ALREADY_EXIST_ERROR", wantField: "email"},
		{name: "duplicate username", err: duplicateUserError("username"), wantStatus: http.StatusConflict, wantCode: "ALREADY_EXIST_ERROR", wantField: "username"},
		{name: "internal", err: errors.New("db down"), wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&fakeUserStore{updateErr: tt.err})

			rec := serveUpdateUser(t, handler, 42, "42", `{"color":"green"}`)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			body := decodeUsersErrorEnvelope(t, rec)
			if body.Error.Code != tt.wantCode {
				t.Fatalf("expected %s, got %q", tt.wantCode, body.Error.Code)
			}
			if tt.wantField != "" {
				if _, ok := body.Error.Fields[tt.wantField]; !ok {
					t.Fatalf("expected field error for %q, got %+v", tt.wantField, body.Error.Fields)
				}
			}
		})
	}
}

func TestSetPasswordSuccess(t *testing.T) {
	for _, tt := range []struct {
		name string
		body string
	}{
		{name: "without confirmation", body: `{"current_password":"old-correct-horse","new_password":"new-correct-horse"}`},
		{name: "with confirmation", body: `{"current_password":"old-correct-horse","new_password":"new-correct-horse","new_password_confirm":"new-correct-horse"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldHash := mustHashPassword(t, "old-correct-horse")
			store := &fakeUserStore{users: map[int64]models.User{
				42: {ID: 42, PasswordHash: oldHash},
			}}
			rec := serveSetPassword(t, NewHandler(store), 42, tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if !store.passwordUpdated || store.passwordUpdatedID != 42 {
				t.Fatalf("expected password update for user 42, got updated=%v id=%d", store.passwordUpdated, store.passwordUpdatedID)
			}
			if store.passwordExpectedHash != oldHash {
				t.Fatalf("store did not receive the previously loaded hash")
			}
			if store.passwordNewHash == "new-correct-horse" || !auth.CheckPassword(store.passwordNewHash, "new-correct-horse") {
				t.Fatalf("store did not receive a matching bcrypt hash")
			}
			response := rec.Body.String()
			for _, secret := range []string{"old-correct-horse", "new-correct-horse", oldHash, store.passwordNewHash} {
				if strings.Contains(response, secret) {
					t.Fatalf("response leaked password material: %s", response)
				}
			}
			var body struct {
				Data struct {
					Message string `json:"message"`
				} `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Data.Message != "Password has been changed" {
				t.Fatalf("unexpected success message %q", body.Data.Message)
			}
		})
	}
}

func TestSetPasswordRejectsInvalidCurrentPassword(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {ID: 42, PasswordHash: mustHashPassword(t, "old-correct-horse")},
	}}
	rec := serveSetPassword(t, NewHandler(store), 42, `{"current_password":"wrong-password","new_password":"new-correct-horse"}`)

	assertSetPasswordFieldError(t, rec, http.StatusUnauthorized, "INVALID_CURRENT_PASSWORD", "current-password")
	if store.passwordUpdated {
		t.Fatalf("password update should not be called")
	}
}

func TestSetPasswordRejectsUnchangedPasswordBeforeCurrentRules(t *testing.T) {
	for _, password := range []string{"old-correct-horse", "short", "password"} {
		t.Run(password, func(t *testing.T) {
			store := &fakeUserStore{users: map[int64]models.User{
				42: {ID: 42, PasswordHash: mustHashPassword(t, password)},
			}}
			body := `{"current_password":` + strconv.Quote(password) + `,"new_password":` + strconv.Quote(password) + `}`
			rec := serveSetPassword(t, NewHandler(store), 42, body)

			assertSetPasswordFieldError(t, rec, http.StatusConflict, "PASSWORD_UNCHANGED", "new-password")
			if store.passwordUpdated {
				t.Fatalf("password update should not be called")
			}
		})
	}
}

func TestSetPasswordValidatesNewPassword(t *testing.T) {
	oldHash := mustHashPassword(t, "old-correct-horse")
	tests := []struct {
		name     string
		password string
	}{
		{name: "too short", password: "short"},
		{name: "too long", password: strings.Repeat("x", 73)},
		{name: "too common", password: "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeUserStore{users: map[int64]models.User{42: {ID: 42, PasswordHash: oldHash}}}
			body := `{"current_password":"old-correct-horse","new_password":` + strconv.Quote(tt.password) + `}`
			rec := serveSetPassword(t, NewHandler(store), 42, body)

			assertSetPasswordFieldError(t, rec, http.StatusBadRequest, "VALIDATION_ERROR", "new-password")
			if store.passwordUpdated {
				t.Fatalf("password update should not be called")
			}
		})
	}
}

func TestSetPasswordRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		fields []string
	}{
		{name: "empty object", body: `{}`, fields: []string{"current-password", "new-password"}},
		{name: "missing current", body: `{"new_password":"new-correct-horse"}`, fields: []string{"current-password"}},
		{name: "empty current", body: `{"current_password":"","new_password":"new-correct-horse"}`, fields: []string{"current-password"}},
		{name: "missing new", body: `{"current_password":"old-correct-horse"}`, fields: []string{"new-password"}},
		{name: "empty new", body: `{"current_password":"old-correct-horse","new_password":""}`, fields: []string{"new-password"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveSetPassword(t, NewHandler(&fakeUserStore{}), 42, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			body := decodeUsersErrorEnvelope(t, rec)
			if body.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
			}
			if len(body.Error.Fields) != len(tt.fields) {
				t.Fatalf("expected fields %v, got %+v", tt.fields, body.Error.Fields)
			}
			for _, field := range tt.fields {
				if _, ok := body.Error.Fields[field]; !ok {
					t.Fatalf("expected field %q, got %+v", field, body.Error.Fields)
				}
			}
		})
	}
}

func TestSetPasswordRejectsNonStringFields(t *testing.T) {
	invalidValues := []string{"null", "123", "true", `[]`, `{}`}
	for _, field := range []struct {
		request  string
		response string
	}{
		{request: "current_password", response: "current-password"},
		{request: "new_password", response: "new-password"},
		{request: "new_password_confirm", response: "confirm-new-password"},
	} {
		for _, invalid := range invalidValues {
			name := field.request + "=" + invalid
			t.Run(name, func(t *testing.T) {
				body := map[string]string{
					"current_password":     `"old-correct-horse"`,
					"new_password":         `"new-correct-horse"`,
					"new_password_confirm": `"new-correct-horse"`,
				}
				body[field.request] = invalid
				jsonBody := `{"current_password":` + body["current_password"] + `,"new_password":` + body["new_password"] + `,"new_password_confirm":` + body["new_password_confirm"] + `}`
				rec := serveSetPassword(t, NewHandler(&fakeUserStore{}), 42, jsonBody)
				assertSetPasswordFieldError(t, rec, http.StatusBadRequest, "VALIDATION_ERROR", field.response)
			})
		}
	}
}

func TestSetPasswordConfirmation(t *testing.T) {
	store := &fakeUserStore{users: map[int64]models.User{
		42: {ID: 42, PasswordHash: mustHashPassword(t, "old-correct-horse")},
	}}
	rec := serveSetPassword(t, NewHandler(store), 42, `{"current_password":"old-correct-horse","new_password":"new-correct-horse","new_password_confirm":"different-password"}`)

	assertSetPasswordFieldError(t, rec, http.StatusBadRequest, "VALIDATION_ERROR", "confirm-new-password")
	if store.oauthChecked || store.passwordUpdated {
		t.Fatalf("confirmation must be rejected before loading password data")
	}
}

func TestSetPasswordRejectsInvalidJSONShape(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed", body: `{"current_password":`},
		{name: "multiple documents", body: `{}` + `{}`},
		{name: "unknown field", body: `{"current_password":"old-correct-horse","new_password":"new-correct-horse","password":"secret"}`},
		{name: "hyphen current", body: `{"current-password":"old-correct-horse","new_password":"new-correct-horse"}`},
		{name: "hyphen new", body: `{"current_password":"old-correct-horse","new-password":"new-correct-horse"}`},
		{name: "hyphen confirm", body: `{"current_password":"old-correct-horse","new_password":"new-correct-horse","confirm-new-password":"new-correct-horse"}`},
		{name: "oversized", body: `{"current_password":"` + strings.Repeat("x", 1<<20) + `","new_password":"new-correct-horse"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveSetPassword(t, NewHandler(&fakeUserStore{}), 42, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeUsersErrorEnvelope(t, rec).Error.Code; got != "BAD_REQUEST" {
				t.Fatalf("expected BAD_REQUEST, got %q", got)
			}
		})
	}
}

func TestSetPasswordRejectsOAuthUser(t *testing.T) {
	store := &fakeUserStore{
		users:      map[int64]models.User{42: {ID: 42, PasswordHash: mustHashPassword(t, "old-correct-horse")}},
		oauthUsers: map[int64]bool{42: true},
	}
	rec := serveSetPassword(t, NewHandler(store), 42, `{"current_password":"old-correct-horse","new_password":"new-correct-horse"}`)

	assertSetPasswordFieldError(t, rec, http.StatusBadRequest, "VALIDATION_ERROR", "new-password")
	if store.passwordUpdated {
		t.Fatalf("password update should not be called for OAuth user")
	}
}

func TestSetPasswordRejectsMissingAuthContext(t *testing.T) {
	handler := NewHandler(&fakeUserStore{})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/new-password", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	handler.SetPassword(rec, req)

	if rec.Code != http.StatusUnauthorized || decodeUsersErrorEnvelope(t, rec).Error.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetPasswordMapsStoreErrors(t *testing.T) {
	oldHash := mustHashPassword(t, "old-correct-horse")
	tests := []struct {
		name       string
		store      *fakeUserStore
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{name: "user not found", store: &fakeUserStore{}, wantStatus: http.StatusNotFound, wantCode: "NOT_FOUND"},
		{name: "user lookup", store: &fakeUserStore{err: errors.New("db down")}, wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
		{name: "oauth lookup", store: &fakeUserStore{users: map[int64]models.User{42: {ID: 42, PasswordHash: oldHash}}, oauthErr: errors.New("db down")}, wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
		{name: "password update", store: &fakeUserStore{users: map[int64]models.User{42: {ID: 42, PasswordHash: oldHash}}, passwordErr: errors.New("db down")}, wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
		{name: "concurrent password update", store: &fakeUserStore{users: map[int64]models.User{42: {ID: 42, PasswordHash: oldHash}}, passwordErr: ErrPasswordChanged}, wantStatus: http.StatusUnauthorized, wantCode: "INVALID_CURRENT_PASSWORD", wantField: "current-password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveSetPassword(t, NewHandler(tt.store), 42, `{"current_password":"old-correct-horse","new_password":"new-correct-horse"}`)
			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			body := decodeUsersErrorEnvelope(t, rec)
			if body.Error.Code != tt.wantCode {
				t.Fatalf("expected %s, got %q", tt.wantCode, body.Error.Code)
			}
			if tt.wantField != "" {
				if _, ok := body.Error.Fields[tt.wantField]; !ok {
					t.Fatalf("expected field %q, got %+v", tt.wantField, body.Error.Fields)
				}
			}
		})
	}
}

func TestSetPasswordLocalizesResponses(t *testing.T) {
	oldHash := mustHashPassword(t, "old-correct-horse")
	tests := []struct {
		name       string
		language   string
		current    string
		wantStatus int
		wantText   string
	}{
		{name: "German success", language: "de", current: "old-correct-horse", wantStatus: http.StatusOK, wantText: "Passwort wurde geändert"},
		{name: "French invalid current", language: "fr", current: "wrong-password", wantStatus: http.StatusUnauthorized, wantText: "Mot de passe actuel invalide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeUserStore{users: map[int64]models.User{42: {ID: 42, PasswordHash: oldHash}}}
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/new-password", strings.NewReader(`{"current_password":`+strconv.Quote(tt.current)+`,"new_password":"new-correct-horse"}`))
			req.Header.Set("Accept-Language", tt.language)
			rec := httptest.NewRecorder()
			auth.DevAuthenticateAs(42)(http.HandlerFunc(NewHandler(store).SetPassword)).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), tt.wantText) {
				t.Fatalf("expected %d containing %q, got %d: %s", tt.wantStatus, tt.wantText, rec.Code, rec.Body.String())
			}
		})
	}
}

type fakeUserStore struct {
	users                map[int64]models.User
	list                 []models.User
	total                int
	err                  error
	countErr             error
	updateErr            error
	passwordErr          error
	oauthUsers           map[int64]bool
	oauthErr             error
	updated              bool
	updatedID            int64
	updatedParams        UpdateUserParams
	passwordUpdated      bool
	passwordUpdatedID    int64
	passwordExpectedHash string
	passwordNewHash      string
	oauthChecked         bool
	oauthCheckedID       int64
	gotLimit             int
	gotOffset            int
}

func (s *fakeUserStore) ListUsers(_ context.Context, limit, offset int) ([]models.User, error) {
	s.gotLimit = limit
	s.gotOffset = offset

	if s.err != nil {
		return nil, s.err
	}
	return s.list, nil
}

func (s *fakeUserStore) CountUsers(_ context.Context) (int, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	if s.total != 0 {
		return s.total, nil
	}
	return len(s.list), nil
}

func (s *fakeUserStore) FindUserByID(_ context.Context, id int64) (models.User, error) {
	if s.err != nil {
		return models.User{}, s.err
	}
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return models.User{}, ErrUserNotFound
}

func (s *fakeUserStore) UserHasOAuthAccount(_ context.Context, id int64) (bool, error) {
	s.oauthChecked = true
	s.oauthCheckedID = id
	if s.oauthErr != nil {
		return false, s.oauthErr
	}
	return s.oauthUsers[id], nil
}

func (s *fakeUserStore) UpdateUser(_ context.Context, id int64, params UpdateUserParams) (models.User, error) {
	s.updated = true
	s.updatedID = id
	s.updatedParams = params

	if s.updateErr != nil {
		return models.User{}, s.updateErr
	}

	u := models.User{ID: id}
	if existing, ok := s.users[id]; ok {
		u = existing
	}
	if params.Email != nil {
		u.Email = *params.Email
	}
	if params.Username != nil {
		u.Username = *params.Username
	}
	if params.FirstName != nil {
		u.FirstName = *params.FirstName
	}
	if params.LastName != nil {
		u.LastName = *params.LastName
	}
	if params.ClearProfilePicture {
		u.ProfilePicture = ""
	}
	if params.Color != nil {
		u.Color = *params.Color
	}
	return u, nil
}

func (s *fakeUserStore) UpdatePasswordHash(_ context.Context, id int64, expectedHash, newHash string) error {
	s.passwordUpdated = true
	s.passwordUpdatedID = id
	s.passwordExpectedHash = expectedHash
	s.passwordNewHash = newHash
	if s.passwordErr != nil {
		return s.passwordErr
	}

	u, ok := s.users[id]
	if !ok || u.PasswordHash != expectedHash {
		return ErrPasswordChanged
	}
	u.PasswordHash = newHash
	s.users[id] = u
	return nil
}

func serveUpdateUser(t *testing.T, handler *Handler, authenticatedUserID int64, pathID string, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+pathID, strings.NewReader(body))
	req.SetPathValue("id", pathID)
	rec := httptest.NewRecorder()

	auth.DevAuthenticateAs(authenticatedUserID)(http.HandlerFunc(handler.UpdateUser)).ServeHTTP(rec, req)
	return rec
}

func serveSetPassword(t *testing.T, handler *Handler, authenticatedUserID int64, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/new-password", strings.NewReader(body))
	rec := httptest.NewRecorder()
	auth.DevAuthenticateAs(authenticatedUserID)(http.HandlerFunc(handler.SetPassword)).ServeHTTP(rec, req)
	return rec
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return hash
}

func assertSetPasswordFieldError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode, wantField string) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("expected %d, got %d: %s", wantStatus, rec.Code, rec.Body.String())
	}
	body := decodeUsersErrorEnvelope(t, rec)
	if body.Error.Code != wantCode {
		t.Fatalf("expected %s, got %q", wantCode, body.Error.Code)
	}
	if _, ok := body.Error.Fields[wantField]; !ok {
		t.Fatalf("expected field %q, got %+v", wantField, body.Error.Fields)
	}
}

func assertStringPointer(t *testing.T, name string, got *string, want string) {
	t.Helper()

	if got == nil {
		t.Fatalf("expected %s to be set to %q, got nil", name, want)
	}
	if *got != want {
		t.Fatalf("expected %s %q, got %q", name, want, *got)
	}
}

func assertRawJSONField(t *testing.T, fields map[string]json.RawMessage, field string, want string) {
	t.Helper()

	raw, ok := fields[field]
	if !ok {
		t.Fatalf("expected %s field, got fields: %+v", field, fields)
	}
	if string(raw) != want {
		t.Fatalf("expected %s to be %s, got %s", field, want, raw)
	}
}

func decodeUsersErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Error struct {
		Code   string `json:"code"`
		Fields map[string]struct {
			Message string `json:"message"`
		} `json:"fields"`
	} `json:"error"`
} {
	t.Helper()

	var body struct {
		Error struct {
			Code   string `json:"code"`
			Fields map[string]struct {
				Message string `json:"message"`
			} `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	return body
}
