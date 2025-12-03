package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// ExitCode represents CLI exit codes.
type ExitCode int

// Exit codes for different error types.
const (
	ExitSuccess         ExitCode = 0
	ExitGeneralError    ExitCode = 1
	ExitNotFound        ExitCode = 2
	ExitAlreadyExists   ExitCode = 3
	ExitDirtyWorkspace  ExitCode = 4
	ExitConfigError     ExitCode = 5
	ExitGitError        ExitCode = 6
	ExitUnknownResource ExitCode = 7
)

// CLIError represents an error with additional CLI context.
type CLIError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// exitCodeForError returns the appropriate exit code for an error.
func exitCodeForError(err error) ExitCode {
	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) {
		return ExitGeneralError
	}

	switch canopyErr.Code {
	case cerrors.ErrWorkspaceNotFound, cerrors.ErrRepoNotFound:
		return ExitNotFound
	case cerrors.ErrWorkspaceExists, cerrors.ErrRepoAlreadyExists:
		return ExitAlreadyExists
	case cerrors.ErrRepoNotClean:
		return ExitDirtyWorkspace
	case cerrors.ErrConfigInvalid:
		return ExitConfigError
	case cerrors.ErrGitOperationFailed:
		return ExitGitError
	case cerrors.ErrUnknownRepository:
		return ExitUnknownResource
	default:
		return ExitGeneralError
	}
}

// userFriendlyMessage returns a user-friendly message for an error.
func userFriendlyMessage(err error) string {
	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) {
		return err.Error()
	}

	// For CanopyError, return the message portion without the code prefix
	return canopyErr.Message
}

// formatErrorJSON formats an error as JSON for --json output.
func formatErrorJSON(err error) string {
	cliErr := CLIError{
		Message: userFriendlyMessage(err),
	}

	var canopyErr *cerrors.CanopyError
	if errors.As(err, &canopyErr) {
		cliErr.Code = string(canopyErr.Code)
		if canopyErr.Cause != nil {
			cliErr.Details = canopyErr.Cause.Error()
		}
	}

	data, marshalErr := json.MarshalIndent(cliErr, "", "  ")
	if marshalErr != nil {
		return fmt.Sprintf(`{"message": %q}`, err.Error())
	}

	return string(data)
}

// handleError processes an error and exits with appropriate code.
// If jsonOutput is true, outputs JSON error format.
func handleError(err error, jsonOutput bool) {
	if err == nil {
		return
	}

	if jsonOutput {
		fmt.Fprintln(os.Stderr, formatErrorJSON(err)) //nolint:forbidigo // error output
	} else {
		fmt.Fprintln(os.Stderr, "Error:", userFriendlyMessage(err)) //nolint:forbidigo // error output
	}

	os.Exit(int(exitCodeForError(err)))
}

// isCanopyError checks if the error is a typed CanopyError.
func isCanopyError(err error) bool {
	var canopyErr *cerrors.CanopyError
	return errors.As(err, &canopyErr)
}

// getCanopyError extracts a CanopyError from an error chain, if present.
func getCanopyError(err error) *cerrors.CanopyError {
	var canopyErr *cerrors.CanopyError
	if errors.As(err, &canopyErr) {
		return canopyErr
	}
	return nil
}
