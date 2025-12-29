## ADDED Requirements

### Requirement: Progress Indicators for Long Operations
Long-running CLI operations SHALL display progress indicators to provide feedback to users.

#### Scenario: Bulk sync with progress
- **WHEN** user runs `workspace sync --pattern`
- **AND** operation includes multiple workspaces
- **THEN** progress bar MUST show completion percentage
- **AND** current workspace name MUST be displayed

#### Scenario: Non-interactive environment
- **WHEN** output is not a TTY
- **THEN** progress bar MUST be disabled
- **AND** simple status messages MUST be shown instead

#### Scenario: Progress with --no-progress flag
- **WHEN** user specifies `--no-progress` flag
- **THEN** progress indicators MUST be disabled
- **AND** only final results MUST be shown

#### Scenario: Cancellation during progress
- **WHEN** user presses Ctrl+C during operation
- **THEN** progress bar MUST show "Cancelled"
- **AND** partial results MUST be reported
