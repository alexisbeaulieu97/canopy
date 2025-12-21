## MODIFIED Requirements

### Requirement: List Workspaces
The `workspace list` command SHALL display all active workspaces. When `--status` flag is provided, the command SHALL also display git status for each repo.

#### Scenario: List active workspaces
- **GIVEN** active workspaces `PROJ-1` and `PROJ-2` exist
- **WHEN** I run `canopy workspace list`
- **THEN** the output SHALL include both `PROJ-1` and `PROJ-2`

#### Scenario: List with status flag
- **GIVEN** workspace `PROJ-1` exists with repos `repo-a` (dirty) and `repo-b` (2 commits ahead)
- **WHEN** I run `canopy workspace list --status`
- **THEN** output SHALL show `PROJ-1` with status indicators
- **AND** `repo-a` SHALL show dirty indicator
- **AND** `repo-b` SHALL show "2 ahead" indicator

#### Scenario: List with status and timeout
- **GIVEN** workspace with a slow/unresponsive repo
- **WHEN** I run `canopy workspace list --status --timeout 5s`
- **AND** a repo exceeds 5 seconds
- **THEN** that repo SHALL show "timeout" status
- **AND** other repos SHALL display normally

#### Scenario: List with status JSON output
- **WHEN** I run `canopy workspace list --status --json`
- **THEN** output SHALL be valid JSON following the standard envelope format
- **AND** each workspace in `data.workspaces` SHALL include `repos` array with status per repo
