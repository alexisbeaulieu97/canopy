package errors_test

import (
	"errors"
	"fmt"
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
