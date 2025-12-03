# cli Specification

## Purpose
TBD - created by archiving change initialize-canopy. Update Purpose after archive.
## Requirements
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

### Requirement: Workspace Git Command
The `workspace git` command SHALL execute arbitrary git commands across all repos in a workspace. The command SHALL support the following options:
- `--parallel`: Execute git commands concurrently across all repos (default: sequential)
- `--continue-on-error`: Continue execution in remaining repos if one fails (default: stop on first error)

#### Scenario: Run git fetch in all repos
- **GIVEN** workspace `PROJ-1` exists with repos `repo-a` and `repo-b`
- **WHEN** I run `canopy workspace git PROJ-1 fetch --all`
- **THEN** `git fetch --all` SHALL be executed in `repo-a`
- **AND** `git fetch --all` SHALL be executed in `repo-b`
- **AND** output from each repo SHALL be displayed with repo name header

#### Scenario: Run git status in all repos
- **GIVEN** workspace `PROJ-1` exists with multiple repos
- **WHEN** I run `canopy workspace git PROJ-1 status`
- **THEN** `git status` output SHALL be shown for each repo
- **AND** repos SHALL be clearly separated in output

#### Scenario: Sequential execution (default)
- **GIVEN** workspace with repos `repo-a`, `repo-b`, `repo-c`
- **WHEN** I run `canopy workspace git PROJ-1 pull`
- **THEN** git pull SHALL execute in `repo-a` first
- **AND** after `repo-a` completes, `repo-b` SHALL execute
- **AND** after `repo-b` completes, `repo-c` SHALL execute

#### Scenario: Parallel execution
- **GIVEN** workspace with multiple repos
- **WHEN** I run `canopy workspace git PROJ-1 fetch --parallel`
- **THEN** git fetch SHALL execute concurrently in all repos
- **AND** output SHALL be collected and displayed per repo

#### Scenario: Stop on first error
- **GIVEN** workspace with repos where one has merge conflicts
- **WHEN** I run `canopy workspace git PROJ-1 pull`
- **AND** the first repo fails
- **THEN** execution SHALL stop immediately
- **AND** exit code SHALL be non-zero

#### Scenario: Continue on error
- **GIVEN** workspace with repos where one will fail
- **WHEN** I run `canopy workspace git PROJ-1 pull --continue-on-error`
- **AND** one repo fails
- **THEN** execution SHALL continue to remaining repos
- **AND** summary SHALL show which repos failed
- **AND** exit code SHALL be non-zero

