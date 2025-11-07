package logger

import (
	"context"
	"log/slog"
	"os"
)

// NewLogger creates a new slog.Logger instance.
// If production is true, it will use JSON format for structured logging.
// Otherwise, it will use a human-readable text format for development.
func NewLogger(production bool) *slog.Logger {
	var handler slog.Handler
	
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// Add source location for debugging
		AddSource: false,
	}
	
	if production {
		// JSON handler for production - easy to parse by monitoring tools
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text handler for development - human readable
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	
	return slog.New(handler)
}

// SetDefault sets the given logger as the default logger for the slog package.
func SetDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}

// FromContext retrieves the logger from the context.
// If no logger is found, it returns the default logger.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithLogger adds the logger to the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

type contextKey string

const loggerKey contextKey = "logger"
