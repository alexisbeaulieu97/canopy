# Workspace Management Specification Deltas

## ADDED Requirements

### Requirement: Workspace Archiving
The system SHALL support archiving workspaces to preserve metadata while removing active worktrees, and restoring them later.

#### Scenario: Archive active workspace
- **WHEN** user runs `canopy workspace archive PROJ-123`
- **THEN** workspace metadata is moved to ~/.canopy/archives/
- **AND** all worktrees are removed from workspaces_root
- **AND** canonical repositories remain untouched
- **AND** workspace no longer appears in active list

#### Scenario: List archived workspaces
- **WHEN** user runs `canopy workspace list --archived`
- **THEN** system displays list of archived workspaces with archive dates
- **AND** shows original repo list for each

#### Scenario: Restore archived workspace
- **WHEN** user runs `canopy workspace restore PROJ-123`
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

#### Scenario: Close with archive flags
- **WHEN** user runs `canopy workspace close PROJ-123 --archive`
- **THEN** the system archives the workspace instead of deleting
- **AND** `--no-archive` overrides prompts to delete directly

#### Scenario: Close default controlled by config
- **GIVEN** `workspace_close_default: archive`
- **WHEN** user closes a workspace without flags in non-interactive mode
- **THEN** the workspace is archived instead of deleted
