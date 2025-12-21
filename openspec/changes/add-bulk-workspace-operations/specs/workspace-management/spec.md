## ADDED Requirements

### Requirement: Bulk Workspace Operations
The system SHALL support batch operations on multiple workspaces matching a pattern.

#### Scenario: Close workspaces by pattern
- **WHEN** user runs `canopy workspace close --pattern "^PROJ-"`
- **THEN** the system SHALL list all matching workspaces
- **AND** prompt for confirmation before proceeding
- **AND** close each matching workspace in sequence
- **AND** display a summary of successes and failures

#### Scenario: Sync workspaces by pattern
- **WHEN** user runs `canopy workspace sync --pattern "^FEATURE-"`
- **THEN** the system SHALL sync all matching workspaces in parallel
- **AND** display a summary table with results for each workspace

#### Scenario: Close all workspaces
- **WHEN** user runs `canopy workspace close --all`
- **THEN** the system SHALL treat this as `--pattern ".*"`
- **AND** require explicit confirmation

#### Scenario: Invalid pattern
- **WHEN** user provides an invalid regex pattern
- **THEN** the system SHALL return INVALID_ARGUMENT error
- **AND** display the regex parse error

### Requirement: Bulk Operation Safety
The system SHALL protect against accidental bulk operations on production workspaces.

#### Scenario: Bulk close requires confirmation
- **WHEN** user runs bulk close without --force
- **THEN** the system SHALL show affected workspaces
- **AND** require explicit y/n confirmation

#### Scenario: Force bulk close
- **WHEN** user runs `canopy workspace close --pattern "..." --force`
- **THEN** the system SHALL proceed without confirmation
- **AND** skip safety checks for dirty/unpushed repos
