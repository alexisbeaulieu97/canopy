# Change: Add Git Hooks Integration

## Why
Development teams often have standard git hooks for linting, formatting, or pre-commit checks. Currently, users must manually copy hooks to each worktree's .git/hooks/ directory. Automating hook setup via configuration ensures consistency across all workspaces and reduces setup friction.

## What Changes
- Add `git_hooks` section to config.yaml for defining hook scripts
- Automatically install configured hooks when creating workspace worktrees
- Support all standard git hooks (pre-commit, pre-push, etc.)
- Add `yard workspace sync-hooks <ID>` command to update hooks in existing workspaces
- Support both inline scripts and file paths for hook definitions
- Make hooks executable automatically

## Impact
- Affected specs: `specs/git-operations/spec.md` (new)
- Affected code:
  - `internal/config/config.go` - Add GitHooks configuration
  - `internal/workspaces/service.go:130-137` - Install hooks after worktree creation
  - `internal/gitx/hooks.go` (new) - Hook installation utilities
  - `cmd/yard/workspace.go` - Add sync-hooks command
