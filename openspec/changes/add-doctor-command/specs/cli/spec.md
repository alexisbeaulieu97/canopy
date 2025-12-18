## ADDED Requirements

### Requirement: Doctor Command
The `canopy doctor` command SHALL validate the environment and configuration, reporting issues with actionable guidance.

#### Scenario: All checks pass
- **GIVEN** a properly configured Canopy environment
- **WHEN** I run `canopy doctor`
- **THEN** the output SHALL show all checks as passing
- **AND** the exit code SHALL be 0

#### Scenario: Git not installed
- **GIVEN** git is not installed or not in PATH
- **WHEN** I run `canopy doctor`
- **THEN** the output SHALL show an error for git availability
- **AND** suggest installing git
- **AND** the exit code SHALL be 2

#### Scenario: Invalid config file
- **GIVEN** `~/.canopy/config.yaml` contains invalid YAML
- **WHEN** I run `canopy doctor`
- **THEN** the output SHALL show an error for config validation
- **AND** include the parse error details
- **AND** the exit code SHALL be 2

#### Scenario: Missing directories with fix flag
- **GIVEN** `projects_root` directory does not exist
- **WHEN** I run `canopy doctor --fix`
- **THEN** the system SHALL create the missing directory
- **AND** report that it was auto-fixed
- **AND** the exit code SHALL be 0

#### Scenario: JSON output for scripting
- **WHEN** I run `canopy doctor --json`
- **THEN** the output SHALL be valid JSON
- **AND** include an array of check results with name, status, and message

#### Scenario: Warning for stale canonical repos
- **GIVEN** a canonical repo has not been fetched in over 30 days
- **WHEN** I run `canopy doctor`
- **THEN** the output SHALL show a warning for the stale repo
- **AND** the exit code SHALL be 1
