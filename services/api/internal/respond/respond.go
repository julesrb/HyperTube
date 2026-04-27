package respond

import (
	"encoding/json"
	"net/http"
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

func List(w http.ResponseWriter, status int, data any, total int) {
	JSON(w, status, envelope{
		Data: data,
		Meta: &Meta{Total: total, Page: 1, PerPage: total},
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
