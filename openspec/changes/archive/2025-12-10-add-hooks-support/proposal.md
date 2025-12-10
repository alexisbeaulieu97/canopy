# Change: Add Lifecycle Hooks Support

## Why
Users need automated setup/teardown tasks (npm install, go mod download, IDE/env config) after workspace creation and before deletion. Post-create and pre-close hooks enable this automation without manual intervention.

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
- **Risk**: Medium - Executes arbitrary commands; see `design.md` for security mitigations, threat model, and failure handling
