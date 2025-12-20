## 1. Preparation
- [x] 1.1 Identify shared helpers used across subcommands (getApp, output formatting)
- [x] 1.2 Create `presenters.go` with output formatting helpers
- [x] 1.3 Ensure all existing tests pass before refactoring

## 2. Extract Subcommands
- [x] 2.1 Extract `workspace new` to `workspace_new.go`
- [x] 2.2 Extract `workspace list` to `workspace_list.go`
- [x] 2.3 Extract `workspace close` to `workspace_close.go`
- [x] 2.4 Extract `workspace view` to `workspace_view.go`
- [x] 2.5 Extract `workspace rename` to `workspace_rename.go`
- [x] 2.6 Extract `workspace export/import` to `workspace_export.go`
- [x] 2.7 Extract `workspace repo add/remove` to `workspace_repo.go`
- [x] 2.8 Extract `workspace git/switch/update/path/reopen` to appropriate files

## 3. Cleanup
- [x] 3.1 Remove extracted code from `workspace.go`
- [x] 3.2 Keep only parent command definition and `init()` in `workspace.go`
- [x] 3.3 Verify all imports are correct in new files
- [x] 3.4 Run `go build` and `go test` to verify no regressions

## 4. Documentation
- [x] 4.1 Update `docs/architecture.md` to reflect new file structure
- [x] 4.2 Add comments in each file explaining its responsibility
