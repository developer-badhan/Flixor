package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/handler"
	"github.com/developer-badhan/Flixor/internal/middleware"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/internal/service"
)

/** 
 * New creates and configures the Gin engine with all routes registered.
 * As we add phases, new route groups get registered here — main.go never changes.
 * New creates and configures the Gin engine with all routes registered.
 * Every new phase adds one route group registration here — nothing else changes.
*/
func New(db *config.DB, cfg *config.Config) *gin.Engine {
	engine := gin.New()

	/** 
	* Attach middleware to every request:
	* Logger — prints method, path, status, latency for every request
	* Recovery — catches any panic and returns 500 instead of crashing the server
	*/
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// Register all route groups
	registerHealthRoutes(engine)

	// ── API v1 group ──────────────────────────────────────────────────
	// All feature routes live under /api/v1.
	// Versioning from day one means we can add /api/v2 later without
	// breaking existing clients.
	v1 := engine.Group("/api/v1")

	// ── Auth routes (Phase 1) ─────────────────────────────────────────
	registerAuthRoutes(v1, db, cfg)

	// Phase 2 will add: registerMovieRoutes(v1, db, cfg)
	// Phase 5 will add: registerUserRoutes(v1, db, cfg)

	return engine
}

/**
 * registerHealthRoutes sets up the health check endpoint.
 * This route has no auth, no business logic — it just proves the server is alive.
 * Load balancers and monitoring tools ping this to know if the service is healthy.
*/
func registerHealthRoutes(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message" : "Gin Web server is running",
			"status":    "ok",
			"service":   "flixor-api",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
}

/**
 * registerAuthRoutes wires auth dependencies and registers all auth endpoints.
 * Dependency construction follows the chain:
 * repository → service → handler
 * Each layer receives only what it needs.
*/
func registerAuthRoutes(v1 *gin.RouterGroup, db *config.DB, cfg *config.Config) {
	// Build the dependency chain bottom-up
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	// Public auth routes — no JWT required
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	/**
	 * Protected auth routes — JWT required
	 * middleware.Auth() validates the token before the handler runs.
	 * If the token is missing or invalid, the handler never executes.
	*/
	protected := v1.Group("/auth")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	{
		protected.GET("/me", authHandler.Me)
	}
}