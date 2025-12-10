# Change: Add Context Propagation to Service Layer

## Why
The service layer currently lacks `context.Context` support, preventing cancellation, timeout propagation, and observability hooks. This is a foundational change that enables future improvements like request tracing, graceful shutdown, and timeout handling for long-running git operations.

## What Changes
- Add `context.Context` as first parameter to all public `Service` methods
- Propagate context through git operations and hook execution
- Add timeout support for git network operations (clone, fetch, push, pull)
- Enable cancellation of parallel git operations
- **BREAKING**: All `Service` method signatures change

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/workspaces/service.go` - All public methods
  - `internal/ports/git.go` - GitOperations interface
  - `internal/gitx/git.go` - Implementation
  - `internal/hooks/executor.go` - Already uses context internally
  - `cmd/canopy/*.go` - All command handlers
  - `internal/tui/commands.go` - TUI commands
- **Risk**: Medium - Breaking change to core interface, but internal-only

