# Change: Add Shell Completions

## Why
Tab completion improves CLI usability by enabling workspace ID and repo name completion, reducing typos, and improving discoverability. Cobra's built-in completion support makes this a quick win.

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
