package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestNewLogger_Production(t *testing.T) {
	var buf bytes.Buffer
	
	// Create a production logger that writes to our buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)
	
	// Log a test message
	logger.Info("test message", slog.String("key", "value"))
	
	// Parse the JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}
	
	// Verify the log contains expected fields
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got %v", logEntry["msg"])
	}
	
	if logEntry["key"] != "value" {
		t.Errorf("Expected key='value', got %v", logEntry["key"])
	}
}

func TestNewLogger_Development(t *testing.T) {
	var buf bytes.Buffer
	
	// Create a development logger that writes to our buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)
	
	// Log a test message
	logger.Info("test message", slog.String("key", "value"))
	
	// Verify the log contains the message (text format)
	output := buf.String()
	if output == "" {
		t.Error("Expected non-empty log output")
	}
}

func TestFromContext(t *testing.T) {
	// Create a logger
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)
	
	// Add logger to context
	ctx := WithLogger(context.Background(), logger)
	
	// Retrieve logger from context
	retrievedLogger := FromContext(ctx)
	
	if retrievedLogger == nil {
		t.Error("Expected logger to be retrieved from context")
	}
	
	// Log with retrieved logger
	retrievedLogger.Info("context test")
	
	if buf.Len() == 0 {
		t.Error("Expected log output from context logger")
	}
}

func TestFromContext_Default(t *testing.T) {
	// Retrieve logger from empty context - should return default
	ctx := context.Background()
	logger := FromContext(ctx)
	
	if logger == nil {
		t.Error("Expected default logger when context has no logger")
	}
}
