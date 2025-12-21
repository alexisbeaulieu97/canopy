## ADDED Requirements

### Requirement: Unified Parallel Execution
The system SHALL use a consistent pattern for parallel operations across all components.

#### Scenario: Parallel execution with worker limit
- **WHEN** executing operations in parallel
- **THEN** the system SHALL respect the configured `parallel_workers` limit
- **AND** not exceed the maximum concurrent operations

#### Scenario: Context cancellation propagation
- **WHEN** the parent context is cancelled
- **THEN** all parallel operations SHALL be cancelled
- **AND** the system SHALL return promptly with cancellation error

#### Scenario: Error aggregation
- **WHEN** multiple parallel operations fail
- **THEN** the system SHALL collect all errors
- **AND** return a summary of failures

## MODIFIED Requirements

### Requirement: Context Propagation
All service methods SHALL propagate the caller's context to child operations.

#### Scenario: Timeout propagation
- **WHEN** caller provides context with timeout
- **THEN** all child operations SHALL respect the timeout
- **AND** return context.DeadlineExceeded when exceeded

#### Scenario: Hook execution with caller context
- **WHEN** hooks are executed during workspace operations
- **THEN** the hook executor SHALL derive timeout from caller context
- **AND** respect per-hook timeout configuration
