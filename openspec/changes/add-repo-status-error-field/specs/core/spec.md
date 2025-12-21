## ADDED Requirements

### Requirement: Repository Status Representation
The system SHALL represent repository status with dedicated fields for status data and errors.

The `RepoStatus` type SHALL include:
- `Name` - Repository name
- `Branch` - Current branch name (empty string if error occurred)
- `IsDirty` - Whether repository has uncommitted changes
- `UnpushedCommits` - Count of commits not pushed to remote
- `BehindRemote` - Count of commits behind remote
- `Error` - Error message if status retrieval failed (empty if successful)

#### Scenario: Successful status retrieval
- **WHEN** repository status is retrieved successfully
- **THEN** the Error field SHALL be empty
- **AND** the Branch field SHALL contain the actual branch name

#### Scenario: Status retrieval timeout
- **WHEN** repository status retrieval times out
- **THEN** the Error field SHALL contain "timeout"
- **AND** the Branch field SHALL be empty

#### Scenario: Status retrieval failure
- **WHEN** repository status retrieval fails with an error
- **THEN** the Error field SHALL contain the error description
- **AND** the Branch field SHALL be empty
