## ADDED Requirements
### Requirement: CLI Layer Responsibilities
The CLI layer (cmd/canopy/) SHALL focus exclusively on user interface concerns:
- Parsing command-line flags and arguments
- Validating user input format (not business rules)
- Calling service layer methods
- Formatting output for display (text, JSON, table)
- Handling user prompts and confirmations

Business logic, orchestration, and domain operations SHALL remain in the service layer (internal/workspaces/).

#### Scenario: Clear separation of concerns
- **GIVEN** a workspace command is executed
- **WHEN** the command processes
- **THEN** flag parsing SHALL occur in CLI layer
- **AND** business validation SHALL occur in service layer
- **AND** domain operations SHALL occur in service layer
- **AND** output formatting SHALL occur in CLI layer

#### Scenario: Subcommand file organization
- **GIVEN** the CLI codebase
- **WHEN** reviewing file structure
- **THEN** each subcommand SHALL have its own file
- **AND** shared output helpers SHALL be in `presenters.go`
- **AND** parent command SHALL be in `workspace.go`
