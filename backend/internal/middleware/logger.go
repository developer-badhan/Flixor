package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/developer-badhan/Flixor/pkg/response"
)

/**
 * RequestLogger is a Gin middleware that does two things per request:
 * 1. Injects a unique request_id into the Gin context (and response header)
 *    so every log line for this request can be correlated in your log aggregator.
 * 2. After the handler chain finishes, emits a structured JSON log line with:
 *    - method, path, status, latency, client IP, request_id
 *    - Log level is chosen by status code: ≥500 → error, ≥400 → warn, else info
 * Mount this as the FIRST middleware in your router chain.
*/
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		
		// 1. Generate & inject Request ID 
		requestID := uuid.New().String()
		c.Set(response.RequestIDKey, requestID)

		// Expose it in the response header for client-side tracing
		c.Header("X-Request-ID", requestID)

		// 2. Record start time 
		start := time.Now()

		// 3. Execute the rest of the handler chain 
		c.Next()

		// 4. Collect metrics after handlers complete 
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath() 
		if path == "" {
			path = c.Request.URL.Path 
		}
		clientIP := c.ClientIP()

		// 5. Build the structured log event 
		event := log.With().
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Str("client_ip", clientIP).
			Int("status", status).
			Dur("latency", latency).
			Logger()

		// 6. Choose log level based on HTTP status 
		switch {
		case status >= 500:
			event.Error().Msg("Server error")
		case status >= 400:
			event.Warn().Msg("Client error")
		default:
			event.Info().Msg("Request handled")
		}
	}
}
