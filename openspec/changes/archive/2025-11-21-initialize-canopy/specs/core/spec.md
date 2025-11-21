# Spec: Core Engine

## ADDED Requirements

### Requirement: Manage Canonical Repos
The system SHALL maintain a directory of canonical git repositories (bare clones) to serve as the source for worktrees.

#### Scenario: Clone missing repo
- **GIVEN** a requested repo URL that is not in `projects_root`
- **WHEN** a workspace is created using that repo
- **THEN** the system SHALL clone it into `projects_root` first

### Requirement: Create Workspace Worktrees
The system SHALL create git worktrees for workspace branches from canonical repositories.

#### Scenario: Create worktree for workspace
- **GIVEN** a canonical repo `repo-a` exists in `projects_root`
- **WHEN** I create a workspace `PROJ-1` involving `repo-a`
- **THEN** the system SHALL create a worktree at `workspaces_root/PROJ-1/repo-a`
- **AND** the worktree SHALL be on branch `PROJ-1`

### Requirement: Safe Deletion
The system SHALL prevent accidental data loss when closing workspaces.

#### Scenario: Block deletion on dirty state
- **GIVEN** a workspace `PROJ-1` with uncommitted changes in `repo-a`
- **WHEN** I try to close the workspace without force flag
- **THEN** the operation SHALL fail with a warning about uncommitted changes

#### Scenario: Force deletion
- **GIVEN** a workspace `PROJ-1` with uncommitted changes
- **WHEN** I close the workspace with `--force` flag
- **THEN** the system SHALL delete the workspace and all worktrees
