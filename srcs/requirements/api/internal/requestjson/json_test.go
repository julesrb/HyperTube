package requestjson

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONObjectAcceptsValidJSONObject(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"Alice"}`))
	rec := httptest.NewRecorder()

	body, ok := DecodeJSONObject(rec, req, map[string]struct{}{"name": {}})
	if !ok {
		t.Fatalf("expected decode to succeed, got status %d: %s", rec.Code, rec.Body.String())
	}

	value, ok := DecodeString(body["name"])
	if !ok || value != "Alice" {
		t.Fatalf("expected name Alice, got value=%q ok=%v", value, ok)
	}
}

func TestDecodeJSONObjectRejectsMalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":`))
	rec := httptest.NewRecorder()

	if _, ok := DecodeJSONObject(rec, req, map[string]struct{}{"name": {}}); ok {
		t.Fatal("expected decode to fail")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDecodeJSONObjectRejectsUnknownField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"Alice","admin":true}`))
	rec := httptest.NewRecorder()

	if _, ok := DecodeJSONObject(rec, req, map[string]struct{}{"name": {}}); ok {
		t.Fatal("expected decode to fail")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDecodeJSONObjectRejectsMultipleJSONDocuments(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"Alice"} {}`))
	rec := httptest.NewRecorder()

	if _, ok := DecodeJSONObject(rec, req, map[string]struct{}{"name": {}}); ok {
		t.Fatal("expected decode to fail")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDecodeStringAcceptsString(t *testing.T) {
	value, ok := DecodeString(json.RawMessage(`"Alice"`))
	if !ok {
		t.Fatal("expected string decode to succeed")
	}
	if value != "Alice" {
		t.Fatalf("expected Alice, got %q", value)
	}
}

func TestDecodeStringRejectsNull(t *testing.T) {
	if _, ok := DecodeString(json.RawMessage(`null`)); ok {
		t.Fatal("expected null to be rejected")
	}
}

func TestDecodeStringRejectsNumber(t *testing.T) {
	if _, ok := DecodeString(json.RawMessage(`42`)); ok {
		t.Fatal("expected number to be rejected")
	}
}

func TestDecodeJSONRejectsMultipleJSONDocuments(t *testing.T) {
	var dst struct {
		Name string `json:"name"`
	}
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"Alice"} {}`))
	rec := httptest.NewRecorder()

	if DecodeJSON(rec, req, &dst) {
		t.Fatal("expected decode to fail")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
