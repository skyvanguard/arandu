package logging

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo}, // default
		{"", slog.LevelInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != "info" {
		t.Errorf("Level = %v, want info", cfg.Level)
	}
	if cfg.Format != "text" {
		t.Errorf("Format = %v, want text", cfg.Format)
	}
	if cfg.Output != "stdout" {
		t.Errorf("Output = %v, want stdout", cfg.Output)
	}
}

func TestLoggerWithBuffer(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Log output should contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Log output should contain 'key=value', got: %s", output)
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	logger := WithFields(Fields{
		"user_id": "123",
		"action":  "test",
	})

	logger.Info("test with fields")

	output := buf.String()
	if !strings.Contains(output, "user_id=123") {
		t.Errorf("Log output should contain 'user_id=123', got: %s", output)
	}
	if !strings.Contains(output, "action=test") {
		t.Errorf("Log output should contain 'action=test', got: %s", output)
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	testErr := errors.New("test error")
	logger := WithError(testErr)

	logger.Error("operation failed")

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Errorf("Log output should contain error message, got: %s", output)
	}
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	ctx := context.WithValue(context.Background(), RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, FlowIDKey, int64(456))

	logger := WithContext(ctx)
	logger.Info("context test")

	output := buf.String()
	if !strings.Contains(output, "request_id=req-123") {
		t.Errorf("Log output should contain request_id, got: %s", output)
	}
	if !strings.Contains(output, "flow_id=456") {
		t.Errorf("Log output should contain flow_id, got: %s", output)
	}
}

func TestLogRequest(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	LogRequest("GET", "/api/flows", 200, 50*time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "http_request") {
		t.Errorf("Log output should contain 'http_request', got: %s", output)
	}
	if !strings.Contains(output, "method=GET") {
		t.Errorf("Log output should contain 'method=GET', got: %s", output)
	}
	if !strings.Contains(output, "status=200") {
		t.Errorf("Log output should contain 'status=200', got: %s", output)
	}
}

func TestLogProviderCall(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	LogProviderCall("openai", "GenerateContent", 100*time.Millisecond, nil)

	output := buf.String()
	if !strings.Contains(output, "provider_call") {
		t.Errorf("Log output should contain 'provider_call', got: %s", output)
	}
	if !strings.Contains(output, "provider=openai") {
		t.Errorf("Log output should contain 'provider=openai', got: %s", output)
	}
}

func TestLogProviderCallWithError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	testErr := errors.New("rate limited")
	LogProviderCall("openai", "GenerateContent", 100*time.Millisecond, testErr)

	output := buf.String()
	if !strings.Contains(output, "rate limited") {
		t.Errorf("Log output should contain error, got: %s", output)
	}
	if !strings.Contains(output, "level=ERROR") {
		t.Errorf("Log should be at ERROR level, got: %s", output)
	}
}

func TestLogDockerOp(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	LogDockerOp("create", "abc123", 200*time.Millisecond, nil)

	output := buf.String()
	if !strings.Contains(output, "docker_operation") {
		t.Errorf("Log output should contain 'docker_operation', got: %s", output)
	}
	if !strings.Contains(output, "container_id=abc123") {
		t.Errorf("Log output should contain container_id, got: %s", output)
	}
}

func TestLogTask(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	Logger = slog.New(handler)

	LogTask(123, "terminal", "completed", 500*time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "task_execution") {
		t.Errorf("Log output should contain 'task_execution', got: %s", output)
	}
	if !strings.Contains(output, "task_id=123") {
		t.Errorf("Log output should contain task_id, got: %s", output)
	}
	if !strings.Contains(output, "task_type=terminal") {
		t.Errorf("Log output should contain task_type, got: %s", output)
	}
}
