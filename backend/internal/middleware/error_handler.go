package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/developer-badhan/Flixor/pkg/apperror"
	"github.com/developer-badhan/Flixor/pkg/response"
)


/**
 * ErrorHandler is a Gin middleware that provides two things:
 * 1. Panic recovery — catches any unhandled panic in the handler chain,
 *    logs the full stack trace server-side, and returns a clean 500 to the client
 *    instead of crashing the server.
 * 2. Error formatting — handlers signal errors by calling c.Error(err).
 *    After all handlers run, this middleware inspects c.Errors, formats them
 *    through the response envelope, and writes the appropriate JSON response.
 *
 * Mount this AFTER RequestLogger so every error log includes the request_id.
 *
 * Handler usage pattern:
 *
 * movie, err := svc.GetByID(id)
 * if err != nil {
 * 	    _ = c.Error(err)   // attach to context
 * 	    return             // stop this handler
 * 	}
 * 	response.OK(c, "Movie fetched", movie)
 */
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Panic recovery 
		defer func() {
			if rec := recover(); rec != nil {
				requestID, _ := c.Get(response.RequestIDKey)

				// Log the panic with as much context as possible (server-side only)
				log.Error().
					Str("request_id", fmt.Sprintf("%v", requestID)).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Interface("panic", rec).
					Msg("Unhandled panic recovered")

				// Return a clean 500 — never expose the panic detail to clients
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Success: false,
					Message: apperror.ErrInternal.Message,
					Error: &response.Error{
						Code: apperror.ErrInternal.Code,
					},
					RequestID: fmt.Sprintf("%v", requestID),
				})
			}
		}()

		// Run handlers 
		c.Next()

		// Process attached errors 
		// Gin allows multiple errors to be attached; we handle the last one.
		if len(c.Errors) == 0 {
			return
		}

		// Grab the last error — in our architecture each handler attaches at most one
		ginErr := c.Errors.Last()
		requestID, _ := c.Get(response.RequestIDKey)
		rid := fmt.Sprintf("%v", requestID)

		// Check if it's our domain AppError 
		if appErr, ok := apperror.IsAppError(ginErr.Err); ok {
			// Log internal cause server-side (if any), hide from client
			if appErr.Internal != nil {
				log.Error().
					Err(appErr.Internal).
					Str("request_id", rid).
					Str("code", appErr.Code).
					Msg("Application error")
			} else if appErr.HTTPStatus >= 500 {
				log.Error().
					Str("request_id", rid).
					Str("code", appErr.Code).
					Msg(appErr.Message)
			} else {
				log.Warn().
					Str("request_id", rid).
					Str("code", appErr.Code).
					Msg(appErr.Message)
			}

			response.Fail(c, appErr)
			return
		}

		// Unknown / raw error — treat as 500 
		log.Error().
			Err(ginErr.Err).
			Str("request_id", rid).
			Msg("Unhandled error reached error middleware")

		response.InternalError(c)
	}
}
