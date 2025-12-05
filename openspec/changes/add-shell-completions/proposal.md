# Change: Add Shell Completions

## Why
Tab completion significantly improves CLI usability:
- Complete workspace IDs without typing full names
- Complete repo names from registry
- Complete subcommands and flags
- Reduces typos and improves discoverability

Cobra has built-in completion support, making this a quick win.

## What Changes
- Add `canopy completion <shell>` command (bash, zsh, fish, powershell)
- Enable dynamic completion for:
  - Workspace IDs in `workspace` subcommands
  - Repository names in `repo` subcommands
  - Registry aliases
- Document installation in README

## Impact
- **Affected specs**: `specs/cli/spec.md`
- **Affected code**:
  - `cmd/canopy/completion.go` - New file for completion command
  - `cmd/canopy/workspace.go` - Add completion functions
  - `cmd/canopy/repo.go` - Add completion functions
  - `README.md` - Add installation instructions
- **Risk**: Very Low - Additive feature using Cobra built-ins
