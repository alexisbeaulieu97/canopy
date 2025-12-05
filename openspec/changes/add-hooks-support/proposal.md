# Change: Add Lifecycle Hooks Support

## Why
Users often need to run setup commands after creating a workspace:
- `npm install` in JavaScript projects
- `go mod download` in Go projects
- IDE configuration scripts
- Environment setup

Currently, users must manually run these after `canopy workspace new`. Hooks enable automation:
- Post-create hooks run after workspace creation
- Pre-close hooks run before workspace deletion
- Per-workspace or global hooks

## What Changes
- Add `hooks` section to config.yaml:
  ```yaml
  hooks:
    post_create:
      - command: "npm install"
        repos: ["frontend"]  # optional filter
      - command: "go mod download"
        repos: ["backend"]
    pre_close:
      - command: "git stash"
  ```
- Execute hooks at appropriate lifecycle points
- Add `--no-hooks` flag to skip execution
- Log hook output with clear attribution

## Impact
- **Affected specs**: `specs/workspace-management/spec.md`
- **Affected code**:
  - `internal/config/config.go` - Add Hook types
  - `internal/workspaces/service.go` - Execute hooks in lifecycle methods
  - `cmd/canopy/workspace.go` - Add `--no-hooks` flag
- **Risk**: Medium - Executes arbitrary commands, needs security consideration
