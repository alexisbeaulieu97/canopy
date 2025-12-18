# Architecture Overview

This document describes the architecture of Canopy for developers who want to understand or contribute to the codebase.

## Design Philosophy

Canopy follows **hexagonal architecture** (also known as ports and adapters), which provides:

- **Testability**: Core business logic can be tested without external dependencies
- **Flexibility**: Adapters can be swapped without changing core logic
- **Maintainability**: Clear separation of concerns

## Package Structure

```
internal/
├── app/          # Application container and dependency injection
├── config/       # Configuration loading and validation
├── domain/       # Core domain models (no dependencies)
├── errors/       # Typed error definitions
├── giturl/       # Git URL parsing utilities
├── gitx/         # Git operations adapter (go-git implementation)
├── hooks/        # Hook execution adapter
├── logging/      # Structured logging
├── mocks/        # Mock implementations for testing
├── output/       # CLI output formatting
├── ports/        # Interface definitions (the "ports")
├── storage/      # Workspace storage adapter
├── testutil/     # Test utilities
├── tui/          # Terminal UI components
├── validation/   # Input validation
└── workspaces/   # Core business logic (the "core")
```

## Architectural Layers

### 1. Domain Layer (`internal/domain`)

Pure data structures with no external dependencies:

```go
// Core domain models
type Workspace struct {
    ID         string
    BranchName string
    Repos      []Repo
}

type Repo struct {
    Name string
    URL  string
}
```

### 2. Ports Layer (`internal/ports`)

Interfaces defining how the core interacts with the outside world:

```go
// Configuration access
type ConfigProvider interface {
    GetProjectsRoot() string
    GetWorkspacesRoot() string
    GetReposForWorkspace(id string) []string
    // ...
}

// Git operations
type GitOperations interface {
    EnsureCanonical(ctx context.Context, url, name string) error
    CreateWorktree(repoName, path, branch string) error
    Status(path string) (isDirty bool, unpushed, behind int, branch string, err error)
    // ...
}

// Workspace persistence
type WorkspaceStorage interface {
    Create(ctx context.Context, ws domain.Workspace) error
    List(ctx context.Context) ([]domain.Workspace, error)
    Load(ctx context.Context, id string) (*domain.Workspace, error)
    // ...
}
```

### 3. Core Layer (`internal/workspaces`)

Business logic that orchestrates operations through ports:

```go
type Service struct {
    config    ports.ConfigProvider
    gitEngine ports.GitOperations
    wsEngine  ports.WorkspaceStorage
    // ...
}

func (s *Service) CreateWorkspace(ctx context.Context, id string, repos []domain.Repo) (string, error) {
    // Validate inputs
    // Create workspace via wsEngine
    // Clone repos via gitEngine
    // Run hooks
}
```

### 4. Adapters Layer

Concrete implementations of ports:

- **`internal/gitx`**: Git operations using go-git
- **`internal/storage`**: File-based workspace storage
- **`internal/config`**: YAML configuration via Viper
- **`internal/hooks`**: Shell command execution

## Key Interfaces

### ConfigProvider

Provides access to configuration values:

```go
type ConfigProvider interface {
    GetProjectsRoot() string      // Where canonical repos are stored
    GetWorkspacesRoot() string    // Where workspaces are created
    GetClosedRoot() string        // Where archived workspaces go
    GetReposForWorkspace(id string) []string  // Pattern matching
    GetHooks() config.Hooks       // Lifecycle hooks config
    GetKeybindings() config.Keybindings  // TUI keybindings
}
```

### GitOperations

Abstracts git operations:

```go
type GitOperations interface {
    // Canonical repository management
    EnsureCanonical(ctx context.Context, url, name string) (*git.Repository, error)
    Clone(ctx context.Context, url, name string) error
    Fetch(ctx context.Context, name string) error
    List() ([]string, error)

    // Worktree operations
    CreateWorktree(repoName, path, branch string) error
    Status(path string) (isDirty bool, unpushed, behind int, branch string, err error)
    Push(ctx context.Context, path, branch string) error
    Pull(ctx context.Context, path string) error
}
```

### WorkspaceStorage

Handles workspace persistence:

```go
type WorkspaceStorage interface {
    Create(ctx context.Context, ws domain.Workspace) error
    Save(ctx context.Context, ws domain.Workspace) error
    Load(ctx context.Context, id string) (*domain.Workspace, error)
    List(ctx context.Context) ([]domain.Workspace, error)
    Delete(ctx context.Context, id string) error
    Close(ctx context.Context, id string, closedAt time.Time) (*domain.ClosedWorkspace, error)
}
```

## Data Flow

### Creating a Workspace

```
1. CLI parses command: canopy workspace new TICKET-123 --repos backend
   ↓
2. App container provides Service with injected dependencies
   ↓
3. Service.CreateWorkspace() orchestrates:
   a. Validate workspace ID (validation package)
   b. Resolve repo names to URLs (RepoResolver)
   c. Create workspace directory (WorkspaceStorage.Create)
   d. Clone canonical repos (GitOperations.EnsureCanonical)
   e. Create worktrees (GitOperations.CreateWorktree)
   f. Execute hooks (HookExecutor.ExecuteHooks)
   ↓
4. Return result to CLI for display
```

### Directory Structure on Disk

```
~/.canopy/
├── config.yaml           # User configuration
├── registry.yaml         # Repository aliases
├── projects/             # Canonical bare repositories
│   ├── backend/          # Bare clone of backend repo
│   │   └── worktrees/    # Git worktree tracking
│   └── frontend/         # Bare clone of frontend repo
│       └── worktrees/    # Git worktree tracking
├── workspaces/           # Active workspaces
│   └── TICKET-123/       # Workspace directory
│       ├── .canopy.yaml  # Workspace metadata
│       ├── backend/      # Git worktree (linked to canonical)
│       └── frontend/     # Git worktree (linked to canonical)
└── closed/               # Archived workspace metadata
    └── TICKET-100/
        └── .canopy.yaml
```

## Git Worktree Architecture

Canopy uses **true git worktrees** to share git objects between the canonical repository and workspace checkouts. This provides:

- **Disk Efficiency**: Worktrees share git objects with the canonical repo, avoiding duplicate storage
- **Correct Remotes**: Each worktree's origin points to the real upstream URL for push/pull operations

### How It Works

1. **Canonical Repository Setup**
   - When a repository is first cloned, Canopy stores the upstream URL in the git config
   - The URL is stored under `canopy.upstreamUrl` in the bare repo's config

2. **Worktree Creation**
   - Uses `git worktree add -b <branch> <path>` to create true worktrees
   - After creation, the worktree's origin remote is configured to point to the upstream URL
   - Branch tracking is set up for proper push/pull behavior

3. **Worktree Cleanup**
   - When closing a workspace, `git worktree remove` is called for each repo
   - Orphan detection includes `git worktree prune` to clean up stale references

### Migration Notes

Workspaces created before this implementation used full clones instead of true worktrees. To benefit from disk savings, users should:

1. Close the old workspace
2. Create a new workspace with the same repositories

Existing workspaces continue to function normally.

## Error Handling

Canopy uses typed errors for predictable handling:

```go
// Define error codes
const (
    ErrWorkspaceNotFound ErrorCode = "WORKSPACE_NOT_FOUND"
    ErrRepoNotFound      ErrorCode = "REPO_NOT_FOUND"
    // ...
)

// Create errors with context
err := cerrors.NewWorkspaceNotFound(id)

// Check error types
if errors.Is(err, cerrors.WorkspaceNotFound) {
    // Handle workspace not found
}
```

## Testing Strategy

### Unit Tests

Test core logic with mock dependencies:

```go
func TestCreateWorkspace(t *testing.T) {
    // Create mock implementations
    mockGit := mocks.NewGitOperations()
    mockStorage := mocks.NewWorkspaceStorage()

    // Create service with mocks
    svc := workspaces.NewService(cfg, mockGit, mockStorage, logger)

    // Test behavior
    _, err := svc.CreateWorkspace(ctx, "test-ws", repos)
    require.NoError(t, err)

    // Verify interactions
    mockStorage.AssertCalled(t, "Create", ...)
}
```

### Integration Tests

Test full flow with real git operations:

```go
func TestIntegration_WorkspaceLifecycle(t *testing.T) {
    // Set up test environment
    tmpDir := t.TempDir()

    // Create real service
    svc := setupTestService(tmpDir)

    // Test full workflow
    _, err := svc.CreateWorkspace(ctx, "test", repos)
    require.NoError(t, err)

    // Verify results on disk
    // ...
}
```

## Extension Points

### Adding New Commands

1. Create command file in `cmd/canopy/`
2. Register in main command tree
3. Use `app.App` for service access

### Adding New Ports

1. Define interface in `internal/ports/`
2. Create adapter in appropriate package
3. Add to `app.App` container
4. Inject into services that need it

### Custom Git Backends

Implement `ports.GitOperations` interface:

```go
type CustomGitBackend struct {
    // ...
}

func (g *CustomGitBackend) EnsureCanonical(ctx context.Context, url, name string) (*git.Repository, error) {
    // Custom implementation
}
// ... implement other methods
```

## See Also

- [Configuration Reference](configuration.md)
- [Error Codes](error-codes.md)
- [Contributing Guide](../CONTRIBUTING.md)
