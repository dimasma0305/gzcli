// Package server provides HTTP server functionality for running CTF challenges
// with WebSocket support, health checks, and rate limiting.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

// RunServer starts the HTTP server with all components
func RunServer(host string, port int) error {
	// Initialize components
	log.Info("Initializing server components...")

	// Create challenge manager and discover challenges
	challengeManager := NewChallengeManager()
	if err := challengeManager.DiscoverChallenges(); err != nil {
		return fmt.Errorf("failed to discover challenges: %w", err)
	}

	// Create executor
	executor := NewExecutor()

	// Create voting manager
	voting := NewVotingManager()

	// Create rate limiter
	rateLimiter := NewRateLimiter()

	// Create WebSocket manager
	wsManager := NewWSManager(challengeManager, executor, voting, rateLimiter)

	// Create health monitor
	healthMonitor := NewHealthMonitor(challengeManager, executor, wsManager)
	healthMonitor.Start()

	// Create HTTP server
	httpServer := NewServer(challengeManager, wsManager)
	if err := httpServer.LoadTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Setup routes
	mux := httpServer.SetupRoutes()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("┌────────────────────────────────────────────────┐")
	log.Info("│  GZCLI Challenge Launcher Server              │")
	log.Info("├────────────────────────────────────────────────┤")
	log.Info("│  Server:     http://%s:%d                 ", host, port)
	log.Info("│  Challenges: %d discovered                     ", challengeManager.GetChallengeCount())
	log.Info("└────────────────────────────────────────────────┘")
	log.Info("")
	log.Info("Available challenges:")
	for _, challenge := range challengeManager.ListChallenges() {
		log.Info("  • %s", challenge.Name)
		log.Info("    URL: http://%s:%d/%s", host, port, challenge.Slug)
	}
	log.Info("")
	log.Info("Press Ctrl+C to stop the server")

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Channel to listen for interrupt signals.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Info("Start shutdown... signal: %v", sig)

		// Give outstanding requests a deadline for completion.
		if err := GracefulShutdown(srv, 5*time.Second); err != nil {
			log.Error("Graceful shutdown failed: %v", err)
			if err := srv.Close(); err != nil {
				log.Error("Error forcing server close: %v", err)
			}
		}
	}

	// Cleanup on shutdown
	healthMonitor.Stop()

	// Stop all running challenges
	log.Info("Stopping all running challenges...")
	for _, challenge := range challengeManager.ListChallenges() {
		if challenge.GetStatus() != StatusStopped {
			log.Info("Stopping challenge: %s", challenge.Name)
			if err := executor.Stop(challenge); err != nil {
				log.Error("Failed to stop challenge %s: %v", challenge.Name, err)
			} else {
				challenge.SetStatus(StatusStopped)
				log.Info("Challenge %s stopped successfully", challenge.Name)
			}
		}
	}

	return nil
}

// GracefulShutdown performs a graceful server shutdown
func GracefulShutdown(srv *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Info("Shutting down server...")

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Info("Server shutdown complete")
	return nil
}
