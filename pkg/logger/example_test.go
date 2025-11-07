package logger_test

import (
	"log/slog"
	"os"

	"github.com/Loviiin/ponto-api-go/pkg/logger"
)

// ExampleNewLogger demonstrates how to create and use a structured logger
func ExampleNewLogger() {
	// Create a production logger (JSON format)
	log := logger.NewLogger(true)
	
	// Log various types of messages with structured fields
	log.Info("application started",
		slog.String("version", "1.0.0"),
		slog.String("environment", "production"))
	
	log.Warn("high memory usage",
		slog.Int("usage_mb", 512),
		slog.Float64("percentage", 85.5))
	
	// Logs are output in JSON format for easy parsing by monitoring tools
}

// ExampleMiddleware demonstrates the HTTP logging middleware
func ExampleMiddleware() {
	// In your main.go, add the middleware to your Gin router:
	// 
	// log := logger.NewLogger(isProduction)
	// router := gin.Default()
	// router.Use(logger.Middleware(log))
	//
	// This will automatically log all HTTP requests with:
	// - method
	// - path
	// - status code
	// - duration
	// - client IP
	// - user agent
}

// ExampleNewLogger_development demonstrates development mode logging
func ExampleNewLogger_development() {
	// Create a development logger (human-readable text format)
	log := logger.NewLogger(false)
	
	log.Info("server starting",
		slog.String("port", "8083"),
		slog.String("host", "localhost"))
	
	// In development mode, logs are easier to read for debugging
}

// Example showing how to use the logger based on environment variable
func Example_environmentBased() {
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	log := logger.NewLogger(isProduction)
	logger.SetDefault(log)
	
	// Now you can use slog.Info, slog.Error, etc. anywhere in your app
	slog.Info("application configured",
		slog.Bool("production", isProduction))
}
