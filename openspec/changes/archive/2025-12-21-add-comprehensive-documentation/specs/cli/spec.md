## ADDED Requirements

### Requirement: Error Code Documentation
All error codes SHALL be documented for scripting and automation purposes.

#### Scenario: Error code reference
- **GIVEN** a user wants to script canopy commands
- **WHEN** they need to handle specific errors
- **THEN** documentation SHALL list all error codes
- **AND** documentation SHALL map error codes to exit codes
- **AND** documentation SHALL provide handling examples

### Requirement: Command Help Text
All CLI commands SHALL have comprehensive help text accessible via `--help`.

#### Scenario: Help includes examples
- **GIVEN** a user runs `canopy workspace new --help`
- **THEN** the output SHALL include usage examples
- **AND** the output SHALL explain all flags

#### Scenario: Help includes error handling
- **GIVEN** a user runs `canopy --help`
- **THEN** the output SHALL mention where to find error code documentation

