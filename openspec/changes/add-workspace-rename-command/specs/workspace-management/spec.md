```markdown
## ADDED Requirements

### Requirement: Rename Workspace
The `workspace rename` command SHALL rename a workspace and update its metadata.

#### Scenario: Simple rename
- **GIVEN** workspace `OLD-123` exists
- **WHEN** I run `canopy workspace rename OLD-123 NEW-456`
- **THEN** the workspace directory SHALL be renamed to `NEW-456`
- **AND** workspace.yaml SHALL reflect ID `NEW-456`

#### Scenario: Rename with branch update
- **GIVEN** workspace `OLD-123` has branch `OLD-123` in all repos
- **WHEN** I run `canopy workspace rename OLD-123 NEW-456 --rename-branches`
- **THEN** branches in all repos SHALL be renamed to `NEW-456`

#### Scenario: Conflict detection
- **GIVEN** workspaces `WS-A` and `WS-B` both exist
- **WHEN** I run `canopy workspace rename WS-A WS-B`
- **THEN** the command SHALL fail with "workspace WS-B already exists"
```
