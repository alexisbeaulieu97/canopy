# Design: Canopy Architecture

## Architecture Overview

Canopy follows a modular architecture to separate concerns between the CLI/TUI, the domain logic, and the underlying git/filesystem operations.

### Components

1.  **Config Layer (`internal/config`)**
    - Uses Viper to load config from `~/.config/canopy/config.yaml`.
    - Defines `projects_root` (canonical clones) and `workspaces_root` (workspaces).

2.  **Git Engine (`internal/gitx`)**
    - Wraps `go-git`.
    - Manages "Canonical Repos" (bare/mirror clones).
    - Creates Worktrees for workspaces.
    - **Constraint**: Must handle worktree creation safely, ensuring branches are named consistently (`<TICKET-ID>`).

3.  **Workspace Engine (`internal/workspace`)**
    - Manages the directory structure in `workspaces_root`.
    - Persists workspace metadata (`workspace.yaml`) containing the list of repos and other context.

4.  **Domain Layer (`internal/domain`)**
    - `Workspace`: Struct representing a workspace (ID, Slug, Repos).
    - `Repo`: Struct representing a repository (URL, Name).

5.  **CLI (`cmd/canopy`)**
    - Cobra commands mapping to domain actions.

6.  **TUI (`internal/tui`)**
    - Bubble Tea model.
    - **State**: List of workspaces, current selection.
    - **Actions**: Open (shell), Close (delete).

## Data Flow

1.  **New Workspace**:
    - User runs `canopy workspace new PROJ-123 --repos repo-a`.
    - `Workspace` creates `workspaces/PROJ-123`.
    - `Gitx` ensures `repo-a` is cloned in `projects_root`.
    - `Gitx` creates worktree for `repo-a` in `workspaces/PROJ-123/repo-a` on branch `PROJ-123`.
    - `Workspace` saves `workspace.yaml`.

2.  **Close Workspace**:
    - User runs `canopy workspace close PROJ-123`.
    - `Gitx` checks status (dirty/unpushed).
    - If clean (or forced), `Gitx` removes worktrees and deletes branch `PROJ-123`.
    - `Workspace` removes `workspaces/PROJ-123`.
