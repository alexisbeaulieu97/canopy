## ADDED Requirements

### Requirement: Parallel Repository Operations
Repository operations during workspace creation SHALL execute in parallel with bounded concurrency.

#### Scenario: Parallel EnsureCanonical execution
- **WHEN** creating a workspace with multiple repositories
- **THEN** EnsureCanonical operations SHALL execute in parallel
- **AND** the number of concurrent operations SHALL be limited by `parallel_workers` config
- **AND** worktree creation SHALL wait for the corresponding EnsureCanonical to complete

#### Scenario: Configurable worker count
- **GIVEN** config has `parallel_workers: 6`
- **WHEN** creating a workspace with 10 repositories
- **THEN** at most 6 EnsureCanonical operations SHALL run concurrently

#### Scenario: Default worker count
- **GIVEN** `parallel_workers` is not configured
- **WHEN** creating a workspace with multiple repositories
- **THEN** the default of 4 concurrent operations SHALL be used

#### Scenario: Error handling in parallel mode
- **GIVEN** a workspace creation with 4 repositories
- **WHEN** one EnsureCanonical operation fails
- **THEN** remaining operations SHALL be cancelled (fail-fast)
- **AND** successfully cloned repositories SHALL be cleaned up
- **AND** the error message SHALL indicate which repository failed

#### Scenario: Context cancellation propagates to workers
- **WHEN** workspace creation context is cancelled
- **THEN** all parallel operations SHALL receive cancellation
- **AND** the operation SHALL return promptly with a cancellation error

