package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/router"
)

func main() {

	// Step 1: Load config 
	cfg := config.Load()

	// Step 2: Connect to MongoDB 
	db := config.ConnectDB(cfg)

	// Disconnect cleanly when main() returns for any reason.
	// defer runs even on panic — connection is always released.
	defer db.Disconnect()

	/**
	 * Step 3: Build the router 
	 * Creates Gin engine with all routes registered.
	 * We pass db here so future route groups can receive it.
	*/
	r := router.New()

	/** 
	 * Step 4: Configure the HTTP server 
	 * We configure net/http.Server explicitly rather than calling
	 * r.Run() so we can set timeouts. r.Run() has no timeout support
	 * which is unsafe in production — slow clients can hold connections open forever.
	*/
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,  // max time to read the full request
		WriteTimeout: 10 * time.Second,  // max time to write the full response
		IdleTimeout:  60 * time.Second,  // max time to keep idle connections alive
	}

	/** 
	 * Step 5: Start server in a goroutine 
	 * Running in a goroutine lets main() continue to the shutdown
	 * listener below. If we called server.ListenAndServe() directly,
	 * main() would block here and never reach the graceful shutdown logic.
	*/
	go func() {
		log.Printf("Flixor API running on port %s (env: %s)", cfg.Port, cfg.AppEnv)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	/** 
	 * Step 6: Graceful shutdown 
	 * Block here until the OS sends SIGINT (Ctrl+C) or SIGTERM (Docker stop).
	 * Without this, Ctrl+C would kill the process instantly — any in-flight
	 * requests would be cut off mid-response.
	*/
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received — draining connections...")

	// Give in-flight requests 5 seconds to finish before forcing shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown due to timeout: %v", err)
	}

	log.Println("Server stopped cleanly")
}