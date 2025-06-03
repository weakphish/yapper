package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var (
	logger *slog.Logger
	file   *os.File
)

// Init initializes the logger with a file destination
func Init() error {
	// Create log directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create logs directory in ~/.config/yapper/logs
	logDir := filepath.Join(homeDir, ".config", "yapper", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp in name
	timestamp := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(logDir, fmt.Sprintf("yapper-%s.log", timestamp))
	
	file, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create JSON handler that writes to the file
	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		AddSource: true,
	})

	// Create the logger
	logger = slog.New(handler)

	// Log that logger was initialized
	logger.Info("logger initialized", "file", logFilePath)
	return nil
}

// Close closes the log file
func Close() error {
	if file != nil {
		logger.Info("closing logger")
		return file.Close()
	}
	return nil
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if logger != nil {
		logger.Debug(msg, args...)
	}
}

// Info logs an info message
func Info(msg string, args ...any) {
	if logger != nil {
		logger.Info(msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if logger != nil {
		logger.Warn(msg, args...)
	}
}

// Error logs an error message
func Error(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
}

// GetLogger returns the slog.Logger instance
func GetLogger() *slog.Logger {
	return logger
}