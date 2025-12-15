package errors_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

func TestCanopyError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *cerrors.CanopyError
		expected string
	}{
		{
			name:     "without cause",
			err:      cerrors.NewWorkspaceNotFound("my-ws"),
			expected: "WORKSPACE_NOT_FOUND: workspace my-ws not found",
		},
		{
			name:     "with cause",
			err:      cerrors.WrapGitError(fmt.Errorf("network error"), "clone"),
			expected: "GIT_OPERATION_FAILED: git clone failed: network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCanopyError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := cerrors.WrapGitError(cause, "push")

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestCanopyError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "same code matches",
			err:    cerrors.NewWorkspaceNotFound("ws1"),
			target: cerrors.WorkspaceNotFound,
			want:   true,
		},
		{
			name:   "different code does not match",
			err:    cerrors.NewWorkspaceNotFound("ws1"),
			target: cerrors.RepoNotFound,
			want:   false,
		},
		{
			name:   "two workspace not found errors match",
			err:    cerrors.NewWorkspaceNotFound("ws1"),
			target: cerrors.NewWorkspaceNotFound("ws2"),
			want:   true,
		},
		{
			name:   "wrapped error matches sentinel",
			err:    fmt.Errorf("outer: %w", cerrors.NewWorkspaceNotFound("ws1")),
			target: cerrors.WorkspaceNotFound,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanopyError_As(t *testing.T) {
	wrappedErr := fmt.Errorf("outer: %w", cerrors.NewWorkspaceNotFound("my-ws"))

	var canopyErr *cerrors.CanopyError
	if !errors.As(wrappedErr, &canopyErr) {
		t.Fatal("errors.As() returned false")
	}

	if canopyErr.Code != cerrors.ErrWorkspaceNotFound {
		t.Errorf("Code = %q, want %q", canopyErr.Code, cerrors.ErrWorkspaceNotFound)
	}

	if canopyErr.Context["workspace_id"] != "my-ws" {
		t.Errorf("Context[workspace_id] = %q, want %q", canopyErr.Context["workspace_id"], "my-ws")
	}
}

func TestCanopyError_WithContext(t *testing.T) {
	err := cerrors.NewRepoNotFound("my-repo").WithContext("path", "/some/path")

	if err.Context["repo_name"] != "my-repo" {
		t.Errorf("Context[repo_name] = %q, want %q", err.Context["repo_name"], "my-repo")
	}

	if err.Context["path"] != "/some/path" {
		t.Errorf("Context[path] = %q, want %q", err.Context["path"], "/some/path")
	}
}

func TestCanopyError_WithContextDoesNotMutateOriginal(t *testing.T) {
	original := cerrors.NewRepoNotFound("my-repo")
	modified := original.WithContext("extra", "value")

	// Verify original is not mutated
	if _, ok := original.Context["extra"]; ok {
		t.Error("WithContext mutated the original error")
	}

	// Verify modified has the new context
	if modified.Context["extra"] != "value" {
		t.Errorf("modified.Context[extra] = %q, want %q", modified.Context["extra"], "value")
	}

	// Verify modified still has original context
	if modified.Context["repo_name"] != "my-repo" {
		t.Errorf("modified.Context[repo_name] = %q, want %q", modified.Context["repo_name"], "my-repo")
	}
}

func TestCanopyError_WithContextDoesNotMutateSentinel(t *testing.T) {
	// Verify that calling WithContext on a sentinel doesn't corrupt it
	modified := cerrors.WorkspaceNotFound.WithContext("key", "value")

	// Sentinel should remain unchanged
	if len(cerrors.WorkspaceNotFound.Context) > 0 {
		t.Error("WithContext mutated the sentinel error")
	}

	// Modified should have the context
	if modified.Context["key"] != "value" {
		t.Errorf("modified.Context[key] = %q, want %q", modified.Context["key"], "value")
	}
}

func TestNewWorkspaceNotFound(t *testing.T) {
	err := cerrors.NewWorkspaceNotFound("test-ws")

	if err.Code != cerrors.ErrWorkspaceNotFound {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrWorkspaceNotFound)
	}

	if err.Context["workspace_id"] != "test-ws" {
		t.Errorf("Context[workspace_id] = %q, want %q", err.Context["workspace_id"], "test-ws")
	}
}

func TestNewWorkspaceExists(t *testing.T) {
	err := cerrors.NewWorkspaceExists("existing-ws")

	if err.Code != cerrors.ErrWorkspaceExists {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrWorkspaceExists)
	}

	if err.Context["workspace_id"] != "existing-ws" {
		t.Errorf("Context[workspace_id] = %q, want %q", err.Context["workspace_id"], "existing-ws")
	}
}

func TestNewRepoNotClean(t *testing.T) {
	err := cerrors.NewRepoNotClean("dirty-repo", "close")

	if err.Code != cerrors.ErrRepoNotClean {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrRepoNotClean)
	}

	if err.Context["repo_name"] != "dirty-repo" {
		t.Errorf("Context[repo_name] = %q, want %q", err.Context["repo_name"], "dirty-repo")
	}

	if err.Context["action"] != "close" {
		t.Errorf("Context[action] = %q, want %q", err.Context["action"], "close")
	}
}

func TestNewRepoAlreadyExists(t *testing.T) {
	err := cerrors.NewRepoAlreadyExists("my-repo", "my-ws")

	if err.Code != cerrors.ErrRepoAlreadyExists {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrRepoAlreadyExists)
	}

	if err.Context["repo_name"] != "my-repo" {
		t.Errorf("Context[repo_name] = %q, want %q", err.Context["repo_name"], "my-repo")
	}

	if err.Context["workspace_id"] != "my-ws" {
		t.Errorf("Context[workspace_id] = %q, want %q", err.Context["workspace_id"], "my-ws")
	}
}

func TestNewUnknownRepository(t *testing.T) {
	t.Run("user requested", func(t *testing.T) {
		err := cerrors.NewUnknownRepository("unknown-repo", true)
		if err.Code != cerrors.ErrUnknownRepository {
			t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrUnknownRepository)
		}

		if err.Context["identifier"] != "unknown-repo" {
			t.Errorf("Context[identifier] = %q, want %q", err.Context["identifier"], "unknown-repo")
		}
	})

	t.Run("not user requested", func(t *testing.T) {
		err := cerrors.NewUnknownRepository("unknown-repo", false)
		if err.Code != cerrors.ErrUnknownRepository {
			t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrUnknownRepository)
		}
	})
}

func TestWrapGitError(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := cerrors.WrapGitError(cause, "fetch")

	if err.Code != cerrors.ErrGitOperationFailed {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrGitOperationFailed)
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}

	if err.Context["operation"] != "fetch" {
		t.Errorf("Context[operation] = %q, want %q", err.Context["operation"], "fetch")
	}
}

func TestNewConfigInvalid(t *testing.T) {
	err := cerrors.NewConfigInvalid("missing projects_root")

	if err.Code != cerrors.ErrConfigInvalid {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrConfigInvalid)
	}

	if err.Context["detail"] != "missing projects_root" {
		t.Errorf("Context[detail] = %q, want %q", err.Context["detail"], "missing projects_root")
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("some error")
	err := cerrors.Wrap(cerrors.ErrWorkspaceNotFound, "custom message", cause)

	if err.Code != cerrors.ErrWorkspaceNotFound {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrWorkspaceNotFound)
	}

	if err.Message != "custom message" {
		t.Errorf("Message = %q, want %q", err.Message, "custom message")
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors can be used for matching
	err := cerrors.NewWorkspaceNotFound("test")

	if !errors.Is(err, cerrors.WorkspaceNotFound) {
		t.Error("WorkspaceNotFound sentinel should match")
	}

	if errors.Is(err, cerrors.RepoNotFound) {
		t.Error("RepoNotFound sentinel should not match WorkspaceNotFound")
	}
}

func TestNewNotInWorkspace(t *testing.T) {
	err := cerrors.NewNotInWorkspace("/some/path")

	if err.Code != cerrors.ErrNotInWorkspace {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrNotInWorkspace)
	}

	if err.Context["path"] != "/some/path" {
		t.Errorf("Context[path] = %q, want %q", err.Context["path"], "/some/path")
	}

	if !errors.Is(err, cerrors.NotInWorkspace) {
		t.Error("NotInWorkspace sentinel should match")
	}
}

func TestNewCommandFailed(t *testing.T) {
	cause := fmt.Errorf("exit code 1")
	err := cerrors.NewCommandFailed("git push", cause)

	if err.Code != cerrors.ErrCommandFailed {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrCommandFailed)
	}

	if err.Context["command"] != "git push" {
		t.Errorf("Context[command] = %q, want %q", err.Context["command"], "git push")
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}

	if !errors.Is(err, cerrors.CommandFailed) {
		t.Error("CommandFailed sentinel should match")
	}
}

func TestNewInvalidArgument(t *testing.T) {
	err := cerrors.NewInvalidArgument("branch", "branch name is required")

	if err.Code != cerrors.ErrInvalidArgument {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrInvalidArgument)
	}

	if err.Context["argument"] != "branch" {
		t.Errorf("Context[argument] = %q, want %q", err.Context["argument"], "branch")
	}

	if err.Context["detail"] != "branch name is required" {
		t.Errorf("Context[detail] = %q, want %q", err.Context["detail"], "branch name is required")
	}

	if !errors.Is(err, cerrors.InvalidArgument) {
		t.Error("InvalidArgument sentinel should match")
	}
}

func TestNewOperationCancelled(t *testing.T) {
	err := cerrors.NewOperationCancelled("workspace creation")

	if err.Code != cerrors.ErrOperationCancelled {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrOperationCancelled)
	}

	if err.Context["operation"] != "workspace creation" {
		t.Errorf("Context[operation] = %q, want %q", err.Context["operation"], "workspace creation")
	}

	if !errors.Is(err, cerrors.OperationCancelled) {
		t.Error("OperationCancelled sentinel should match")
	}
}

func TestNewIOFailed(t *testing.T) {
	cause := fmt.Errorf("permission denied")
	err := cerrors.NewIOFailed("create directory", cause)

	if err.Code != cerrors.ErrIOFailed {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrIOFailed)
	}

	if err.Context["operation"] != "create directory" {
		t.Errorf("Context[operation] = %q, want %q", err.Context["operation"], "create directory")
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}

	if !errors.Is(err, cerrors.IOFailed) {
		t.Error("IOFailed sentinel should match")
	}
}

func TestNewRegistryError(t *testing.T) {
	cause := fmt.Errorf("file not found")
	err := cerrors.NewRegistryError("save", "could not write file", cause)

	if err.Code != cerrors.ErrRegistryError {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrRegistryError)
	}

	if err.Context["operation"] != "save" {
		t.Errorf("Context[operation] = %q, want %q", err.Context["operation"], "save")
	}

	if err.Context["detail"] != "could not write file" {
		t.Errorf("Context[detail] = %q, want %q", err.Context["detail"], "could not write file")
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}

	if !errors.Is(err, cerrors.RegistryError) {
		t.Error("RegistryError sentinel should match")
	}
}

func TestNewInternalError(t *testing.T) {
	cause := fmt.Errorf("unexpected nil pointer")
	err := cerrors.NewInternalError("app not initialized", cause)

	if err.Code != cerrors.ErrInternalError {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrInternalError)
	}

	if err.Context["detail"] != "app not initialized" {
		t.Errorf("Context[detail] = %q, want %q", err.Context["detail"], "app not initialized")
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}

	if !errors.Is(err, cerrors.InternalError) {
		t.Error("InternalError sentinel should match")
	}
}

func TestNewRepoInUse(t *testing.T) {
	workspaces := []string{"ws1", "ws2"}
	err := cerrors.NewRepoInUse("my-repo", workspaces)

	if err.Code != cerrors.ErrRepoInUse {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrRepoInUse)
	}

	if err.Context["repo_name"] != "my-repo" {
		t.Errorf("Context[repo_name] = %q, want %q", err.Context["repo_name"], "my-repo")
	}

	if !errors.Is(err, cerrors.RepoInUse) {
		t.Error("RepoInUse sentinel should match")
	}
}

func TestNewWorkspaceMetadataError(t *testing.T) {
	cause := fmt.Errorf("invalid json")
	err := cerrors.NewWorkspaceMetadataError("my-ws", "read", cause)

	if err.Code != cerrors.ErrWorkspaceMetadata {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrWorkspaceMetadata)
	}

	if err.Context["workspace_id"] != "my-ws" {
		t.Errorf("Context[workspace_id] = %q, want %q", err.Context["workspace_id"], "my-ws")
	}

	if err.Context["operation"] != "read" {
		t.Errorf("Context[operation] = %q, want %q", err.Context["operation"], "read")
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}

	if !errors.Is(err, cerrors.WorkspaceMetadata) {
		t.Error("WorkspaceMetadata sentinel should match")
	}
}

func TestNewNoReposConfigured(t *testing.T) {
	err := cerrors.NewNoReposConfigured("empty-ws")

	if err.Code != cerrors.ErrNoReposConfigured {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrNoReposConfigured)
	}

	if err.Context["workspace_id"] != "empty-ws" {
		t.Errorf("Context[workspace_id] = %q, want %q", err.Context["workspace_id"], "empty-ws")
	}

	if !errors.Is(err, cerrors.NoReposConfigured) {
		t.Error("NoReposConfigured sentinel should match")
	}
}

func TestNewMissingBranchConfig(t *testing.T) {
	err := cerrors.NewMissingBranchConfig("my-ws")

	if err.Code != cerrors.ErrMissingBranchConfig {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrMissingBranchConfig)
	}

	if err.Context["workspace_id"] != "my-ws" {
		t.Errorf("Context[workspace_id] = %q, want %q", err.Context["workspace_id"], "my-ws")
	}

	if !errors.Is(err, cerrors.MissingBranchConfig) {
		t.Error("MissingBranchConfig sentinel should match")
	}
}

func TestNewConfigValidation(t *testing.T) {
	err := cerrors.NewConfigValidation("workspace_root", "value cannot be empty")

	if err.Code != cerrors.ErrConfigValidation {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrConfigValidation)
	}

	if !strings.Contains(err.Message, "workspace_root") {
		t.Errorf("Message should contain field name, got %q", err.Message)
	}

	if !strings.Contains(err.Message, "value cannot be empty") {
		t.Errorf("Message should contain detail, got %q", err.Message)
	}

	if !errors.Is(err, cerrors.ConfigValidation) {
		t.Error("ConfigValidation sentinel should match")
	}
}

func TestNewPathInvalid(t *testing.T) {
	err := cerrors.NewPathInvalid("/invalid/path", "path contains invalid characters")

	if err.Code != cerrors.ErrPathInvalid {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrPathInvalid)
	}

	if !strings.Contains(err.Message, "/invalid/path") {
		t.Errorf("Message should contain path, got %q", err.Message)
	}

	if !strings.Contains(err.Message, "path contains invalid characters") {
		t.Errorf("Message should contain reason, got %q", err.Message)
	}

	if err.Context["path"] != "/invalid/path" {
		t.Errorf("Context[path] = %q, want %q", err.Context["path"], "/invalid/path")
	}

	if !errors.Is(err, cerrors.PathInvalid) {
		t.Error("PathInvalid sentinel should match")
	}
}

func TestNewPathNotDirectory(t *testing.T) {
	err := cerrors.NewPathNotDirectory("/some/file.txt")

	if err.Code != cerrors.ErrPathNotDirectory {
		t.Errorf("Code = %q, want %q", err.Code, cerrors.ErrPathNotDirectory)
	}

	if !strings.Contains(err.Message, "/some/file.txt") {
		t.Errorf("Message should contain path, got %q", err.Message)
	}

	if err.Context["path"] != "/some/file.txt" {
		t.Errorf("Context[path] = %q, want %q", err.Context["path"], "/some/file.txt")
	}

	if !errors.Is(err, cerrors.PathNotDirectory) {
		t.Error("PathNotDirectory sentinel should match")
	}
}
