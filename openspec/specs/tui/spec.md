# tui Specification

## Purpose
Interactive terminal UI for workspace management. Provides navigable list views, detail panels, keyboard shortcuts, status indicators, stale detection, disk usage display, behind-remote indicators, health status color coding, customizable keybindings, and reusable UI components.
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

### Requirement: Stale Workspace Detection
The TUI SHALL identify and visually indicate workspaces that haven't been modified recently based on configurable threshold.

#### Scenario: Display stale indicator
- **WHEN** workspace last modified date exceeds configured threshold
- **THEN** workspace is marked with stale indicator in list
- **AND** user can see at-a-glance which workspaces are inactive

### Requirement: Disk Usage Display
The TUI SHALL show disk space used by each workspace and total usage across all workspaces.

#### Scenario: Per-workspace disk usage
- **WHEN** TUI displays workspace list
- **THEN** each workspace shows disk usage in human-readable format
- **AND** usage is calculated from all worktrees in workspace

#### Scenario: Total disk usage summary
- **WHEN** TUI is open
- **THEN** header or footer shows total disk usage across all workspaces
- **AND** total workspace count is displayed

### Requirement: Behind-Remote Status
The TUI SHALL indicate when workspace repos are behind their remote branches.

#### Scenario: Show behind-remote indicator
- **WHEN** workspace has repos with commits available on remote
- **THEN** behind-remote badge is shown in list
- **AND** detail view shows commit count behind per repo

### Requirement: Health Status Indicators
The TUI SHALL use color coding to indicate workspace health status.

#### Scenario: Clean workspace indicator
- **WHEN** workspace has no uncommitted changes and no unpushed commits
- **THEN** workspace is shown in green
- **AND** user can quickly identify healthy workspaces

#### Scenario: Dirty workspace indicator
- **WHEN** workspace has uncommitted or unpushed changes
- **THEN** workspace is shown in red
- **AND** detail view shows which repos are dirty

#### Scenario: Needs attention indicator
- **WHEN** workspace is behind remote or stale
- **THEN** workspace is shown in yellow
- **AND** user can identify workspaces needing sync

### Requirement: Customizable Keyboard Bindings
The TUI SHALL support user-configurable keyboard bindings.

#### Scenario: Default keybindings work
- **GIVEN** no keybinding configuration exists
- **WHEN** user launches TUI
- **THEN** default keybindings are active (q=quit, j/k=navigate, etc.)

#### Scenario: Custom keybinding from config
- **GIVEN** config contains `tui.keybindings.delete: "d"`
- **WHEN** user presses "d" in TUI
- **THEN** delete action is triggered

#### Scenario: Override default keybinding
- **GIVEN** config contains `tui.keybindings.quit: "x"`
- **WHEN** user presses "x" in TUI
- **THEN** quit action is triggered
- **AND** "q" no longer triggers quit

#### Scenario: Keybinding conflict detection
- **GIVEN** config assigns same key to two actions
- **WHEN** config is validated
- **THEN** error is reported: "keybinding conflict: 'd' assigned to both 'delete' and 'details'"

#### Scenario: Invalid keybinding rejected
- **GIVEN** config contains invalid keybinding value
- **WHEN** config is validated
- **THEN** error is reported with the invalid value

### Requirement: Reusable TUI Components
The TUI SHALL provide reusable UI components that can be shared across views.

#### Scenario: StatusBadge renders workspace state
- **WHEN** rendering a workspace item
- **THEN** the StatusBadge component renders the appropriate state (dirty, clean, stale, error)
- **AND** styling is consistent across all views using the component

#### Scenario: ConfirmDialog handles user confirmation
- **WHEN** a destructive action requires confirmation
- **THEN** the ConfirmDialog component displays the prompt
- **AND** handles yes/no response with callbacks

#### Scenario: WorkspaceListItem renders workspace entry
- **WHEN** displaying a workspace in a list
- **THEN** the WorkspaceListItem component renders name, status, and metadata
- **AND** styling is consistent with the delegate pattern

