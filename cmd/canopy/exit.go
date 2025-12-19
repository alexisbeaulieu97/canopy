package main

import "fmt"

// ExitCodeError is an error that carries a specific exit code.
// This allows RunE functions to signal non-standard exit codes
// without calling os.Exit directly, preserving Cobra's cleanup.
type ExitCodeError struct {
	Code    int
	Message string
}

// Error implements the error interface.
func (e *ExitCodeError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return fmt.Sprintf("exit code %d", e.Code)
}

// NewExitCodeError creates an ExitCodeError with the given code and message.
func NewExitCodeError(code int, message string) *ExitCodeError {
	return &ExitCodeError{
		Code:    code,
		Message: message,
	}
}
