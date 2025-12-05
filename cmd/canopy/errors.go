package main

import (
	"encoding/json"
	"errors"
	"fmt"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// ExitCode represents CLI exit codes.
type ExitCode int

// Exit codes for different error types.
const (
	ExitSuccess          ExitCode = 0
	ExitGeneralError     ExitCode = 1
	ExitNotFound         ExitCode = 2
	ExitAlreadyExists    ExitCode = 3
	ExitDirtyWorkspace   ExitCode = 4
	ExitConfigError      ExitCode = 5
	ExitGitError         ExitCode = 6
	ExitUnknownResource  ExitCode = 7
	ExitNotInWorkspace   ExitCode = 8
	ExitInvalidArgument  ExitCode = 9
	ExitIOError          ExitCode = 10
	ExitRegistryError    ExitCode = 11
	ExitCommandFailed    ExitCode = 12
	ExitInternalError    ExitCode = 13
	ExitRepoInUse        ExitCode = 14
	ExitMetadataError    ExitCode = 15
	ExitNoReposConfig    ExitCode = 16
	ExitMissingBranch    ExitCode = 17
	ExitOperationAborted ExitCode = 18
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
	case cerrors.ErrNotInWorkspace:
		return ExitNotInWorkspace
	case cerrors.ErrInvalidArgument:
		return ExitInvalidArgument
	case cerrors.ErrIOFailed:
		return ExitIOError
	case cerrors.ErrRegistryError:
		return ExitRegistryError
	case cerrors.ErrCommandFailed:
		return ExitCommandFailed
	case cerrors.ErrInternalError:
		return ExitInternalError
	case cerrors.ErrRepoInUse:
		return ExitRepoInUse
	case cerrors.ErrWorkspaceMetadata:
		return ExitMetadataError
	case cerrors.ErrNoReposConfigured:
		return ExitNoReposConfig
	case cerrors.ErrMissingBranchConfig:
		return ExitMissingBranch
	case cerrors.ErrOperationCancelled:
		return ExitOperationAborted
	default:
		return ExitGeneralError
	}
}

// userFriendlyMessage returns a user-friendly message for an error.
func userFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

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
