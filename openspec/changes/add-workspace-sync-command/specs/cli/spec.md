```markdown
## ADDED Requirements

### Requirement: Sync Workspace
The `workspace sync` command SHALL fetch and pull all repos within a workspace.

#### Scenario: Sync all repos
- **GIVEN** workspace `PROJ-123` exists with repos `backend` and `frontend`
- **WHEN** I run `canopy workspace sync PROJ-123`
- **THEN** the system SHALL run `git fetch --all` in each repo
- **AND** run `git pull` in each repo
- **AND** display success/failure status per repo

#### Scenario: Fetch only
- **WHEN** I run `canopy workspace sync PROJ-123 --fetch-only`
- **THEN** the system SHALL only run `git fetch --all` without pulling

#### Scenario: Sync with rebase
- **WHEN** I run `canopy workspace sync PROJ-123 --rebase`
- **THEN** the system SHALL use `git pull --rebase` instead of merge
```
