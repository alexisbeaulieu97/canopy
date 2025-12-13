# Tasks: Add Context Propagation

## Implementation Checklist

### 1. Update Interfaces
- [x] 1.1 Update `ports.GitOperations` interface:
  - Add context to `Clone(ctx, url, name)`
  - Add context to `Fetch(ctx, name)`
  - Add context to `Push(ctx, path, branch)`
  - Add context to `Pull(ctx, path)`
  - Add context to `EnsureCanonical(ctx, url, name)`
- [x] 1.2 Update `ports.WorkspaceStorage` interface (if any async operations)
- [x] 1.3 Update mock implementations in `internal/mocks/`

### 2. Update Git Engine
- [x] 2.1 Update `gitx.GitEngine.Clone` to use context with go-git
- [x] 2.2 Update `gitx.GitEngine.Fetch` to use context
- [x] 2.3 Update `gitx.GitEngine.Push` to use context
- [x] 2.4 Update `gitx.GitEngine.Pull` to use context
- [x] 2.5 Update `gitx.GitEngine.EnsureCanonical` to use context
- [x] 2.6 Add default timeout constant (5 minutes)
- [x] 2.7 Update `RunCommand` to respect context cancellation

### 3. Update Service Layer
- [x] 3.1 Update `Service.CreateWorkspace(ctx, ...)` 
- [x] 3.2 Update `Service.CreateWorkspaceWithOptions(ctx, ...)`
- [x] 3.3 Update `Service.CloseWorkspace(ctx, ...)`
- [x] 3.4 Update `Service.CloseWorkspaceKeepMetadata(ctx, ...)`
- [x] 3.5 Update `Service.AddRepoToWorkspace(ctx, ...)`
- [x] 3.6 Update `Service.RemoveRepoFromWorkspace(ctx, ...)`
- [x] 3.7 Update `Service.PushWorkspace(ctx, ...)`
- [x] 3.8 Update `Service.RunGitInWorkspace(ctx, ...)`
- [x] 3.9 Update `Service.SwitchBranch(ctx, ...)`
- [x] 3.10 Update `Service.RestoreWorkspace(ctx, ...)`
- [x] 3.11 Update `Service.AddCanonicalRepo(ctx, ...)`
- [x] 3.12 Update `Service.RemoveCanonicalRepo(ctx, ...)`
- [x] 3.13 Update `Service.SyncCanonicalRepo(ctx, ...)`
- [x] 3.14 Update `Service.ExportWorkspace(ctx, ...)`
- [x] 3.15 Update `Service.ImportWorkspace(ctx, ...)`

### 4. Update Parallel Operations
- [x] 4.1 Update `runGitParallel` to cancel on context done
- [x] 4.2 Update `runGitSequential` to check context between iterations
- [x] 4.3 Ensure goroutines respect context cancellation

### 5. Update CLI Commands
- [x] 5.1 Update workspace commands to pass `cmd.Context()`
- [x] 5.2 Update repo commands to pass `cmd.Context()`
- [x] 5.3 Update status command
- [x] 5.4 Update check command

### 6. Update TUI
- [x] 6.1 Create background context for TUI operations
- [x] 6.2 Update TUI commands to pass context
- [x] 6.3 Cancel pending operations on TUI quit

### 7. Testing
- [x] 7.1 Update service tests with context
- [x] 7.2 Add timeout test cases
- [x] 7.3 Add cancellation test cases
- [x] 7.4 Update mock tests

