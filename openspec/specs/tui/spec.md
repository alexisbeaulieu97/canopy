# tui Specification

## Purpose
Interactive terminal UI for workspace management. Provides navigable list views, detail panels, keyboard shortcuts, and status indicators.
## Requirements
### Requirement: Interactive List
The TUI SHALL display a navigable list of workspaces.

#### Scenario: Navigate workspace list
- **GIVEN** a list of workspaces exists
- **WHEN** I press Down/Up arrows
- **THEN** the selection highlight SHALL move to the next/previous workspace

### Requirement: Detail View
The TUI SHALL show details for the selected workspace.

#### Scenario: View workspace details
- **GIVEN** I have selected workspace `PROJ-1` in the list
- **WHEN** I press Enter
- **THEN** the TUI SHALL display the list of repos and their git status for `PROJ-1`

### Requirement: Keyboard Shortcuts
The TUI SHALL expose keyboard shortcuts for common workspace actions.

#### Scenario: Push selected workspace
- **WHEN** I press `p`
- **THEN** the TUI asks for confirmation
- **AND** confirming pushes all repos for the selected workspace

#### Scenario: Open workspace in editor
- **WHEN** I press `o`
- **THEN** the selected workspace opens in `$VISUAL` or `$EDITOR`

#### Scenario: Filter workspaces
- **WHEN** I press `/`
- **THEN** search mode activates and filters the list by ID substring
- **WHEN** I press `s`
- **THEN** the list toggles to show only stale workspaces

#### Scenario: Close workspace
- **WHEN** I press `c`
- **THEN** the TUI asks for confirmation before closing the selected workspace

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

### Requirement: State-Based View Management
The TUI SHALL use explicit state objects to manage view modes rather than boolean flags.

#### Scenario: List view state
- **WHEN** the TUI is initialized
- **THEN** the view state SHALL be `ListViewState`
- **THEN** key events SHALL be handled by the list state handler

#### Scenario: Detail view state transition
- **WHEN** the user presses the details key on a workspace
- **THEN** the view state SHALL transition to `DetailViewState`
- **THEN** the selected workspace SHALL be stored in the state

#### Scenario: Confirm state transition
- **WHEN** the user initiates a destructive action (close, push)
- **THEN** the view state SHALL transition to `ConfirmViewState`
- **THEN** the action and target SHALL be stored in the state

#### Scenario: State exit
- **WHEN** the user presses cancel in a non-list state
- **THEN** the view state SHALL transition back to `ListViewState`

### Requirement: Workspace Data Encapsulation
The TUI SHALL encapsulate workspace data and caches in a dedicated component.

#### Scenario: Status cache access
- **WHEN** the TUI needs to display workspace status
- **THEN** it SHALL access status through the workspace model
- **THEN** the cache lookup SHALL be encapsulated

#### Scenario: Filter application
- **WHEN** the user toggles the stale filter
- **THEN** the workspace model SHALL apply the filter
- **THEN** the filtered items SHALL be returned to the view

