## ADDED Requirements

### Requirement: Workspace Sync Command
The `workspace sync` command SHALL pull updates for all repos in a workspace and display a formatted summary.

#### Scenario: Sync workspace with updates available
- **GIVEN** workspace `PROJ-1` exists with repos `repo-a` and `repo-b`
- **AND** remote has new commits for both repos
- **WHEN** I run `canopy workspace sync PROJ-1`
- **THEN** the system SHALL pull updates for each repo
- **AND** output SHALL display a summary table with repo name, status, and commit count

#### Scenario: Sync workspace already up-to-date
- **GIVEN** workspace `PROJ-1` exists with repos at latest commits
- **WHEN** I run `canopy workspace sync PROJ-1`
- **THEN** output SHALL show each repo as "up-to-date"
- **AND** summary SHALL indicate "0 updated"

#### Scenario: Sync with repo error
- **GIVEN** workspace with a repo that has merge conflicts
- **WHEN** I run `canopy workspace sync PROJ-1`
- **THEN** the failed repo SHALL be marked with error status
- **AND** other repos SHALL still be synced
- **AND** summary SHALL show failure count
- **AND** exit code SHALL be non-zero

#### Scenario: Sync with timeout
- **GIVEN** a repo with slow/unresponsive remote
- **WHEN** I run `canopy workspace sync PROJ-1 --timeout 30s`
- **AND** a repo exceeds 30 seconds
- **THEN** that repo SHALL be marked as "timed out"
- **AND** other repos SHALL complete normally

#### Scenario: Sync with JSON output
- **WHEN** I run `canopy workspace sync PROJ-1 --json`
- **THEN** output SHALL be valid JSON following standard envelope
- **AND** `data.repos` SHALL contain per-repo sync results
