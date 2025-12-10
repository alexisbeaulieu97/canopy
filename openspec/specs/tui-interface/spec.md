# tui-interface Specification

## Purpose
Interactive TUI enhancements for workspace visibility and quick actions. Includes stale detection, disk usage display, behind-remote indicators, and keyboard shortcuts.
## Requirements
### Requirement: Stale Workspace Detection
The TUI SHALL identify and visually indicate workspaces that haven't been modified recently based on configurable threshold.

#### Scenario: Display stale indicator
- **WHEN** workspace last modified date exceeds configured threshold
- **THEN** workspace is marked with stale indicator in list
- **AND** user can see at-a-glance which workspaces are inactive

#### Scenario: Filter stale workspaces
- **WHEN** user presses 's' key in TUI
- **THEN** list filters to show only stale workspaces
- **AND** pressing 's' again toggles back to all workspaces

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

### Requirement: Quick Actions
The TUI SHALL provide keyboard shortcuts for common workspace operations.

#### Scenario: Push all repos in workspace
- **WHEN** user selects workspace and presses 'p' key
- **THEN** confirmation prompt appears
- **AND** confirming pushes all repos to remote
- **AND** progress/results are shown

#### Scenario: Open workspace in editor
- **WHEN** user selects workspace and presses 'o' key
- **THEN** workspace directory opens in $EDITOR
- **AND** TUI remains active or exits based on editor type

### Requirement: Workspace Search and Filtering
The TUI SHALL support searching workspaces by ID and filtering by status.

#### Scenario: Search workspaces
- **WHEN** user presses '/' key
- **THEN** search input appears at bottom
- **AND** list filters in real-time as user types
- **AND** pressing Enter accepts filter, Esc cancels

#### Scenario: Clear search filter
- **WHEN** search is active and user presses Esc
- **THEN** filter is cleared
- **AND** full workspace list is restored

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

