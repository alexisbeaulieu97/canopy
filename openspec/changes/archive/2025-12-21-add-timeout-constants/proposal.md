# Change: Add Named Timeout Constants

## Why

Magic numbers for timeouts appear in the codebase without clear documentation of their purpose or rationale. For example, `cmd/canopy/repo.go:106` uses `30*time.Second` for cleanup operations. Named constants improve code readability and make it easier to adjust timeouts consistently if needed.

## What Changes

- Define named constants for all timeout values used in the codebase
- Replace magic timeout numbers with named constants
- Document the rationale for each timeout value

## Impact

- Affected specs: core-architecture (code clarity)
- Affected code:
  - `internal/config/constants.go` or similar (new constants)
  - `cmd/canopy/repo.go` (use constant instead of magic number)
- Risk: Very low - constant extraction only
