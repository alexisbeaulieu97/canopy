## ADDED Requirements

### Requirement: Typed Action Constants
TUI actions SHALL be represented by typed constants rather than string literals.

#### Scenario: Action type definition
- **WHEN** defining user actions that trigger confirmation dialogs
- **THEN** actions SHALL be defined as typed constants
- **AND** the type SHALL be `Action string` for readability
- **AND** constants SHALL be exported for use across TUI packages

#### Scenario: No magic strings in action handling
- **WHEN** checking which action was requested in a confirmation dialog
- **THEN** the comparison SHALL use typed constants (e.g., `action == ActionClose`)
- **AND** string literals SHALL NOT be used for action comparisons

