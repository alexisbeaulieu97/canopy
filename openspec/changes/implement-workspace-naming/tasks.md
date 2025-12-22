## 1. Configuration
- [x] 1.1 Add template validation at config load time
- [x] 1.2 Add `ComputeWorkspaceDir(id string)` method to config
- [x] 1.3 Validate template produces valid directory names

## 2. Workspace Creation
- [x] 2.1 Update `create.go` to use computed directory name
- [x] 2.2 Update storage layer to respect naming template
- [x] 2.3 Centralize directory creation (remove double-mkdir)

## 3. Workspace Lookup
- [x] 3.1 Ensure lookup works with custom naming
- [x] 3.2 Store original ID in metadata for reverse lookup

## 4. CLI Enhancements
- [x] 4.1 Show computed directory name in `config validate`
- [x] 4.2 Add example preview with sample workspace ID

## 5. Testing
- [x] 5.1 Add unit tests for template computation
- [x] 5.2 Add integration tests for custom naming
- [x] 5.3 Test edge cases (special chars, long names)

## 6. Documentation
- [x] 6.1 Update docs/configuration.md with working examples
- [x] 6.2 Document available template variables
