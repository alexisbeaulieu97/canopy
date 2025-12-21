## MODIFIED Requirements

### Requirement: Workspace Directory Naming
The system SHALL use the `workspace_naming` template to compute workspace directory names.

#### Scenario: Default naming template
- **WHEN** workspace_naming is "{{.ID}}" (default)
- **AND** user creates workspace "PROJ-123"
- **THEN** the directory SHALL be `workspaces_root/PROJ-123/`

#### Scenario: Custom naming template
- **WHEN** workspace_naming is "ws-{{.ID}}"
- **AND** user creates workspace "PROJ-123"
- **THEN** the directory SHALL be `workspaces_root/ws-PROJ-123/`

#### Scenario: Invalid template
- **WHEN** workspace_naming contains invalid template syntax
- **THEN** the system SHALL return CONFIG_VALIDATION error at startup

#### Scenario: Template produces invalid directory name
- **WHEN** template produces a name with path separators or invalid characters
- **THEN** the system SHALL return CONFIG_VALIDATION error

## ADDED Requirements

### Requirement: Template Preview
The system SHALL show computed workspace directory in config validation.

#### Scenario: Config validate shows naming preview
- **WHEN** user runs `canopy config validate`
- **THEN** the output SHALL include the fully resolved example directory path
- **AND** the path SHALL be computed by applying the configured `workspace_naming` template to "EXAMPLE-123"
- **AND** the output SHALL show the complete path under `workspaces_root` (for example: `workspaces_root/EXAMPLE-123/`)
