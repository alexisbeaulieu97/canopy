## ADDED Requirements

### Requirement: Extensible Repository Resolution
The system SHALL use a Strategy pattern for repository resolution to enable extensibility.

#### Scenario: URL resolution strategy
- **WHEN** a repository identifier starts with a URL scheme (http://, https://, git@, ssh://, git://, file://)
- **THEN** the URL strategy SHALL handle resolution
- **AND** the repository name SHALL be derived from the URL path

#### Scenario: Registry resolution strategy
- **WHEN** a repository identifier matches a registered alias
- **THEN** the registry strategy SHALL return the registered URL
- **AND** the alias SHALL be used as the repository name

#### Scenario: GitHub shorthand resolution strategy
- **WHEN** a repository identifier contains exactly one slash (owner/repo format)
- **AND** neither segment is empty
- **THEN** the GitHub shorthand strategy SHALL construct a GitHub HTTPS URL
- **AND** the repo segment SHALL be used as the repository name

#### Scenario: Strategy chain execution
- **WHEN** resolving a repository identifier
- **THEN** strategies SHALL be tried in default order: URL → Registry → GitHub shorthand
- **AND** the first strategy that returns a successful match SHALL be used (first-match wins)
- **AND** if a strategy matches but encounters an error during resolution, the chain SHALL abort with that error
- **AND** if no strategy matches the input format, an `UnknownRepository` error SHALL be returned

#### Scenario: Strategy precedence override
- **WHEN** the resolver is configured with a custom strategy order
- **THEN** the custom order SHALL override the default precedence
- **AND** strategies not in the custom list SHALL be excluded from resolution

### Requirement: Shared Git URL Utilities
The system SHALL provide a shared package for Git URL parsing with the following operations:
- **Scheme detection**: Determine if a string is a valid Git URL
- **Repository name extraction**: Extract the repo name from a URL
- **Alias derivation**: Generate a default alias from a URL

#### Scenario: URL scheme detection
- **WHEN** checking if a string is a Git URL
- **THEN** the utility SHALL recognize: http://, https://, ssh://, git://, git@, file://
- **AND** the utility SHALL return false for plain strings without URL schemes

#### Scenario: Repository name extraction
- **WHEN** extracting a repository name from a URL
- **THEN** the utility SHALL handle SCP-style URLs (git@host:owner/repo.git)
- **AND** the utility SHALL handle standard URLs (https://host/owner/repo.git)
- **AND** the utility SHALL strip .git suffix if present
- **AND** the utility SHALL return the last non-empty path segment
- **AND** the utility SHALL return empty string for invalid/empty input

#### Scenario: Alias derivation
- **WHEN** deriving an alias from a URL
- **THEN** the utility SHALL extract the repository name
- **AND** the utility SHALL convert to lowercase
- **AND** the utility SHALL return a non-empty string suitable for use as a registry alias
