# Change: Extract Test Helpers to Shared Package

## Why
Test helper functions like `createRepoWithCommit`, `runGit`, and `runGitOutput` are duplicated across multiple test files (`service_test.go`, `git_test.go`). This duplication leads to maintenance burden and inconsistent behavior. Extracting to a shared package improves maintainability and consistency.

## What Changes
- Create `internal/testutil/` package for shared test helpers
- Extract git test helpers (repo creation, git commands)
- Extract filesystem helpers (temp dirs, file creation)
- Update existing tests to use shared helpers
- Add documentation for test utilities

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/testutil/git.go` - New file for git helpers
  - `internal/testutil/fs.go` - New file for filesystem helpers
  - `internal/workspaces/service_test.go` - Use shared helpers
  - `internal/gitx/git_test.go` - Use shared helpers
  - `internal/workspace/workspace_test.go` - Use shared helpers (if applicable)
- **Risk**: Very Low - Test-only refactor, no production code changes

