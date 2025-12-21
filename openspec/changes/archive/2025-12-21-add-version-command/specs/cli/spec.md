## ADDED Requirements

### Requirement: Version Command
The CLI SHALL provide a `version` command to display build information.

#### Scenario: Display version information
- **WHEN** user runs `canopy version`
- **THEN** the output SHALL include:
  - Version string (from git tag or "dev")
  - Git commit hash (short form)
  - Build date in ISO 8601 format
  - Go version used for compilation

#### Scenario: JSON version output
- **WHEN** user runs `canopy version --json`
- **THEN** the output SHALL be valid JSON
- **AND** SHALL include version, commit, buildDate, and goVersion fields

#### Scenario: Version flag on root command
- **WHEN** user runs `canopy --version`
- **THEN** the version string SHALL be printed
- **AND** the program SHALL exit with code 0

### Requirement: Version Embedding at Build Time
The version information SHALL be embedded at build time using Go linker flags.

#### Scenario: Tagged release build
- **GIVEN** the repository has a git tag `v1.2.3`
- **WHEN** the binary is built with ldflags
- **THEN** `canopy version` SHALL display `v1.2.3`

#### Scenario: Development build
- **GIVEN** no ldflags are provided during build
- **WHEN** `canopy version` is run
- **THEN** version SHALL display "dev"
- **AND** commit SHALL display "unknown"

