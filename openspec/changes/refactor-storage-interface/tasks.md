## 1. Interface Definition
- [ ] 1.1 Update `ports.WorkspaceStorage` interface in `internal/ports/storage.go`
- [ ] 1.2 Add `context.Context` as first parameter to all methods
- [ ] 1.3 Change `Create` signature to `Create(ctx, ws domain.Workspace) error`
- [ ] 1.4 Change `Save` signature to `Save(ctx, ws domain.Workspace) error`
- [ ] 1.5 Change `Load` signature to `Load(ctx, id string) (*domain.Workspace, error)`
- [ ] 1.6 Change `Close` signature to `Close(ctx, id string, closedAt time.Time) (*domain.ClosedWorkspace, error)`
- [ ] 1.7 Change `Rename` signature to `Rename(ctx, oldID, newID string) error`
- [ ] 1.8 Change `List` return type to `[]domain.Workspace`
- [ ] 1.9 Change `DeleteClosed` signature to `DeleteClosed(ctx, id string, closedAt time.Time) error`
- [ ] 1.10 Remove `LoadByID` (now redundant with new `Load`)

## 2. Mock Implementation
- [ ] 2.1 Update `internal/mocks/storage.go` to match new interface

## 3. Storage Implementation
- [ ] 3.1 Update `Engine.Create` to accept `domain.Workspace`
- [ ] 3.2 Update `Engine.Save` to look up directory by ID internally
- [ ] 3.3 Update `Engine.Load` to accept ID and resolve directory internally
- [ ] 3.4 Update `Engine.Close` to accept ID and resolve directory internally
- [ ] 3.5 Update `Engine.Rename` to accept old/new IDs
- [ ] 3.6 Update `Engine.List` to return slice
- [ ] 3.7 Update `Engine.DeleteClosed` to accept ID and timestamp
- [ ] 3.8 Add internal `resolveDirectory(id string) (string, error)` helper
- [ ] 3.9 Add internal `resolveClosedDirectory(id string, closedAt time.Time) (string, error)` helper
- [ ] 3.10 Remove `LoadByID` method

## 4. Service Layer Updates
- [ ] 4.1 Update `service.go` Create calls to pass domain object
- [ ] 4.2 Update `service.go` Save calls to pass workspace (ID used internally)
- [ ] 4.3 Update `service.go` Load calls to use new signature
- [ ] 4.4 Update `service.go` Close calls to use ID
- [ ] 4.5 Update `service.go` Rename calls to use IDs
- [ ] 4.6 Update `git_service.go` Save calls
- [ ] 4.7 Remove dirName tracking where no longer needed
- [ ] 4.8 Pass context.Context through all storage calls

## 5. Test Updates
- [ ] 5.1 Update `service_test.go` to use new interface methods
- [ ] 5.2 Update any direct Engine usage in tests
- [ ] 5.3 Update test helpers if needed

## 6. Verification
- [ ] 6.1 Run `go build ./...` to verify compilation
- [ ] 6.2 Run `go test ./...` to verify tests pass
- [ ] 6.3 Run `golangci-lint run` to check for issues
