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

### Requirement: Orphan Remediation
The system SHALL provide remediation suggestions and commands for detected orphans.

#### Scenario: Suggest remediation for missing canonical repo
- **GIVEN** orphan detected due to missing canonical repo
- **WHEN** orphan report is displayed
- **THEN** system suggests: "Run 'canopy repo add <url>' to restore, or 'canopy workspace remove-repo PROJ-123 backend' to remove reference"

#### Scenario: Suggest remediation for missing worktree directory
- **GIVEN** orphan detected due to missing worktree directory
- **WHEN** orphan report is displayed
- **THEN** system suggests: "Run 'canopy workspace repair PROJ-123' to recreate worktree, or 'canopy workspace remove-repo PROJ-123 backend' to remove reference"

#### Scenario: Auto-repair orphans
- **WHEN** user runs `canopy check --orphans --repair`
- **THEN** system attempts to recreate missing worktrees from canonical repos
- **AND** reports which orphans were fixed and which require manual intervention

### Requirement: Repo Removal Warnings
The system SHALL warn when removing canonical repos that are referenced by workspaces.

#### Scenario: Warn before removing in-use repo
- **GIVEN** canonical repo `backend` is used by workspaces `PROJ-1` and `PROJ-2`
- **WHEN** user runs `canopy repo remove backend`
- **THEN** system displays warning: "Repository 'backend' is used by 2 workspaces: PROJ-1, PROJ-2. These will become orphaned."
- **AND** prompts for confirmation

#### Scenario: Force remove in-use repo
- **GIVEN** canonical repo `backend` is used by workspaces
- **WHEN** user runs `canopy repo remove backend --force`
- **THEN** repo is removed without confirmation
- **AND** warning is still logged

#### Scenario: Remove unused repo
- **GIVEN** canonical repo `unused` is not referenced by any workspace
- **WHEN** user runs `canopy repo remove unused`
- **THEN** repo is removed without warning
