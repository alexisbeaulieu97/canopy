# Change: Add Dry Run Mode

## Why
Destructive operations like `workspace close` and `repo remove` permanently delete data. Users want to:
- Preview what would be deleted before committing
- Verify the correct workspace/repo is targeted
- Safely explore commands without consequences

A `--dry-run` flag provides this safety net.

## What Changes
- Add `--dry-run` flag to destructive commands:
  - `canopy workspace close <ID>`
  - `canopy repo remove <NAME>`
- Dry run shows what would happen without executing
- Output clearly indicates it's a preview
- Works with `--json` for scripted verification

## Impact
- **Affected specs**: `specs/cli/spec.md`
- **Affected code**:
  - `cmd/canopy/workspace.go` - Add --dry-run to close
  - `cmd/canopy/repo.go` - Add --dry-run to remove
  - `internal/workspaces/service.go` - Add dry run variants or parameter
- **Risk**: Very Low - Additive safety feature
