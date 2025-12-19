## 1. Preparation
- [ ] 1.1 Identify shared helpers used across subcommands (getApp, output formatting)
- [ ] 1.2 Create `presenters.go` with output formatting helpers
- [ ] 1.3 Ensure all existing tests pass before refactoring

## 2. Extract Subcommands
- [ ] 2.1 Extract `workspace new` to `workspace_new.go`
- [ ] 2.2 Extract `workspace list` to `workspace_list.go`
- [ ] 2.3 Extract `workspace close` to `workspace_close.go`
- [ ] 2.4 Extract `workspace view` to `workspace_view.go`
- [ ] 2.5 Extract `workspace rename` to `workspace_rename.go`
- [ ] 2.6 Extract `workspace export/import` to `workspace_export.go`
- [ ] 2.7 Extract `workspace repo add/remove` to `workspace_repo.go`
- [ ] 2.8 Extract `workspace git/switch/update/path/reopen` to appropriate files

## 3. Cleanup
- [ ] 3.1 Remove extracted code from `workspace.go`
- [ ] 3.2 Keep only parent command definition and `init()` in `workspace.go`
- [ ] 3.3 Verify all imports are correct in new files
- [ ] 3.4 Run `go build` and `go test` to verify no regressions

## 4. Documentation
- [ ] 4.1 Update `docs/architecture.md` to reflect new file structure
- [ ] 4.2 Add comments in each file explaining its responsibility
