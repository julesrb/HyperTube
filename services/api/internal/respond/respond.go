package respond

import (
	"encoding/json"
	"net/http"
	"reflect"
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
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

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
