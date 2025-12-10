# core Specification Delta

## MODIFIED Requirements

### Requirement: Two-Phase Config Validation
Config validation SHALL be split into pure value validation and environment validation phases.

#### Scenario: ValidateValues checks config values
- **WHEN** `ValidateValues()` is called
- **THEN** it validates `CloseDefault` is "delete" or "archive"
- **AND** validates regex patterns compile
- **AND** validates `StaleThresholdDays >= 0`
- **AND** does not check filesystem paths

#### Scenario: ValidateEnvironment checks paths
- **WHEN** `ValidateEnvironment()` is called
- **THEN** it checks configured paths exist
- **AND** checks paths are directories
- **AND** optionally checks paths are writable

#### Scenario: Validate runs both phases
- **WHEN** `Validate()` is called
- **THEN** it calls `ValidateValues()` first
- **AND** if values pass, calls `ValidateEnvironment()`
- **AND** value errors are reported before environment errors

#### Scenario: Test with invalid values
- **GIVEN** a config with invalid `CloseDefault` value
- **WHEN** `ValidateValues()` is called
- **THEN** validation fails without checking filesystem
- **AND** error message identifies the invalid field
