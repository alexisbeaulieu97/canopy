# Proposal: Initialize Canopy

## Why
Canopy needs a foundational architecture to deliver its core value proposition: workspace-based worktrees for managing multi-repo development. This initialization establishes the "walking skeleton" that enables iteration on advanced features.

## What Changes

### Project Structure
- Initialize Go module `github.com/alexisbeaulieu97/canopy`
- Create directory structure (`cmd/canopy`, `internal/*`)
- Add `golangci-lint` configuration

### Core Engines
- **Config**: Viper-based configuration with `~/.canopy/config.yaml`
- **Logging**: charmbracelet/log for structured logging
- **Git Engine**: go-git wrapper for canonical repos and worktrees
- **Workspace Engine**: Metadata management for workspaces

### CLI Commands
- `canopy init`: Generate default configuration
- `canopy workspace new`: Create workspace with worktrees
- `canopy workspace list`: List active workspaces
- `canopy workspace close`: Remove workspace and worktrees
- `canopy status`: Show current workspace status

### TUI
- Interactive workspace list with Bubble Tea
- Workspace detail view with repo status

## Impact
- **Affected specs**: core, tui, cli
