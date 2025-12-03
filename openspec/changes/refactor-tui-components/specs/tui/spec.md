## MODIFIED Requirements

### Requirement: TUI Code Organization
The TUI package SHALL be organized into focused modules for maintainability.

#### Scenario: Module separation
- **WHEN** a developer works on the view layer, **THEN** they SHALL find view logic in `view.go`
- **WHEN** a developer works on state updates, **THEN** they SHALL find update logic in `update.go`
- **WHEN** a developer works on message types, **THEN** they SHALL find them in `messages.go`
