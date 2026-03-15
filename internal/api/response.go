package api

import (
	"encoding/json"
	"net/http"

	"github.com/ed007183/llmgopher/pkg/llm"
)

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError writes an OpenAI-compatible error response.
func WriteError(w http.ResponseWriter, status int, msg, errType string) {
	WriteJSON(w, status, llm.APIError{
		Error: llm.APIErrorBody{
			Message: msg,
			Type:    errType,
		},
	})
}
