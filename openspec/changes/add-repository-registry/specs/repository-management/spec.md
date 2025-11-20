# Repository Management Specification Deltas

## ADDED Requirements

### Requirement: Repository Registry
The system SHALL maintain a registry of repository aliases at `~/.yard/repos.yaml` that maps short names to full Git URLs with optional metadata.

#### Scenario: Register repository with alias
- **WHEN** user runs `yard repo register backend https://github.com/org/backend-api`
- **THEN** entry is created in repos.yaml with alias "backend" and given URL
- **AND** subsequent workspace creations can use "backend" as shorthand

#### Scenario: Registry file doesn't exist
- **WHEN** system attempts to load registry and file doesn't exist
- **THEN** empty registry is created in memory
- **AND** no error is reported
- **AND** file is created on first registration

#### Scenario: Use registry alias in workspace creation
- **WHEN** user runs `yard workspace new PROJ-123 --repos backend,frontend`
- **THEN** system resolves "backend" and "frontend" from registry
- **AND** workspace is created with corresponding URLs
- **AND** repos are cloned correctly

### Requirement: Repository Resolution Priority
The system SHALL resolve repository identifiers using a prioritized resolution chain: URL → registry → owner/repo format → error.

#### Scenario: Full URL takes precedence
- **WHEN** user provides `--repos https://github.com/org/repo`
- **THEN** system uses URL directly without checking registry
- **AND** workspace is created with given URL

#### Scenario: Registry alias resolution
- **WHEN** user provides `--repos backend` and "backend" exists in registry
- **THEN** system uses registered URL for backend
- **AND** workspace is created successfully

#### Scenario: Owner/repo format fallback
- **WHEN** user provides `--repos myorg/myrepo` and it's not in registry
- **THEN** system constructs URL as `https://github.com/myorg/myrepo`
- **AND** workspace is created with constructed URL

#### Scenario: Unknown repository error
- **WHEN** user provides `--repos unknown` and it's not in registry or URL format
- **THEN** system returns error suggesting `yard repo register unknown <url>`
- **AND** no workspace is created

### Requirement: Auto-Registration on Clone
The system SHALL automatically register repositories in the registry when cloned via `yard repo add`.

#### Scenario: Clone auto-registers with derived alias
- **WHEN** user runs `yard repo add https://github.com/org/backend-api`
- **THEN** repository is cloned to canonical storage
- **AND** entry is created in registry with alias "backend-api"
- **AND** user is notified of the generated alias

#### Scenario: Custom alias for clone
- **WHEN** user runs `yard repo add https://github.com/org/backend-api --alias be`
- **THEN** repository is cloned to canonical storage
- **AND** entry is created in registry with alias "be"
- **AND** user can use "be" in future workspace commands

#### Scenario: Skip auto-registration
- **WHEN** user runs `yard repo add <url> --no-register`
- **THEN** repository is cloned to canonical storage
- **AND** no registry entry is created
- **AND** user must use full URL in future

### Requirement: Registry Alias Uniqueness
The system SHALL enforce unique aliases within the registry and handle collisions gracefully.

#### Scenario: Duplicate alias error on manual registration
- **WHEN** user runs `yard repo register backend <url>` and "backend" already exists
- **THEN** system returns error indicating alias is taken
- **AND** shows existing URL for that alias
- **AND** suggests using --force flag to overwrite

#### Scenario: Auto-registration collision handling
- **WHEN** system auto-registers a repo and derived alias conflicts
- **THEN** system appends "-2" (or "-3", etc.) to make it unique
- **AND** notifies user of modified alias
- **AND** registration succeeds with unique name

### Requirement: Registry Management Commands
The system SHALL provide commands to list, register, unregister, and inspect registry entries.

#### Scenario: List all registry entries
- **WHEN** user runs `yard repo list-registry`
- **THEN** system displays table of aliases, URLs, and descriptions
- **AND** entries are sorted alphabetically by alias
- **AND** includes entry count in output

#### Scenario: Unregister alias
- **WHEN** user runs `yard repo unregister backend`
- **THEN** entry is removed from registry
- **AND** repos.yaml is updated
- **AND** confirmation message is shown
- **AND** canonical repo remains in storage (not deleted)

#### Scenario: Show registry entry details
- **WHEN** user runs `yard repo show backend`
- **THEN** system displays full details including URL, branch, description, and tags
- **AND** indicates if canonical repo exists in storage

### Requirement: Registry Metadata Support
The system SHALL support optional metadata for registry entries including default branch, description, and tags.

#### Scenario: Register with metadata
- **WHEN** user runs `yard repo register backend <url> --branch develop --description "Backend API"`
- **THEN** entry includes URL, default_branch, and description
- **AND** metadata is persisted to repos.yaml

#### Scenario: Register with tags
- **WHEN** user runs `yard repo register backend <url> --tags api,golang,backend`
- **THEN** entry includes tags list
- **AND** tags can be used for filtering in future

#### Scenario: Filter registry by tags
- **WHEN** user runs `yard repo list-registry --tags backend`
- **THEN** only entries with "backend" tag are shown
- **AND** count reflects filtered results
