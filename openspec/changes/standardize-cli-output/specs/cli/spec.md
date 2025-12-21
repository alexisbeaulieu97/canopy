## ADDED Requirements

### Requirement: Centralized Output Styling
The system SHALL use centralized style definitions for CLI output.

#### Scenario: Color output in TTY
- **WHEN** output is to a TTY
- **THEN** colors SHALL be applied using lipgloss styles

#### Scenario: Plain output in non-TTY
- **WHEN** output is not to a TTY (piped or redirected)
- **THEN** ANSI color codes SHALL be omitted

#### Scenario: Consistent separator formatting
- **WHEN** displaying section separators
- **THEN** the system SHALL use consistent width and character

## MODIFIED Requirements

### Requirement: JSON Error Output
The system SHALL provide consistent JSON error format across all commands.

#### Scenario: Command fails with error
- **WHEN** any command fails with --json flag
- **THEN** the output SHALL include `success: false`
- **AND** include `error.code` with the CanopyError code
- **AND** include `error.message` with user-friendly message
- **AND** optionally include `error.context` with structured data

#### Scenario: Successful command output
- **WHEN** any command succeeds with --json flag
- **THEN** the output SHALL include `success: true`
- **AND** include command-specific data payload
