# TUI Interface Specification Deltas

## ADDED Requirements

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
