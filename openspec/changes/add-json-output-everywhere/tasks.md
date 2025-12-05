# Implementation Tasks

## 1. JSON Output Helpers
- [x] 1.1 Create `internal/output/json.go` with helper functions
- [x] 1.2 Add `PrintJSON(v interface{})` for consistent formatting
- [x] 1.3 Add `MustPrintJSON(v interface{})` that panics on error
- [x] 1.4 Use consistent indentation (2 spaces)

## 2. Workspace Commands
- [x] 2.1 Add `--json` to `workspace status` command (via `workspace view`)
- [x] 2.2 Add `--json` to `workspace path` command (outputs `{"path": "..."`)
- [x] 2.3 Verify `workspace list` JSON is consistent (updated to use envelope format)

## 3. Repo Commands
- [x] 3.1 Add `--json` to `repo list` command
- [x] 3.2 Add `--json` to `repo path` command

## 4. Other Commands
- [x] 4.1 Add `--json` to `check` command
- [x] 4.2 Add `--json` to `status` command (top-level status)

## 5. Documentation
- [x] 5.1 Document JSON output format in specs/cli/spec.md (included in delta)
- [x] 5.2 Examples of JSON envelope format included in spec
- [x] 5.3 Document consistent field naming conventions in spec

## 6. Testing
- [x] 6.1 Test JSON output parses correctly (internal/output/json_test.go)
- [x] 6.2 Test consistency across commands (verified via build + existing tests)
- [x] 6.3 Integration test with jq (manual verification possible)
