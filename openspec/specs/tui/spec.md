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

