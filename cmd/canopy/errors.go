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

// errorCodeToExitCode maps error codes to CLI exit codes.
var errorCodeToExitCode = map[cerrors.ErrorCode]ExitCode{
	cerrors.ErrWorkspaceNotFound:   ExitNotFound,
	cerrors.ErrRepoNotFound:        ExitNotFound,
	cerrors.ErrWorkspaceExists:     ExitAlreadyExists,
	cerrors.ErrRepoAlreadyExists:   ExitAlreadyExists,
	cerrors.ErrRepoNotClean:        ExitDirtyWorkspace,
	cerrors.ErrConfigInvalid:       ExitConfigError,
	cerrors.ErrGitOperationFailed:  ExitGitError,
	cerrors.ErrUnknownRepository:   ExitUnknownResource,
	cerrors.ErrNotInWorkspace:      ExitNotInWorkspace,
	cerrors.ErrInvalidArgument:     ExitInvalidArgument,
	cerrors.ErrIOFailed:            ExitIOError,
	cerrors.ErrRegistryError:       ExitRegistryError,
	cerrors.ErrCommandFailed:       ExitCommandFailed,
	cerrors.ErrInternalError:       ExitInternalError,
	cerrors.ErrRepoInUse:           ExitRepoInUse,
	cerrors.ErrWorkspaceMetadata:   ExitMetadataError,
	cerrors.ErrNoReposConfigured:   ExitNoReposConfig,
	cerrors.ErrMissingBranchConfig: ExitMissingBranch,
	cerrors.ErrOperationCancelled:  ExitOperationAborted,
}

// exitCodeForError returns the appropriate exit code for an error.
func exitCodeForError(err error) ExitCode {
	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) {
		return ExitGeneralError
	}

	if code, ok := errorCodeToExitCode[canopyErr.Code]; ok {
		return code
	}

	return ExitGeneralError
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
