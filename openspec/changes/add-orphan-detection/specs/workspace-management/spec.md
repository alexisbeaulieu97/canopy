# workspace-management Specification Delta

## ADDED Requirements

### Requirement: Orphan Detection
The system SHALL detect and report orphaned worktrees that reference missing canonical repos or directories.

#### Scenario: Detect missing canonical repo
- **GIVEN** workspace `PROJ-123` references repo `backend` in metadata
- **AND** canonical repo `backend` does not exist
- **WHEN** user runs `canopy check --orphans`
- **THEN** system reports orphan: "PROJ-123/backend: canonical repo 'backend' not found"

#### Scenario: Detect missing worktree directory
- **GIVEN** workspace metadata references worktree at `workspaces/PROJ-123/backend`
- **AND** directory does not exist
- **WHEN** user runs `canopy check --orphans`
- **THEN** system reports orphan: "PROJ-123/backend: worktree directory missing"

#### Scenario: No orphans found
- **GIVEN** all workspace worktrees are valid
- **WHEN** user runs `canopy check --orphans`
- **THEN** system reports "No orphaned worktrees found"

#### Scenario: TUI shows orphan warning
- **GIVEN** workspace has orphaned worktrees
- **WHEN** workspace is displayed in TUI
- **THEN** warning indicator is shown next to workspace name
- **AND** detail view shows orphan reasons
