# Workspace Management Specification Deltas

## ADDED Requirements

### Requirement: Workspace Closing
The system SHALL support closing workspaces to preserve metadata while removing active worktrees, and reopening them later.

#### Scenario: Close active workspace
- **WHEN** user runs `canopy workspace close PROJ-123 --archive`
- **THEN** workspace metadata is moved to the configured `closed_root`
- **AND** all worktrees are removed from workspaces_root
- **AND** canonical repositories remain untouched
- **AND** workspace no longer appears in active list

#### Scenario: List closed workspaces
- **WHEN** user runs `canopy workspace list --closed`
- **THEN** system displays list of closed workspaces with close dates
- **AND** shows original repo list for each

#### Scenario: Reopen closed workspace
- **WHEN** user runs `canopy workspace reopen PROJ-123`
- **THEN** workspace directory is recreated in workspaces_root
- **AND** worktrees are recreated from canonical repos on the recorded branch
- **AND** workspace appears in active list again
- **AND** closed entry is removed (or marked as reopened)

#### Scenario: Close nonexistent workspace
- **WHEN** user attempts to close workspace that doesn't exist
- **THEN** system returns error "workspace not found"
- **AND** no changes are made

#### Scenario: Restore to existing workspace conflict
- **WHEN** user attempts to restore workspace ID that already exists actively
- **THEN** system returns error suggesting --force or different ID
- **AND** no existing workspace is modified

#### Scenario: Close with keep/delete flags
- **WHEN** user runs `canopy workspace close PROJ-123 --keep`
- **THEN** the system stores the workspace instead of deleting
- **AND** `--delete` overrides prompts to delete directly

#### Scenario: Close default controlled by config
- **GIVEN** `workspace_close_default: archive`
- **WHEN** user closes a workspace without flags in non-interactive mode
- **THEN** the workspace is stored instead of deleted
