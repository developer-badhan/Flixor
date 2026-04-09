package router

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/handler"
	"github.com/developer-badhan/Flixor/internal/middleware"
)

// startTime is the time when the server was started
var startTime = time.Now()

/**
 * SetupRoutes wires all API routes to the provided Gin engine.
 * The router relies on the injected handlers and doesn't handle server config.
 */
func SetupRoutes(
	r *gin.Engine,
	authHandler 			*handler.AuthHandler,
	movieHandler 			*handler.MovieHandler,
	streamHandler 			*handler.StreamHandler,
	interactionHandler 		*handler.InteractionHandler,
	recommendationHandler 	*handler.RecommendationHandler,
	analyticsHandler 		*handler.AnalyticsHandler,
	jwtSecret 				string,
	db 						*config.DB,
) {
	// Live check
	r.GET("/live", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "alive",
		})
	})	
	
	// Ready check
	r.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := db.Client.Ping(ctx, nil)
		if err != nil {
			c.JSON(503, gin.H{
				"status": "not_ready",
				"db":     "down",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ready",
			"db":     "up",
		})
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "flixor-api",
			"timestamp": time.Now().UTC(),
			"uptime":    time.Since(startTime).String(),
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
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/logout-all", authHandler.LogoutAll)
	}

	// Movie & Stream routes
	movies := v1.Group("/movies")
	{
		// Public endpoints — no auth required
		movies.GET("", movieHandler.GetMovies)
		movies.GET("/search", movieHandler.SearchMovies)	
		movies.GET("/:id", movieHandler.GetMovieByID)

		// Protected sub-group
		protected := movies.Group("/")
		protected.Use(middleware.Auth(jwtSecret))
		{
			protected.POST("/sync", movieHandler.SyncMovies)
		}
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

	// Analytics routes
	analytics := v1.Group("/analytics")
	analytics.Use(middleware.Auth(jwtSecret))
	{
		analytics.GET("/trending", analyticsHandler.GetTrending)
		analytics.GET("/most-watched", analyticsHandler.GetMostWatched)
		analytics.GET("/top-genres", analyticsHandler.GetTopGenres)
		analytics.GET("/stats", analyticsHandler.GetPlatformStats)
	}
}