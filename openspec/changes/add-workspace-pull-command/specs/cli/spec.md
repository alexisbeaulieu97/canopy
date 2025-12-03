```markdown
## ADDED Requirements

### Requirement: Pull Workspace
The `workspace pull` command SHALL pull all repos within a workspace.

#### Scenario: Pull all repos
- **GIVEN** workspace `PROJ-123` exists with repos `backend` and `frontend`
- **WHEN** I run `canopy workspace pull PROJ-123`
- **THEN** the system SHALL run `git pull` in each repo
- **AND** display success/failure status per repo

#### Scenario: Pull with rebase
- **WHEN** I run `canopy workspace pull PROJ-123 --rebase`
- **THEN** the system SHALL run `git pull --rebase` in each repo

#### Scenario: Continue on error
- **GIVEN** `backend` repo has merge conflicts
- **WHEN** I run `canopy workspace pull PROJ-123 --continue-on-error`
- **THEN** the system SHALL continue pulling `frontend`
- **AND** report which repos failed
```
