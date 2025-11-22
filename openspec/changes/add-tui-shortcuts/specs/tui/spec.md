# TUI Specification Deltas

## ADDED Requirements

### Requirement: Git Operation Shortcuts
The TUI SHALL provide keyboard shortcuts for common git operations on the selected workspace.

#### Scenario: Fetch all repos
- **GIVEN** workspace is selected in list
- **WHEN** user presses `f` key
- **THEN** git fetch SHALL execute in all repos
- **AND** loading spinner SHALL display during operation
- **AND** result message SHALL show on completion

#### Scenario: Pull all repos
- **GIVEN** workspace is selected in list
- **WHEN** user presses `P` (shift+p) key
- **THEN** git pull SHALL execute in all repos
- **AND** loading spinner SHALL display during operation

#### Scenario: Push all repos with confirmation
- **GIVEN** workspace is selected in list
- **WHEN** user presses `p` key
- **THEN** confirmation prompt SHALL appear
- **AND** confirming with `y` SHALL push all repos
- **AND** declining with `n` SHALL cancel

### Requirement: External Application Shortcuts
The TUI SHALL provide shortcuts to open workspaces in external applications.

#### Scenario: Open in browser
- **GIVEN** workspace is selected in list
- **WHEN** user presses `g` key
- **THEN** default browser SHALL open with repo remote URLs

#### Scenario: Open in editor
- **GIVEN** workspace is selected in list
- **WHEN** user presses `o` key
- **THEN** workspace directory SHALL open in `$EDITOR`

### Requirement: Filter Toggle Shortcuts
The TUI SHALL provide shortcuts to toggle workspace list filters.

#### Scenario: Toggle dirty filter
- **WHEN** user presses `D` key
- **THEN** list SHALL show only dirty workspaces
- **AND** pressing `D` again SHALL clear the filter

#### Scenario: Toggle behind-remote filter
- **WHEN** user presses `B` key
- **THEN** list SHALL show only workspaces behind remote
- **AND** pressing `B` again SHALL clear the filter

#### Scenario: Combined filters
- **GIVEN** dirty filter is active
- **WHEN** user presses `B` key
- **THEN** both filters SHALL be active
- **AND** list SHALL show workspaces matching both criteria

### Requirement: Help Overlay
The TUI SHALL display a help overlay showing all available keyboard shortcuts.

#### Scenario: Show help
- **WHEN** user presses `?` key
- **THEN** help overlay SHALL appear
- **AND** overlay SHALL list all available shortcuts with descriptions

#### Scenario: Dismiss help
- **GIVEN** help overlay is visible
- **WHEN** user presses `?` or `esc`
- **THEN** help overlay SHALL close
- **AND** normal list view SHALL resume

### Requirement: Refresh Shortcut
The TUI SHALL provide a shortcut to refresh the workspace list.

#### Scenario: Refresh list
- **WHEN** user presses `r` key
- **THEN** workspace list SHALL reload from disk
- **AND** loading spinner SHALL display during refresh
- **AND** current selection SHALL be preserved if still valid
