# Workspace Management Spec (Delta)

## ADDED Requirements

### Requirement: Rename Workspace Command
The tool SHALL provide a command to rename workspaces.

#### Scenario: Rename workspace successfully
- **WHEN** user runs `canopy workspace rename <old-id> <new-id>`
- **THEN** workspace directory SHALL be renamed atomically
- **AND** workspace.yaml SHALL be updated with new ID

#### Scenario: Old workspace does not exist
- **WHEN** user runs rename with non-existent old-id
- **THEN** command SHALL fail with error message
- **AND** no filesystem changes SHALL occur

### Requirement: Rename Validation
The tool SHALL validate the new workspace ID before renaming.

#### Scenario: Validate new ID format
- **WHEN** rename is requested
- **THEN** tool SHALL validate new ID contains no path traversal characters
- **AND** tool SHALL validate new ID uses valid characters only

#### Scenario: New ID conflicts with existing workspace
- **WHEN** new-id matches an existing workspace
- **THEN** command SHALL abort with conflict error
- **AND** original workspace SHALL remain unchanged

### Requirement: Atomic Directory Rename
The tool SHALL rename workspace directories atomically.

#### Scenario: Atomic rename operation
- **WHEN** rename is executed
- **THEN** directory rename SHALL be atomic (all-or-nothing)
- **AND** intermediate states SHALL not be visible

### Requirement: Optional Branch Rename
The tool SHALL optionally rename branches in all repos within the workspace.

#### Scenario: Rename branches with flag
- **WHEN** user provides `--rename-branches` flag
- **THEN** branches matching old workspace ID SHALL be renamed in all repos
- **AND** remote tracking branch updates SHALL be attempted

#### Scenario: Skip branch rename by default
- **WHEN** `--rename-branches` flag is not provided
- **THEN** repo branches SHALL remain unchanged

### Requirement: Confirmation and Force
The tool SHALL prompt for confirmation before renaming.

#### Scenario: Confirmation prompt
- **WHEN** rename is requested without `--force`
- **THEN** confirmation prompt SHALL appear
- **AND** user must confirm to proceed

#### Scenario: Skip confirmation with force
- **WHEN** `--force` flag is provided
- **THEN** rename SHALL proceed without confirmation prompt

### Requirement: Rollback on Failure
The tool SHALL rollback changes if rename fails partway through.

#### Scenario: Rollback on partial failure
- **WHEN** rename fails after partial completion
- **THEN** all changes SHALL be rolled back
- **AND** workspace SHALL be restored to original state
