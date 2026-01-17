package executor

import (
	"testing"
)

func TestTerminalName(t *testing.T) {
	tests := []struct {
		name     string
		flowID   int64
		expected string
	}{
		{
			name:     "basic flow ID",
			flowID:   1,
			expected: "arandu-terminal-1",
		},
		{
			name:     "large flow ID",
			flowID:   123456789,
			expected: "arandu-terminal-123456789",
		},
		{
			name:     "zero flow ID",
			flowID:   0,
			expected: "arandu-terminal-0",
		},
		{
			name:     "negative flow ID",
			flowID:   -1,
			expected: "arandu-terminal--1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TerminalName(tt.flowID)
			if result != tt.expected {
				t.Errorf("TerminalName(%d) = %q, want %q", tt.flowID, result, tt.expected)
			}
		})
	}
}

func TestBrowserName(t *testing.T) {
	expected := "arandu-browser"
	result := BrowserName()
	if result != expected {
		t.Errorf("BrowserName() = %q, want %q", result, expected)
	}
}

func TestLogTypeConstants(t *testing.T) {
	// Verify log type constants have expected values
	tests := []struct {
		logType  LogType
		expected string
	}{
		{LogTypeInput, "input"},
		{LogTypeOutput, "output"},
		{LogTypeSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(string(tt.logType), func(t *testing.T) {
			if string(tt.logType) != tt.expected {
				t.Errorf("LogType constant = %q, want %q", tt.logType, tt.expected)
			}
		})
	}
}

func TestDefaultImageConstant(t *testing.T) {
	expected := "debian:latest"
	if defaultImage != expected {
		t.Errorf("defaultImage = %q, want %q", defaultImage, expected)
	}
}

func TestPortConstant(t *testing.T) {
	expected := "9222"
	if port != expected {
		t.Errorf("port = %q, want %q", port, expected)
	}
}

func TestSummaryWordCountConstant(t *testing.T) {
	expected := 10
	if SummaryWordCount != expected {
		t.Errorf("SummaryWordCount = %d, want %d", SummaryWordCount, expected)
	}
}
