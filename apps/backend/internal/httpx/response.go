// Package httpx contains small HTTP helpers: JSON responses, error mapping and
// middleware shared across all backend modules.
package httpx

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("httpx: encode response: %v", err)
	}
}

// NoContent writes a 204 response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes an error response. *APIError values keep their status/code;
// anything else becomes a generic 500 so we never leak internals to clients.
func Error(w http.ResponseWriter, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		JSON(w, apiErr.Status, apiErr)
		return
	}
	log.Printf("httpx: unhandled error: %v", err)
	JSON(w, http.StatusInternalServerError, NewError(http.StatusInternalServerError, "internal", "internal server error"))
}

// Decode reads and validates a JSON request body into dst.
func Decode(r *http.Request, dst any) error {
	if r.Body == nil {
		return ErrBadRequest("empty request body")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return ErrBadRequest("invalid JSON body: " + err.Error())
	}
	return nil
}
