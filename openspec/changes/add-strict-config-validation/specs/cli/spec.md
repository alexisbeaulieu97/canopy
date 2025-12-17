## ADDED Requirements

### Requirement: Strict Config Validation
The system SHALL validate configuration files strictly to catch typos and invalid values early.

#### Scenario: Unknown config field rejected
- **GIVEN** config file contains `parrallel_workers: 8` (typo)
- **WHEN** config is loaded
- **THEN** system returns error: "unknown config field 'parrallel_workers', did you mean 'parallel_workers'?"
- **AND** suggests similar known fields when possible

#### Scenario: Valid config accepted
- **GIVEN** config file contains only valid, known fields
- **WHEN** config is loaded
- **THEN** config is loaded successfully
- **AND** all values are applied as specified

#### Scenario: Hook timeout validation
- **GIVEN** config contains hook with `timeout: -5`
- **WHEN** config is loaded
- **THEN** system returns error: "hook timeout must be positive"

#### Scenario: Config validate command
- **WHEN** user runs `canopy config validate`
- **THEN** system loads and validates config
- **AND** reports any validation errors
- **AND** exits with code 0 if valid, non-zero if invalid

#### Scenario: Config validate with path
- **WHEN** user runs `canopy config validate --config /path/to/config.yaml`
- **THEN** system validates the specified config file
- **AND** does not use default config search paths
