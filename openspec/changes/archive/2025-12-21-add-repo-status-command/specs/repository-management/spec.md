# repository-management Specification Delta

## ADDED Requirements

### Requirement: Repository Status Information
The system SHALL provide detailed status information for canonical repositories.

#### Scenario: View single repo status
- **WHEN** user runs `canopy repo status backend`
- **THEN** system displays repo name, path, disk usage, last fetch time
- **AND** shows count and list of workspaces using this repo

#### Scenario: View all repo statuses
- **WHEN** user runs `canopy repo status` (no argument)
- **THEN** system displays status for all canonical repos
- **AND** output is formatted as a table

#### Scenario: Repo status JSON output
- **WHEN** user runs `canopy repo status --json`
- **THEN** output is valid JSON
- **AND** includes all status fields

#### Scenario: Last fetch time from FETCH_HEAD
- **WHEN** repo status is queried
- **THEN** last fetch time is determined from `.git/FETCH_HEAD` mtime
- **AND** displays "never fetched" if file doesn't exist
