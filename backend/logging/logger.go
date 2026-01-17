package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// ContextKey is a custom type for context keys to satisfy staticcheck SA1029
type ContextKey string

// Context key constants for logging
const (
	RequestIDKey ContextKey = "request_id"
	FlowIDKey    ContextKey = "flow_id"
	UserIDKey    ContextKey = "user_id"
)

// Logger is the global logger instance
var Logger *slog.Logger

// Config holds logger configuration
type Config struct {
	// Level is the minimum log level (debug, info, warn, error)
	Level string
	// Format is the output format (json, text)
	Format string
	// Output is the output destination (stdout, stderr, or file path)
	Output string
	// AddSource adds source file and line to logs
	AddSource bool
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:     "info",
		Format:    "text",
		Output:    "stdout",
		AddSource: false,
	}
}

// Init initializes the global logger
func Init(cfg Config) error {
	level := parseLevel(cfg.Level)
	output, err := parseOutput(cfg.Output)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	return nil
}

// parseLevel converts a string level to slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// parseOutput converts an output string to io.Writer
func parseOutput(output string) (io.Writer, error) {
	switch output {
	case "stdout", "":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	}
}

// Fields is a map of key-value pairs for structured logging
type Fields map[string]any

// Debug logs a debug message
func Debug(msg string, fields ...any) {
	Logger.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...any) {
	Logger.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...any) {
	Logger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...any) {
	Logger.Error(msg, fields...)
}

// WithContext returns a logger with context values
func WithContext(ctx context.Context) *slog.Logger {
	// Extract common context values
	logger := Logger

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.With("request_id", requestID)
	}

	if flowID, ok := ctx.Value(FlowIDKey).(int64); ok {
		logger = logger.With("flow_id", flowID)
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		logger = logger.With("user_id", userID)
	}

	return logger
}

// WithFields returns a logger with additional fields
func WithFields(fields Fields) *slog.Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return Logger.With(args...)
}

// WithError returns a logger with an error field
func WithError(err error) *slog.Logger {
	return Logger.With("error", err.Error())
}

// LogRequest logs an HTTP request
func LogRequest(method, path string, statusCode int, duration time.Duration, fields ...any) {
	args := []any{
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ms", duration.Milliseconds(),
	}
	args = append(args, fields...)
	Logger.Info("http_request", args...)
}

// LogProviderCall logs an LLM provider API call
func LogProviderCall(provider, operation string, duration time.Duration, err error, fields ...any) {
	args := []any{
		"provider", provider,
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
	}
	if err != nil {
		args = append(args, "error", err.Error())
	}
	args = append(args, fields...)

	if err != nil {
		Logger.Error("provider_call", args...)
	} else {
		Logger.Info("provider_call", args...)
	}
}

// LogDockerOp logs a Docker operation
func LogDockerOp(operation, containerID string, duration time.Duration, err error, fields ...any) {
	args := []any{
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
	}
	if containerID != "" {
		args = append(args, "container_id", containerID)
	}
	if err != nil {
		args = append(args, "error", err.Error())
	}
	args = append(args, fields...)

	if err != nil {
		Logger.Error("docker_operation", args...)
	} else {
		Logger.Info("docker_operation", args...)
	}
}

// LogTask logs a task execution
func LogTask(taskID int64, taskType, status string, duration time.Duration, fields ...any) {
	args := []any{
		"task_id", taskID,
		"task_type", taskType,
		"status", status,
		"duration_ms", duration.Milliseconds(),
	}
	args = append(args, fields...)
	Logger.Info("task_execution", args...)
}

// LogPanic logs a panic with stack trace
func LogPanic(recovered any) {
	stack := make([]byte, 4096)
	n := runtime.Stack(stack, false)
	Logger.Error("panic_recovered",
		"panic", recovered,
		"stack", string(stack[:n]),
	)
}

func init() {
	// Initialize with default config (stdout never fails)
	_ = Init(DefaultConfig())
}
