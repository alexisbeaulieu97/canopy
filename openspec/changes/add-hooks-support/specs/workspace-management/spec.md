# workspace-management Specification Delta

## ADDED Requirements

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
