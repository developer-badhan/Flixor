package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/developer-badhan/Flixor/pkg/apperror"
)

/*
 * Package response provides a single, consistent JSON envelope for every API response.
 * Every endpoint must go through one of these helpers — no handler should call
 * c.JSON directly with arbitrary structures. This ensures:
 *   - Clients always know what shape to expect
 *   - Error codes are always machine-readable
 *   - Request IDs are always present for tracing
*/

/**	
 * Envelope types :
 * 
 * Response is the top-level JSON envelope returned for every API call.
 * 
*/
type Response struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Error     *Error `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

/**
 * Error is the structured error block inside the envelope.
 * Clients can rely on Code for programmatic handling.
*/
type Error struct {
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

/**
 * Context key for request ID
*/
const RequestIDKey = "request_id"

/**
 * Success helpers
 * OK sends a 200 OK response with data.
 * Usage:
 * 	response.OK(c, "Movies fetched", gin.H{"movies": movies})
*/
func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Created sends a 201 Created response.
func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// NoContent sends a 204 No Content response (no body).
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

/**
 * Error helpers:
 * Fail sends a structured error response from an *apperror.AppError.
 * This is the primary path — call it from the global error middleware.
*/
func Fail(c *gin.Context, err *apperror.AppError) {
	c.JSON(err.HTTPStatus, Response{
		Success: false,
		Message: err.Message,
		Error: &Error{
			Code:    err.Code,
			Details: err.Details,
		},
		RequestID: getRequestID(c),
	})
}

/**
 * InternalError is a convenience for unexpected errors that don't have an AppError.
 * It hides the raw Go error from the client and logs it separately.
*/
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: apperror.ErrInternal.Message,
		Error: &Error{
			Code: apperror.ErrInternal.Code,
		},
		RequestID: getRequestID(c),
	})
}

/**
 * Internal helpers 
 * getRequestID extracts the request ID from the context.
 * It is used by all the response helpers to include the request ID in the response.
*/
func getRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}
