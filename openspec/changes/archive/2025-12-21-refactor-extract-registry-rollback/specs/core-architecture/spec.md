## ADDED Requirements

### Requirement: Registry Transaction Safety

Registry modifications SHALL use atomic save-with-rollback semantics to prevent partial state on failure.

The save-with-rollback pattern SHALL:
- Attempt to persist registry changes
- On failure, execute the provided rollback function
- Log rollback failures without masking the original error
- Return the original save error to the caller

#### Scenario: Successful save

- **WHEN** a registry modification is saved successfully
- **THEN** no rollback SHALL be attempted
- **AND** the function SHALL return nil

#### Scenario: Save failure with successful rollback

- **WHEN** a registry save fails
- **THEN** the rollback function SHALL be executed
- **AND** the original save error SHALL be returned

#### Scenario: Save failure with rollback failure

- **WHEN** a registry save fails and rollback also fails
- **THEN** the rollback failure SHALL be logged
- **AND** the original save error SHALL be returned (not masked)
