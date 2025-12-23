## ADDED Requirements

### Requirement: Workspace Health Check Command
The CLI SHALL provide a `doctor workspace` subcommand for comprehensive workspace health analysis.

#### Scenario: Check all workspaces
- **WHEN** user runs `canopy doctor workspace`
- **THEN** all active workspaces MUST be checked
- **AND** health status MUST be reported for each

#### Scenario: Check specific workspace
- **WHEN** user runs `canopy doctor workspace PROJ-123`
- **THEN** only that workspace MUST be checked
- **AND** detailed health report MUST be shown

#### Scenario: Auto-fix issues
- **WHEN** user runs `canopy doctor workspace --fix`
- **AND** fixable issues are found
- **THEN** fixes MUST be applied
- **AND** fix actions MUST be logged

#### Scenario: JSON output
- **WHEN** user runs `canopy doctor workspace --json`
- **THEN** health check results MUST be JSON formatted
- **AND** include check name, status, description, and fixable flag
