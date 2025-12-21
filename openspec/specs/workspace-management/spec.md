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

### Requirement: Workspace Metadata Versioning
Workspace metadata files SHALL include a version field to enable schema evolution and migration.

#### Scenario: New workspace includes version
- **WHEN** a new workspace is created, **THEN** the workspace.yaml SHALL include `version: 1` matching the current schema version

#### Scenario: Load workspace without version (legacy)
- **WHEN** loading a workspace.yaml without a version field, **THEN** the version SHALL default to 0, the workspace SHALL be treated as compatible, and an automatic migration SHALL be applied to upgrade to the current version (migration failure SHALL abort load and surface an error)

#### Scenario: Load workspace with unknown future version
- **WHEN** loading a workspace.yaml with version higher than current, **THEN** a warning SHALL be logged including the version, read operations SHALL return all fields including unknown/future fields, write operations SHALL only modify known fields and MUST reject attempts to modify unknown fields, the original workspace version and any unknown fields MUST be preserved unchanged when persisting writes, and known fields SHALL be validated normally

#### Scenario: Save workspace after migration
- **WHEN** saving a workspace that was migrated from an older version (0 or lower than current), **THEN** the workspace SHALL be saved with the current schema version

### Requirement: Export/Import Version Compatibility
Workspace export and import SHALL validate version compatibility.

#### Scenario: Export includes version
- **WHEN** exporting a workspace, **THEN** the export file SHALL include the workspace schema version in the `version` field

#### Scenario: Import validates version
- **WHEN** importing a workspace export file, **THEN** the version SHALL be validated against supported versions, imports from compatible versions SHALL succeed, and imports from incompatible versions SHALL fail with a clear error message

### Requirement: Workspace Rename
The system SHALL support renaming active workspaces.

#### Scenario: Rename workspace
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID`
- **THEN** workspace directory is renamed from `OLD-ID` to `NEW-ID`
- **AND** workspace metadata is updated with new ID
- **AND** success message is displayed

#### Scenario: Rename with branch rename
- **GIVEN** workspace has branch named `OLD-ID`
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID`
- **THEN** branch is also renamed to `NEW-ID` (default behavior)
- **AND** `--no-rename-branch` flag disables branch rename

#### Scenario: Rename to existing ID fails
- **GIVEN** workspace `NEW-ID` already exists
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID`
- **THEN** error is returned: "workspace 'NEW-ID' already exists"
- **AND** no changes are made

#### Scenario: Rename with force overwrites
- **GIVEN** workspace `NEW-ID` already exists
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID --force`
- **THEN** existing `NEW-ID` workspace is deleted
- **AND** `OLD-ID` is renamed to `NEW-ID`

#### Scenario: Rename closed workspace fails
- **GIVEN** workspace `OLD-ID` is closed
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID`
- **THEN** error is returned: "cannot rename closed workspace; reopen first with 'workspace open'"
- **AND** no changes are made

### Requirement: Workspace Templates
The system SHALL support user-defined workspace templates that specify default repositories and configuration for common workspace types.

#### Scenario: Define template in config
- **WHEN** user adds template to config.yaml with name, repos, and description
- **THEN** template is available for workspace creation
- **AND** template appears in `canopy template list` output

#### Scenario: Create workspace from template
- **WHEN** user runs `canopy workspace new PROJ-123 --template fullstack`
- **THEN** workspace is created with repositories defined in "fullstack" template
- **AND** default branch from template is used if no explicit branch specified
- **AND** workspace is ready for use

#### Scenario: Template with explicit repos
- **WHEN** user runs `canopy workspace new PROJ-123 --template backend --repos extra-lib`
- **THEN** workspace includes both template repos (backend, common) AND extra-lib
- **AND** all repos are cloned successfully

#### Scenario: Unknown template error
- **WHEN** user runs `canopy workspace new PROJ-123 --template nonexistent`
- **THEN** system returns error listing available templates
- **AND** no workspace is created

### Requirement: Template Configuration Format
Templates SHALL be defined in config.yaml with name, repos, optional default branch, and optional description.

#### Scenario: Parse template from config
- **WHEN** config.yaml contains templates section with valid YAML
- **THEN** templates are loaded into Config.Templates map
- **AND** each template is accessible by name

#### Scenario: Template with all fields
- **WHEN** template includes repos, default_branch, description, and setup_commands
- **THEN** all fields are parsed correctly
- **AND** template validation succeeds

#### Scenario: Invalid template configuration
- **WHEN** template in config.yaml has missing required fields (e.g., no repos)
- **THEN** config validation fails with clear error message
- **AND** indicates which template and field is problematic

### Requirement: Template Listing and Inspection
The system SHALL provide commands to list available templates and show template details.

#### Scenario: List all templates
- **WHEN** user runs `canopy template list`
- **THEN** system displays table of template names and descriptions
- **AND** shows repo count for each template
- **AND** templates are sorted alphabetically

#### Scenario: Show template details
- **WHEN** user runs `canopy template show fullstack`
- **THEN** system displays template name, description, repos list, and default branch
- **AND** indicates which repos are available in registry/canonical storage

### Requirement: Template Setup Commands
Templates SHALL support optional setup commands that execute after workspace creation to configure the environment.

#### Scenario: Execute template setup commands
- **WHEN** workspace is created from template with setup_commands defined
- **THEN** each command executes in workspace directory after repos are cloned
- **AND** commands run in order specified
- **AND** output is shown to user

#### Scenario: Setup command failure
- **WHEN** template setup command returns non-zero exit code
- **THEN** user is warned but workspace creation continues
- **AND** subsequent setup commands still execute
- **AND** workspace is marked as partially initialized

### Requirement: Template Composition
Templates SHALL be composable, allowing explicit repos to be added to template repos.

#### Scenario: Additive repository specification
- **WHEN** user specifies both --template and --repos flags
- **THEN** final repo list is union of template repos and explicit repos
- **AND** duplicates are removed (if same repo specified in both)
- **AND** all unique repos are included in workspace

#### Scenario: Template overrides branch
- **WHEN** template specifies default_branch but user provides --branch flag
- **THEN** user's explicit branch takes precedence
- **AND** template branch is ignored

### Requirement: Close Safety Checks
The system SHALL verify workspaces are safe to close before removing worktrees.

#### Scenario: Block close with uncommitted changes
- **GIVEN** workspace `PROJ-123` has a repo with uncommitted changes
- **WHEN** user runs `canopy workspace close PROJ-123`
- **THEN** system returns error indicating which repo has uncommitted changes
- **AND** no worktrees are removed

#### Scenario: Block close with unpushed commits
- **GIVEN** workspace `PROJ-123` has a repo with committed but unpushed changes
- **WHEN** user runs `canopy workspace close PROJ-123`
- **THEN** system returns error indicating which repo has unpushed commits
- **AND** no worktrees are removed

#### Scenario: Force close bypasses safety checks
- **GIVEN** workspace `PROJ-123` has repos with uncommitted or unpushed changes
- **WHEN** user runs `canopy workspace close PROJ-123 --force`
- **THEN** workspace is closed despite safety warnings
- **AND** worktrees are removed

#### Scenario: Preview shows safety status
- **GIVEN** workspace `PROJ-123` has repos with various states
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run`
- **THEN** preview shows which repos have uncommitted changes
- **AND** preview shows which repos have unpushed commits
- **AND** preview indicates whether close would be blocked

