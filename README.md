# Yardmaster

> Workspace-centric worktrees for humans who live in JIRA and git.

Yardmaster (`yard`) is a CLI/TUI tool that manages per-ticket workspaces. It creates isolated directories for each workspace, containing git worktrees for all relevant repositories, while keeping canonical clones centralized.

## Features

- **Workspaces**: Create a dedicated folder for each ticket/task (e.g., `~/workspaces/PROJ-123`).
- **Git Worktrees**: Automatically create worktrees for multiple repos on the workspace branch.
- **Centralized Storage**: Canonical repos are stored in `~/projects` (configurable) and never re-cloned.
- **TUI**: Interactive terminal UI for managing workspaces.
- **Shell Integration**: Easily `cd` into workspaces or open them in your editor.

## Getting Started

### 1. Installation

```bash
go install github.com/alexisbeaulieu97/yard/cmd/yard@latest
```

### 2. Initialization

Initialize the configuration file:

```bash
yard init
```

This creates `~/.yard/config.yaml` with default settings.

### 3. Add Repositories

Add the repositories you work with frequently:

```bash
yard repo add https://github.com/myorg/backend.git
yard repo add https://github.com/myorg/frontend.git
```

### 4. Create Your First Workspace

Create a workspace for a ticket (e.g., `PROJ-123`) and include specific repos:

```bash
yard workspace new PROJ-123 --repos backend,frontend
```

This will:
1. Create `~/workspaces/PROJ-123` (or similar, based on naming config).
2. Create worktrees for `backend` and `frontend` inside that folder.
3. Checkout a branch named `PROJ-123` (or custom branch if specified).

## Usage

### Workspaces

- **Create**: `yard workspace new <ID> [flags]`
  - `--repos`: Comma-separated list of repos.
  - `--branch`: Custom branch name (defaults to ID).
  - `--slug`: Optional slug for directory naming.
- **List**: `yard workspace list`
- **View**: `yard workspace view <ID>`
- **Path**: `yard workspace path <ID>` (prints absolute path)
- **Sync**: `yard workspace sync <ID>` (pulls all repos)
- **Close**: `yard workspace close <ID>` (removes workspace and worktrees)

### Repositories

- **List**: `yard repo list`
- **Add**: `yard repo add <URL>`
- **Remove**: `yard repo remove <NAME>`
- **Sync**: `yard repo sync <NAME>` (fetches updates)

### TUI

Launch the interactive UI:

```bash
yard tui
```

- **Enter**: View details / Open workspace (if shell integration active).
- **s**: Sync workspace.
- **c**: Close workspace.

## Configuration

Edit `~/.yard/config.yaml`:

```yaml
projects_root: ~/projects
workspaces_root: ~/workspaces
workspace_naming: "{{.ID}}__{{.Slug}}"
```

### Advanced Configuration

#### Workspace Naming

You can customize how workspace directories are named using Go templates:

```yaml
workspace_naming: "{{.ID}}"           # Result: PROJ-123
workspace_naming: "{{.ID}}-{{.Slug}}" # Result: PROJ-123-fix-bug
```

#### Auto-Repositories (Regex)

Automatically include repositories based on the workspace ID pattern:

```yaml
workspace_patterns:
  - regex: "^BACK-.*"
    repos: ["backend", "common-lib"]
  - regex: "^FRONT-.*"
    repos: ["frontend", "ui-kit"]
```

With this config, `yard workspace new BACK-456` will automatically include `backend` and `common-lib`.