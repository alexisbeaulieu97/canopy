# Tasks: Remove Deprecated Code

## Implementation Checklist

### Phase 1: Audit Deprecated Code
- [ ] Search for `Deprecated:` comments in codebase
- [ ] List all deprecated items:
  - [ ] `workspace.ClosedWorkspace` alias
  - [ ] (Add others if found)
- [ ] Identify consumers of deprecated code

### Phase 2: Update Consumers
- [ ] Search for `workspace.ClosedWorkspace` usage
- [ ] Replace with `domain.ClosedWorkspace`
- [ ] Update imports as needed

### Phase 3: Remove Deprecated Code
- [ ] Remove `ClosedWorkspace` type alias from `workspace/workspace.go`
- [ ] Remove associated deprecated comment

### Phase 4: Verify
- [ ] Run `go build ./...` to ensure no compile errors
- [ ] Run `go test ./...` to ensure tests pass
- [ ] Run `golangci-lint run` to check for issues
