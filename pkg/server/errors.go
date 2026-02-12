package server

import (
	"net/http"
)

// HTTPError represents an error with an associated HTTP status code.
type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

// NewHTTPError creates a new HTTPError.
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

// Error400 creates a 400 Bad Request error.
func Error400(message string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, message)
}

// Error401 creates a 401 Unauthorized error.
func Error401(message string) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message)
}

// Error403 creates a 403 Forbidden error.
func Error403(message string) *HTTPError {
	return NewHTTPError(http.StatusForbidden, message)
}

// Error404 creates a 404 Not Found error.
func Error404(message string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, message)
}

// Error500 creates a 500 Internal Server Error.
func Error500(message string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message)
}

// WrapError wraps an existing error into an HTTPError with the given code.
func WrapError(code int, err error) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: err.Error(),
	}
}
