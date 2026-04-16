## MODIFIED Requirements

### Requirement: Search Filter
The TUI SHALL support searching workspaces by ID using built-in list filtering.

#### Scenario: Search workspaces
- **WHEN** user presses `/` key
- **THEN** search input SHALL appear
- **AND** list SHALL filter in real-time as user types
- **AND** pressing Enter SHALL accept filter
- **AND** pressing Esc SHALL cancel search

#### Scenario: Search repo names, branches, and status keywords
- **GIVEN** workspace metadata and cached status are loaded
- **WHEN** the user searches for a repo name, branch name, or keyword such as `dirty`, `unpushed`, `behind`, `stale`, `error`, `locked`, or `orphan`
- **THEN** matching workspaces SHALL remain visible even if the workspace ID itself does not match

## ADDED Requirements

### Requirement: Bulk Action Summaries
The TUI SHALL report explicit bulk-operation outcomes for push, sync, and close actions.

#### Scenario: Partial bulk success
- **GIVEN** a bulk action targets multiple workspaces
- **AND** one or more workspaces fail while others succeed
- **WHEN** the action completes
- **THEN** the TUI SHALL display a summary containing succeeded count, failed count, and failed workspace IDs
- **AND** the summary SHALL not collapse to only the first error

### Requirement: Error Recovery Feedback
The TUI SHALL clear stale error banners after later successful reloads and operations.

#### Scenario: Successful reload clears stale error
- **GIVEN** the header shows a previous error
- **WHEN** a subsequent workspace reload succeeds
- **THEN** the stale error banner SHALL be cleared

#### Scenario: Successful operation clears stale error
- **GIVEN** the header shows a previous error
- **WHEN** a later push, sync, close, or open-editor action succeeds
- **THEN** the stale error banner SHALL be cleared

### Requirement: Actionable Zero-Repo Detail State
The TUI SHALL provide actionable guidance for workspaces that contain zero repositories.

#### Scenario: Zero-repo detail state
- **GIVEN** the selected workspace contains no repositories
- **WHEN** the user opens the detail view
- **THEN** the repositories section SHALL show an explicit empty state
- **AND** it SHALL suggest `canopy workspace repo add <workspace> <repo>`
