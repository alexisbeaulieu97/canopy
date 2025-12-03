```markdown
## ADDED Requirements

### Requirement: Typed Errors
The system SHALL use typed errors for domain-specific error conditions.

#### Scenario: Workspace not found error
- **GIVEN** workspace `MISSING-123` does not exist
- **WHEN** any command references `MISSING-123`
- **THEN** the error SHALL be of type `ErrWorkspaceNotFound`
- **AND** the error message SHALL include the workspace ID

#### Scenario: Unclean workspace error
- **GIVEN** workspace `DIRTY-123` has uncommitted changes
- **WHEN** I run `canopy workspace close DIRTY-123`
- **THEN** the error SHALL be of type `ErrUncleanWorkspace`
- **AND** the error message SHALL list dirty repos

#### Scenario: Error codes in JSON
- **WHEN** a command fails with `--json` flag
- **THEN** the output SHALL include an `"error_code"` field
- **AND** the code SHALL map to the error type (e.g., `"WORKSPACE_NOT_FOUND"`)
```
