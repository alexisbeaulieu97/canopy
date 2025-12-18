## ADDED Requirements

### Requirement: Transactional Operation Integrity
Mutating operations SHALL maintain system consistency by cleaning up partial changes on failure.

#### Scenario: Create workspace fails during clone
- **WHEN** `CreateWorkspace` successfully creates the workspace directory
- **AND** clone operation fails for one repo
- **THEN** the system SHALL remove the workspace directory
- **AND** no metadata file SHALL remain
- **AND** the operation SHALL return an error

#### Scenario: Add repo fails during metadata update
- **WHEN** `AddRepoToWorkspace` successfully creates the worktree
- **AND** metadata update fails
- **THEN** the system SHALL remove the created worktree
- **AND** the workspace metadata SHALL remain unchanged
- **AND** the operation SHALL return an error

#### Scenario: Restore workspace fails during recreation
- **WHEN** `RestoreWorkspace` begins restoration
- **AND** worktree creation fails
- **THEN** the closed workspace entry SHALL remain intact
- **AND** no partial workspace directory SHALL remain
- **AND** the user can retry the restore operation

#### Scenario: Rollback actions logged
- **WHEN** an operation fails and rollback is triggered
- **THEN** the system SHALL log each rollback action at debug level
- **AND** include the original error and cleanup status
