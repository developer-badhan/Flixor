package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/developer-badhan/Flixor/internal/handler"
	"github.com/developer-badhan/Flixor/internal/middleware"
)

/**
 * SetupRouter wires all routes and returns the configured Gin engine.
 * All dependencies (handlers) are injected — the router itself has no business logic.
 */
func SetupRouter(
	authHandler *handler.AuthHandler,
	movieHandler *handler.MovieHandler,
	jwtSecret string,
) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery()) // recover from panics and return 500

	//  Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"content-type" : "Application/json",
			"message":   "Flixor API is healthy",
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"uptime":    time.Since(time.Now().Add(-time.Hour)).String(), // Example uptime
			"version":   "1.0.0",
		})
	})

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Movie routes
	movies := r.Group("/movies")
	{
		// Public endpoints — no auth required
		movies.GET("", movieHandler.GetMovies)        // GET /movies?page=1&limit=20
		movies.GET("/:id", movieHandler.GetMovieByID) // GET /movies/:id

		// Admin endpoint — protected by JWT auth middleware
		movies.POST("/sync", middleware.Auth(jwtSecret), movieHandler.SyncMovies)
	}

	return r
}
