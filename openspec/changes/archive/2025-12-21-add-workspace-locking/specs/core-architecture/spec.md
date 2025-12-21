## ADDED Requirements

### Requirement: Workspace Locking
The system SHALL prevent concurrent mutating operations on the same workspace using file-based locks.

#### Scenario: Lock acquired for create operation
- **WHEN** a workspace creation is initiated
- **THEN** the system SHALL acquire an exclusive lock for the workspace ID
- **AND** release the lock when the operation completes (success or failure)

#### Scenario: Concurrent operations blocked
- **GIVEN** workspace `PROJ-1` has an active lock held by another process
- **WHEN** a second process attempts to close `PROJ-1`
- **THEN** the second process SHALL wait up to `lock_timeout` for the lock
- **AND** fail with `ErrWorkspaceLocked` if timeout expires

#### Scenario: Read operations not locked
- **GIVEN** workspace `PROJ-1` has an active lock
- **WHEN** I run `canopy workspace list` or `canopy workspace status PROJ-1`
- **THEN** the operation SHALL complete without waiting for the lock

#### Scenario: Stale lock cleanup
- **GIVEN** a lock file exists that is older than `lock_stale_threshold`
- **WHEN** another operation attempts to acquire the lock
- **THEN** the system SHALL remove the stale lock
- **AND** SHALL acquire a fresh lock

#### Scenario: Lock released on failure
- **GIVEN** an operation holds a lock on workspace `PROJ-1`
- **WHEN** the operation fails with an error
- **THEN** the lock SHALL be released
- **AND** subsequent operations SHALL proceed without waiting
