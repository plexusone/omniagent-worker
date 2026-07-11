package worker

import (
	"fmt"
)

// WorkerError provides structured error information.
type WorkerError struct {
	// Code is a machine-readable error code.
	Code string `json:"code"`

	// Message is a human-readable error message.
	Message string `json:"message"`

	// Details contains additional error context.
	Details any `json:"details,omitempty"`

	// Cause is the underlying error, if any.
	Cause error `json:"-"`
}

// Error codes.
const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeExecution    = "EXECUTION_ERROR"
	ErrCodeTimeout      = "TIMEOUT_ERROR"
	ErrCodeLLM          = "LLM_ERROR"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeUnavailable  = "UNAVAILABLE"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeCancelled    = "CANCELLED"
	ErrCodeRateLimited  = "RATE_LIMITED"
	ErrCodeUnauthorized = "UNAUTHORIZED"
)

// Error implements the error interface.
func (e *WorkerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *WorkerError) Unwrap() error {
	return e.Cause
}

// NewError creates a new WorkerError.
func NewError(code, message string) *WorkerError {
	return &WorkerError{
		Code:    code,
		Message: message,
	}
}

// NewErrorWithCause creates a new WorkerError with an underlying cause.
func NewErrorWithCause(code, message string, cause error) *WorkerError {
	return &WorkerError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates a validation error.
func NewValidationError(message string) *WorkerError {
	return NewError(ErrCodeValidation, message)
}

// NewExecutionError creates an execution error.
func NewExecutionError(message string, cause error) *WorkerError {
	return NewErrorWithCause(ErrCodeExecution, message, cause)
}

// NewTimeoutError creates a timeout error.
func NewTimeoutError(message string) *WorkerError {
	return NewError(ErrCodeTimeout, message)
}

// NewLLMError creates an LLM error.
func NewLLMError(message string, cause error) *WorkerError {
	return NewErrorWithCause(ErrCodeLLM, message, cause)
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(message string) *WorkerError {
	return NewError(ErrCodeNotFound, message)
}

// IsWorkerError checks if an error is a WorkerError.
func IsWorkerError(err error) bool {
	_, ok := err.(*WorkerError)
	return ok
}

// AsWorkerError converts an error to a WorkerError.
// If the error is already a WorkerError, it is returned as-is.
// Otherwise, a new WorkerError is created wrapping the original error.
func AsWorkerError(err error) *WorkerError {
	if err == nil {
		return nil
	}
	if we, ok := err.(*WorkerError); ok {
		return we
	}
	return NewErrorWithCause(ErrCodeInternal, err.Error(), err)
}
