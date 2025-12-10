# Canopy Documentation

Welcome to the Canopy documentation. This guide will help you get started and make the most of Canopy's workspace management features.

## Quick Links

| Document | Description |
|----------|-------------|
| [Quick Start](quick-start.md) | Get up and running in 5 minutes |
| [Usage Guide](usage.md) | Complete workflow and command reference |
| [Configuration](configuration.md) | All configuration options explained |
| [Hooks](hooks.md) | Automate tasks with lifecycle hooks |

## What is Canopy?

Canopy is a CLI/TUI tool that manages isolated git workspaces for your development work. It creates dedicated directories for each task, containing git worktrees for all relevant repositories, while keeping canonical clones centralized.

### Key Concepts

- **Workspace**: An isolated directory containing worktrees for one or more repositories, all on the same branch
- **Canonical Repository**: A bare git clone stored centrally in `projects_root`, used as the source for all worktrees
- **Worktree**: A git worktree linked to a canonical repository, allowing you to work on multiple branches simultaneously

### The Canopy Metaphor

Think of Canopy as your vantage point above the forest. Just as a forest canopy provides a bird's-eye view of the trees and branches below, this tool gives you an elevated perspective to see and organize all your git workspaces. The TUI provides a literal canopy-level dashboard where you can survey your entire development landscape.

## Getting Help

- Run `canopy --help` for command reference
- Run `canopy <command> --help` for command-specific help
- Report issues at [GitHub Issues](https://github.com/alexisbeaulieu97/canopy/issues)
