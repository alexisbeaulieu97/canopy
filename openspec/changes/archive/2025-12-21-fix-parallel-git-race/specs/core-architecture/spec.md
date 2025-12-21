## ADDED Requirements

### Requirement: Parallel Git Operations Early Termination
When running git commands in parallel with `continueOnError=false`, the system SHALL cancel remaining operations after the first failure.

#### Scenario: First error cancels pending operations
- **GIVEN** a workspace with 5 repositories
- **AND** parallel git execution is enabled
- **AND** `continueOnError` is false
- **WHEN** the first repository operation fails
- **THEN** pending operations SHALL be cancelled
- **AND** running operations SHALL be signalled to stop
- **AND** the function SHALL return the first error

#### Scenario: All operations complete when continueOnError is true
- **GIVEN** a workspace with 5 repositories
- **AND** parallel git execution is enabled
- **AND** `continueOnError` is true
- **WHEN** some repository operations fail
- **THEN** all operations SHALL complete
- **AND** all results (success and failure) SHALL be returned

#### Scenario: No race conditions in result collection
- **GIVEN** parallel git execution
- **WHEN** multiple goroutines write results concurrently
- **THEN** all results SHALL be collected without data races
- **AND** `go test -race` SHALL pass

