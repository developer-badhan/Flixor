package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/**
 * CORSMiddleware handles Cross-Origin Resource Sharing (CORS) for the API.
 * 
 * It sets the necessary headers to allow the frontend to communicate with the backend
 * and handles the HTTP OPTIONS preflight requests.
 */
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In production, you would replace "*" with your actual frontend URL (e.g., https://flixor.app)
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		/**
		 * If the request is an OPTIONS request, we return 204 (No Content) immediately.
		 * Browsers send an OPTIONS request before certain types of requests (like those with headers).
		 */
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
