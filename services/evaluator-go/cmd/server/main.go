package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/aristath/arduino-trader/services/evaluator-go/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	// Get number of workers from environment or use default (num CPUs)
	numWorkers := runtime.NumCPU()
	if numWorkers < 2 {
		numWorkers = 2
	}

	log.Printf("Starting Go Evaluation Service on port %s with %d workers", port, numWorkers)

	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Create batch evaluator
	batchEvaluator := handlers.NewBatchEvaluator(numWorkers)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", handlers.HealthCheck)

		// Batch evaluation
		v1.POST("/evaluate/batch", batchEvaluator.EvaluateBatch)
	}

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
