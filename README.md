# Canopy

> A bird's-eye view of your git workspaces

**Canopy** is a CLI/TUI tool that manages isolated workspaces for your development work. It creates dedicated directories for each workspace, containing git worktrees for all relevant repositories, while keeping canonical clones centralized.

## The Metaphor

Think of **Canopy** as your vantage point above the forest. Just as a canopy provides a bird's-eye view of the trees and branches below, this tool gives you an elevated perspective to see and organize all your git workspaces and branches. The TUI provides a literal canopy-level dashboard where you can survey your entire development landscape—multiple repositories (trees), their branches, and the workspaces where you tend them. You're not lost in the undergrowth; you're managing from above with clarity and control.

## Features

- **Workspaces**: Create a dedicated folder for each workspace (e.g., `~/workspaces/feature-auth`, `~/workspaces/PROJ-123`).
- **Git Worktrees**: Automatically create worktrees for multiple repos on the workspace branch.
- **Centralized Storage**: Canonical repos are stored in `~/projects` (configurable) and never re-cloned.
- **TUI**: Interactive terminal UI for managing workspaces.
- **Shell Integration**: Easily `cd` into workspaces or open them in your editor.

## Getting Started

### 1. Installation

```bash
go install github.com/alexisbeaulieu97/canopy/cmd/canopy@latest
```

### 2. Initialization

Initialize the configuration file:

```bash
canopy init
```

This creates `~/.canopy/config.yaml` with default settings.

### 3. Add Repositories

Add the repositories you work with frequently:

```bash
canopy repo add https://github.com/myorg/backend.git
canopy repo add https://github.com/myorg/frontend.git
```

### 4. Create Your First Workspace

Create a workspace (e.g., `PROJ-123` for a workspace, or `feature-auth` for a feature) and include specific repos:

```bash
canopy workspace new PROJ-123 --repos backend,frontend
```

This will:
1. Create `~/workspaces/PROJ-123` (or similar, based on naming config).
2. Create worktrees for `backend` and `frontend` inside that folder.
3. Checkout a branch named `PROJ-123` (or custom branch if specified).

## Usage

### Workspaces

- **Create**: `canopy workspace new <ID> [flags]`
  - `--repos`: Comma-separated list of repos.
  - `--branch`: Custom branch name (defaults to ID).
  - `--slug`: Optional slug for directory naming.
- **List**: `canopy workspace list`
- **View**: `canopy workspace view <ID>`
- **Path**: `canopy workspace path <ID>` (prints absolute path)
- **Sync**: _removed_ (use `workspace git <ID> pull` once implemented)
- **List closed**: `canopy workspace list --closed`
- **Close**: `canopy workspace close <ID> [--keep|--delete]` (prompts to keep metadata in TTY; flags override)
- **Reopen**: `canopy workspace reopen <ID>` (recreates worktrees from closed entry)

### Repositories

- **List**: `canopy repo list`
- **Add**: `canopy repo add <URL>`
- **Remove**: `canopy repo remove <NAME>`
- **Sync**: `canopy repo sync <NAME>` (fetches updates)

#### Repository Registry

Use short aliases instead of full URLs:

- **Register**: `canopy repo register <alias> <url> [--branch develop --description "..."] [--tags api,go]`
- **Unregister**: `canopy repo unregister <alias>`
- **List registry**: `canopy repo list-registry [--tags backend]`
- **Show entry**: `canopy repo show <alias>`

`canopy repo add` auto-registers a sensible alias (override with `--alias` or skip with `--no-register`).

### TUI

Launch the interactive UI:

```bash
canopy tui
```

- **Enter**: View details / Open workspace (if shell integration active).
- **s**: Sync workspace.
- **c**: Close workspace.

## Configuration

Edit `~/.canopy/config.yaml`:

```yaml
projects_root: ~/projects
workspaces_root: ~/workspaces
closed_root: ~/.canopy/closed
workspace_close_default: delete # or keep
```

- `workspace_close_default: delete` keeps delete behavior; set to `keep` to make keeping metadata the default when no flags are supplied. Explicit flags (`--keep` / `--delete`) always win.

### Advanced Configuration

See `docs/configuration.md` for a concise list of config keys.

#### Auto-Repositories (Regex)

Automatically include repositories based on the workspace ID pattern:

```yaml
workspace_patterns:
  - regex: "^BACK-.*"
    repos: ["backend", "common-lib"]
  - regex: "^FRONT-.*"
    repos: ["frontend", "ui-kit"]
```

With this config, `canopy workspace new BACK-456` will automatically include `backend` and `common-lib`.
