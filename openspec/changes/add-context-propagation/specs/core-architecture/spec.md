## MODIFIED Requirements

### Requirement: Context Propagation
All service methods that perform I/O operations or may take significant time SHALL accept a context.Context parameter for cancellation and timeout handling.

#### Scenario: Cancellation during workspace lookup
- **WHEN** a workspace lookup is in progress
- **AND** the parent context is cancelled
- **THEN** the lookup operation MUST terminate promptly
- **AND** return a context.Cancelled error

#### Scenario: Timeout during workspace operations
- **WHEN** a workspace operation exceeds its timeout
- **THEN** the operation MUST terminate
- **AND** return a context.DeadlineExceeded error

#### Scenario: Ctrl+C during long operation
- **WHEN** user presses Ctrl+C during a long-running operation
- **THEN** all child operations MUST receive cancellation signal
- **AND** terminate gracefully
