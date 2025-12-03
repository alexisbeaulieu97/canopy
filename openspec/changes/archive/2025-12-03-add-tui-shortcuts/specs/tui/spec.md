# TUI Specification Deltas

Note: This change documents shortcuts already implemented in `enhance-status-dashboard`.

## ADDED Requirements

### Requirement: Push Shortcut
The TUI SHALL provide a keyboard shortcut to push all repos in the selected workspace.

#### Scenario: Push all repos with confirmation
- **GIVEN** workspace is selected in list
- **WHEN** user presses `p` key
- **THEN** confirmation prompt SHALL appear
- **AND** confirming with `y` SHALL push all repos
- **AND** declining with `n` SHALL cancel
- **AND** loading spinner SHALL display during push

### Requirement: Open Editor Shortcut
The TUI SHALL provide a shortcut to open workspaces in the user's editor.

#### Scenario: Open in editor
- **GIVEN** workspace is selected in list
- **WHEN** user presses `o` key
- **THEN** workspace directory SHALL open in `$VISUAL` or `$EDITOR`

#### Scenario: No editor configured
- **GIVEN** neither `$VISUAL` nor `$EDITOR` is set
- **WHEN** user presses `o` key
- **THEN** error message SHALL display explaining how to set editor

### Requirement: Stale Filter Shortcut
The TUI SHALL provide a shortcut to toggle the stale workspace filter.

#### Scenario: Toggle stale filter
- **WHEN** user presses `s` key
- **THEN** list SHALL show only stale workspaces
- **AND** pressing `s` again SHALL clear the filter
- **AND** header SHALL indicate active filter

### Requirement: Search Filter
The TUI SHALL support searching workspaces by ID using built-in list filtering.

#### Scenario: Search workspaces
- **WHEN** user presses `/` key
- **THEN** search input SHALL appear
- **AND** list SHALL filter in real-time as user types
- **AND** pressing Enter SHALL accept filter
- **AND** pressing Esc SHALL cancel search

### Requirement: Close Shortcut
The TUI SHALL provide a keyboard shortcut to close the selected workspace.

#### Scenario: Close workspace with confirmation
- **GIVEN** workspace is selected in list
- **WHEN** user presses `c` key
- **THEN** confirmation prompt SHALL appear
- **AND** confirming with `y` SHALL close the workspace
- **AND** declining with `n` SHALL cancel

```
