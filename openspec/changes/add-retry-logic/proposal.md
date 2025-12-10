# Change: Add Retry Logic for Git Network Operations

## Why
Git network operations (clone, fetch, push, pull) can fail due to transient network issues, rate limiting, or temporary server unavailability. Currently, these failures require manual retry. Adding automatic retry with exponential backoff improves reliability and user experience.

## What Changes
- Add retry wrapper for git network operations
- Implement exponential backoff with jitter
- Make retry configurable (max attempts, initial delay)
- Add retry-specific logging
- Preserve original error on final failure

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/gitx/retry.go` - New file for retry logic
  - `internal/gitx/git.go` - Wrap network operations
  - `internal/config/config.go` - Add retry configuration (optional)
- **Risk**: Low - Improves reliability, configurable behavior

