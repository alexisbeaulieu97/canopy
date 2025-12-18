# Canopy

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/github/actions/workflow/status/alexisbeaulieu97/canopy/test.yml?label=tests)](https://github.com/alexisbeaulieu97/canopy/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexisbeaulieu97/canopy)](https://goreportcard.com/report/github.com/alexisbeaulieu97/canopy)

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

### Using Go Install (Simple)

```bash
go install github.com/alexisbeaulieu97/canopy/cmd/canopy@latest
```

### From Source (With Version Info)

```bash
git clone https://github.com/alexisbeaulieu97/canopy.git
cd canopy
make install
```

This embeds version, commit hash, and build date into the binary.

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
| [Error Codes](docs/error-codes.md) | Error codes for scripting |
| [Architecture](docs/architecture.md) | Technical architecture overview |

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
| `canopy workspace rename <OLD> <NEW>` | Rename a workspace |
| `canopy workspace branch <ID> <BRANCH>` | Switch branch for all repos |
| `canopy workspace sync <ID>` | Pull updates for all repositories |
| `canopy workspace git <ID> <git-args...>` | Run git command across all repos |
| `canopy workspace export <ID>` | Export workspace definition |
| `canopy workspace import <file>` | Import workspace from file |
| `canopy workspace repo add <ID> <REPO>` | Add a repository to workspace |
| `canopy workspace repo remove <ID> <REPO>` | Remove a repository from workspace |

**Flags for `workspace new`:**
- `--repos` — Comma-separated list of repositories
- `--branch` — Custom branch name (defaults to ID)
- `--print-path` — Print the created workspace path
- `--no-hooks` — Skip post_create hooks
- `--hooks-only` — Run post_create hooks without creating workspace

**Flags for `workspace close`:**
- `--keep` — Keep metadata for later restoration
- `--delete` — Delete without keeping metadata
- `--force` — Force close even with uncommitted changes
- `--dry-run` — Preview what would be deleted
- `--no-hooks` — Skip pre_close hooks
- `--hooks-only` — Run pre_close hooks without closing workspace

**Flags for `workspace sync`:**
- `--timeout` — Timeout for each repository sync (default: 60s)
- `--json` — Output in JSON format

### Repositories

| Command | Description |
|---------|-------------|
| `canopy repo list` | List cloned repositories |
| `canopy repo add <URL>` | Clone and register a repository |
| `canopy repo remove <NAME>` | Remove a repository |
| `canopy repo sync <NAME>` | Fetch updates from remote |
| `canopy repo path <NAME>` | Print canonical repository path |

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
| `canopy version` | Print version information |

**Version output example:**

```text
canopy version v1.0.0
commit: abc1234
built: 2025-01-15T10:30:00Z
go: go1.21.0
```

Use `canopy version --json` for machine-readable output or `canopy --version` for a short version string.

## Configuration

Configuration is stored in `~/.canopy/config.yaml`:

```yaml
projects_root: ~/projects
workspaces_root: ~/workspaces
closed_root: ~/.canopy/closed
workspace_close_default: delete  # default; set to "archive" to keep metadata

defaults:
  workspace_patterns:
    - pattern: "^PROJ-"
      repos: ["backend", "frontend"]
```

See [Configuration Reference](docs/configuration.md) for all options.

## Troubleshooting

### Common Issues

**Workspace creation fails with "unknown repository"**

The repository must be cloned or registered first:
```bash
# Clone the repository
canopy repo add https://github.com/myorg/repo.git

# Or register an alias
canopy repo register repo https://github.com/myorg/repo.git
```

**Git operations timeout**

Configure retry settings in `~/.canopy/config.yaml`:
```yaml
git:
  retry:
    max_attempts: 5
    initial_delay: "2s"
    max_delay: "60s"
```

**"Repository has uncommitted changes" error**

Either commit/stash changes or use `--force`:
```bash
canopy workspace close PROJ-123 --force
```

**Configuration validation errors**

Run `canopy check` to validate your configuration:
```bash
canopy check
```

For machine-readable error handling, see the [Error Codes Reference](docs/error-codes.md).

## License

[MIT](LICENSE)
