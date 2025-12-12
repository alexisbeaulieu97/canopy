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
- **THEN** strategies SHALL be tried in configured order
- **AND** the first successful resolution SHALL be returned
- **AND** if no strategy matches, an error SHALL be returned

### Requirement: Shared Git URL Utilities
The system SHALL provide a shared package for Git URL parsing and manipulation.

#### Scenario: URL scheme detection
- **WHEN** checking if a string is a Git URL
- **THEN** the utility SHALL recognize: http://, https://, ssh://, git://, git@, file://

#### Scenario: Repository name extraction
- **WHEN** extracting a repository name from a URL
- **THEN** the utility SHALL handle SCP-style URLs (git@host:path)
- **AND** the utility SHALL strip .git suffix
- **AND** the utility SHALL return the last path segment
