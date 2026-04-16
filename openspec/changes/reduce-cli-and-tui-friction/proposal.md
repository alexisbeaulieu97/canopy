# Change: Reduce CLI and TUI friction

## Why

Several high-frequency flows still create avoidable friction through dead-end workspace creation, noisy runtime failures, low-signal list/detail output, and TUI feedback that does not match real triage workflows.

## Problem Details

- `workspace new` can create empty, dead-end workspaces
- runtime command failures still leak Cobra usage noise
- list/detail surfaces hide useful triage context
- TUI search and bulk feedback are narrower than the actual user workflow

## What Changes

- **BREAKING** Guard workspace creation when no repositories resolve and return actionable guidance
- Suppress Cobra usage for runtime/domain failures while preserving parse-time help
- Expand list/detail output with richer metadata and multi-signal status summaries
- Improve TUI search, bulk-operation summaries, and stale error clearing
- Update docs to reflect the new first-run and recovery paths

## Impact

- Affected specs: `cli`, `tui`
- Affected code:
  - `cmd/canopy`
  - `internal/tui`
  - `internal/errors`
  - `README.md`, `docs/quick-start.md`, `docs/usage.md`
