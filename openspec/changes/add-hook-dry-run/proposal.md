# Change: Add Hook Dry-Run Mode

## Why
Users configuring hooks have no safe way to verify their commands before execution. A typo or misconfigured hook can cause unexpected side effects. A dry-run mode would let users preview what commands would run without actually executing them, making hook debugging safer and faster.

## What Changes
- Add `--dry-run-hooks` flag to commands that trigger hooks (`workspace new`, `workspace close`)
- When enabled, print the hook commands that would execute without running them
- Show resolved variables (workspace ID, repo name, branch) in the preview
- Support `canopy hooks list` command to show configured hooks and their triggers
- Add `canopy hooks test <hook-name> --workspace <id>` to dry-run a specific hook

## Impact
- Affected specs: `specs/cli/spec.md`
- Affected code:
  - `internal/hooks/executor.go` - Add dry-run mode to executor
  - `cmd/canopy/workspace.go` - Add `--dry-run-hooks` flag
  - `cmd/canopy/hooks.go` (new) - Add hooks subcommand
- Breaking changes: None (additive only)
