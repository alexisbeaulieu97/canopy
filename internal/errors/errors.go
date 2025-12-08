// Package errors provides typed errors for the canopy application.
package errors

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrorCode identifies the type of error.
type ErrorCode string

// Error codes for domain errors.
const (
	ErrWorkspaceNotFound   ErrorCode = "WORKSPACE_NOT_FOUND"
	ErrWorkspaceExists     ErrorCode = "WORKSPACE_EXISTS"
	ErrRepoNotFound        ErrorCode = "REPO_NOT_FOUND"
	ErrRepoNotClean        ErrorCode = "REPO_NOT_CLEAN"
	ErrRepoAlreadyExists   ErrorCode = "REPO_ALREADY_EXISTS"
	ErrGitOperationFailed  ErrorCode = "GIT_OPERATION_FAILED"
	ErrConfigInvalid       ErrorCode = "CONFIG_INVALID"
	ErrUnknownRepository   ErrorCode = "UNKNOWN_REPOSITORY"
	ErrNotInWorkspace      ErrorCode = "NOT_IN_WORKSPACE"
	ErrCommandFailed       ErrorCode = "COMMAND_FAILED"
	ErrInvalidArgument     ErrorCode = "INVALID_ARGUMENT"
	ErrOperationCancelled  ErrorCode = "OPERATION_CANCELLED"
	ErrIOFailed            ErrorCode = "IO_FAILED"
	ErrRegistryError       ErrorCode = "REGISTRY_ERROR"
	ErrInternalError       ErrorCode = "INTERNAL_ERROR"
	ErrRepoInUse           ErrorCode = "REPO_IN_USE"
	ErrWorkspaceMetadata   ErrorCode = "WORKSPACE_METADATA_ERROR"
	ErrNoReposConfigured   ErrorCode = "NO_REPOS_CONFIGURED"
	ErrMissingBranchConfig ErrorCode = "MISSING_BRANCH_CONFIG"
	ErrHookFailed          ErrorCode = "HOOK_FAILED"
	ErrHookTimeout         ErrorCode = "HOOK_TIMEOUT"
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

// WithContext returns a copy of the error with additional context key-value pairs.
// This creates a shallow copy to avoid mutating sentinel errors.
func (e *CanopyError) WithContext(key, value string) *CanopyError {
	newContext := make(map[string]string)
	for k, v := range e.Context {
		newContext[k] = v
	}

	newContext[key] = value

	return &CanopyError{
		Code:    e.Code,
		Message: e.Message,
		Cause:   e.Cause,
		Context: newContext,
	}
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

// NewNotInWorkspace creates an error for when a command is run outside a workspace.
func NewNotInWorkspace(path string) *CanopyError {
	return &CanopyError{
		Code:    ErrNotInWorkspace,
		Message: "not inside a workspace",
		Context: map[string]string{"path": path},
	}
}

// NewCommandFailed creates an error for when a command execution fails.
func NewCommandFailed(command string, cause error) *CanopyError {
	return &CanopyError{
		Code:    ErrCommandFailed,
		Message: fmt.Sprintf("command failed: %s", command),
		Cause:   cause,
		Context: map[string]string{"command": command},
	}
}

// NewInvalidArgument creates an error for invalid input arguments.
func NewInvalidArgument(name, detail string) *CanopyError {
	return &CanopyError{
		Code:    ErrInvalidArgument,
		Message: fmt.Sprintf("invalid argument %s: %s", name, detail),
		Context: map[string]string{"argument": name, "detail": detail},
	}
}

// NewOperationCancelled creates an error for user-cancelled operations.
func NewOperationCancelled(operation string) *CanopyError {
	return &CanopyError{
		Code:    ErrOperationCancelled,
		Message: fmt.Sprintf("operation cancelled: %s", operation),
		Context: map[string]string{"operation": operation},
	}
}

// NewIOFailed creates an error for IO operation failures.
func NewIOFailed(operation string, cause error) *CanopyError {
	return &CanopyError{
		Code:    ErrIOFailed,
		Message: fmt.Sprintf("IO operation failed: %s", operation),
		Cause:   cause,
		Context: map[string]string{"operation": operation},
	}
}

// NewRegistryError creates an error for registry operations.
func NewRegistryError(operation, detail string, cause error) *CanopyError {
	return &CanopyError{
		Code:    ErrRegistryError,
		Message: fmt.Sprintf("registry %s failed: %s", operation, detail),
		Cause:   cause,
		Context: map[string]string{"operation": operation, "detail": detail},
	}
}

// NewInternalError creates an error for unexpected internal failures.
func NewInternalError(detail string, cause error) *CanopyError {
	return &CanopyError{
		Code:    ErrInternalError,
		Message: fmt.Sprintf("internal error: %s", detail),
		Cause:   cause,
		Context: map[string]string{"detail": detail},
	}
}

// NewRepoInUse creates an error for when a repo is used by workspaces.
func NewRepoInUse(name string, workspaces []string) *CanopyError {
	return &CanopyError{
		Code:    ErrRepoInUse,
		Message: fmt.Sprintf("repository %s is used by workspaces: %s. Use --force to remove", name, strings.Join(workspaces, ", ")),
		Context: map[string]string{"repo_name": name},
	}
}

// NewWorkspaceMetadataError creates an error for workspace metadata operations.
func NewWorkspaceMetadataError(workspaceID, operation string, cause error) *CanopyError {
	return &CanopyError{
		Code:    ErrWorkspaceMetadata,
		Message: fmt.Sprintf("failed to %s workspace metadata for %s", operation, workspaceID),
		Cause:   cause,
		Context: map[string]string{"workspace_id": workspaceID, "operation": operation},
	}
}

// NewNoReposConfigured creates an error for workspaces with no repos.
func NewNoReposConfigured(workspaceID string) *CanopyError {
	return &CanopyError{
		Code:    ErrNoReposConfigured,
		Message: fmt.Sprintf("no repositories configured for workspace %s", workspaceID),
		Context: map[string]string{"workspace_id": workspaceID},
	}
}

// NewMissingBranchConfig creates an error for missing branch configuration.
func NewMissingBranchConfig(workspaceID string) *CanopyError {
	return &CanopyError{
		Code:    ErrMissingBranchConfig,
		Message: fmt.Sprintf("workspace %s has no branch set in metadata", workspaceID),
		Context: map[string]string{"workspace_id": workspaceID},
	}
}

// NewHookFailed creates an error for a failed hook execution.
func NewHookFailed(index int, command string, exitCode int, repoName, stderr string) *CanopyError {
	ctx := map[string]string{
		"index":     fmt.Sprintf("%d", index),
		"command":   command,
		"exit_code": fmt.Sprintf("%d", exitCode),
	}

	if repoName != "" {
		ctx["repo_name"] = repoName
	}

	if stderr != "" {
		ctx["stderr"] = stderr
	}

	msg := fmt.Sprintf("hook[%d] failed", index)
	if repoName != "" {
		msg = fmt.Sprintf("hook[%d] failed in repo '%s'", index, repoName)
	}

	return &CanopyError{
		Code:    ErrHookFailed,
		Message: msg,
		Context: ctx,
	}
}

// NewHookTimeout creates an error for a hook that timed out.
func NewHookTimeout(index int, command string, timeout time.Duration) *CanopyError {
	return &CanopyError{
		Code:    ErrHookTimeout,
		Message: fmt.Sprintf("hook[%d] timed out after %s", index, timeout),
		Context: map[string]string{
			"index":   fmt.Sprintf("%d", index),
			"command": command,
			"timeout": timeout.String(),
		},
	}
}

// Sentinel errors for use with errors.Is().
var (
	WorkspaceNotFound   = &CanopyError{Code: ErrWorkspaceNotFound}
	WorkspaceExists     = &CanopyError{Code: ErrWorkspaceExists}
	RepoNotFound        = &CanopyError{Code: ErrRepoNotFound}
	RepoNotClean        = &CanopyError{Code: ErrRepoNotClean}
	RepoAlreadyExists   = &CanopyError{Code: ErrRepoAlreadyExists}
	GitOperationFailed  = &CanopyError{Code: ErrGitOperationFailed}
	ConfigInvalid       = &CanopyError{Code: ErrConfigInvalid}
	UnknownRepository   = &CanopyError{Code: ErrUnknownRepository}
	NotInWorkspace      = &CanopyError{Code: ErrNotInWorkspace}
	CommandFailed       = &CanopyError{Code: ErrCommandFailed}
	InvalidArgument     = &CanopyError{Code: ErrInvalidArgument}
	OperationCancelled  = &CanopyError{Code: ErrOperationCancelled}
	IOFailed            = &CanopyError{Code: ErrIOFailed}
	RegistryError       = &CanopyError{Code: ErrRegistryError}
	InternalError       = &CanopyError{Code: ErrInternalError}
	RepoInUse           = &CanopyError{Code: ErrRepoInUse}
	WorkspaceMetadata   = &CanopyError{Code: ErrWorkspaceMetadata}
	NoReposConfigured   = &CanopyError{Code: ErrNoReposConfigured}
	MissingBranchConfig = &CanopyError{Code: ErrMissingBranchConfig}
	HookFailed          = &CanopyError{Code: ErrHookFailed}
	HookTimeout         = &CanopyError{Code: ErrHookTimeout}
)
