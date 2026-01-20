package errors

import (
	"errors"
	"fmt"
)

// Error types for consistent error handling across the application
var (
	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput indicates invalid user input
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates missing or invalid authentication
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the user doesn't have permission
	ErrForbidden = errors.New("forbidden")

	// ErrInternal indicates an internal server error
	ErrInternal = errors.New("internal error")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("timeout")

	// ErrUnavailable indicates a service is unavailable
	ErrUnavailable = errors.New("service unavailable")

	// ErrRateLimit indicates rate limiting was triggered
	ErrRateLimit = errors.New("rate limit exceeded")
)

// AppError represents an application error with context
type AppError struct {
	// Op is the operation being performed
	Op string
	// Err is the underlying error
	Err error
	// Message is a user-friendly message
	Message string
	// Code is an optional error code
	Code string
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Op
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(op string, err error, message string) *AppError {
	return &AppError{
		Op:      op,
		Err:     err,
		Message: message,
	}
}

// Wrap wraps an error with operation context
func Wrap(op string, err error) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Op:  op,
		Err: err,
	}
}

// WrapWithMessage wraps an error with operation context and a user message
func WrapWithMessage(op string, err error, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Op:      op,
		Err:     err,
		Message: message,
	}
}

// Is reports whether any error in err's chain matches target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalidInput checks if the error is an invalid input error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsInternal checks if the error is an internal error
func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsUnavailable checks if the error is a service unavailable error
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}

// ProviderError represents an error from an LLM provider
type ProviderError struct {
	Provider  string
	Op        string
	Err       error
	Retryable bool
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider %s: %s: %v", e.Provider, e.Op, e.Err)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new provider error
func NewProviderError(provider, op string, err error, retryable bool) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Op:        op,
		Err:       err,
		Retryable: retryable,
	}
}

// DockerError represents an error from Docker operations
type DockerError struct {
	Op          string
	ContainerID string
	Err         error
}

func (e *DockerError) Error() string {
	if e.ContainerID != "" {
		return fmt.Sprintf("docker %s (container %s): %v", e.Op, e.ContainerID, e.Err)
	}
	return fmt.Sprintf("docker %s: %v", e.Op, e.Err)
}

func (e *DockerError) Unwrap() error {
	return e.Err
}

// NewDockerError creates a new Docker error
func NewDockerError(op, containerID string, err error) *DockerError {
	return &DockerError{
		Op:          op,
		ContainerID: containerID,
		Err:         err,
	}
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d errors occurred: %v", len(e.Errors), e.Errors[0])
}

// Add adds an error to the multi-error
func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// HasErrors returns true if there are any errors
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// ErrorOrNil returns nil if there are no errors
func (e *MultiError) ErrorOrNil() error {
	if !e.HasErrors() {
		return nil
	}
	return e
}
