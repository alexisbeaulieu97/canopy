# Workspace Management Specification Deltas

## ADDED Requirements

### Requirement: Workspace Archiving
The system SHALL support archiving workspaces to preserve metadata while removing active worktrees, and restoring them later.

#### Scenario: Archive active workspace
- **WHEN** user runs `yard workspace archive PROJ-123`
- **THEN** workspace metadata is moved to ~/.yard/archives/
- **AND** all worktrees are removed from workspaces_root
- **AND** canonical repositories remain untouched
- **AND** workspace no longer appears in active list

#### Scenario: List archived workspaces
- **WHEN** user runs `yard workspace list --archived`
- **THEN** system displays list of archived workspaces with archive dates
- **AND** shows original repo list for each

#### Scenario: Restore archived workspace
- **WHEN** user runs `yard workspace restore PROJ-123`
- **THEN** workspace directory is recreated in workspaces_root
- **AND** worktrees are recreated from canonical repos on archived branch
- **AND** workspace appears in active list again
- **AND** archive entry is removed (or marked as restored)

#### Scenario: Archive nonexistent workspace
- **WHEN** user attempts to archive workspace that doesn't exist
- **THEN** system returns error "workspace not found"
- **AND** no changes are made

#### Scenario: Restore to existing workspace conflict
- **WHEN** user attempts to restore workspace ID that already exists actively
- **THEN** system returns error suggesting --force or different ID
- **AND** no existing workspace is modified
