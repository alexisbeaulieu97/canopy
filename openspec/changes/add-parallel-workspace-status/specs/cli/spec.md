## ADDED Requirements
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
