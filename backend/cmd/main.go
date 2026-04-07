package main

import (
	"log"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/handler"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/internal/service"
	"github.com/developer-badhan/Flixor/internal/router"
)

/** 
 * Main entry point for the Flixor backend server.
 * 
 * Steps:
 * 1. Load environment variables from .env file (if exists).
 * 2. Load application configuration.
 * 3. Connect to MongoDB and defer disconnection.
 * 4. Initialize repositories for users and movies.
 * 5. Initialize services for authentication and movie management
 * 6. Initialize HTTP handlers for auth and movies.
 * 7. Set up the Gin router with the handlers and JWT middleware
 * 8. Start the HTTP server on the specified port.
 * 
 * Note: The server will log a message when it starts and if it fails to start.
*/
func main() {
	// Load config (also loads .env internally)
	cfg := config.Load()

	// Initialize Gemini client
	geminiClient := service.NewGeminiClient(cfg.GEMINI_API_KEY)

	// Connect to MongoDB 
	db := config.ConnectDB(cfg)
	log.Println("✅ Connected to MongoDB")

	// Disconnect cleanly when main() returns for any reason.
	defer db.Disconnect()

	// Repositories 
	userRepo  := repository.NewUserRepository(db)
	movieRepo := repository.NewMovieRepository(db.Database)
	streamRepo := repository.NewStreamRepository(db.Database)
	interactionRepo := repository.NewInteractionRepository(db.Database)
	recoRepo := repository.NewRecommendationRepository(db.Database)
	analyticsRepo := repository.NewAnalyticsRepository(db.Database)

	// Services 
	authSvc  := service.NewAuthService(userRepo, cfg)
	movieSvc := service.NewMovieService(movieRepo)
	streamSvc := service.NewStreamService(streamRepo)
	interactionSvc := service.NewInteractionService(interactionRepo)
	recoSvc := service.NewRecommendationService(recoRepo, geminiClient)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)

	// Handlers 
	authHandler  := handler.NewAuthHandler(authSvc)
	movieHandler := handler.NewMovieHandler(movieSvc)
	streamHandler := handler.NewStreamHandler(streamSvc)
	interactionHandler := handler.NewInteractionHandler(interactionSvc)
	recoHandler := handler.NewRecommendationHandler(recoSvc)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)

	// Setup router with all handlers and JWT middleware
	r := router.SetupRouter(authHandler, movieHandler, streamHandler, interactionHandler, recoHandler, analyticsHandler, cfg.JWTSecret)

	// Set trusted proxies to nil to fix the warning
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// Start server 
	// port := os.Getenv("PORT")
	port := config.Load().Port
	if port == "" {
		port = "5000"
	}

	log.Printf("🚀 Flixor server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

