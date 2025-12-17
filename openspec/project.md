# Project Context

## Purpose
**Canopy** (binary `canopy`) is a workspace manager for engineers working across multiple repositories. It solves the problem of juggling multiple branches and repositories by creating isolated workspace directories containing git worktrees, while keeping canonical clones centralized. Common use cases include feature development, bug fixes, experiments, or workspace-based workflows (e.g., JIRA).

## Tech Stack
- **Language**: Go
- **CLI**: Cobra
- **Config**: Viper
- **TUI**: Bubble Tea, Bubbles, Lipgloss
- **Git**: go-git
- **Logging**: charmbracelet/log
- **Filesystem**: afero (for testing)

## Project Conventions

### Code Style
- Standard Go formatting (`gofmt`).
- Linter: `golangci-lint`.

### Architecture Patterns
- **Hexagonal/Clean Architecture**:
    - `cmd/`: Entry points.
    - `internal/domain/`: Core logic (Workspace, Repo entities).
    - `internal/gitx/`: Git adapter.
    - `internal/storage/`: Filesystem adapter.
    - `internal/config/`: Configuration adapter.
    - `internal/tui/`: UI adapter.

### Testing Strategy
- Unit tests for domain and adapters.
- `afero` for filesystem mocking.
- Integration tests for git operations.

### Git Workflow
- Feature branches.
- Conventional Commits.

## Domain Context
- **Canonical Repo**: A bare or mirror clone managed by Canopy in `projects_root`.
- **Workspace**: A directory in `workspaces_root` containing one or more repository worktrees (e.g., `workspaces/feature-auth`, `workspaces/PROJ-123`).
- **Worktree**: A git worktree checked out to a branch, typically matching the workspace name.

## Important Constraints
- **No shelling out to git**: Use `go-git` for core operations to ensure portability and testability.
- **Safe Deletion**: `workspace close` must verify no unpushed/uncommitted changes before deletion.
