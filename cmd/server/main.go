// Package main provides the Astrometry API Server
//
//	@title						Astrometry API Server
//	@version					0.1.0
//	@description				REST API for astronomical image plate-solving using Astrometry.net
//	@termsOfService				http://swagger.io/terms/
//
//	@contact.name				API Support
//	@contact.url				https://github.com/DiarmuidKelly/astrometry-api-server
//
//	@license.name				GPL-3.0
//	@license.url				https://www.gnu.org/licenses/gpl-3.0.en.html
//
//	@host						localhost:8080
//	@BasePath					/
//	@schemes					http
//
//	@tag.name					Analysis
//	@tag.description			Image analysis and FOV calculation
//	@tag.name					Solving
//	@tag.description			Plate-solving operations
//	@tag.name					Health
//	@tag.description			Server health and status
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/DiarmuidKelly/astrometry-api-server/docs"
	"github.com/DiarmuidKelly/astrometry-api-server/internal/handlers"
	"github.com/DiarmuidKelly/astrometry-api-server/internal/middleware"
	client "github.com/DiarmuidKelly/astrometry-go-client"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Configuration from environment
	indexPath := getEnv("ASTROMETRY_INDEX_PATH", "/data/indexes")
	port := getEnv("PORT", "8080")
	containerName := getEnv("ASTROMETRY_CONTAINER_NAME", "astrometry-solver")
	maxUploadSize := int64(50 * 1024 * 1024) // 50MB default

	// Create astrometry client with docker exec mode
	// Note: Docker socket access required for containerized deployment
	// See SECURITY.md for security considerations
	config := &client.ClientConfig{
		IndexPath:     indexPath,
		Timeout:       5 * time.Minute,
		TempDir:       "/shared-data",
		UseDockerExec: true,
		ContainerName: containerName,
	}

	astrometryClient, err := client.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create astrometry client: %v", err)
	}

	// Create handlers
	solveHandler := handlers.NewSolveHandler(astrometryClient, maxUploadSize)
	analyseHandler := handlers.NewAnalyseHandler(maxUploadSize)
	healthHandler := handlers.NewHealthHandler()

	// Setup router
	mux := http.NewServeMux()
	mux.Handle("/solve", middleware.Logger(middleware.CORS(solveHandler)))
	mux.Handle("/analyse", middleware.Logger(middleware.CORS(analyseHandler)))
	mux.Handle("/health", middleware.Logger(healthHandler))

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Minute, // Long timeout for solving
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting Astrometry API Server on port %s", port)
		log.Printf("Using index path: %s", indexPath)
		log.Printf("Using docker exec mode with container: %s", containerName)
		log.Printf("Swagger UI available at: http://localhost:%s/swagger/", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
