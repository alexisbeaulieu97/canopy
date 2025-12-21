## ADDED Requirements

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
