package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/developer-badhan/Flixor/pkg/response"
)

/**
 * RequestIDMiddleware injects a unique UUID into every request.
 * 
 * This allows us to trace a specific request through all our logs and 
 * provides a reference for users if they encounter an error.
 */
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Generate a new unique ID
		requestID := uuid.New().String()

		// 2. Set it in the Gin context so handlers and response helpers can access it
		c.Set(response.RequestIDKey, requestID)

		// 3. Set it in the response header so the client can see/log it
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
