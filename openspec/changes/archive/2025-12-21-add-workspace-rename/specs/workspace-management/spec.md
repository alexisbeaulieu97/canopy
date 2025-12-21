# workspace-management Specification Delta

## ADDED Requirements

### Requirement: Workspace Rename
The system SHALL support renaming active workspaces.

#### Scenario: Rename workspace
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID`
- **THEN** workspace directory is renamed from `OLD-ID` to `NEW-ID`
- **AND** workspace metadata is updated with new ID
- **AND** success message is displayed

#### Scenario: Rename with branch rename
- **GIVEN** workspace has branch named `OLD-ID`
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID`
- **THEN** branch is also renamed to `NEW-ID` (default behavior)
- **AND** `--no-rename-branch` flag disables branch rename

#### Scenario: Rename to existing ID fails
- **GIVEN** workspace `NEW-ID` already exists
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID`
- **THEN** error is returned: "workspace 'NEW-ID' already exists"
- **AND** no changes are made

#### Scenario: Rename with force overwrites
- **GIVEN** workspace `NEW-ID` already exists
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID --force`
- **THEN** existing `NEW-ID` workspace is deleted
- **AND** `OLD-ID` is renamed to `NEW-ID`

#### Scenario: Rename closed workspace fails
- **GIVEN** workspace `OLD-ID` is closed
- **WHEN** user runs `canopy workspace rename OLD-ID NEW-ID`
- **THEN** error is returned: "cannot rename closed workspace; reopen first with 'workspace open'"
- **AND** no changes are made
