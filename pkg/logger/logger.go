package logger

import (
	"log/slog"
	"os"
)

// NewLogger creates a new structured logger
func NewLogger(level slog.Level) *slog.Logger {
	var handler slog.Handler
	
	// Use JSON handler for production, text handler for development
	if os.Getenv("TESSLEBOX_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}
	
	return slog.New(handler)
}

// Info logs an info message with attributes
func Info(logger *slog.Logger, msg string, attrs ...any) {
	logger.Info(msg, attrs...)
}

// Error logs an error message with attributes
func Error(logger *slog.Logger, msg string, err error, attrs ...any) {
	if err != nil {
		logger.Error(msg, append(attrs, slog.String("error", err.Error()))...)
	} else {
		logger.Error(msg, attrs...)
	}
}

// Warn logs a warning message with attributes
func Warn(logger *slog.Logger, msg string, attrs ...any) {
	logger.Warn(msg, attrs...)
}

// Debug logs a debug message with attributes
func Debug(logger *slog.Logger, msg string, attrs ...any) {
	logger.Debug(msg, attrs...)
}