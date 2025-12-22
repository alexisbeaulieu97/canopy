package main

import (
	"errors"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
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
	ExitWorkspaceLocked  ExitCode = 19
	ExitUnpushedCommits  ExitCode = 20
	ExitTimeout          ExitCode = 21
	ExitHookFailed       ExitCode = 22
	ExitPathError        ExitCode = 23
)

// errorCodeToExitCode maps error codes to CLI exit codes.
var errorCodeToExitCode = map[cerrors.ErrorCode]ExitCode{
	cerrors.ErrWorkspaceNotFound:      ExitNotFound,
	cerrors.ErrRepoNotFound:           ExitNotFound,
	cerrors.ErrWorkspaceExists:        ExitAlreadyExists,
	cerrors.ErrRepoAlreadyExists:      ExitAlreadyExists,
	cerrors.ErrRepoNotClean:           ExitDirtyWorkspace,
	cerrors.ErrConfigInvalid:          ExitConfigError,
	cerrors.ErrConfigValidation:       ExitConfigError,
	cerrors.ErrGitOperationFailed:     ExitGitError,
	cerrors.ErrUnknownRepository:      ExitUnknownResource,
	cerrors.ErrNotInWorkspace:         ExitNotInWorkspace,
	cerrors.ErrInvalidArgument:        ExitInvalidArgument,
	cerrors.ErrIOFailed:               ExitIOError,
	cerrors.ErrRegistryError:          ExitRegistryError,
	cerrors.ErrCommandFailed:          ExitCommandFailed,
	cerrors.ErrInternalError:          ExitInternalError,
	cerrors.ErrRepoInUse:              ExitRepoInUse,
	cerrors.ErrWorkspaceMetadata:      ExitMetadataError,
	cerrors.ErrNoReposConfigured:      ExitNoReposConfig,
	cerrors.ErrMissingBranchConfig:    ExitMissingBranch,
	cerrors.ErrOperationCancelled:     ExitOperationAborted,
	cerrors.ErrWorkspaceLocked:        ExitWorkspaceLocked,
	cerrors.ErrRepoHasUnpushedCommits: ExitUnpushedCommits,
	cerrors.ErrOperationTimeout:       ExitTimeout,
	cerrors.ErrHookFailed:             ExitHookFailed,
	cerrors.ErrHookTimeout:            ExitTimeout,
	cerrors.ErrPathInvalid:            ExitPathError,
	cerrors.ErrPathNotDirectory:       ExitPathError,
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
	return output.FormatErrorJSON(err)
}
