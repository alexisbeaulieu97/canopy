package validation_test

import (
	"errors"
	"strings"
	"testing"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

func TestValidateWorkspaceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      string
		wantErr bool
		errCode cerrors.ErrorCode
	}{
		// Valid cases
		{name: "simple id", id: "my-workspace", wantErr: false},
		{name: "with numbers", id: "workspace123", wantErr: false},
		{name: "with underscores", id: "my_workspace", wantErr: false},
		{name: "with dots", id: "my.workspace", wantErr: false},
		{name: "unicode characters", id: "workspace-日本語", wantErr: false},
		{name: "single character", id: "a", wantErr: false},
		{name: "max length", id: strings.Repeat("a", 255), wantErr: false},

		// Invalid cases
		{name: "empty", id: "", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "leading whitespace", id: " workspace", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "trailing whitespace", id: "workspace ", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "exceeds max length", id: strings.Repeat("a", 256), wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains forward slash", id: "my/workspace", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains backslash", id: "my\\workspace", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "parent dir reference", id: "..", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "parent dir in path", id: "my..workspace", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "control character", id: "workspace\x00name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "newline", id: "workspace\nname", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "tab", id: "workspace\tname", wantErr: true, errCode: cerrors.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validation.ValidateWorkspaceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWorkspaceID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errCode != "" {
				var canopyErr *cerrors.CanopyError
				if errors.As(err, &canopyErr) {
					if canopyErr.Code != tt.errCode {
						t.Errorf("ValidateWorkspaceID(%q) error code = %v, want %v", tt.id, canopyErr.Code, tt.errCode)
					}
				} else {
					t.Errorf("ValidateWorkspaceID(%q) expected CanopyError, got %T", tt.id, err)
				}
			}
		})
	}
}

func TestValidateBranchName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		branch  string
		wantErr bool
		errCode cerrors.ErrorCode
	}{
		// Valid cases
		{name: "empty (allowed)", branch: "", wantErr: false},
		{name: "simple branch", branch: "feature", wantErr: false},
		{name: "with slash", branch: "feature/my-branch", wantErr: false},
		{name: "with numbers", branch: "feature-123", wantErr: false},
		{name: "with underscore", branch: "feature_branch", wantErr: false},
		{name: "release branch", branch: "release/v1.0.0", wantErr: false},

		// Invalid cases - reserved names
		{name: "HEAD reserved", branch: "HEAD", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "head reserved", branch: "head", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "FETCH_HEAD reserved", branch: "FETCH_HEAD", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "ORIG_HEAD reserved", branch: "ORIG_HEAD", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "MERGE_HEAD reserved", branch: "MERGE_HEAD", wantErr: true, errCode: cerrors.ErrInvalidArgument},

		// Invalid cases - git ref rules
		{name: "starts with dot", branch: ".branch", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "ends with dot", branch: "branch.", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "double dots", branch: "branch..name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "ends with .lock", branch: "branch.lock", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains @{", branch: "branch@{", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains tilde", branch: "branch~1", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains caret", branch: "branch^1", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains colon", branch: "branch:name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains question", branch: "branch?name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains asterisk", branch: "branch*name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains bracket", branch: "branch[name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains backslash", branch: "branch\\name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains space", branch: "branch name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains tab", branch: "branch\tname", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains null", branch: "branch\x00name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "starts with slash", branch: "/branch", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "ends with slash", branch: "branch/", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "consecutive slashes", branch: "branch//name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "exceeds max length", branch: strings.Repeat("a", 256), wantErr: true, errCode: cerrors.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validation.ValidateBranchName(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranchName(%q) error = %v, wantErr %v", tt.branch, err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errCode != "" {
				var canopyErr *cerrors.CanopyError
				if errors.As(err, &canopyErr) {
					if canopyErr.Code != tt.errCode {
						t.Errorf("ValidateBranchName(%q) error code = %v, want %v", tt.branch, canopyErr.Code, tt.errCode)
					}
				} else {
					t.Errorf("ValidateBranchName(%q) expected CanopyError, got %T", tt.branch, err)
				}
			}
		})
	}
}

func TestValidateRepoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repoName string
		wantErr  bool
		errCode  cerrors.ErrorCode
	}{
		// Valid cases
		{name: "simple name", repoName: "my-repo", wantErr: false},
		{name: "with numbers", repoName: "repo123", wantErr: false},
		{name: "with underscores", repoName: "my_repo", wantErr: false},
		{name: "with dots", repoName: "my.repo", wantErr: false},
		{name: "max length", repoName: strings.Repeat("a", 255), wantErr: false},

		// Invalid cases
		{name: "empty", repoName: "", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "exceeds max length", repoName: strings.Repeat("a", 256), wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains forward slash", repoName: "my/repo", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "contains backslash", repoName: "my\\repo", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "parent dir reference", repoName: "..", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "parent dir embedded", repoName: "my..repo", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "control character", repoName: "repo\x00name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "newline", repoName: "repo\nname", wantErr: true, errCode: cerrors.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validation.ValidateRepoName(tt.repoName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepoName(%q) error = %v, wantErr %v", tt.repoName, err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errCode != "" {
				var canopyErr *cerrors.CanopyError
				if errors.As(err, &canopyErr) {
					if canopyErr.Code != tt.errCode {
						t.Errorf("ValidateRepoName(%q) error code = %v, want %v", tt.repoName, canopyErr.Code, tt.errCode)
					}
				} else {
					t.Errorf("ValidateRepoName(%q) expected CanopyError, got %T", tt.repoName, err)
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errCode cerrors.ErrorCode
	}{
		// Valid cases
		{name: "simple path", path: "subdir", wantErr: false},
		{name: "nested path", path: "subdir/nested", wantErr: false},
		{name: "with dots in name", path: "file.txt", wantErr: false},
		{name: "relative current dir", path: ".", wantErr: false},

		// Invalid cases
		{name: "empty", path: "", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "absolute unix path", path: "/etc/passwd", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "absolute windows path", path: "C:\\Windows", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "parent directory", path: "..", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "parent in path", path: "../secret", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "parent in nested path", path: "subdir/../secret", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "control character", path: "path\x00name", wantErr: true, errCode: cerrors.ErrInvalidArgument},
		{name: "newline in path", path: "path\nname", wantErr: true, errCode: cerrors.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validation.ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errCode != "" {
				var canopyErr *cerrors.CanopyError
				if errors.As(err, &canopyErr) {
					if canopyErr.Code != tt.errCode {
						t.Errorf("ValidatePath(%q) error code = %v, want %v", tt.path, canopyErr.Code, tt.errCode)
					}
				} else {
					t.Errorf("ValidatePath(%q) expected CanopyError, got %T", tt.path, err)
				}
			}
		})
	}
}

// Fuzz tests for security-critical validators

func FuzzValidateWorkspaceID(f *testing.F) {
	// Seed with interesting inputs
	f.Add("valid-workspace")
	f.Add("")
	f.Add("../secret")
	f.Add("/etc/passwd")
	f.Add("workspace\x00name")
	f.Add(strings.Repeat("a", 300))
	f.Add(" leading")
	f.Add("trailing ")
	f.Add("with/slash")
	f.Add("with\\backslash")

	f.Fuzz(func(t *testing.T, id string) {
		// Should never panic
		err := validation.ValidateWorkspaceID(id)

		// If valid, verify invariants
		if err == nil {
			if id == "" {
				t.Error("empty ID should be invalid")
			}
			if strings.Contains(id, "/") || strings.Contains(id, "\\") {
				t.Error("ID with path separators should be invalid")
			}
			if strings.Contains(id, "..") {
				t.Error("ID with .. should be invalid")
			}
			if len(id) > 255 {
				t.Error("ID exceeding 255 chars should be invalid")
			}
			if strings.TrimSpace(id) != id {
				t.Error("ID with leading/trailing whitespace should be invalid")
			}
		}
	})
}

func FuzzValidateBranchName(f *testing.F) {
	// Seed with interesting inputs
	f.Add("feature")
	f.Add("")
	f.Add("HEAD")
	f.Add(".hidden")
	f.Add("branch.lock")
	f.Add("branch@{")
	f.Add("branch~1")
	f.Add("branch\x00name")
	f.Add(strings.Repeat("a", 300))

	f.Fuzz(func(t *testing.T, name string) {
		// Should never panic
		_ = validation.ValidateBranchName(name)
	})
}

func FuzzValidateRepoName(f *testing.F) {
	// Seed with interesting inputs
	f.Add("my-repo")
	f.Add("")
	f.Add("../secret")
	f.Add("repo\x00name")
	f.Add(strings.Repeat("a", 300))
	f.Add("with/slash")

	f.Fuzz(func(t *testing.T, name string) {
		// Should never panic
		err := validation.ValidateRepoName(name)

		// If valid, verify invariants
		if err == nil {
			if name == "" {
				t.Error("empty repo name should be invalid")
			}
			if strings.Contains(name, "/") || strings.Contains(name, "\\") {
				t.Error("repo name with path separators should be invalid")
			}
			if strings.Contains(name, "..") {
				t.Error("repo name with .. should be invalid")
			}
			if len(name) > 255 {
				t.Error("repo name exceeding 255 chars should be invalid")
			}
		}
	})
}

func FuzzValidatePath(f *testing.F) {
	// Seed with interesting inputs
	f.Add("subdir")
	f.Add("")
	f.Add("../secret")
	f.Add("/etc/passwd")
	f.Add("C:\\Windows")
	f.Add("path\x00name")
	f.Add("normal/nested/path")

	f.Fuzz(func(t *testing.T, path string) {
		// Should never panic
		err := validation.ValidatePath(path)

		// If valid, verify invariants
		if err == nil {
			if path == "" {
				t.Error("empty path should be invalid")
			}
			if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
				t.Error("absolute path should be invalid")
			}
			if strings.Contains(path, "..") {
				t.Error("path with .. should be invalid")
			}
		}
	})
}
