## ADDED Requirements

### Requirement: Standard Library Concurrency Patterns
Concurrent operations SHALL use standard Go concurrency patterns from `golang.org/x/sync` where applicable, rather than custom implementations.

#### Scenario: Parallel repo operations use errgroup
- **WHEN** multiple repositories need concurrent operations (e.g., EnsureCanonical)
- **THEN** the implementation SHALL use `errgroup.Group` for coordination
- **AND** bounded concurrency SHALL be configured via `SetLimit()`

#### Scenario: Fail-fast on first error
- **WHEN** a concurrent operation fails
- **THEN** remaining operations SHALL be cancelled via errgroup's context
- **AND** the first error SHALL be returned to the caller
