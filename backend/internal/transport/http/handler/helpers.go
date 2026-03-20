package handler

import (
	"encoding/json"
	"net/http"

	"meetings-editor/internal/transport/http/model"
)

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func respondError(w http.ResponseWriter, status int, message string, details any) {
	respond(w, status, model.ErrorResponse{Message: message, Details: details})
}
