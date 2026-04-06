package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

/** 
 * New creates and configures the Gin engine with all routes registered.
 * As we add phases, new route groups get registered here — main.go never changes.
*/
func New() *gin.Engine {
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

	// Phase 1 will add: registerAuthRoutes(engine)
	// Phase 2 will add: registerMovieRoutes(engine, db)
	// Each phase adds one line here — nothing else changes

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