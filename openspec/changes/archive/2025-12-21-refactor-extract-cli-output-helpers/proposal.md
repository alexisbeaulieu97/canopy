# Change: Extract CLI Output Helpers

## Why

CLI output formatting is inconsistent across commands. There are 64+ `fmt.Printf` statements in `workspace.go` and `repo.go` with varying message styles, some with paths, some without, some with status indicators, some plain. This makes maintaining a consistent user experience difficult and leads to subtle inconsistencies.

## What Changes

- Create `internal/output/cli.go` with standardized output helpers
- Replace direct `fmt.Printf` calls in CLI commands with helper functions
- Establish consistent message patterns:
  - Success messages: action completed
  - Info messages: neutral information
  - Action with path: action completed with filesystem path
- Maintain backward compatibility with existing output (no breaking changes to scripts parsing output)

## Impact

- Affected specs: cli (output consistency)
- Affected code:
  - `internal/output/cli.go` (new)
  - `cmd/canopy/workspace.go` (~30 fmt.Printf calls)
  - `cmd/canopy/repo.go` (~15 fmt.Printf calls)
  - `cmd/canopy/status.go`, `check.go` (minor updates)
- Risk: Low - refactoring only, no behavioral changes
