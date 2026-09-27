package httpx

import "net/http"

// APIError is a structured error carrying an HTTP status and a stable code.
type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string { return e.Message }

// NewError builds an APIError.
func NewError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

// Common errors reused across handlers.
var (
	ErrBadRequest   = func(msg string) *APIError { return NewError(http.StatusBadRequest, "bad_request", msg) }
	ErrUnauthorized = func(msg string) *APIError { return NewError(http.StatusUnauthorized, "unauthorized", msg) }
	ErrForbidden    = func(msg string) *APIError { return NewError(http.StatusForbidden, "forbidden", msg) }
	ErrNotFound     = func(msg string) *APIError { return NewError(http.StatusNotFound, "not_found", msg) }
	ErrConflict     = func(msg string) *APIError { return NewError(http.StatusConflict, "conflict", msg) }
	ErrTooMany      = func(msg string) *APIError { return NewError(http.StatusTooManyRequests, "rate_limited", msg) }
	ErrInternal     = func(msg string) *APIError { return NewError(http.StatusInternalServerError, "internal", msg) }
)
