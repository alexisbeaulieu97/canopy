# Usage Guide

This guide covers common workflows and detailed command usage for Canopy.

## Table of Contents

- [Typical Workflow](#typical-workflow)
- [Workspace Management](#workspace-management)
- [Repository Management](#repository-management)
- [Working with the TUI](#working-with-the-tui)
- [Tips and Best Practices](#tips-and-best-practices)

## Typical Workflow

### 1. Create a Workspace

When starting work on a task (e.g., `PROJ-123`):

```bash
canopy workspace new PROJ-123 --repos backend,frontend
```

This creates:
- A directory at `~/workspaces/PROJ-123/` (configurable)
- Git worktrees for each repository inside
- A branch named `PROJ-123` in each repository

### 2. Work in the Workspace

```bash
cd ~/workspaces/PROJ-123
ls
# backend/  frontend/

# Make changes in both repos
cd backend && git status
```

All repositories are automatically on the `PROJ-123` branch.

### 3. Check Status

View the status of all workspaces:

```bash
canopy status
```

Or check a specific workspace:

```bash
canopy workspace view PROJ-123
```

### 4. Push Changes

Use standard git commands inside each worktree:

```bash
cd ~/workspaces/PROJ-123/backend
git add .
git commit -m "feat: implement feature"
git push origin PROJ-123
```

### 5. Close the Workspace

When finished:

```bash
# Archive for later restoration
canopy workspace close PROJ-123 --archive

# Or delete completely
canopy workspace close PROJ-123 --no-archive
```

### 6. Restore if Needed

Reopen an archived workspace:

```bash
canopy workspace reopen PROJ-123
```

## Workspace Management

### Creating Workspaces

**Basic creation:**
```bash
canopy workspace new PROJ-123 --repos backend,frontend
```

**With custom branch name:**
```bash
canopy workspace new PROJ-123 --repos backend --branch feature/auth
```

### Listing Workspaces

```bash
# List active workspaces
canopy workspace list

# List closed workspaces
canopy workspace list --closed

# JSON output for scripting
canopy workspace list --json
```

### Viewing Workspace Details

```bash
canopy workspace view PROJ-123
```

Shows:
- Workspace ID and path
- Branch name
- Included repositories
- Creation date

### Getting Workspace Path

Useful for scripting or shell integration:

```bash
canopy workspace path PROJ-123
# /home/user/workspaces/PROJ-123

# Use in scripts
cd "$(canopy workspace path PROJ-123)"
```

### Closing Workspaces

```bash
# Interactive (prompts in TTY)
canopy workspace close PROJ-123

# Keep metadata for later restoration
canopy workspace close PROJ-123 --keep

# Delete completely
canopy workspace close PROJ-123 --delete

# Run pre_close hooks only (don't actually close)
canopy workspace close PROJ-123 --hooks-only

# Force close even with uncommitted changes or unpushed commits
canopy workspace close PROJ-123 --force
```

#### Safety Checks

Before closing a workspace, Canopy verifies that all repositories are in a safe state:

1. **No uncommitted changes** - All changes must be committed
2. **No unpushed commits** - All commits must be pushed to the remote

If either check fails, the close operation is blocked. Use `--force` to bypass these safety checks (use with caution, as unpushed work may be lost).

The `--dry-run` flag shows what would happen, including warnings for any repos with uncommitted changes or unpushed commits.

### Reopening Archived Workspaces

```bash
canopy workspace reopen PROJ-123
```

This recreates worktrees from the archived metadata.

## Repository Management

### Adding Repositories

```bash
# Add by URL
canopy repo add https://github.com/myorg/backend.git

# Add with custom alias
canopy repo add https://github.com/myorg/backend.git --alias api

# Add without registering alias
canopy repo add https://github.com/myorg/backend.git --no-register
```

### Listing Repositories

```bash
canopy repo list
```

### Syncing Repositories

Fetch updates from remote:

```bash
canopy repo sync backend
```

### Checking Repository Status

```bash
# Status of all repositories
canopy repo status

# Status of specific repository
canopy repo status backend
```

### Using the Registry

The registry maps short aliases to full repository URLs:

```bash
# Register an alias
canopy repo register api https://github.com/myorg/backend.git

# List registry entries
canopy repo list-registry

# Show entry details
canopy repo show api

# Remove alias
canopy repo unregister api
```

## Working with the TUI

Launch the interactive terminal UI:

```bash
canopy tui
```

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate list |
| `Enter` | View workspace details |
| `/` | Search/filter workspaces |
| `Esc` | Clear search / go back |

### Actions

| Key | Action |
|-----|--------|
| `o` | Open workspace in editor |
| `c` | Close selected workspace |
| `p` | Push workspace changes |
| `s` | Toggle stale workspace filter |
| `q` | Quit |

### Customizing Keybindings

See [Configuration - TUI Keybindings](configuration.md#tui-keybindings).

## Tips and Best Practices

### Use Workspace Patterns

Configure automatic repo assignment in `~/.canopy/config.yaml`:

```yaml
defaults:
  workspace_patterns:
    - pattern: "^BACK-"
      repos: ["backend", "common-lib"]
    - pattern: "^FRONT-"
      repos: ["frontend", "ui-kit"]
```

Now `canopy workspace new BACK-123` automatically includes the right repos.

### Set Up Lifecycle Hooks

Automate setup tasks with [hooks](hooks.md):

```yaml
hooks:
  post_create:
    - command: "cd {{.WorkspacePath}}/backend && make setup"
  pre_close:
    - command: "cd {{.WorkspacePath}} && docker-compose down"
```

### Shell Integration

Add to your `.bashrc` or `.zshrc`:

```bash
# Quick workspace navigation
cw() {
    cd "$(canopy workspace path "$1")"
}

# Create and enter workspace
nw() {
    canopy workspace new "$@" && cw "$1"
}
```

### Archive by Default

If you often reopen workspaces, set archiving as default:

```yaml
workspace_close_default: archive
```

Then use `--no-archive` when you want to delete permanently.
