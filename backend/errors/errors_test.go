package errors

import (
	"errors"
	"testing"
)

func TestAppError(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "with message",
			appErr: &AppError{
				Op:      "CreateFlow",
				Message: "flow creation failed",
			},
			expected: "CreateFlow: flow creation failed",
		},
		{
			name: "with underlying error",
			appErr: &AppError{
				Op:  "CreateFlow",
				Err: errors.New("database error"),
			},
			expected: "CreateFlow: database error",
		},
		{
			name: "op only",
			appErr: &AppError{
				Op: "CreateFlow",
			},
			expected: "CreateFlow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appErr.Error(); got != tt.expected {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap("TestOp", originalErr)

	if wrapped == nil {
		t.Fatal("Wrap() returned nil")
	}

	if wrapped.Op != "TestOp" {
		t.Errorf("Wrap() Op = %v, want TestOp", wrapped.Op)
	}

	if !errors.Is(wrapped, originalErr) {
		t.Error("Wrap() should preserve original error")
	}
}

func TestWrapNil(t *testing.T) {
	wrapped := Wrap("TestOp", nil)

	if wrapped != nil {
		t.Error("Wrap(nil) should return nil")
	}
}

func TestIsNotFound(t *testing.T) {
	err := Wrap("TestOp", ErrNotFound)

	if !IsNotFound(err) {
		t.Error("IsNotFound() should return true")
	}

	otherErr := Wrap("TestOp", ErrInternal)
	if IsNotFound(otherErr) {
		t.Error("IsNotFound() should return false for other errors")
	}
}

func TestProviderError(t *testing.T) {
	err := NewProviderError("openai", "GenerateContent", errors.New("rate limited"), true)

	if err.Provider != "openai" {
		t.Errorf("Provider = %v, want openai", err.Provider)
	}

	if !err.Retryable {
		t.Error("Retryable should be true")
	}

	expected := "provider openai: GenerateContent: rate limited"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestDockerError(t *testing.T) {
	err := NewDockerError("create", "abc123", errors.New("image not found"))

	expected := "docker create (container abc123): image not found"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestDockerErrorNoContainer(t *testing.T) {
	err := NewDockerError("pull", "", errors.New("network error"))

	expected := "docker pull: network error"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("email", "invalid format")

	expected := "validation error: email: invalid format"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestMultiError(t *testing.T) {
	multi := &MultiError{}

	if multi.HasErrors() {
		t.Error("HasErrors() should return false for empty MultiError")
	}

	if multi.ErrorOrNil() != nil {
		t.Error("ErrorOrNil() should return nil for empty MultiError")
	}

	multi.Add(errors.New("error 1"))
	multi.Add(errors.New("error 2"))
	multi.Add(nil) // nil should be ignored

	if len(multi.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(multi.Errors))
	}

	if !multi.HasErrors() {
		t.Error("HasErrors() should return true")
	}

	if multi.ErrorOrNil() == nil {
		t.Error("ErrorOrNil() should return error")
	}
}

func TestMultiErrorSingle(t *testing.T) {
	multi := &MultiError{}
	multi.Add(errors.New("single error"))

	expected := "single error"
	if multi.Error() != expected {
		t.Errorf("Error() = %v, want %v", multi.Error(), expected)
	}
}
