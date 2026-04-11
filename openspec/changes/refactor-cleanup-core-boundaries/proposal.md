# Change: Refactor cleanup across core boundaries and presentation helpers

## Why
The current codebase is functionally sound, but several internal cleanup opportunities are now blocking maintainability:

- core packages still depend on concrete config types in a few places
- some CLI and TUI helper layers only forward to other packages
- unused abstractions remain in production packages with test-only coverage
- duplicated formatting and bulk-operation orchestration make future changes harder

This refactor keeps behavior stable while reducing duplication, tightening dependency boundaries, and removing abandoned code paths.

## What Changes
- replace `internal/ports` dependencies on concrete `config.*` types with port-owned DTOs/interfaces
- wire hook execution entirely through app-level dependency injection and reuse the hook executor for template setup commands
- remove production-unused output and TUI abstractions
- consolidate byte-size and relative-time formatting behind shared internal helpers
- flatten thin TUI compatibility wrappers
- split oversized command and output files by concern and centralize shared bulk-operation helpers

## Impact
- Affected specs: `core-architecture`
- Affected code:
  - `internal/ports/`
  - `internal/config/`
  - `internal/workspaces/`
  - `internal/hooks/`
  - `internal/output/`
  - `internal/tui/`
  - `cmd/canopy/`
