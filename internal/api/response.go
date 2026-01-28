package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type ListResponse[T any] struct {
	Data   []T `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func Error(w http.ResponseWriter, status int, err string, message string) {
	JSON(w, status, ErrorResponse{Error: err, Message: message})
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "bad_request", message)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "not_found", message)
}

func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
}
