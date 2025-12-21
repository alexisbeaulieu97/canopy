# cli Specification

## Purpose
Defines the command-line interface for Canopy, including workspace and repository commands, global flags, output formatting (JSON/text), error handling patterns, and exit code conventions. This spec ensures consistent CLI behavior across all commands for both interactive and scripted usage.
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
The `workspace list` command SHALL display all active workspaces. When `--status` flag is provided, the command SHALL also display git status for each repo.

#### Scenario: List active workspaces
- **GIVEN** active workspaces `PROJ-1` and `PROJ-2` exist
- **WHEN** I run `canopy workspace list`
- **THEN** the output SHALL include both `PROJ-1` and `PROJ-2`

#### Scenario: List with status flag
- **GIVEN** workspace `PROJ-1` exists with repos `repo-a` (dirty) and `repo-b` (2 commits ahead)
- **WHEN** I run `canopy workspace list --status`
- **THEN** output SHALL show `PROJ-1` with status indicators
- **AND** `repo-a` SHALL show dirty indicator
- **AND** `repo-b` SHALL show "2 ahead" indicator

#### Scenario: List with status and timeout
- **GIVEN** workspace with a slow/unresponsive repo
- **WHEN** I run `canopy workspace list --status --timeout 5s`
- **AND** a repo exceeds 5 seconds
- **THEN** that repo SHALL show "timeout" status
- **AND** other repos SHALL display normally

#### Scenario: List with status JSON output
- **WHEN** I run `canopy workspace list --status --json`
- **THEN** output SHALL be valid JSON following the standard envelope format
- **AND** each workspace in `data.workspaces` SHALL include `repos` array with status per repo

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

### Requirement: Error Code Documentation
All error codes SHALL be documented for scripting and automation purposes.

#### Scenario: Error code reference
- **GIVEN** a user wants to script canopy commands
- **WHEN** they need to handle specific errors
- **THEN** documentation SHALL list all error codes
- **AND** documentation SHALL map error codes to exit codes
- **AND** documentation SHALL provide handling examples

### Requirement: Command Help Text
All CLI commands SHALL have comprehensive help text accessible via `--help`.

#### Scenario: Help includes examples
- **GIVEN** a user runs `canopy workspace new --help`
- **THEN** the output SHALL include usage examples
- **AND** the output SHALL explain all flags

#### Scenario: Help includes error handling
- **GIVEN** a user runs `canopy --help`
- **THEN** the output SHALL mention where to find error code documentation

### Requirement: Configuration Path Override
The CLI SHALL support overriding the default configuration file path via flag or environment variable.

#### Scenario: Override config with flag
- **WHEN** user runs `canopy --config /path/to/config.yaml workspace list`
- **THEN** the configuration SHALL be loaded from `/path/to/config.yaml`
- **AND** the default config path SHALL be ignored

#### Scenario: Override config with environment variable
- **WHEN** `CANOPY_CONFIG=/path/to/config.yaml` is set
- **AND** user runs `canopy workspace list` without `--config` flag
- **THEN** the configuration SHALL be loaded from `/path/to/config.yaml`

#### Scenario: Flag takes precedence over environment variable
- **WHEN** `CANOPY_CONFIG=/env/config.yaml` is set
- **AND** user runs `canopy --config /flag/config.yaml workspace list`
- **THEN** the configuration SHALL be loaded from `/flag/config.yaml`
- **AND** the environment variable SHALL be ignored

#### Scenario: Default config when no override
- **WHEN** `--config` flag is not provided
- **AND** `CANOPY_CONFIG` environment variable is not set
- **THEN** the configuration SHALL be loaded from `~/.canopy/config.yaml`

#### Scenario: Config file not found error
- **WHEN** `--config /nonexistent/config.yaml` is specified
- **AND** the file does not exist
- **THEN** an error SHALL be returned indicating the config file was not found
- **AND** the error message SHALL include the attempted path

### Requirement: Doctor Command
The `canopy doctor` command SHALL validate the environment and configuration, reporting issues with actionable guidance.

#### Scenario: All checks pass
- **GIVEN** a properly configured Canopy environment
- **WHEN** I run `canopy doctor`
- **THEN** the output SHALL show all checks as passing
- **AND** the exit code SHALL be 0

#### Scenario: Git not installed
- **GIVEN** git is not installed or not in PATH
- **WHEN** I run `canopy doctor`
- **THEN** the output SHALL show an error for git availability
- **AND** the output SHALL suggest installing git
- **AND** the exit code SHALL be 2

#### Scenario: Invalid config file
- **GIVEN** `~/.canopy/config.yaml` contains invalid YAML
- **WHEN** I run `canopy doctor`
- **THEN** the output SHALL show an error for config validation
- **AND** the output SHALL include the parse error details
- **AND** the exit code SHALL be 2

#### Scenario: Missing directories with fix flag
- **GIVEN** `projects_root` directory does not exist
- **WHEN** I run `canopy doctor --fix`
- **THEN** the system SHALL create the missing directory
- **AND** the output SHALL report that it was auto-fixed
- **AND** the exit code SHALL be 0

#### Scenario: JSON output for scripting
- **GIVEN** canopy is configured and ready to run checks
- **WHEN** I run `canopy doctor --json`
- **THEN** the output SHALL be valid JSON
- **AND** the output SHALL include an array of check results with name, status, and message

#### Scenario: Warning for stale canonical repos
- **GIVEN** a canonical repo has not been fetched in over 30 days
- **WHEN** I run `canopy doctor`
- **THEN** the output SHALL show a warning for the stale repo
- **AND** the exit code SHALL be 1

### Requirement: Hook Dry-Run Mode
Commands that execute hooks SHALL support a `--dry-run-hooks` flag to preview hook execution without side effects.

#### Scenario: Dry-run shows resolved commands
- **GIVEN** a `post_create` hook configured as `echo "Created ${WORKSPACE_ID}"`
- **WHEN** I run `canopy workspace new PROJ-1 --dry-run-hooks`
- **THEN** the output SHALL show `echo "Created PROJ-1"` without executing it
- **AND** the workspace SHALL still be created (only hooks are dry-run)

#### Scenario: Dry-run with multiple repos
- **GIVEN** a hook configured with `repos: [api, worker]`
- **WHEN** I run `canopy workspace new PROJ-1 --repos api,worker,frontend --dry-run-hooks`
- **THEN** the output SHALL show the hook would run for `api` and `worker`
- **AND** the output SHALL indicate `frontend` is excluded from this hook

#### Scenario: Dry-run JSON output
- **WHEN** I run `canopy workspace new PROJ-1 --dry-run-hooks --json`
- **THEN** the JSON output SHALL include a `hooks_preview` array
- **AND** each entry SHALL contain `event`, `command`, `repos`, and `would_execute` fields

### Requirement: Hooks List Command
The `canopy hooks list` command SHALL display all configured hooks and their triggers.

#### Scenario: List configured hooks
- **GIVEN** hooks configured for `post_create` and `pre_close` events
- **WHEN** I run `canopy hooks list`
- **THEN** the output SHALL show each hook with its event, command, and repo filter

#### Scenario: No hooks configured
- **GIVEN** no hooks are configured
- **WHEN** I run `canopy hooks list`
- **THEN** the output SHALL indicate no hooks are configured
- **AND** the output SHALL suggest how to add hooks in config.yaml

### Requirement: Hooks Test Command
The `canopy hooks test <event> --workspace <id>` command SHALL dry-run hooks for a specific event against an existing workspace.

#### Scenario: Test post_create hooks
- **GIVEN** a `post_create` hook configured as `echo "Setup ${WORKSPACE_ID}"`
- **AND** workspace `PROJ-1` exists
- **WHEN** I run `canopy hooks test post_create --workspace PROJ-1`
- **THEN** the output SHALL show `echo "Setup PROJ-1"` with resolved variables
- **AND** no commands SHALL be executed
- **AND** the exit code SHALL be 0

#### Scenario: Test with invalid event
- **WHEN** I run `canopy hooks test invalid_event --workspace PROJ-1`
- **THEN** the output SHALL show an error indicating valid events are `post_create` and `pre_close`
- **AND** the exit code SHALL be non-zero

#### Scenario: Test with non-existent workspace
- **WHEN** I run `canopy hooks test post_create --workspace NONEXISTENT`
- **THEN** the output SHALL show a workspace not found error
- **AND** the exit code SHALL be non-zero

### Requirement: Parallel Workspace Status
The `workspace list --status` command SHALL fetch workspace status concurrently for improved performance.

#### Scenario: Parallel status fetching
- **GIVEN** 10 workspaces exist
- **WHEN** user runs `canopy workspace list --status`
- **THEN** status SHALL be fetched concurrently using worker pool
- **AND** output order SHALL be deterministic (sorted by workspace ID)
- **AND** worker count SHALL respect `parallel_workers` configuration

#### Scenario: Sequential status fallback
- **GIVEN** workspaces exist
- **WHEN** user runs `canopy workspace list --status --sequential-status`
- **THEN** status SHALL be fetched sequentially
- **AND** output SHALL match parallel mode output exactly

### Requirement: Strict Config Validation
The system SHALL validate configuration files strictly to catch typos and invalid values early.

#### Scenario: Unknown config field rejected
- **GIVEN** config file contains `parrallel_workers: 8` (typo)
- **WHEN** config is loaded
- **THEN** system returns error: "unknown config field 'parrallel_workers', did you mean 'parallel_workers'?"
- **AND** suggests similar known fields when possible

#### Scenario: Valid config accepted
- **GIVEN** config file contains only valid, known fields
- **WHEN** config is loaded
- **THEN** config is loaded successfully
- **AND** all values are applied as specified

#### Scenario: Hook timeout validation
- **GIVEN** config contains hook with `timeout: -5`
- **WHEN** config is loaded
- **THEN** system returns error: "hook timeout must be positive"

#### Scenario: Config validate command
- **WHEN** user runs `canopy config validate`
- **THEN** system loads and validates config
- **AND** reports any validation errors
- **AND** exits with code 0 if valid, non-zero if invalid

#### Scenario: Config validate with path
- **WHEN** user runs `canopy config validate --config /path/to/config.yaml`
- **THEN** system validates the specified config file
- **AND** does not use default config search paths

### Requirement: Version Command
The CLI SHALL provide a `version` command to display build information.

#### Scenario: Display version information
- **WHEN** user runs `canopy version`
- **THEN** the output SHALL include:
  - Version string (from git tag or "dev")
  - Git commit hash (short form)
  - Build date in ISO 8601 format
  - Go version used for compilation

#### Scenario: JSON version output
- **WHEN** user runs `canopy version --json`
- **THEN** the output SHALL be valid JSON
- **AND** SHALL include version, commit, buildDate, and goVersion fields

#### Scenario: Version flag on root command
- **WHEN** user runs `canopy --version`
- **THEN** the version string SHALL be printed
- **AND** the program SHALL exit with code 0

### Requirement: Version Embedding at Build Time
The version information SHALL be embedded at build time using Go linker flags.

#### Scenario: Tagged release build
- **GIVEN** the repository has a git tag `v1.2.3`
- **WHEN** the binary is built with ldflags
- **THEN** `canopy version` SHALL display `v1.2.3`

#### Scenario: Development build
- **GIVEN** no ldflags are provided during build
- **WHEN** `canopy version` is run
- **THEN** version SHALL display "dev"
- **AND** commit SHALL display "unknown"

### Requirement: Workspace Sync Command
The `workspace sync` command SHALL pull updates for all repos in a workspace and display a formatted summary.

#### Scenario: Sync workspace with updates available
- **GIVEN** workspace `PROJ-1` exists with repos `repo-a` and `repo-b`
- **AND** remote has new commits for both repos
- **WHEN** I run `canopy workspace sync PROJ-1`
- **THEN** the system SHALL pull updates for each repo
- **AND** output SHALL display a summary table with repo name, status, and commit count

#### Scenario: Sync workspace already up-to-date
- **GIVEN** workspace `PROJ-1` exists with repos at latest commits
- **WHEN** I run `canopy workspace sync PROJ-1`
- **THEN** output SHALL show each repo as "up-to-date"
- **AND** summary SHALL indicate "0 updated"

#### Scenario: Sync with repo error
- **GIVEN** workspace with a repo that has merge conflicts
- **WHEN** I run `canopy workspace sync PROJ-1`
- **THEN** the failed repo SHALL be marked with error status
- **AND** other repos SHALL still be synced
- **AND** summary SHALL show failure count
- **AND** exit code SHALL be non-zero

#### Scenario: Sync with timeout
- **GIVEN** a repo with slow/unresponsive remote
- **WHEN** I run `canopy workspace sync PROJ-1 --timeout 30s`
- **AND** a repo exceeds 30 seconds
- **THEN** that repo SHALL be marked as "timed out"
- **AND** other repos SHALL complete normally

#### Scenario: Sync with JSON output
- **WHEN** I run `canopy workspace sync PROJ-1 --json`
- **THEN** output SHALL be valid JSON following standard envelope
- **AND** `data.repos` SHALL contain per-repo sync results

### Requirement: CLI Output Consistency

The CLI SHALL use standardized output helpers for all user-facing messages to ensure consistent formatting across commands.

Output helpers SHALL provide the following message types:
- Success messages for completed actions
- Info messages for neutral information
- Warning messages for non-fatal issues
- Path-aware messages that include filesystem locations

#### Scenario: Success message format

- **WHEN** a CLI command completes successfully, **THEN** the output SHALL follow the pattern: "[Action] [target]" (e.g., "Created workspace foo")

#### Scenario: Success message with path

- **WHEN** a CLI command creates or modifies a filesystem resource, **THEN** the output SHALL include the path: "[Action] [target] in [path]"

#### Scenario: Consistent verb usage

- **WHEN** displaying success messages, **THEN** past tense verbs SHALL be used (Created, Closed, Removed, Renamed, Added)

