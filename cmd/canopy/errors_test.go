package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

func TestExitCodeForError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ExitCode
	}{
		{
			name:     "workspace not found",
			err:      cerrors.NewWorkspaceNotFound("test-ws"),
			expected: ExitNotFound,
		},
		{
			name:     "repo not found",
			err:      cerrors.NewRepoNotFound("test-repo"),
			expected: ExitNotFound,
		},
		{
			name:     "workspace exists",
			err:      cerrors.NewWorkspaceExists("test-ws"),
			expected: ExitAlreadyExists,
		},
		{
			name:     "repo already exists",
			err:      cerrors.NewRepoAlreadyExists("test-repo", "test-ws"),
			expected: ExitAlreadyExists,
		},
		{
			name:     "repo not clean",
			err:      cerrors.NewRepoNotClean("test-repo", "close"),
			expected: ExitDirtyWorkspace,
		},
		{
			name:     "config invalid",
			err:      cerrors.NewConfigInvalid("missing field"),
			expected: ExitConfigError,
		},
		{
			name:     "git operation failed",
			err:      cerrors.WrapGitError(fmt.Errorf("network error"), "push"),
			expected: ExitGitError,
		},
		{
			name:     "unknown repository",
			err:      cerrors.NewUnknownRepository("unknown", true),
			expected: ExitUnknownResource,
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("some error"),
			expected: ExitGeneralError,
		},
		{
			name:     "wrapped canopy error",
			err:      fmt.Errorf("outer: %w", cerrors.NewWorkspaceNotFound("test-ws")),
			expected: ExitNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exitCodeForError(tt.err)
			if got != tt.expected {
				t.Errorf("exitCodeForError() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{
			name:     "workspace not found",
			err:      cerrors.NewWorkspaceNotFound("test-ws"),
			contains: "workspace test-ws not found",
		},
		{
			name:     "repo not clean",
			err:      cerrors.NewRepoNotClean("dirty-repo", "close"),
			contains: "dirty-repo has uncommitted changes",
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("some generic error"),
			contains: "some generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := userFriendlyMessage(tt.err)
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("userFriendlyMessage() = %q, want containing %q", msg, tt.contains)
			}
		})
	}
}

func TestFormatErrorJSON(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantCode     string
		wantContains string
	}{
		{
			name:         "canopy error with code",
			err:          cerrors.NewWorkspaceNotFound("test-ws"),
			wantCode:     "WORKSPACE_NOT_FOUND",
			wantContains: "workspace test-ws not found",
		},
		{
			name:         "canopy error with cause",
			err:          cerrors.WrapGitError(fmt.Errorf("network error"), "push"),
			wantCode:     "GIT_OPERATION_FAILED",
			wantContains: "network error",
		},
		{
			name:         "generic error",
			err:          fmt.Errorf("generic error"),
			wantContains: "generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json := formatErrorJSON(tt.err)
			if tt.wantCode != "" && !strings.Contains(json, tt.wantCode) {
				t.Errorf("formatErrorJSON() missing code %q in %s", tt.wantCode, json)
			}
			if !strings.Contains(json, tt.wantContains) {
				t.Errorf("formatErrorJSON() missing %q in %s", tt.wantContains, json)
			}
		})
	}
}

func TestIsCanopyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "canopy error",
			err:  cerrors.NewWorkspaceNotFound("test"),
			want: true,
		},
		{
			name: "wrapped canopy error",
			err:  fmt.Errorf("outer: %w", cerrors.NewWorkspaceNotFound("test")),
			want: true,
		},
		{
			name: "generic error",
			err:  fmt.Errorf("generic"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCanopyError(tt.err); got != tt.want {
				t.Errorf("isCanopyError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCanopyError(t *testing.T) {
	t.Run("returns canopy error", func(t *testing.T) {
		err := cerrors.NewWorkspaceNotFound("test")
		got := getCanopyError(err)
		if got == nil {
			t.Fatal("getCanopyError() = nil, want non-nil")
		}
		if got.Code != cerrors.ErrWorkspaceNotFound {
			t.Errorf("Code = %q, want %q", got.Code, cerrors.ErrWorkspaceNotFound)
		}
	})

	t.Run("extracts from wrapped error", func(t *testing.T) {
		wrapped := fmt.Errorf("outer: %w", cerrors.NewRepoNotFound("test"))
		got := getCanopyError(wrapped)
		if got == nil {
			t.Fatal("getCanopyError() = nil, want non-nil")
		}
		if got.Code != cerrors.ErrRepoNotFound {
			t.Errorf("Code = %q, want %q", got.Code, cerrors.ErrRepoNotFound)
		}
	})

	t.Run("returns nil for generic error", func(t *testing.T) {
		err := errors.New("generic")
		if got := getCanopyError(err); got != nil {
			t.Errorf("getCanopyError() = %v, want nil", got)
		}
	})
}
