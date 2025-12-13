## ADDED Requirements

### Requirement: Context Propagation in Service Layer
All public Service methods SHALL accept `context.Context` as their first parameter to enable cancellation, timeout, and observability.

#### Scenario: Service method accepts context
- **WHEN** a CLI command calls a Service method
- **THEN** the command SHALL pass `cmd.Context()` as the first argument
- **AND** the Service SHALL propagate context to downstream operations

#### Scenario: Context cancellation stops operation
- **WHEN** context is cancelled during a git operation
- **THEN** the operation SHALL return an error
- **AND** any spawned goroutines SHALL terminate

#### Scenario: Context timeout for network operations
- **WHEN** a git network operation exceeds the context deadline
- **THEN** the operation SHALL return a context deadline exceeded error
- **AND** partial operations SHALL be cleaned up where possible

### Requirement: Git Operations Context Support
Git network operations (Clone, Fetch, Push, Pull) SHALL respect context cancellation and deadlines.

#### Scenario: Clone respects context timeout
- **WHEN** `GitOperations.Clone(ctx, url, name)` is called with a deadline
- **AND** the clone exceeds the deadline
- **THEN** the clone SHALL be aborted
- **AND** a context deadline exceeded error SHALL be returned

#### Scenario: Fetch respects context cancellation
- **WHEN** `GitOperations.Fetch(ctx, name)` is called
- **AND** the context is cancelled during fetch
- **THEN** the fetch SHALL stop
- **AND** a context cancelled error SHALL be returned

#### Scenario: Parallel operations cancel on context done
- **WHEN** `RunGitInWorkspace` executes in parallel mode
- **AND** the context is cancelled
- **THEN** all pending goroutines SHALL be signalled to stop
- **AND** the function SHALL return promptly

### Requirement: Default Network Timeout
Network-bound git operations SHALL use a default timeout when no deadline is set on the context.

#### Scenario: Default timeout applied
- **GIVEN** a context with no deadline
- **WHEN** a git network operation is initiated
- **THEN** a default timeout of 5 minutes SHALL be applied
- **AND** operations exceeding this timeout SHALL fail

#### Scenario: Explicit deadline overrides default
- **GIVEN** a context with an explicit deadline
- **WHEN** a git network operation is initiated
- **THEN** the explicit deadline SHALL be used
- **AND** the default timeout SHALL NOT apply

