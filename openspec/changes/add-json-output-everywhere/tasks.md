# Implementation Tasks

## 1. JSON Output Helpers
- [ ] 1.1 Create `internal/output/json.go` with helper functions
- [ ] 1.2 Add `PrintJSON(v interface{})` for consistent formatting
- [ ] 1.3 Add `MustPrintJSON(v interface{})` that panics on error
- [ ] 1.4 Use consistent indentation (2 spaces)

## 2. Workspace Commands
- [ ] 2.1 Add `--json` to `workspace status` command
- [ ] 2.2 Add `--json` to `workspace path` command (outputs `{"path": "..."`)
- [ ] 2.3 Verify `workspace list` JSON is consistent

## 3. Repo Commands
- [ ] 3.1 Add `--json` to `repo list` command
- [ ] 3.2 Add `--json` to `repo status` command (if implementing)

## 4. Other Commands
- [ ] 4.1 Add `--json` to `check` command
- [ ] 4.2 Add `--json` to `template list` (if implementing templates)

## 5. Documentation
- [ ] 5.1 Document JSON output format in README
- [ ] 5.2 Add examples of using with jq
- [ ] 5.3 Document consistent field naming conventions

## 6. Testing
- [ ] 6.1 Test JSON output parses correctly
- [ ] 6.2 Test consistency across commands
- [ ] 6.3 Integration test with jq
