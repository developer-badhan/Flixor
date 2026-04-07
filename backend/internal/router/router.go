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
	authHandler        		*handler.AuthHandler,
	movieHandler       		*handler.MovieHandler,
	streamHandler      		*handler.StreamHandler,
	interactionHandler 		*handler.InteractionHandler,
	recommendationHandler   *handler.RecommendationHandler,
	jwtSecret          		string,
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

	// API V1 group
	v1 := r.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Movie & Stream routes
	movies := v1.Group("/movies")
	movies.Use(middleware.Auth(jwtSecret))
	{
		// Public endpoints — no auth required
		movies.GET("", movieHandler.GetMovies)        
		movies.GET("/:id", movieHandler.GetMovieByID) 
		movies.GET("/search", movieHandler.SearchMovies)

		// Admin endpoint — protected by JWT auth middleware
		movies.POST("/sync", movieHandler.SyncMovies)
	}

	// Single movie stream endpoint
	movie := v1.Group("/movie")
	movie.Use(middleware.Auth(jwtSecret))
	{
		movie.GET("/stream/:id", streamHandler.GetStream)

	}

	// Interaction routes
	interactions := v1.Group("/interactions")
	interactions.Use(middleware.Auth(jwtSecret))
	{
		interactions.GET("/watchlist", interactionHandler.GetWatchlist)
		interactions.POST("/watchlist/:id", interactionHandler.AddToWatchlist)
		interactions.DELETE("/watchlist/:id", interactionHandler.RemoveFromWatchlist)
		interactions.POST("/like/:id", interactionHandler.ReactToMovie)
		interactions.POST("/dislike/:id", interactionHandler.ReactToMovie)
		interactions.GET("/history", interactionHandler.GetHistory)
	}

	// Recommendation routes
	recommendations := v1.Group("/recommendations")
	recommendations.Use(middleware.Auth(jwtSecret))
	{
		recommendations.GET("/", recommendationHandler.GetRecommendations)
	}

	return r
}
