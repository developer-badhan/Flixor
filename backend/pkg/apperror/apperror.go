package apperror

import (
	"fmt"
	"net/http"
)

/**
 * AppError is the single error type used across the entire application.
 * Every layer (handler, service, repository) returns this type so the
 * global error middleware can always extract the HTTP status and error code.
*/
type AppError struct {
	// Code is a machine-readable string, e.g. "NOT_FOUND", "VALIDATION_ERROR"
	Code string `json:"code"`

	// Message is a human-readable description safe to return to the client
	Message string `json:"message"`

	// Details holds optional extra context, e.g. field-level validation errors
	Details any `json:"details,omitempty"`

	// HTTPStatus is the HTTP status code — NOT serialized into JSON
	HTTPStatus int `json:"-"`

	// Internal is the underlying Go error for logging — NOT sent to client
	Internal error `json:"-"`
}

/**
 * Error implements the error interface so AppError can be used anywhere
 * a standard error is expected.
*/
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap allows errors.Is / errors.As to traverse the chain.
func (e *AppError) Unwrap() error {
	return e.Internal
}

/**
 * WithInternal attaches an underlying Go error for server-side logging
 * without leaking it to the client. Returns the same *AppError for chaining.
 *
 * Usage:
 * The internal error is not sent to the client, but it is logged server-side.
 * apperror.ErrInternal.WithInternal(err)
*/
func (e *AppError) WithInternal(err error) *AppError {
	// Clone so we don't mutate the shared sentinel
	clone := *e
	clone.Internal = err
	return &clone
}

// WithDetails attaches extra context (e.g. validation field errors).
func (e *AppError) WithDetails(details any) *AppError {
	clone := *e
	clone.Details = details
	return &clone
}

// WithMessage overrides the human-readable message.
func (e *AppError) WithMessage(msg string) *AppError {
	clone := *e
	clone.Message = msg
	return &clone
}

/**
 * Predefined sentinel errors
 * Use these everywhere instead of creating ad-hoc errors.
 * ErrBadRequest: 400 Bad Request
 * ErrValidation: 422 Unprocessable Entity
 * ErrUnauthorized: 401 Unauthorized
 * ErrForbidden: 403 Forbidden
 * ErrNotFound: 404 Not Found
 * ErrConflict: 409 Conflict
 * ErrInternal: 500 Internal Server Error
 * ErrServiceUnavailable: 503 Service Unavailable
 * ErrTooManyRequests: 429 Too Many Requests
*/
var (
	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "Invalid request",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrValidation = &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "One or more fields are invalid",
		HTTPStatus: http.StatusUnprocessableEntity,
	}

	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "You do not have permission to perform this action",
		HTTPStatus: http.StatusForbidden,
	}

	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "The requested resource was not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrConflict = &AppError{
		Code:       "CONFLICT",
		Message:    "Resource already exists",
		HTTPStatus: http.StatusConflict,
	}

	ErrInternal = &AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "An unexpected error occurred. Please try again later.",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrServiceUnavailable = &AppError{
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "The service is temporarily unavailable",
		HTTPStatus: http.StatusServiceUnavailable,
	}

	ErrTooManyRequests = &AppError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    "Rate limit exceeded. Please slow down.",
		HTTPStatus: http.StatusTooManyRequests,
	}
)

/**
 * New creates a custom AppError on the fly when none of the sentinels fit.
 * 
 * Usage:
 * apperror.New(http.StatusBadRequest, "INVALID_TOKEN", "JWT token is expired")
*/
func New(httpStatus int, code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

/**
 * IsAppError checks whether an error is an *AppError and returns it.
 * Useful in middleware to distinguish domain errors from unexpected panics.
*/
func IsAppError(err error) (*AppError, bool) {
	ae, ok := err.(*AppError)
	return ae, ok
}
