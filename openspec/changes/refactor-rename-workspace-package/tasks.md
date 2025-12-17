## 1. Package Rename
- [ ] 1.1 Rename `internal/workspace/` directory to `internal/storage/`
- [ ] 1.2 Update package declaration in all files from `package workspace` to `package storage`
- [ ] 1.3 Rename `workspace.go` to `storage.go` (primary file)

## 2. Import Updates
- [ ] 2.1 Update import in `internal/app/app.go`
- [ ] 2.2 Update import in `internal/workspaces/service_test.go`
- [ ] 2.3 Search for any other references and update

## 3. Documentation
- [ ] 3.1 Update `openspec/project.md` architecture section

## 4. Verification
- [ ] 4.1 Run `go build ./...` to verify no broken imports
- [ ] 4.2 Run `go test ./...` to verify tests pass
- [ ] 4.3 Run `golangci-lint run` to check for issues
