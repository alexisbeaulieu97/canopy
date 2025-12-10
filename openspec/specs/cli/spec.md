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

### Requirement: Dry Run Mode for Destructive Commands
Destructive commands SHALL support `--dry-run` flag to preview changes without executing them.

#### Scenario: Dry run workspace close
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run`
- **THEN** the system displays what would be deleted
- **AND** shows workspace directory, affected repos, and total size
- **AND** no filesystem changes occur

#### Scenario: Dry run repo remove
- **WHEN** user runs `canopy repo remove backend --dry-run`
- **THEN** the system displays what would be removed
- **AND** shows directory path, size, and workspaces that would become orphaned
- **AND** no filesystem changes occur

#### Scenario: Dry run with JSON output
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run --json`
- **THEN** the preview is output as JSON following the standard envelope schema
- **AND** preview data is nested under `data.preview`
- **AND** includes all affected paths and sizes

#### Scenario: Dry run with force flag
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run --force`
- **THEN** the system still only previews
- **AND** force flag does not bypass dry-run

### Requirement: Global JSON Output Flag
All CLI commands SHALL support a `--json` flag for machine-parseable output.

#### Scenario: Enable JSON output
- **WHEN** user runs any command with `--json` flag
- **THEN** output SHALL be valid JSON
- **AND** output SHALL use consistent structure across all commands

#### Scenario: JSON output for scripting
- **WHEN** `--json` flag is provided
- **THEN** output SHALL be parseable by tools like `jq`
- **AND** no human-readable decorations SHALL be included

### Requirement: JSON Error Handling
Errors SHALL be output as structured JSON when `--json` flag is set.

#### Scenario: Error in JSON mode
- **WHEN** command fails with `--json` flag
- **THEN** output SHALL include `"success": false`
- **AND** `"error"` object SHALL contain `code`, `message`, and `context` fields

#### Scenario: Success in JSON mode
- **WHEN** command succeeds with `--json` flag
- **THEN** output SHALL include `"success": true`
- **AND** `"data"` field SHALL contain command-specific results

### Requirement: Consistent JSON Structure
All JSON output SHALL follow a standard envelope format.

#### Scenario: Standard envelope format
- **WHEN** any command outputs JSON
- **THEN** response SHALL have top-level `success`, `data`, and `error` fields
- **AND** `data` SHALL be null on error, `error` SHALL be null on success

### Requirement: Typed Error Returns
CLI commands SHALL return typed errors from `internal/errors` rather than `fmt.Errorf` strings.

#### Scenario: Command returns typed error
- **WHEN** a CLI command encounters an error condition
- **THEN** it returns a typed error (e.g., `ErrWorkspaceNotFound`, `ErrInvalidArgument`)
- **AND** the error can be inspected with `errors.Is()` or `errors.As()`

#### Scenario: Error includes context
- **WHEN** a typed error is returned
- **THEN** it includes contextual information (workspace ID, repo name, etc.)
- **AND** the error message is user-friendly

### Requirement: Exit Code Mapping
The CLI SHALL return consistent exit codes mapped from error types.

#### Scenario: Normal success returns 0
- **WHEN** a command completes successfully
- **THEN** the exit code is 0

#### Scenario: Typed errors map to exit codes
- **WHEN** a command returns a typed error
- **THEN** the exit code is determined by the error type
- **AND** `ErrWorkspaceNotFound` returns exit code 1
- **AND** `ErrInvalidArgument` returns exit code 2
- **AND** `ErrConfigInvalid` returns exit code 3

