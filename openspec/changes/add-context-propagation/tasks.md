# Tasks: Add Context Propagation

## Implementation Checklist

### 1. Update Interfaces
- [ ] 1.1 Update `ports.GitOperations` interface:
  - Add context to `Clone(ctx, url, name)`
  - Add context to `Fetch(ctx, name)`
  - Add context to `Push(ctx, path, branch)`
  - Add context to `Pull(ctx, path)`
  - Add context to `EnsureCanonical(ctx, url, name)`
- [ ] 1.2 Update `ports.WorkspaceStorage` interface (if any async operations)
- [ ] 1.3 Update mock implementations in `internal/mocks/`

### 2. Update Git Engine
- [ ] 2.1 Update `gitx.GitEngine.Clone` to use context with go-git
- [ ] 2.2 Update `gitx.GitEngine.Fetch` to use context
- [ ] 2.3 Update `gitx.GitEngine.Push` to use context
- [ ] 2.4 Update `gitx.GitEngine.Pull` to use context
- [ ] 2.5 Update `gitx.GitEngine.EnsureCanonical` to use context
- [ ] 2.6 Add default timeout constant (5 minutes)
- [ ] 2.7 Update `RunCommand` to respect context cancellation

### 3. Update Service Layer
- [ ] 3.1 Update `Service.CreateWorkspace(ctx, ...)` 
- [ ] 3.2 Update `Service.CreateWorkspaceWithOptions(ctx, ...)`
- [ ] 3.3 Update `Service.CloseWorkspace(ctx, ...)`
- [ ] 3.4 Update `Service.CloseWorkspaceKeepMetadata(ctx, ...)`
- [ ] 3.5 Update `Service.AddRepoToWorkspace(ctx, ...)`
- [ ] 3.6 Update `Service.RemoveRepoFromWorkspace(ctx, ...)`
- [ ] 3.7 Update `Service.PushWorkspace(ctx, ...)`
- [ ] 3.8 Update `Service.RunGitInWorkspace(ctx, ...)`
- [ ] 3.9 Update `Service.SwitchBranch(ctx, ...)`
- [ ] 3.10 Update `Service.RestoreWorkspace(ctx, ...)`
- [ ] 3.11 Update `Service.AddCanonicalRepo(ctx, ...)`
- [ ] 3.12 Update `Service.RemoveCanonicalRepo(ctx, ...)`
- [ ] 3.13 Update `Service.SyncCanonicalRepo(ctx, ...)`
- [ ] 3.14 Update `Service.ExportWorkspace(ctx, ...)`
- [ ] 3.15 Update `Service.ImportWorkspace(ctx, ...)`

### 4. Update Parallel Operations
- [ ] 4.1 Update `runGitParallel` to cancel on context done
- [ ] 4.2 Update `runGitSequential` to check context between iterations
- [ ] 4.3 Ensure goroutines respect context cancellation

### 5. Update CLI Commands
- [ ] 5.1 Update workspace commands to pass `cmd.Context()`
- [ ] 5.2 Update repo commands to pass `cmd.Context()`
- [ ] 5.3 Update status command
- [ ] 5.4 Update check command

### 6. Update TUI
- [ ] 6.1 Create background context for TUI operations
- [ ] 6.2 Update TUI commands to pass context
- [ ] 6.3 Cancel pending operations on TUI quit

### 7. Testing
- [ ] 7.1 Update service tests with context
- [ ] 7.2 Add timeout test cases
- [ ] 7.3 Add cancellation test cases
- [ ] 7.4 Update mock tests

