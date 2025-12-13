## ADDED Requirements

### Requirement: Automatic Retry for Git Network Operations
Git network operations SHALL automatically retry on transient failures using exponential backoff.

#### Scenario: Transient failure triggers retry
- **GIVEN** a git clone operation
- **WHEN** the operation fails with a network timeout
- **THEN** the operation SHALL be retried
- **AND** subsequent attempts SHALL use exponential backoff

#### Scenario: Permanent failure does not retry
- **GIVEN** a git clone operation
- **WHEN** the operation fails with authentication error (401/403)
- **THEN** the operation SHALL NOT be retried
- **AND** the error SHALL be returned immediately

#### Scenario: Max attempts exceeded
- **GIVEN** retry configuration with max_attempts=3
- **WHEN** all 3 attempts fail with transient errors
- **THEN** the final error SHALL be returned
- **AND** the error message SHALL indicate retry exhaustion

### Requirement: Exponential Backoff with Jitter
Retry delays SHALL use exponential backoff with random jitter to prevent thundering herd.

#### Scenario: Backoff calculation
- **GIVEN** initial_delay=1s and multiplier=2
- **WHEN** calculating delay for attempt N
- **THEN** base delay SHALL be initial_delay * (multiplier ^ (N-1))
- **AND** jitter SHALL be applied (±25% of base delay)
- **AND** delay SHALL not exceed max_delay

#### Scenario: Jitter prevents synchronized retries
- **GIVEN** multiple concurrent operations failing
- **WHEN** retries are scheduled
- **THEN** retry times SHALL be randomized
- **AND** NOT synchronized to the same instant

### Requirement: Retry Logging
Retry attempts SHALL be logged for debugging and observability.

#### Scenario: Log retry attempt
- **WHEN** a retry is attempted
- **THEN** an Info-level log message SHALL be emitted
- **AND** the log SHALL include attempt number, operation, and delay

#### Scenario: Log final failure
- **WHEN** all retry attempts are exhausted
- **THEN** a Warning-level log message SHALL be emitted
- **AND** the log SHALL include total attempts and final error

### Requirement: Context-Aware Retry
Retry operations SHALL respect context cancellation and deadlines.

#### Scenario: Context cancelled during backoff
- **GIVEN** a retry operation waiting for backoff delay
- **WHEN** the context is cancelled
- **THEN** the retry SHALL be aborted immediately
- **AND** a context cancelled error SHALL be returned

#### Scenario: Context deadline during retry
- **GIVEN** a context with 5-second deadline
- **AND** retry backoff would exceed deadline
- **THEN** retry SHALL be skipped
- **AND** the last error SHALL be returned

