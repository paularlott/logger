package main

// This is an example of how to create an application-level log package
// Copy this pattern to your application as internal/log/log.go

import (
	"flag"
	"fmt"
	"os"

	"github.com/paularlott/logger"
	logslog "github.com/paularlott/logger/slog"
)

// Example: Application-level log package
var defaultLogger logger.Logger

func init() {
	// Initialize with default configuration
	defaultLogger = logslog.New(logslog.Config{
		Level:  "info",
		Format: "console",
		Writer: os.Stdout,
	})
}

// Configure sets up the logger
func Configure(level, format string) {
	defaultLogger = logslog.New(logslog.Config{
		Level:  level,
		Format: format,
		Writer: os.Stdout,
	})
}

// Package-level functions for convenience
func Info(msg string, keysAndValues ...any) {
	defaultLogger.Info(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...any) {
	defaultLogger.Debug(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...any) {
	defaultLogger.Warn(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...any) {
	defaultLogger.Error(msg, keysAndValues...)
}

func With(key string, value any) logger.Logger {
	return defaultLogger.With(key, value)
}

func WithError(err error) logger.Logger {
	return defaultLogger.WithError(err)
}

func GetLogger() logger.Logger {
	return defaultLogger
}

// Example: A service that accepts the logger interface
type UserService struct {
	log logger.Logger
}

func NewUserService(log logger.Logger) *UserService {
	if log == nil {
		log = logger.NewNullLogger()
	}
	return &UserService{
		log: log.WithGroup("user-service"),
	}
}

func (s *UserService) Login(userID int, username string) error {
	s.log.Info("user logging in", "user_id", userID, "username", username)

	// Simulate error
	if userID == 0 {
		err := fmt.Errorf("invalid user ID")
		s.log.WithError(err).Error("login failed")
		return err
	}

	s.log.Info("login successful", "user_id", userID)
	return nil
}

func main() {
	// Parse command line flags
	level := flag.String("log-level", "info", "Log level (trace|debug|info|warn|error)")
	format := flag.String("log-format", "console", "Log format (console|json)")
	flag.Parse()

	// Configure logging
	Configure(*level, *format)

	Info("application starting", "version", "1.0.0")

	// Create service with logger
	svc := NewUserService(GetLogger())

	// Demonstrate various logging features
	Info("demonstrating logging features")

	// Simple logging
	Debug("debug message", "detail", "some detail")
	Info("info message", "count", 42)
	Warn("warning message", "deprecated", "oldFeature")

	// With contextual fields
	reqLog := With("request_id", "abc123").With("ip", "192.168.1.1")
	reqLog.Info("processing request")
	reqLog.Debug("validated input")

	// Service logging
	if err := svc.Login(123, "john"); err != nil {
		Error("login error", "error", err)
	}

	// Error logging
	if err := svc.Login(0, "invalid"); err != nil {
		WithError(err).Error("operation failed")
	}

	Info("application stopped")
}
