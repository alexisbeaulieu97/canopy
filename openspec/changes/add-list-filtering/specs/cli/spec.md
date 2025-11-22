# CLI Specification Deltas

## MODIFIED Requirements

### Requirement: List Workspaces
The `workspace list` command SHALL display all active workspaces with optional filtering by status.

#### Scenario: List active workspaces
- **GIVEN** active workspaces `PROJ-1` and `PROJ-2` exist
- **WHEN** I run `canopy workspace list`
- **THEN** the output SHALL include both `PROJ-1` and `PROJ-2`

#### Scenario: Filter with include flag
- **GIVEN** workspaces exist with varying statuses (dirty, clean, stale)
- **WHEN** I run `canopy workspace list --include=dirty,stale`
- **THEN** the output SHALL only include workspaces that are dirty OR stale
- **AND** clean workspaces SHALL be excluded

#### Scenario: Filter with exclude flag
- **GIVEN** workspaces exist with varying statuses
- **WHEN** I run `canopy workspace list --exclude=clean`
- **THEN** the output SHALL exclude all clean workspaces
- **AND** dirty, stale, and behind workspaces SHALL be included

#### Scenario: Combined include and exclude
- **GIVEN** workspaces exist with varying statuses
- **WHEN** I run `canopy workspace list --include=dirty,stale --exclude=archived`
- **THEN** the output SHALL include dirty or stale workspaces
- **AND** archived workspaces SHALL be excluded even if dirty or stale

#### Scenario: Config default filters
- **GIVEN** config has `workspace.list.include: ["dirty"]`
- **WHEN** I run `canopy workspace list` without flags
- **THEN** the output SHALL only include dirty workspaces

#### Scenario: CLI overrides config
- **GIVEN** config has `workspace.list.include: ["dirty"]`
- **WHEN** I run `canopy workspace list --include=stale`
- **THEN** the CLI flag SHALL override config
- **AND** only stale workspaces SHALL be shown

## ADDED Requirements

### Requirement: Filter Value Validation
The CLI SHALL validate filter values and reject unknown status types.

#### Scenario: Invalid filter value
- **WHEN** I run `canopy workspace list --include=invalid`
- **THEN** the command SHALL fail with an error
- **AND** the error message SHALL list valid filter values
