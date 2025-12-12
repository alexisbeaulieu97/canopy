## ADDED Requirements

### Requirement: State-Based View Management
The TUI SHALL use explicit state objects to manage view modes rather than boolean flags.

#### Scenario: List view state
- **WHEN** the TUI is initialized
- **THEN** the view state SHALL be `ListViewState`
- **AND** key events SHALL be handled by the list state handler

#### Scenario: Detail view state transition
- **WHEN** the user presses the details key on a workspace
- **THEN** the view state SHALL transition to `DetailViewState`
- **AND** the selected workspace SHALL be stored in the state

#### Scenario: Confirm state transition
- **WHEN** the user initiates a destructive action (close, push)
- **THEN** the view state SHALL transition to `ConfirmViewState`
- **AND** the action and target SHALL be stored in the state

#### Scenario: State exit
- **WHEN** the user presses cancel in a non-list state
- **THEN** the view state SHALL transition back to `ListViewState`

### Requirement: Workspace Data Encapsulation
The TUI SHALL encapsulate workspace data and caches in a dedicated component.

#### Scenario: Status cache access
- **WHEN** the TUI needs to display workspace status
- **THEN** it SHALL access status through the workspace model
- **AND** the cache lookup SHALL be encapsulated

#### Scenario: Filter application
- **WHEN** the user toggles the stale filter
- **THEN** the workspace model SHALL apply the filter
- **AND** the filtered items SHALL be returned to the view
