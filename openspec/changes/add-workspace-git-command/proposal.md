# Change: Add Workspace Git Command

## Why
Users frequently need to run the same git command across all repos in a workspace (fetch, pull, push, stash, checkout). Currently this requires manual iteration or scripting. A `canopy workspace git` subcommand provides a unified way to execute arbitrary git commands across all repos, eliminating the need for multiple specialized commands.

## What Changes
- Add `canopy workspace git <WORKSPACE-ID> <git-args...>` subcommand
- Execute the git command in each repo within the workspace
- Show per-repo output with clear separation
- Support `--parallel` flag for concurrent execution
- Support `--continue-on-error` to not stop on first failure
- Exit with non-zero if any repo fails (unless `--continue-on-error`)

## Impact
- Affected specs: `specs/cli/spec.md`
- Affected code:
  - `cmd/canopy/workspace.go` - Add new `git` subcommand
  - `internal/workspaces/service.go` - Add `RunGitInWorkspace()` method
  - `internal/gitx/git.go` - Add generic command execution helper
