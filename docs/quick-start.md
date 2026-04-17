# Quick Start

Get up and running with Canopy in a few minutes.

## Prerequisites

- **Go 1.24+** — [Install Go](https://go.dev/doc/install)
- **Git** — Must be available in your PATH

## Installation

**Option 1: Install from source**
```bash
git clone https://github.com/alexisbeaulieu97/canopy.git
cd canopy
go install ./cmd/canopy
```

**Option 2: Install directly (requires release tag)**
```bash
go install github.com/alexisbeaulieu97/canopy/cmd/canopy@latest
```

## Initial Setup

### 1. Initialize Configuration

```bash
canopy init
```

This creates `~/.canopy/config.yaml` with sensible defaults.
The generated file contains starter examples, and the command prints the next setup commands you need.

### 2. Add Your Repositories

Register the repositories you work with:

```bash
canopy repo add https://github.com/myorg/backend.git
canopy repo add https://github.com/myorg/frontend.git
```

Canopy clones these once and reuses them for all workspaces.
If you prefer aliases without cloning yet, you can also use `canopy repo register <alias> <repository-url>`.

### 3. Create Your First Workspace

```bash
canopy workspace new PROJ-123 --repos backend,frontend
```

If you run `canopy workspace new` before repos resolve from `--repos`, templates, or workspace patterns, Canopy now fails with guidance instead of creating an empty workspace.

This creates:
- A workspace directory at `~/.canopy/workspaces/PROJ-123/`
- Worktrees for `backend` and `frontend` inside it
- A branch named `PROJ-123` in each repository

### 4. Start Working

```bash
cd ~/.canopy/workspaces/PROJ-123
ls
# backend/  frontend/
```

Both repositories are checked out on the `PROJ-123` branch.

## Common Commands

| Task | Command |
|------|---------|
| List workspaces | `canopy workspace list` |
| Check status | `canopy status` |
| View workspace details | `canopy workspace view PROJ-123` |
| Get workspace path | `canopy workspace path PROJ-123` |
| Close workspace | `canopy workspace close PROJ-123` |
| Launch TUI | `canopy tui` |

## Next Steps

- Read the [Usage Guide](usage.md) for detailed workflows
- Configure [workspace patterns](configuration.md#workspace-patterns) for automatic repo assignment
- Set up [hooks](hooks.md) to automate tasks
