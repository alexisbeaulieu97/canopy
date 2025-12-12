## ADDED Requirements

### Requirement: Direct Workspace Lookup
The workspace storage SHALL support direct lookup by workspace ID without listing all workspaces. The method signature SHALL be `LoadByID(id string) (*domain.Workspace, string, error)` returning the workspace metadata, directory name, and any error.

#### Scenario: Direct lookup by ID
- **WHEN** `LoadByID(id)` is called with a valid workspace ID
- **THEN** the storage SHALL attempt direct path access
- **AND** the method SHALL return `(workspace, dirName, nil)` where `workspace` is the metadata and `dirName` is the directory name

#### Scenario: Direct lookup fallback
- **WHEN** direct path access fails (ID differs from directory name)
- **THEN** the storage SHALL fall back to scanning all workspaces
- **AND** the method SHALL return `(workspace, dirName, nil)` if the workspace exists

#### Scenario: Workspace not found
- **WHEN** `LoadByID(id)` is called with a non-existent workspace ID
- **THEN** the method SHALL return `(nil, "", WorkspaceNotFound)` error

### Requirement: Workspace Metadata Caching
The service layer SHALL cache workspace metadata to reduce filesystem I/O.

#### Scenario: Cache hit
- **WHEN** looking up a workspace that was recently accessed
- **AND** the cache entry has not expired
- **THEN** the cached workspace SHALL be returned
- **AND** no filesystem I/O SHALL occur

#### Scenario: Cache miss
- **WHEN** looking up a workspace not in cache
- **THEN** the workspace SHALL be loaded from storage
- **AND** the result SHALL be added to the cache

#### Scenario: Cache invalidation on write
- **WHEN** a workspace is created, updated, or deleted
- **THEN** the cache entry for that workspace SHALL be invalidated
- **AND** subsequent lookups SHALL reload from storage
