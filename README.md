# Canopy

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> A bird's-eye view of your git workspaces

**Canopy** is a CLI/TUI tool that manages isolated workspaces for your development work. It creates dedicated directories for each workspace, containing git worktrees for all relevant repositories, while keeping canonical clones centralized.

## Table of Contents

- [The Metaphor](#the-metaphor)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Documentation](#documentation)
- [Command Reference](#command-reference)
- [Configuration](#configuration)
- [License](#license)

## The Metaphor

Think of **Canopy** as your vantage point above the forest. Just as a canopy provides a bird's-eye view of the trees and branches below, this tool gives you an elevated perspective to see and organize all your git workspaces and branches. The TUI provides a literal canopy-level dashboard where you can survey your entire development landscape—multiple repositories (trees), their branches, and the workspaces where you tend them.

## Features

- **Isolated Workspaces** — Dedicated directories for each task (e.g., `~/workspaces/PROJ-123`)
- **Git Worktrees** — Automatic worktree creation for multiple repos on the same branch
- **Centralized Storage** — Canonical repos cloned once and reused across all workspaces
- **Interactive TUI** — Terminal UI for managing workspaces at a glance
- **Lifecycle Hooks** — Run custom commands on workspace creation and closure
- **Pattern Matching** — Auto-assign repos to workspaces based on ID patterns

## Installation

```bash
go install github.com/alexisbeaulieu97/canopy/cmd/canopy@latest
```

## Quick Start

```bash
# Initialize configuration
canopy init

# Add repositories you work with
canopy repo add https://github.com/myorg/backend.git
canopy repo add https://github.com/myorg/frontend.git

# Create a workspace
canopy workspace new PROJ-123 --repos backend,frontend

# Start working
cd ~/workspaces/PROJ-123
```

See the [Quick Start Guide](docs/quick-start.md) for a complete walkthrough.

## Documentation

| Guide | Description |
|-------|-------------|
| [Quick Start](docs/quick-start.md) | Get up and running in 5 minutes |
| [Usage Guide](docs/usage.md) | Complete workflow and examples |
| [Configuration](docs/configuration.md) | All configuration options |
| [Hooks](docs/hooks.md) | Automate with lifecycle hooks |

## Command Reference

### Workspaces

| Command | Description |
|---------|-------------|
| `canopy workspace new <ID>` | Create a new workspace |
| `canopy workspace list` | List active workspaces |
| `canopy workspace view <ID>` | View workspace details |
| `canopy workspace path <ID>` | Print workspace path |
| `canopy workspace close <ID>` | Close and optionally archive workspace |
| `canopy workspace reopen <ID>` | Restore an archived workspace |

**Flags for `workspace new`:**
- `--repos` — Comma-separated list of repositories
- `--branch` — Custom branch name (defaults to ID)
- `--slug` — Optional slug for directory naming
- `--hooks-only` — Run post_create hooks without creating workspace

**Flags for `workspace close`:**
- `--archive` — Archive metadata for later restoration
- `--no-archive` — Delete without archiving
- `--hooks-only` — Run pre_close hooks without closing workspace

### Repositories

| Command | Description |
|---------|-------------|
| `canopy repo list` | List cloned repositories |
| `canopy repo add <URL>` | Clone and register a repository |
| `canopy repo remove <NAME>` | Remove a repository |
| `canopy repo sync <NAME>` | Fetch updates from remote |
| `canopy repo status [NAME]` | Show repository status |

### Registry

Use short aliases for repositories:

| Command | Description |
|---------|-------------|
| `canopy repo register <alias> <url>` | Register an alias |
| `canopy repo unregister <alias>` | Remove an alias |
| `canopy repo list-registry` | List all registered aliases |
| `canopy repo show <alias>` | Show registry entry details |

### TUI

```bash
canopy tui
```

| Key | Action |
|-----|--------|
| `Enter` | View workspace details |
| `o` | Open workspace in editor |
| `c` | Close workspace |
| `s` | Toggle stale filter |
| `/` | Search workspaces |
| `q` | Quit |

See [Configuration](docs/configuration.md#tui-keybindings) to customize keybindings.

### Other Commands

| Command | Description |
|---------|-------------|
| `canopy init` | Initialize configuration |
| `canopy status` | Show overall status |
| `canopy check` | Validate configuration |

## Configuration

Configuration is stored in `~/.canopy/config.yaml`:

```yaml
projects_root: ~/projects
workspaces_root: ~/workspaces
archives_root: ~/.canopy/archives
workspace_close_default: archive  # or delete

defaults:
  workspace_patterns:
    - pattern: "^PROJ-"
      repos: ["backend", "frontend"]
```

See [Configuration Reference](docs/configuration.md) for all options.

## License

[MIT](LICENSE)
