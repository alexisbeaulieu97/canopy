## ADDED Requirements

### Requirement: Workspace Metadata Versioning
Workspace metadata files SHALL include a version field to enable schema evolution and migration.

#### Scenario: New workspace includes version
- **WHEN** a new workspace is created, **THEN** the workspace.yaml SHALL include `version: 1` matching the current schema version

#### Scenario: Load workspace without version (legacy)
- **WHEN** loading a workspace.yaml without a version field, **THEN** the version SHALL default to 0, the workspace SHALL be treated as compatible, and an automatic migration SHALL be applied to upgrade to the current version (migration failure SHALL abort load and surface an error)

#### Scenario: Load workspace with unknown future version
- **WHEN** loading a workspace.yaml with version higher than current, **THEN** a warning SHALL be logged including the version, read operations SHALL succeed, write operations SHALL preserve the original version and unknown fields (read-only for unknown fields), and known fields SHALL be validated as usual

#### Scenario: Save workspace after migration
- **WHEN** saving a workspace that was migrated from an older version (0 or lower than current), **THEN** the workspace SHALL be saved with the current schema version

### Requirement: Export/Import Version Compatibility
Workspace export and import SHALL validate version compatibility.

#### Scenario: Export includes version
- **WHEN** exporting a workspace, **THEN** the export file SHALL include the workspace schema version in the `version` field

#### Scenario: Import validates version
- **WHEN** importing a workspace export file, **THEN** the version SHALL be validated against supported versions, imports from compatible versions SHALL succeed, and imports from incompatible versions SHALL fail with a clear error message

