# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-01-15

### Added

- **Workspace Management**
  - Create isolated workspaces with `canopy workspace new <ID>`
  - List active workspaces with `canopy workspace list`
  - View workspace details with `canopy workspace view <ID>`
  - Get workspace path with `canopy workspace path <ID>`
  - Close workspaces with `canopy workspace close <ID>` (supports `--keep` and `--delete` flags)
  - Reopen archived workspaces with `canopy workspace reopen <ID>`
  - Rename workspaces with `canopy workspace rename <old-id> <new-id>`

- **Repository Management**
  - Clone and register repositories with `canopy repo add <URL>`
  - List cloned repositories with `canopy repo list`
  - Remove repositories with `canopy repo remove <NAME>`
  - Sync updates from remote with `canopy repo sync <NAME>`
  - Repository registry for short aliases

- **Registry System**
  - Register aliases with `canopy repo register <alias> <url>`
  - Unregister aliases with `canopy repo unregister <alias>`
  - List registry entries with `canopy repo list-registry`
  - Show registry details with `canopy repo show <alias>`

- **Terminal User Interface (TUI)**
  - Interactive workspace management with `canopy tui`
  - Configurable keybindings
  - Stale workspace filtering
  - Search functionality

- **Lifecycle Hooks**
  - `post_create` hooks for workspace creation
  - `pre_close` hooks for workspace closure
  - Per-repo hook filtering
  - Configurable timeouts

- **Configuration**
  - YAML-based configuration in `~/.canopy/config.yaml`
  - Workspace pattern matching for automatic repo assignment
  - Customizable workspace naming templates
  - Git operation retry settings with exponential backoff

- **Other Commands**
  - `canopy init` - Initialize configuration
  - `canopy status` - Show overall status
  - `canopy check` - Validate configuration
  - `canopy version` - Print version information with build metadata

- **Error Handling**
  - Typed error codes for scripting and automation
  - Machine-readable JSON output with `--json` flag
  - Context-aware error messages

- **Developer Features**
  - Hexagonal architecture for testability
  - Comprehensive test coverage
  - Pure Go git operations using go-git
  - Structured logging

### Security

- Input validation for workspace IDs, branch names, and paths
- Path traversal protection
- Safe directory name handling

[Unreleased]: https://github.com/alexisbeaulieu97/canopy/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/alexisbeaulieu97/canopy/releases/tag/v1.0.0
