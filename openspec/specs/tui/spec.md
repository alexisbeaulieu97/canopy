# tui Specification

## Purpose
TBD - created by archiving change initialize-canopy. Update Purpose after archive.
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
