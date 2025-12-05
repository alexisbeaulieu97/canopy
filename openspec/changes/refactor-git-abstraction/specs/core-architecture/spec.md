# core-architecture Specification Delta

## MODIFIED Requirements

### Requirement: Pure go-git Implementation
The system SHALL use only go-git library for all git operations without shelling out to the git CLI.

#### Scenario: Clone repository with go-git
- **WHEN** a repository needs to be cloned
- **THEN** the operation is performed using go-git's Clone function
- **AND** no external git process is spawned

#### Scenario: Create worktree with go-git
- **WHEN** a worktree needs to be created
- **THEN** the operation is performed using go-git's Worktree API
- **AND** no external git process is spawned

#### Scenario: Fetch updates with go-git
- **WHEN** repository updates need to be fetched
- **THEN** the operation is performed using go-git's Fetch function
- **AND** no external git process is spawned

#### Scenario: Branch operations with go-git
- **WHEN** branch creation or checkout is needed
- **THEN** the operation is performed using go-git's Branch and Checkout APIs
- **AND** no external git process is spawned

### Requirement: Uniform Error Handling for Git Operations
All git operations SHALL return domain errors wrapped with context, without exposing go-git internals.

#### Scenario: Git operation returns domain error
- **WHEN** a git operation fails
- **THEN** the error is wrapped as an internal/errors type
- **AND** the original go-git error is available via errors.Unwrap()
