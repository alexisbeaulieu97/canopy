# Usage Guide

This guide covers common workflows and detailed command usage for Canopy.

## Table of Contents

- [Typical Workflow](#typical-workflow)
- [Workspace Management](#workspace-management)
- [Repository Management](#repository-management)
- [Working with the TUI](#working-with-the-tui)
- [Troubleshooting](#troubleshooting)
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

View the status of your current workspace (when inside a workspace directory):

```bash
canopy status
```

Or view details of a specific workspace:

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
# Keep metadata for later restoration
canopy workspace close PROJ-123 --keep

# Or delete completely
canopy workspace close PROJ-123 --delete
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

# Show git status for each repository (parallel by default)
canopy workspace list --status

# Force sequential status fetching
canopy workspace list --status --sequential-status

# Force parallel status fetching
canopy workspace list --status --parallel-status

# With custom timeout for status check
canopy workspace list --status --timeout 10s

# JSON output for scripting
canopy workspace list --json

# JSON output with status data
canopy workspace list --status --json
```

Status entries include an `Error` field for status failures (e.g., `timeout`). When `Error` is set, `Branch` is empty.

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

# Close all workspaces matching a pattern (prompts for confirmation)
canopy workspace close --pattern "^PROJ-"

# Close all workspaces (equivalent to --pattern ".*")
canopy workspace close --all --force
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

### Renaming Workspaces

```bash
# Rename workspace (also renames branches by default)
canopy workspace rename PROJ-123 PROJ-456

# Rename workspace only, keep branches as-is
canopy workspace rename PROJ-123 PROJ-456 --rename-branch=false
```

### Switching Branches

Switch all repositories in a workspace to a different branch:

```bash
# Switch to existing branch
canopy workspace branch PROJ-123 develop

# Create and switch to new branch
canopy workspace branch PROJ-123 feature/new --create

# Switch branch for all matching workspaces
canopy workspace branch --pattern "^PROJ-" develop
```

### Syncing Workspaces

The `workspace sync` command pulls updates for all repositories in a workspace and displays a curated summary instead of raw git output.

```bash
# Sync all repos in a workspace
canopy workspace sync PROJ-123

# Sync with a custom timeout (default: 60s per repo)
canopy workspace sync PROJ-123 --timeout 30s

# Output JSON for automation
canopy workspace sync PROJ-123 --json

# Sync all workspaces matching a pattern
canopy workspace sync --pattern "^FEATURE-"
```

The output displays a table with:
- **REPOSITORY**: Name of the repositories
- **STATUS**: Outcome (UPDATED, UP-TO-DATE, CONFLICT, TIMEOUT, ERROR)
- **UPDATED**: Number of new commits pulled
- **DETAILS**: Error messages if any

### Running Git Commands Across Repos

Execute any git command in all repositories within a workspace:

```bash
# Check status in all repos
canopy workspace git PROJ-123 status

# Fetch all remotes
canopy workspace git PROJ-123 -- fetch --all

# Run in parallel for faster execution
canopy workspace git PROJ-123 --parallel pull

# Continue even if some repos fail
canopy workspace git PROJ-123 --continue-on-error status
```

### Exporting and Importing Workspaces

Export a workspace definition to share or backup:

```bash
# Export to stdout
canopy workspace export PROJ-123

# Export to file
canopy workspace export PROJ-123 --output workspace.yaml

# Export as JSON
canopy workspace export PROJ-123 --format json
```

Import a workspace from an exported file:

```bash
# Import from file
canopy workspace import workspace.yaml

# Import with different workspace ID
canopy workspace import workspace.yaml --id NEW-WORKSPACE

# Import with different branch
canopy workspace import workspace.yaml --branch develop

# Import from stdin
canopy workspace import - < workspace.yaml
```

### Managing Workspace Repositories

Add or remove repositories from an existing workspace:

```bash
# Add a repository
canopy workspace repo add PROJ-123 backend

# Remove a repository
canopy workspace repo remove PROJ-123 frontend
```

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

View detailed information about canonical repositories, including disk usage, last fetch time, and workspace usage:

```bash
# List status for all repositories
canopy repo status

# Show status for a specific repository
canopy repo status backend

# Output in JSON format
canopy repo status --json
```

Output includes:
- **NAME**: Repository alias
- **SIZE**: Disk usage on the local system
- **LAST FETCH**: Time of the last `repo sync` or `workspace update`
- **WORKSPACES**: Number of active/archived workspaces using this repository

### Getting Repository Path

Print the absolute path of a canonical repository:

```bash
canopy repo path backend
# /home/user/.canopy/projects/backend

# JSON output for scripting
canopy repo path backend --json
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

## Health Checks

### Checking Environment

The `check` command validates your Canopy configuration:

```bash
# Validate configuration
canopy check
```

### Detecting Orphaned Worktrees

Orphaned worktrees are git worktrees that reference missing workspaces or have invalid git directories:

```bash
# Check for orphaned worktrees
canopy check --orphans

# JSON output for scripting
canopy check --orphans --json

# Automatically repair orphaned worktrees by removing them
canopy check --orphans --repair
```

Common orphan scenarios:
- Workspace directory was manually deleted
- Git worktree reference points to non-existent workspace
- Corrupted `.git` file in worktree

The `--repair` flag removes orphaned worktrees safely. Always review the output first without `--repair` to see what would be affected.

## Troubleshooting

### Running Diagnostics

The `doctor` command validates your Canopy environment and reports issues with actionable guidance:

```bash
# Run all checks
canopy doctor

# Output results as JSON for scripting
canopy doctor --json

# Auto-fix simple issues (create missing directories)
canopy doctor --fix
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks passed |
| 1 | Warnings present (non-critical issues) |
| 2 | Errors present (critical issues) |

### Common Issues

**Git not installed or not in PATH**
```
✗ Git Installation: git is not installed or not in PATH
```
Install git from https://git-scm.com/downloads

**Missing directories**
```
✗ Directory: projects_root: directory does not exist: /path/to/projects
```
Run `canopy doctor --fix` to create missing directories, or create them manually.

**Stale repositories**
```
✗ Repo: backend: stale (last fetch: 2024-01-15)
```
Run `canopy repo sync backend` to fetch updates.

**Invalid configuration**
```
✗ Configuration: configuration error
```
Check your `~/.canopy/config.yaml` for syntax errors. The error details will indicate the specific issue.

### Checking Workspace Status

To diagnose issues with a specific workspace:

```bash
# View workspace details
canopy workspace view PROJ-123

# Check git status for all repos (parallel by default)
canopy workspace list --status

# Force sequential status fetching
canopy workspace list --status --sequential-status

# Run git status directly in each repo
canopy workspace git PROJ-123 status
```

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

Then use `--delete` when you want to delete permanently.
