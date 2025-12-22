// Package validation provides centralized input validation functions
// to prevent security issues like path traversal and ensure consistent UX.
package validation

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// Validation constants.
const (
	// MaxWorkspaceIDLength is the maximum allowed length for workspace IDs.
	MaxWorkspaceIDLength = 255

	// MaxBranchNameLength is the maximum allowed length for branch names.
	MaxBranchNameLength = 255

	// MaxRepoNameLength is the maximum allowed length for repository names.
	MaxRepoNameLength = 255
)

// NormalizeWorkspaceDirName validates and normalizes a workspace directory name.
// Returns the cleaned directory name or an error if invalid.
func NormalizeWorkspaceDirName(name string) (string, error) {
	cleaned := filepath.Clean(strings.TrimSpace(name))
	if cleaned == "" || cleaned == "." {
		return "", cerrors.NewInvalidArgument("workspace_dir", "workspace directory name cannot be empty")
	}

	if filepath.IsAbs(cleaned) {
		return "", cerrors.NewInvalidArgument("workspace_dir", "workspace directory name must be relative")
	}

	if strings.ContainsRune(cleaned, '/') || strings.ContainsRune(cleaned, '\\') {
		return "", cerrors.NewInvalidArgument("workspace_dir", "workspace directory name cannot contain path separators")
	}

	if strings.Contains(cleaned, "..") {
		return "", cerrors.NewInvalidArgument("workspace_dir", "workspace directory name cannot contain path traversal sequences (..)")
	}

	if len(cleaned) > MaxWorkspaceIDLength {
		return "", cerrors.NewInvalidArgument("workspace_dir", "workspace directory name exceeds maximum length of 255 characters")
	}

	for _, r := range cleaned {
		if unicode.IsControl(r) {
			return "", cerrors.NewInvalidArgument("workspace_dir", "workspace directory name cannot contain control characters")
		}
	}

	return cleaned, nil
}

// Git ref reserved names that cannot be used as branch names.
var gitReservedNames = map[string]bool{
	"HEAD":             true,
	"head":             true,
	"FETCH_HEAD":       true,
	"ORIG_HEAD":        true,
	"MERGE_HEAD":       true,
	"CHERRY_PICK_HEAD": true,
}

// gitRefInvalidPatterns contains patterns that are invalid in git ref names.
// Based on git-check-ref-format rules.
var gitRefInvalidPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.\.`),            // double dots
	regexp.MustCompile(`^\.`),             // starts with dot
	regexp.MustCompile(`\.$`),             // ends with dot
	regexp.MustCompile(`\.lock$`),         // ends with .lock
	regexp.MustCompile(`@\{`),             // @{ sequence
	regexp.MustCompile(`[\x00-\x1f\x7f]`), // control characters
	regexp.MustCompile(`[~^:?*\[\\]`),     // special characters
	regexp.MustCompile(`\s`),              // whitespace
}

// ValidateWorkspaceID validates a workspace ID string.
// Returns an error if the ID is invalid, nil otherwise.
func ValidateWorkspaceID(id string) error {
	// Check for empty
	if id == "" {
		return cerrors.NewInvalidArgument("workspace-id", "cannot be empty")
	}

	// Check for leading/trailing whitespace
	if strings.TrimSpace(id) != id {
		return cerrors.NewInvalidArgument("workspace-id", "cannot have leading or trailing whitespace")
	}

	// Check length
	if len(id) > MaxWorkspaceIDLength {
		return cerrors.NewInvalidArgument("workspace-id", "exceeds maximum length of 255 characters")
	}

	// Check for path separators (both Unix and Windows)
	if strings.ContainsRune(id, '/') || strings.ContainsRune(id, '\\') {
		return cerrors.NewInvalidArgument("workspace-id", "cannot contain path separators")
	}

	// Check for parent directory reference
	if strings.Contains(id, "..") {
		return cerrors.NewInvalidArgument("workspace-id", "cannot contain path traversal sequences (..)")
	}

	// Check for control characters
	for _, r := range id {
		if unicode.IsControl(r) {
			return cerrors.NewInvalidArgument("workspace-id", "cannot contain control characters")
		}
	}

	return nil
}

// ValidateBranchName validates a git branch name.
// Returns an error if the name is invalid, nil otherwise.
func ValidateBranchName(name string) error {
	// Empty branch names are allowed (will default to workspace ID)
	if name == "" {
		return nil
	}

	// Check length
	if len(name) > MaxBranchNameLength {
		return cerrors.NewInvalidArgument("branch", "exceeds maximum length of 255 characters")
	}

	// Check for reserved names
	if gitReservedNames[name] {
		return cerrors.NewInvalidArgument("branch", "reserved name not allowed: "+name)
	}

	// Check against git ref invalid patterns
	for _, pattern := range gitRefInvalidPatterns {
		if pattern.MatchString(name) {
			return cerrors.NewInvalidArgument("branch", "contains invalid characters or sequences for git refs")
		}
	}

	// Check for leading/trailing slashes
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return cerrors.NewInvalidArgument("branch", "cannot start or end with /")
	}

	// Check for consecutive slashes
	if strings.Contains(name, "//") {
		return cerrors.NewInvalidArgument("branch", "cannot contain consecutive slashes")
	}

	return nil
}

// ValidateRepoName validates a repository name.
// Returns an error if the name is invalid, nil otherwise.
func ValidateRepoName(name string) error {
	// Check for empty
	if name == "" {
		return cerrors.NewInvalidArgument("repo-name", "cannot be empty")
	}

	// Check for leading/trailing whitespace
	if strings.TrimSpace(name) != name {
		return cerrors.NewInvalidArgument("repo-name", "cannot have leading or trailing whitespace")
	}

	// Check length
	if len(name) > MaxRepoNameLength {
		return cerrors.NewInvalidArgument("repo-name", "exceeds maximum length of 255 characters")
	}

	// Check for path separators (repo names should be simple identifiers)
	if strings.ContainsRune(name, '/') || strings.ContainsRune(name, '\\') {
		return cerrors.NewInvalidArgument("repo-name", "cannot contain path separators")
	}

	// Check for parent directory reference
	if strings.Contains(name, "..") {
		return cerrors.NewInvalidArgument("repo-name", "cannot contain path traversal sequences (..)")
	}

	// Check for control characters
	for _, r := range name {
		if unicode.IsControl(r) {
			return cerrors.NewInvalidArgument("repo-name", "cannot contain control characters")
		}
	}

	return nil
}

// ValidatePath validates a path to prevent path traversal attacks.
// The path must be relative and not attempt to escape the expected directory.
// Returns an error if the path is invalid, nil otherwise.
func ValidatePath(path string) error {
	// Check for empty
	if path == "" {
		return cerrors.NewInvalidArgument("path", "cannot be empty")
	}

	// Check for absolute paths (not allowed for user-provided paths)
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return cerrors.NewInvalidArgument("path", "absolute paths not allowed")
	}

	// On Windows, also check for drive letters
	if len(path) >= 2 && path[1] == ':' {
		return cerrors.NewInvalidArgument("path", "absolute paths not allowed")
	}

	// Check for parent directory traversal
	if strings.Contains(path, "..") {
		return cerrors.NewInvalidArgument("path", "path traversal sequences (..) not allowed")
	}

	// Check for control characters
	for _, r := range path {
		if unicode.IsControl(r) {
			return cerrors.NewInvalidArgument("path", "cannot contain control characters")
		}
	}

	return nil
}
