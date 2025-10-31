package logger

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware returns a gin middleware that logs each HTTP request with structured logging.
// It logs: method, path, status code, duration, client IP, and user agent.
func Middleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Log request with structured fields
		logger.Info("http request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.Duration("duration", duration),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("request error",
					slog.String("error", e.Error()),
					slog.Int("error_type", int(e.Type)),
					slog.String("method", method),
					slog.String("path", path),
					slog.Int("status", statusCode),
				)
			}
		}
	}
}
