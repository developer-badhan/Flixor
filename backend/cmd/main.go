package main

import (
	"log"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/handler"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/internal/service"
	"github.com/developer-badhan/Flixor/internal/router"

	"github.com/joho/godotenv"
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
	// 1. Load environment variables 
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from OS environment")
	}

	// 2. Load config 
	cfg := config.Load()

	// 3. Connect to MongoDB 
	db := config.ConnectDB(cfg)
	log.Println("✅ Connected to MongoDB")

	// Disconnect cleanly when main() returns for any reason.
	defer db.Disconnect()

	// 3. Repositories 
	userRepo  := repository.NewUserRepository(db)
	movieRepo := repository.NewMovieRepository(db.Database)
	streamRepo := repository.NewStreamRepository(db.Database)

	// 4. Services 
	authSvc  := service.NewAuthService(userRepo, cfg)
	movieSvc := service.NewMovieService(movieRepo)
	streamSvc := service.NewStreamService(streamRepo)

	// 5. Handlers 
	authHandler  := handler.NewAuthHandler(authSvc)
	movieHandler := handler.NewMovieHandler(movieSvc)
	streamHandler := handler.NewStreamHandler(streamSvc)


	// 6. Setup router with all handlers and JWT middleware
	r := router.SetupRouter(authHandler, movieHandler, streamHandler, cfg.JWTSecret)

	// 7. Start server 
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

