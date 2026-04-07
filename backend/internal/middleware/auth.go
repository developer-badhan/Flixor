package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/developer-badhan/Flixor/pkg/utils"
)

/**
 * Auth returns a Gin middleware function that validates JWT tokens.
 * Usage: router.Use(middleware.Auth(secret)) or per-route:
 *         protected.GET("/profile", middleware.Auth(secret), handler.Profile)
 * 
 *  On success: injects user_id and email into the request context.
 * On failure: aborts with 401 — the handler never runs.
*/
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Extract the token from the Authorization header 
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		// Header must be exactly two parts: "Bearer" and the token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header format must be: Bearer <token>",
			})
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token cannot be empty",
			})
			return
		}

		// Step 2: Validate the token 
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			statusCode := http.StatusUnauthorized
			c.AbortWithStatusJSON(statusCode, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Step 3: Inject claims into request context 
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		// ── Step 4: Pass control to the next handler 
		c.Next()
	}
}