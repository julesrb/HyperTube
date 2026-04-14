package respond

import (
	"encoding/json"
	"net/http"

	"hypertube/api/internal/models"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Data(w http.ResponseWriter, status int, data any) {
	JSON(w, status, models.Response{Data: data})
}

func List(w http.ResponseWriter, status int, data any, total int) {
	JSON(w, status, models.Response{
		Data: data,
		Meta: &models.Meta{Total: total, Page: 1, PerPage: total},
	})
}

func Item(w http.ResponseWriter, status int, data any) {
	JSON(w, status, models.Response{
		Data: data,
	})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, models.ErrorResponse{
		Error: models.ErrorBody{Code: code, Message: message},
	})
}
