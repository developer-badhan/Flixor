package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"


	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/handler"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/internal/router"
	"github.com/developer-badhan/Flixor/internal/middleware"
	"github.com/developer-badhan/Flixor/internal/service"
	"github.com/developer-badhan/Flixor/pkg/cloudinary"
	"github.com/developer-badhan/Flixor/pkg/email"
	"github.com/developer-badhan/Flixor/pkg/logger"
)

/**
 * Main entry point for the Flixor backend server.
 *
 * Steps:
 * 1. Load configuration and initialize logging.
 * 2. Connect to Database (MongoDB).
 * 3. Initialize Repositories, Services, and Handlers.
 * 4. Create and configure the Gin HTTP server framework.
 * 5. Inject dependencies into API routes.
 * 6. Start HTTP server and listen for graceful shutdown.
 */
func main() {
	// Load config (also loads .env internally)
	cfg := config.Load()

	// Initialise structured logger
	logger.Init(cfg.AppEnv) // "development" or "production"
	logger.Info().Str("env", cfg.AppEnv).Msg("Flixor backend starting")

	// Set Gin mode based on environment BEFORE initializing Gin
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gemini client
	geminiClient := service.NewGeminiClient(cfg.GEMINI_API_KEY)

	// Connect to MongoDB
	db := config.ConnectDB(cfg)
	log.Println("✅ Connected to MongoDB")

	// Ensure analytics indexes exist (non-fatal)
	initCtx := context.Background()
	if err := repository.EnsureAnalyticsIndexes(initCtx, db.Database); err != nil {
		log.Printf("Warning: could not ensure analytics indexes: %v", err)
	}

	// Disconnect cleanly when main() returns for any reason.
	defer db.Disconnect()

	// Repositories
	userRepo := repository.NewUserRepository(db.Database)
	movieRepo := repository.NewMovieRepository(db.Database)
	streamRepo := repository.NewStreamRepository(db.Database)
	interactionRepo := repository.NewInteractionRepository(db.Database)
	recoRepo := repository.NewRecommendationRepository(db.Database)
	analyticsRepo := repository.NewAnalyticsRepository(db.Database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db.Database)

	// Initialize Cloudinary client
	cldClient, err := cloudinary.NewClient(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	if err != nil {
		log.Fatalf("failed to initialise Cloudinary: %v", err)
	}

	// Initialize mailer
	mailer := email.NewMailer(cfg.EmailHost, cfg.EmailPort, cfg.EmailUser, cfg.EmailPassword)

	// Ensure DB Indexes (run at startup, idempotent)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := refreshTokenRepo.EnsureIndexes(ctx); err != nil {
		log.Fatalf("failed to create refresh token indexes: %v", err)
	}
	log.Println("✅ Refresh token indexes ensured")

	// Services
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, cfg)
	userSvc := service.NewUserService(userRepo, cldClient, mailer)
	movieSvc := service.NewMovieService(movieRepo)
	streamSvc := service.NewStreamService(streamRepo)
	interactionSvc := service.NewInteractionService(interactionRepo)
	recoSvc := service.NewRecommendationService(recoRepo, geminiClient)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	movieHandler := handler.NewMovieHandler(movieSvc)
	streamHandler := handler.NewStreamHandler(streamSvc)
	interactionHandler := handler.NewInteractionHandler(interactionSvc)
	recoHandler := handler.NewRecommendationHandler(recoSvc)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)

	// ----------------------------------------------------
	// SERVER SETUP
	// ----------------------------------------------------

	// Initialize Gin engine
	r := gin.New()

	// CORS middleware
	r.Use(middleware.CORSMiddleware())
	// Request ID middleware
	r.Use(middleware.RequestIDMiddleware())

	// Global rate limiter: generous for normal browsing
	globalStore := middleware.NewRateLimiterStore(10, 20)
	r.Use(middleware.RateLimit(globalStore))

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Set trusted proxies to nil to fix the warning
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// Setup routes using the initialized Gin engine
	router.SetupRoutes(
		r,
		authHandler,
		userHandler,
		movieHandler,
		streamHandler,
		interactionHandler,
		recoHandler,
		analyticsHandler,
		cfg.JWTSecret,
		db,
	)

	// Define Server Port
	port := cfg.Port
	if port == "" {
		port = "5000"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start in a goroutine so we can listen for shutdown signals
	go func() {
		log.Printf("🚀 Flixor server starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown signal received — draining connections...
	logger.Info().Msg("Shutdown signal received — draining connections...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}
	// Server exited cleanly
	logger.Info().Msg("Server exited cleanly")
}