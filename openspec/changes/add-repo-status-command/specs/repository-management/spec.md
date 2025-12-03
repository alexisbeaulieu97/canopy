```markdown
## ADDED Requirements

### Requirement: Repo Status
The `repo status` command SHALL display status information for canonical repositories.

#### Scenario: List all repos with status
- **GIVEN** canonical repos `backend` and `frontend` exist in projects_root
- **WHEN** I run `canopy repo status`
- **THEN** the output SHALL show each repo with last fetch time and disk size

#### Scenario: Show single repo details
- **WHEN** I run `canopy repo status backend`
- **THEN** the output SHALL show detailed info for `backend`
- **AND** include remote branches
- **AND** include which workspaces use this repo

#### Scenario: Filter stale repos
- **GIVEN** `backend` was fetched 30 days ago
- **AND** `frontend` was fetched 2 days ago
- **WHEN** I run `canopy repo status --stale 7`
- **THEN** only `backend` SHALL be displayed
```
