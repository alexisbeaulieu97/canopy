## ADDED Requirements

### Requirement: Named Timeout Constants

Timeout values SHALL be defined as named constants with documentation explaining their purpose and rationale.

#### Scenario: Cleanup operation timeout

- **WHEN** a cleanup operation requires a timeout context
- **THEN** the code SHALL use a named constant (e.g., `DefaultCleanupTimeout`)
- **AND** the constant SHALL include a documentation comment

#### Scenario: No magic timeout numbers

- **WHEN** reviewing CLI command handlers
- **THEN** no inline magic numbers for timeouts SHALL be present
- **AND** all timeouts SHALL reference named constants
