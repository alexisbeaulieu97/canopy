# workspace-management Specification Delta

## ADDED Requirements

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
