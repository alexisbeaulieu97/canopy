## ADDED Requirements

### Requirement: Workspace Metadata Versioning
Workspace metadata files SHALL include a version field to enable schema evolution and migration.

#### Scenario: New workspace includes version
- **WHEN** a new workspace is created
- **THEN** the workspace.yaml SHALL include `version: 1`
- **AND** the version SHALL match the current schema version

#### Scenario: Load workspace without version (legacy)
- **WHEN** loading a workspace.yaml without a version field
- **THEN** the version SHALL default to 0
- **AND** the workspace SHALL be treated as compatible
- **AND** a migration MAY be applied to upgrade to current version

#### Scenario: Load workspace with unknown future version
- **WHEN** loading a workspace.yaml with version higher than current
- **THEN** a warning SHALL be logged
- **AND** the workspace SHALL be loaded with best-effort compatibility
- **AND** write operations SHALL preserve the original version

#### Scenario: Save workspace updates version
- **WHEN** saving a workspace that was loaded with an older version
- **THEN** the workspace SHALL be saved with the current schema version
- **AND** any necessary migrations SHALL be applied

### Requirement: Export/Import Version Compatibility
Workspace export and import SHALL validate version compatibility.

#### Scenario: Export includes version
- **WHEN** exporting a workspace
- **THEN** the export file SHALL include the workspace schema version
- **AND** the version SHALL be preserved in the `version` field of the export

#### Scenario: Import validates version
- **WHEN** importing a workspace export file
- **THEN** the version SHALL be validated against supported versions
- **AND** imports from compatible versions SHALL succeed
- **AND** imports from incompatible versions SHALL fail with a clear error message

