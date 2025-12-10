# workspace-management Specification

## Purpose
Define workspace closing behavior (store metadata, remove worktrees) and reopening for restoration.
## Requirements
### Requirement: Workspace Closing
The system SHALL support closing workspaces to preserve metadata while removing active worktrees, and reopening them later.

#### Scenario: Close active workspace
- **WHEN** user runs `canopy workspace close PROJ-123 --keep`
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

#### Scenario: Close with keep/delete flags
- **WHEN** user runs `canopy workspace close PROJ-123 --keep`
- **THEN** the system stores the workspace metadata instead of deleting
- **AND** `--delete` overrides prompts to delete directly

#### Scenario: Close default controlled by config
- **GIVEN** `workspace_close_default: archive`
- **WHEN** user closes a workspace without flags in non-interactive mode
- **THEN** the workspace is stored instead of deleted

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

### Requirement: Workspace Export
The system SHALL support exporting workspace definitions to portable files.

#### Scenario: Export workspace to YAML
- **WHEN** user runs `canopy workspace export PROJ-123`
- **THEN** workspace definition is output as YAML to stdout
- **AND** includes workspace ID, branch, repo names, and URLs

#### Scenario: Export workspace to file
- **WHEN** user runs `canopy workspace export PROJ-123 --output ws.yaml`
- **THEN** workspace definition is written to `ws.yaml`

#### Scenario: Export workspace as JSON
- **WHEN** user runs `canopy workspace export PROJ-123 --format json`
- **THEN** workspace definition is output as JSON

### Requirement: Workspace Import
The system SHALL support importing workspace definitions from files.

#### Scenario: Import workspace from file
- **WHEN** user runs `canopy workspace import ws.yaml`
- **THEN** workspace is created from the definition
- **AND** missing canonical repos are cloned
- **AND** worktrees are created for each repo

#### Scenario: Import with ID override
- **WHEN** user runs `canopy workspace import ws.yaml --id NEW-ID`
- **THEN** workspace is created with ID `NEW-ID`
- **AND** original ID in file is ignored

#### Scenario: Import conflict detection
- **GIVEN** workspace `PROJ-123` already exists
- **WHEN** user runs `canopy workspace import ws.yaml` (containing PROJ-123)
- **THEN** error is returned: "workspace 'PROJ-123' already exists"
- **AND** no changes are made

#### Scenario: Import with force overwrites
- **GIVEN** workspace `PROJ-123` already exists
- **WHEN** user runs `canopy workspace import ws.yaml --force`
- **THEN** existing workspace is replaced with imported definition

### Requirement: Lifecycle Hooks
The system SHALL support configurable hooks that run during workspace lifecycle events.

#### Scenario: Post-create hook execution
- **GIVEN** config contains `hooks.post_create` with commands
- **WHEN** user creates a workspace with `canopy workspace new PROJ-123`
- **THEN** post_create hooks execute after worktrees are created
- **AND** hooks run in workspace directory
- **AND** hook output is displayed to user

#### Scenario: Pre-close hook execution
- **GIVEN** config contains `hooks.pre_close` with commands
- **WHEN** user closes a workspace with `canopy workspace close PROJ-123`
- **THEN** pre_close hooks execute before worktrees are removed
- **AND** hooks run in workspace directory

#### Scenario: Hook filtered by repo
- **GIVEN** config contains hook with `repos: [backend]`
- **WHEN** the hook would execute
- **THEN** it only runs for workspaces containing the `backend` repo

#### Scenario: Skip hooks with flag
- **WHEN** user runs `canopy workspace new PROJ-123 --no-hooks`
- **THEN** no hooks are executed
- **AND** workspace is created normally

#### Scenario: Hook failure handling
- **GIVEN** a hook command returns non-zero exit code
- **WHEN** the hook executes
- **THEN** error is logged with hook output
- **AND** subsequent hooks continue (fail-soft by default)

#### Scenario: Run hooks without lifecycle action
- **WHEN** user runs `canopy workspace new PROJ-123 --hooks-only`
- **THEN** post_create hooks execute for the existing workspace
- **AND** no workspace creation or worktree changes occur

#### Scenario: Hooks-only close
- **WHEN** user runs `canopy workspace close PROJ-123 --hooks-only`
- **THEN** pre_close hooks execute for the workspace
- **AND** the workspace remains open

#### Scenario: Reject invalid hook commands
- **GIVEN** a hook command is empty, whitespace-only, or contains newlines
- **WHEN** configuration is validated
- **THEN** validation fails with an error referencing the offending hook

