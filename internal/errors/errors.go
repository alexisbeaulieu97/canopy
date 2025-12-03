// Package errors provides typed errors for the canopy application.
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode identifies the type of error.
type ErrorCode string

// Error codes for domain errors.
const (
	ErrWorkspaceNotFound  ErrorCode = "WORKSPACE_NOT_FOUND"
	ErrWorkspaceExists    ErrorCode = "WORKSPACE_EXISTS"
	ErrRepoNotFound       ErrorCode = "REPO_NOT_FOUND"
	ErrRepoNotClean       ErrorCode = "REPO_NOT_CLEAN"
	ErrRepoAlreadyExists  ErrorCode = "REPO_ALREADY_EXISTS"
	ErrGitOperationFailed ErrorCode = "GIT_OPERATION_FAILED"
	ErrConfigInvalid      ErrorCode = "CONFIG_INVALID"
	ErrUnknownRepository  ErrorCode = "UNKNOWN_REPOSITORY"
)

// CanopyError is a typed error with code, message, cause, and context.
type CanopyError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Context map[string]string
}

// Error implements the error interface.
func (e *CanopyError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for use with errors.Unwrap.
func (e *CanopyError) Unwrap() error {
	return e.Cause
}

// Is checks if the target error has the same error code.
func (e *CanopyError) Is(target error) bool {
	var t *CanopyError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// WithContext adds or updates context key-value pairs.
func (e *CanopyError) WithContext(key, value string) *CanopyError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// NewWorkspaceNotFound creates an error for when a workspace is not found.
func NewWorkspaceNotFound(id string) *CanopyError {
	return &CanopyError{
		Code:    ErrWorkspaceNotFound,
		Message: fmt.Sprintf("workspace %s not found", id),
		Context: map[string]string{"workspace_id": id},
	}
}

// NewWorkspaceExists creates an error for when a workspace already exists.
func NewWorkspaceExists(id string) *CanopyError {
	return &CanopyError{
		Code:    ErrWorkspaceExists,
		Message: fmt.Sprintf("workspace already exists: %s", id),
		Context: map[string]string{"workspace_id": id},
	}
}

// NewRepoNotFound creates an error for when a repository is not found.
func NewRepoNotFound(name string) *CanopyError {
	return &CanopyError{
		Code:    ErrRepoNotFound,
		Message: fmt.Sprintf("repository %s not found", name),
		Context: map[string]string{"repo_name": name},
	}
}

// NewRepoNotClean creates an error for when a repository has uncommitted changes.
func NewRepoNotClean(repoName, action string) *CanopyError {
	return &CanopyError{
		Code:    ErrRepoNotClean,
		Message: fmt.Sprintf("repo %s has uncommitted changes. Use --force to %s", repoName, action),
		Context: map[string]string{"repo_name": repoName, "action": action},
	}
}

// NewRepoAlreadyExists creates an error for when a repo already exists in a workspace.
func NewRepoAlreadyExists(repoName, workspaceID string) *CanopyError {
	return &CanopyError{
		Code:    ErrRepoAlreadyExists,
		Message: fmt.Sprintf("repository %s already exists in workspace %s", repoName, workspaceID),
		Context: map[string]string{"repo_name": repoName, "workspace_id": workspaceID},
	}
}

// NewUnknownRepository creates an error for unresolvable repository identifiers.
func NewUnknownRepository(identifier string, userRequested bool) *CanopyError {
	var msg string
	if userRequested {
		msg = fmt.Sprintf("unknown repository '%s'. Register it first: canopy repo register %s <repository-url>", identifier, identifier)
	} else {
		msg = fmt.Sprintf("unknown repository '%s': provide a URL or registered alias", identifier)
	}
	return &CanopyError{
		Code:    ErrUnknownRepository,
		Message: msg,
		Context: map[string]string{"identifier": identifier},
	}
}

// WrapGitError wraps a git operation error.
func WrapGitError(err error, operation string) *CanopyError {
	return &CanopyError{
		Code:    ErrGitOperationFailed,
		Message: fmt.Sprintf("git %s failed", operation),
		Cause:   err,
		Context: map[string]string{"operation": operation},
	}
}

// NewConfigInvalid creates an error for invalid configuration.
func NewConfigInvalid(detail string) *CanopyError {
	return &CanopyError{
		Code:    ErrConfigInvalid,
		Message: fmt.Sprintf("invalid configuration: %s", detail),
		Context: map[string]string{"detail": detail},
	}
}

// Wrap wraps an error with a CanopyError, preserving the cause.
func Wrap(code ErrorCode, message string, cause error) *CanopyError {
	return &CanopyError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Sentinel errors for use with errors.Is().
var (
	WorkspaceNotFound  = &CanopyError{Code: ErrWorkspaceNotFound}
	WorkspaceExists    = &CanopyError{Code: ErrWorkspaceExists}
	RepoNotFound       = &CanopyError{Code: ErrRepoNotFound}
	RepoNotClean       = &CanopyError{Code: ErrRepoNotClean}
	RepoAlreadyExists  = &CanopyError{Code: ErrRepoAlreadyExists}
	GitOperationFailed = &CanopyError{Code: ErrGitOperationFailed}
	ConfigInvalid      = &CanopyError{Code: ErrConfigInvalid}
	UnknownRepository  = &CanopyError{Code: ErrUnknownRepository}
)
