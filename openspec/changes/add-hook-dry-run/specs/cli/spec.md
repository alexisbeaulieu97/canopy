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
- **AND** indicate `frontend` is excluded from this hook

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
- **AND** suggest how to add hooks in config.yaml
