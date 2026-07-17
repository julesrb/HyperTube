package respond

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hypertube/api/internal/i18n"
)

func TestLocalizedErrorWritesTranslatedEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	req.Header.Set("Accept-Language", "de")
	rec := httptest.NewRecorder()

	LocalizedError(rec, req, http.StatusUnauthorized, "INVALID_CREDENTIALS", i18n.MsgInvalidCredentials)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "INVALID_CREDENTIALS" {
		t.Fatalf("expected INVALID_CREDENTIALS, got %q", body.Error.Code)
	}
	if body.Error.Message != "E-Mail, Benutzername oder Passwort ist ungültig" {
		t.Fatalf("expected German message, got %q", body.Error.Message)
	}
}

func TestValidationErrorWritesFieldEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()

	ValidationError(rec, http.StatusBadRequest, FieldErrors{
		"email": {Message: "Invalid email"},
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Error struct {
			Code    string      `json:"code"`
			Message string      `json:"message"`
			Fields  FieldErrors `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if body.Error.Message != "" {
		t.Fatalf("expected no top-level validation message, got %q", body.Error.Message)
	}
	if body.Error.Fields["email"].Message != "Invalid email" {
		t.Fatalf("expected email field message, got %+v", body.Error.Fields)
	}
}

func TestDataAndListEnvelopes(t *testing.T) {
	t.Run("data", func(t *testing.T) {
		rec := httptest.NewRecorder()

		Data(rec, http.StatusCreated, map[string]string{"message": "ok"})

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var body struct {
			Data map[string]string `json:"data"`
			Meta *Meta             `json:"meta"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Data["message"] != "ok" {
			t.Fatalf("unexpected data envelope: %+v", body.Data)
		}
		if body.Meta != nil {
			t.Fatalf("did not expect meta for data envelope, got %+v", body.Meta)
		}
	})

	t.Run("paginated list", func(t *testing.T) {
		rec := httptest.NewRecorder()

		ListPaginated(rec, http.StatusOK, []string{"a", "b"}, 10, 2, 5)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var body struct {
			Data []string `json:"data"`
			Meta Meta     `json:"meta"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(body.Data) != 2 || body.Meta.Total != 10 || body.Meta.Page != 2 || body.Meta.PerPage != 5 {
			t.Fatalf("unexpected list envelope: data=%+v meta=%+v", body.Data, body.Meta)
		}
	})
}
