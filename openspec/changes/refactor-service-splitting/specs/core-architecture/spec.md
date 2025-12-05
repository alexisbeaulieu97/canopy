# core-architecture Specification Delta

## ADDED Requirements

### Requirement: Single Responsibility Service Components
The workspaces service layer SHALL be composed of focused sub-services, each with a single responsibility.

#### Scenario: RepoResolver handles identifier resolution
- **WHEN** a repo identifier is provided (name, alias, or URL)
- **THEN** the RepoResolver component resolves it to a canonical repo path
- **AND** the resolution logic is isolated from workspace operations

#### Scenario: DiskUsageCalculator handles size computation
- **WHEN** workspace disk usage is requested
- **THEN** the DiskUsageCalculator component computes and caches the result
- **AND** caching logic is isolated from workspace operations

#### Scenario: CanonicalRepoService handles repo management
- **WHEN** canonical repo operations are performed (list, add, remove, sync)
- **THEN** the CanonicalRepoService component handles the operation
- **AND** repo management is isolated from workspace lifecycle
