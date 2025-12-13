## Context

The current service layer has no mechanism for:
- Cancelling long-running operations (e.g., large git clones)
- Setting timeouts on network operations
- Propagating request context for tracing/observability
- Graceful shutdown during TUI operations

Go's `context.Context` is the standard solution for these concerns.

## Goals / Non-Goals

**Goals:**
- Add context.Context to all Service public methods
- Enable cancellation of parallel git operations
- Support timeouts for network-bound git operations
- Maintain backward compatibility in behavior (only signatures change)

**Non-Goals:**
- Adding tracing/metrics (future work)
- Changing error types or return values
- Adding new functionality beyond context support

## Decisions

### Decision: Context as First Parameter
All Service methods will accept `context.Context` as their first parameter, following Go conventions.

**Alternatives considered:**
- Store context in Service struct: Rejected - context should flow with requests, not be stored
- Optional context via functional options: Rejected - adds complexity, context is fundamental

### Decision: GitOperations Interface Update
The `ports.GitOperations` interface will be updated to accept context for network operations only (Clone, Fetch, Push, Pull). Local operations (Status, Checkout) don't need context.

**Rationale:** Only network operations benefit from cancellation/timeout.

### Decision: Default Timeout for Network Operations
Network operations will use a configurable default timeout (5 minutes) when no deadline is set on the context.

**Rationale:** Prevents indefinite hangs on network issues.

## Risks / Trade-offs

- **Risk**: Breaking change to all command handlers
  - **Mitigation**: Internal-only change, no external API
  
- **Risk**: Increased code verbosity
  - **Mitigation**: Context is Go idiom, developers expect it

## Migration Plan

1. Update `ports.GitOperations` interface with context
2. Update `gitx.GitEngine` implementation
3. Update `workspaces.Service` methods one by one
4. Update all CLI command handlers
5. Update TUI commands
6. Update all tests

**Rollback:** Revert commit - no data migration needed

## Open Questions

- Should we add a global timeout configuration option?
- Should context cancellation trigger cleanup of partial operations?

