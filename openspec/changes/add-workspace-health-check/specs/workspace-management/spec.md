## ADDED Requirements

### Requirement: Workspace Health Checks
The system SHALL perform comprehensive health checks on workspaces to detect issues.

#### Scenario: Worktree integrity check
- **WHEN** health check is performed
- **THEN** each worktree's .git file MUST be validated
- **AND** worktree reference back to main repo MUST be verified

#### Scenario: Metadata consistency check
- **WHEN** health check is performed
- **THEN** workspace.yaml MUST be validated against schema
- **AND** repo entries MUST match actual worktrees on disk

#### Scenario: Git config check
- **WHEN** health check is performed
- **THEN** each repo's git config MUST be readable
- **AND** remote URLs MUST be valid format

#### Scenario: Health score calculation
- **WHEN** all checks complete
- **THEN** overall health score MUST be calculated
- **AND** scores MUST be: healthy, warning, or critical
