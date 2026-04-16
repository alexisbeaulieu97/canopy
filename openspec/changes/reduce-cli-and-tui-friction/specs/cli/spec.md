## MODIFIED Requirements

### Requirement: Initialize Config
The `init` command SHALL create the global configuration file and provide a concrete next-step checklist for first-time setup.

#### Scenario: First run
- **GIVEN** no config exists at `~/.canopy/config.yaml`
- **WHEN** I run `canopy init`
- **THEN** the system SHALL create `~/.canopy/config.yaml` with starter configuration
- **AND** the config SHALL use examples rather than implying live repositories
- **AND** the command SHALL print the next commands needed to add repositories and create a workspace

### Requirement: Create Workspace
The `workspace new` command SHALL create a new workspace with worktrees.

#### Scenario: New workspace with repos
- **WHEN** I run `canopy workspace new PROJ-123 --repos repo-a,repo-b`
- **THEN** the system SHALL create a workspace directory at `workspaces_root/PROJ-123`
- **AND** the workspace SHALL contain worktrees for `repo-a` and `repo-b`

#### Scenario: Creation fails when no repositories resolve
- **GIVEN** no repositories were provided explicitly
- **AND** no template repositories resolve
- **AND** no workspace-pattern defaults resolve
- **WHEN** I run `canopy workspace new PROJ-123`
- **THEN** the command SHALL fail with `NO_REPOS_CONFIGURED`
- **AND** the error SHALL explain how to recover with `--repos`, `--template`, repository registration, or `defaults.workspace_patterns`
- **AND** no workspace metadata or worktrees SHALL be created

### Requirement: List Workspaces
The `workspace list` command SHALL display all active workspaces. When `--status` flag is provided, the command SHALL also display git status for each repo.

#### Scenario: List active workspaces
- **GIVEN** active workspaces `PROJ-1` and `PROJ-2` exist
- **WHEN** I run `canopy workspace list`
- **THEN** the output SHALL include both `PROJ-1` and `PROJ-2`

#### Scenario: List with combined status summary
- **GIVEN** workspace `PROJ-1` has dirty, unpushed, and behind-remote repositories
- **WHEN** I run `canopy workspace list --status`
- **THEN** the workspace row SHALL display a compact multi-signal summary
- **AND** the row SHALL not collapse to only the first issue

#### Scenario: List empty workspace explicitly
- **GIVEN** workspace `PROJ-EMPTY` contains zero repositories
- **WHEN** I run `canopy workspace list --status`
- **THEN** the workspace row SHALL display `no repos`

#### Scenario: Status failures stay in-row
- **GIVEN** one workspace status check times out or errors
- **WHEN** I run `canopy workspace list --status`
- **THEN** the failure SHALL be represented in the workspace row or final summary
- **AND** the command SHALL not emit per-workspace warning lines between table rows

### Requirement: Command Help Text
All CLI commands SHALL have comprehensive help text accessible via `--help`.

#### Scenario: Runtime failure after parsing
- **GIVEN** command parsing succeeded
- **WHEN** a command fails at runtime due to configuration, workspace lookup, or domain validation
- **THEN** the CLI SHALL return the error without printing Cobra usage text

### Requirement: Status Command
The `status` command SHALL show the status of the current workspace.

#### Scenario: Run status outside a workspace
- **GIVEN** the current working directory is not inside `workspaces_root`
- **WHEN** I run `canopy status`
- **THEN** the command SHALL fail with `NOT_IN_WORKSPACE`
- **AND** the message SHALL include the current directory
- **AND** the message SHALL suggest a next step such as `canopy workspace list`

### Requirement: View Workspace Details
The `workspace view` command SHALL display detailed workspace metadata and repository state.

#### Scenario: View workspace details
- **GIVEN** workspace `PROJ-123` exists
- **WHEN** I run `canopy workspace view PROJ-123`
- **THEN** the output SHALL include the workspace path, branch, repo count, disk usage, and last modified timestamp
- **AND** the output SHALL include lock state when the workspace is locked
- **AND** the output SHALL include orphaned worktree details when present

#### Scenario: View empty workspace details
- **GIVEN** workspace `PROJ-EMPTY` contains zero repositories
- **WHEN** I run `canopy workspace view PROJ-EMPTY`
- **THEN** the output SHALL show an explicit empty state
- **AND** the empty state SHALL suggest `canopy workspace repo add PROJ-EMPTY <repo>`
