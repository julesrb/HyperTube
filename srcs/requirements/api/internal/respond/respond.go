package respond

import (
	"encoding/json"
	"net/http"
	"reflect"

	"hypertube/api/internal/i18n"
)

type Meta struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type envelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

type errorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message,omitempty"`
	Fields  FieldErrors `json:"fields,omitempty"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type FieldError struct {
	Message string `json:"message"`
}

type FieldErrors map[string]FieldError

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Data(w http.ResponseWriter, status int, data any) {
	JSON(w, status, envelope{Data: data})
}

func List(w http.ResponseWriter, status int, data any) {
	total := listLength(data)
	JSON(w, status, envelope{
		Data: data,
		Meta: &Meta{Total: total, Page: 0, PerPage: total},
	})
}

func ListPaginated(w http.ResponseWriter, status int, data any, total, page, perPage int) {
	JSON(w, status, envelope{
		Data: data,
		Meta: &Meta{Total: total, Page: page, PerPage: perPage},
	})
}

func Item(w http.ResponseWriter, status int, data any) {
	JSON(w, status, envelope{Data: data})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, errorResponse{
		Error: errorBody{Code: code, Message: message},
	})
}

func LocalizedError(w http.ResponseWriter, r *http.Request, status int, code string, message i18n.Message, args ...any) {
	Error(w, status, code, i18n.T(i18n.FromRequest(r), message, args...))
}

func ValidationError(w http.ResponseWriter, status int, fields FieldErrors) {
	ErrorWithFields(w, status, "VALIDATION_ERROR", fields)
}

func ErrorWithFields(w http.ResponseWriter, status int, code string, fields FieldErrors) {
	JSON(w, status, errorResponse{
		Error: errorBody{Code: code, Fields: fields},
	})
}

func FieldValidationError(w http.ResponseWriter, status int, field, message string) {
	ValidationError(w, status, FieldErrors{
		field: {Message: message},
	})
}

func LocalizedFieldValidationError(w http.ResponseWriter, r *http.Request, status int, field string, message i18n.Message, args ...any) {
	FieldValidationError(w, status, field, i18n.T(i18n.FromRequest(r), message, args...))
}

func listLength(data any) int {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return 0
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return v.Len()
	default:
		return 0
	}
}
