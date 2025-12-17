## ADDED Requirements

### Requirement: Implementation-Agnostic Storage Interface
The `WorkspaceStorage` interface SHALL be implementation-agnostic, using domain identifiers (workspace IDs) rather than implementation details (directory names, file paths).

#### Scenario: Create workspace by domain object
- **WHEN** `Create(ctx, workspace)` is called with a domain.Workspace
- **THEN** the storage SHALL persist the workspace
- **AND** the caller SHALL NOT need to specify directory names or paths

#### Scenario: Load workspace by ID
- **WHEN** `Load(ctx, id)` is called with a workspace ID
- **THEN** the storage SHALL return the workspace metadata
- **AND** the caller SHALL NOT need to know the underlying storage path

#### Scenario: Save workspace by domain object
- **WHEN** `Save(ctx, workspace)` is called with a domain.Workspace
- **THEN** the storage SHALL update the persisted workspace using the ID from the domain object
- **AND** the caller SHALL NOT need to provide directory names

#### Scenario: Close workspace by ID
- **WHEN** `Close(ctx, id, closedAt)` is called with a workspace ID
- **THEN** the storage SHALL archive the workspace
- **AND** the caller SHALL NOT need to provide directory names

#### Scenario: Rename workspace by IDs
- **WHEN** `Rename(ctx, oldID, newID)` is called
- **THEN** the storage SHALL update the workspace ID
- **AND** the implementation MAY rename underlying directories as needed

### Requirement: Context Support in Storage Interface
All `WorkspaceStorage` methods SHALL accept `context.Context` as their first parameter to enable cancellation and timeout for I/O operations.

#### Scenario: Storage method accepts context
- **WHEN** a service method calls a storage method
- **THEN** the service SHALL pass its context to the storage method
- **AND** the storage SHALL respect context cancellation

#### Scenario: Context cancellation stops I/O
- **WHEN** context is cancelled during a storage operation
- **THEN** the operation SHALL return promptly
- **AND** an appropriate error SHALL be returned
