## 1. Interface Definition
- [x] 1.1 Update `ports.WorkspaceStorage` interface in `internal/ports/storage.go`
- [x] 1.2 Add `context.Context` as first parameter to all methods
- [x] 1.3 Change `Create` signature to `Create(ctx, ws domain.Workspace) error`
- [x] 1.4 Change `Save` signature to `Save(ctx, ws domain.Workspace) error`
- [x] 1.5 Change `Load` signature to `Load(ctx, id string) (*domain.Workspace, error)`
- [x] 1.6 Change `Close` signature to `Close(ctx, id string, closedAt time.Time) (*domain.ClosedWorkspace, error)`
- [x] 1.7 Change `Rename` signature to `Rename(ctx, oldID, newID string) error`
- [x] 1.8 Change `List` return type to `[]domain.Workspace`
- [x] 1.9 Change `DeleteClosed` signature to `DeleteClosed(ctx, id string, closedAt time.Time) error`
- [x] 1.10 Remove `LoadByID` (now redundant with new `Load`)

## 2. Mock Implementation
- [x] 2.1 Update `internal/mocks/storage.go` to match new interface

## 3. Storage Implementation
- [x] 3.1 Update `Engine.Create` to accept `domain.Workspace`
- [x] 3.2 Update `Engine.Save` to look up directory by ID internally
- [x] 3.3 Update `Engine.Load` to accept ID and resolve directory internally
- [x] 3.4 Update `Engine.Close` to accept ID and resolve directory internally
- [x] 3.5 Update `Engine.Rename` to accept old/new IDs
- [x] 3.6 Update `Engine.List` to return slice
- [x] 3.7 Update `Engine.DeleteClosed` to accept ID and timestamp
- [x] 3.8 Add internal `resolveDirectory(id string) (string, error)` helper
- [x] 3.9 Add internal `resolveClosedDirectory(id string, closedAt time.Time) (string, error)` helper
- [x] 3.10 Remove `LoadByID` method

## 4. Service Layer Updates
- [x] 4.1 Update `service.go` Create calls to pass domain object
- [x] 4.2 Update `service.go` Save calls to pass workspace (ID used internally)
- [x] 4.3 Update `service.go` Load calls to use new signature
- [x] 4.4 Update `service.go` Close calls to use ID
- [x] 4.5 Update `service.go` Rename calls to use IDs
- [x] 4.6 Update `git_service.go` Save calls
- [x] 4.7 Remove dirName tracking where no longer needed
- [x] 4.8 Pass context.Context through all storage calls

## 5. Test Updates
- [x] 5.1 Update `service_test.go` to use new interface methods
- [x] 5.2 Update any direct Engine usage in tests
- [x] 5.3 Update test helpers if needed

## 6. Verification
- [x] 6.1 Run `go build ./...` to verify compilation
- [x] 6.2 Run `go test ./...` to verify tests pass
- [x] 6.3 Run `golangci-lint run` to check for issues
