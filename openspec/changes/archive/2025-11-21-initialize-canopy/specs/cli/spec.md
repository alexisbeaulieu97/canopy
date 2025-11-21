# Spec: CLI

## ADDED Requirements

### Requirement: Initialize Config
The `init` command SHALL create the global configuration file.

#### Scenario: First run
- **GIVEN** no config exists at `~/.canopy/config.yaml`
- **WHEN** I run `canopy init`
- **THEN** the system SHALL create `~/.canopy/config.yaml` with default paths

### Requirement: Create Workspace
The `workspace new` command SHALL create a new workspace with worktrees.

#### Scenario: New workspace with repos
- **WHEN** I run `canopy workspace new PROJ-123 --repos repo-a,repo-b`
- **THEN** the system SHALL create a workspace directory at `workspaces_root/PROJ-123`
- **AND** the workspace SHALL contain worktrees for `repo-a` and `repo-b`

### Requirement: List Workspaces
The `workspace list` command SHALL display all active workspaces.

#### Scenario: List active workspaces
- **GIVEN** active workspaces `PROJ-1` and `PROJ-2` exist
- **WHEN** I run `canopy workspace list`
- **THEN** the output SHALL include both `PROJ-1` and `PROJ-2`
