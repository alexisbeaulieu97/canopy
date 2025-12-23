## ADDED Requirements

### Requirement: TUI Multi-Select Mode
The TUI SHALL support selecting multiple workspaces for bulk operations.

#### Scenario: Toggle selection with Space
- **WHEN** user presses Space on a workspace
- **THEN** workspace selection state MUST toggle
- **AND** visual indicator MUST show selection status

#### Scenario: Select all workspaces
- **WHEN** user presses `a` key
- **THEN** all visible workspaces MUST be selected
- **AND** selection count MUST update

#### Scenario: Deselect all workspaces
- **WHEN** user presses `A` (Shift+a)
- **THEN** all workspaces MUST be deselected
- **AND** selection count MUST show 0

### Requirement: TUI Bulk Operations
The TUI SHALL allow performing actions on multiple selected workspaces.

#### Scenario: Bulk sync selected workspaces
- **WHEN** workspaces are selected
- **AND** user presses `s` (sync key)
- **THEN** confirmation dialog MUST appear
- **AND** upon confirmation, sync MUST run on all selected workspaces

#### Scenario: Bulk close selected workspaces
- **WHEN** workspaces are selected
- **AND** user presses `c` (close key)
- **THEN** confirmation dialog MUST appear with count
- **AND** upon confirmation, close MUST run on all selected workspaces

#### Scenario: Single operation fallback
- **WHEN** no workspaces are selected
- **AND** user presses an action key
- **THEN** action MUST apply to current (highlighted) workspace only
