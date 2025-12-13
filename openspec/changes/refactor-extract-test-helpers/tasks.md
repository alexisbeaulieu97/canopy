# Tasks: Extract Test Helpers to Shared Package

## Implementation Checklist

### 1. Create Test Utility Package
- [x] 1.1 Create `internal/testutil/` directory
- [x] 1.2 Create `internal/testutil/doc.go` with package documentation
- [ ] 1.3 Add build constraint to ensure test-only usage:
  ```go
  //go:build !release
  ```

### 2. Extract Git Helpers
- [x] 2.1 Create `internal/testutil/git.go`
- [x] 2.2 Extract `CreateRepoWithCommit(t, path string)`:
  - Initialize git repo
  - Configure user.email and user.name
  - Create initial commit
- [x] 2.3 Extract `RunGit(t, dir string, args ...string)`:
  - Execute git command
  - Fail test on error
- [x] 2.4 Extract `RunGitOutput(t, dir string, args ...string) string`:
  - Execute git command
  - Return trimmed output
- [x] 2.5 Extract `CloneToBare(t, sourceRepo, destPath string)`:
  - Clone repo as bare
  - Return repository handle

### 3. Extract Filesystem Helpers
- [x] 3.1 Create `internal/testutil/fs.go`
- [x] 3.2 Extract `MustMkdir(t, path string)`:
  - Create directory
  - Fail test on error
- [x] 3.3 Add `MustWriteFile(t, path, content string)`:
  - Write file with content
  - Fail test on error
- [x] 3.4 Add `MustReadFile(t, path string) string`:
  - Read file content
  - Fail test on error

### 4. Extract Test Service Setup
- [ ] 4.1 Create `internal/testutil/service.go`
- [ ] 4.2 Extract `NewTestService(t) *TestServiceDeps`:
  - Create temp directories
  - Initialize config, git engine, workspace engine
  - Return struct with all dependencies
- [ ] 4.3 Add cleanup registration with `t.Cleanup()`

### 5. Update Existing Tests
- [x] 5.1 Update `internal/workspaces/service_test.go`:
  - Import testutil package
  - Replace local helpers with shared ones
  - Remove duplicate helper functions
- [ ] 5.2 Update `internal/gitx/git_test.go`:
  - Import testutil package
  - Replace local helpers with shared ones
- [ ] 5.3 Update any other test files with duplicated helpers

### 6. Documentation
- [x] 6.1 Add godoc comments to all exported helpers
- [x] 6.2 Add usage examples in doc.go
- [ ] 6.3 Update CONTRIBUTING.md (if exists) with test helper guidance

### 7. Verification
- [ ] 7.1 Run all tests to verify no regressions
- [ ] 7.2 Verify no duplicate helper functions remain
- [ ] 7.3 Run linter on new package
