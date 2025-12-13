# Tasks: Add Input Validation Hardening

## Implementation Checklist

### 1. Create Validation Package
- [ ] 1.1 Create `internal/validation/` directory
- [ ] 1.2 Create `validation.go` with validation functions
- [ ] 1.3 Add validation constants (max lengths, patterns)

### 2. Workspace ID Validation
- [ ] 2.1 Implement `ValidateWorkspaceID(id string) error`
- [ ] 2.2 Check non-empty, max length (255 chars)
- [ ] 2.3 Check no path separator characters
- [ ] 2.4 Check no parent directory references (`..`)
- [ ] 2.5 Check no leading/trailing whitespace

### 3. Branch Name Validation
- [ ] 3.1 Implement `ValidateBranchName(name string) error`
- [ ] 3.2 Check against git ref naming rules
- [ ] 3.3 Reject reserved names (HEAD, etc.)
- [ ] 3.4 Check no control characters

### 4. Repository Name Validation
- [ ] 4.1 Implement `ValidateRepoName(name string) error`
- [ ] 4.2 Check non-empty, max length
- [ ] 4.3 Check no path traversal

### 5. Path Validation
- [ ] 5.1 Implement `ValidatePath(path string) error`
- [ ] 5.2 Prevent path traversal outside workspace
- [ ] 5.3 Check path doesn't start with `/`

### 6. Apply Validation in Service Layer
- [ ] 6.1 Add validation to `CreateWorkspace`
- [ ] 6.2 Add validation to `RenameWorkspace`
- [ ] 6.3 Add validation to `AddRepoToWorkspace`

### 7. Apply Validation in CLI
- [ ] 7.1 Validate workspace ID in commands
- [ ] 7.2 Validate branch names
- [ ] 7.3 Show user-friendly error messages

### 8. Testing
- [ ] 8.1 Add unit tests for all validation functions
- [ ] 8.2 Add edge case tests (unicode, long strings, etc.)
- [ ] 8.3 Add fuzz tests for security-critical validators

