# Tasks: Add Input Validation Hardening

## Implementation Checklist

### 1. Create Validation Package
- [x] 1.1 Create `internal/validation/` directory
- [x] 1.2 Create `validation.go` with validation functions
- [x] 1.3 Add validation constants (max lengths, patterns)

### 2. Workspace ID Validation
- [x] 2.1 Implement `ValidateWorkspaceID(id string) error`
- [x] 2.2 Check non-empty, max length (255 chars)
- [x] 2.3 Check no path separator characters
- [x] 2.4 Check no parent directory references (`..`)
- [x] 2.5 Check no leading/trailing whitespace

### 3. Branch Name Validation
- [x] 3.1 Implement `ValidateBranchName(name string) error`
- [x] 3.2 Check against git ref naming rules
- [x] 3.3 Reject reserved names (HEAD, etc.)
- [x] 3.4 Check no control characters

### 4. Repository Name Validation
- [x] 4.1 Implement `ValidateRepoName(name string) error`
- [x] 4.2 Check non-empty, max length
- [x] 4.3 Check no path traversal

### 5. Path Validation
- [x] 5.1 Implement `ValidatePath(path string) error`
- [x] 5.2 Prevent path traversal outside workspace
- [x] 5.3 Check path doesn't start with `/`

### 6. Apply Validation in Service Layer
- [x] 6.1 Add validation to `CreateWorkspace`
- [x] 6.2 Add validation to `RenameWorkspace`
- [x] 6.3 Add validation to `AddRepoToWorkspace`

### 7. Apply Validation in CLI
- [x] 7.1 Validate workspace ID in commands
- [x] 7.2 Validate branch names
- [x] 7.3 Show user-friendly error messages

### 8. Testing
- [x] 8.1 Add unit tests for all validation functions
- [x] 8.2 Add edge case tests (unicode, long strings, etc.)
- [x] 8.3 Add fuzz tests for security-critical validators

