## ADDED Requirements

### Requirement: CLI Output Consistency

The CLI SHALL use standardized output helpers for all user-facing messages to ensure consistent formatting across commands.

Output helpers SHALL provide the following message types:
- Success messages for completed actions
- Info messages for neutral information
- Warning messages for non-fatal issues
- Path-aware messages that include filesystem locations

#### Scenario: Success message format

- **WHEN** a CLI command completes successfully
- **THEN** the output SHALL follow the pattern: "[Action] [target]" (e.g., "Created workspace foo")

#### Scenario: Success message with path

- **WHEN** a CLI command creates or modifies a filesystem resource
- **THEN** the output SHALL include the path: "[Action] [target] in [path]"

#### Scenario: Consistent verb usage

- **WHEN** displaying success messages
- **THEN** past tense verbs SHALL be used (Created, Closed, Removed, Renamed, Added)
