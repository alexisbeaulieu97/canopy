```markdown
## MODIFIED Requirements

### Requirement: TUI Code Organization
The TUI package SHALL be organized into focused modules for maintainability.

#### Scenario: Module separation
- **GIVEN** the TUI codebase is split into multiple files
- **WHEN** a developer works on the view layer
- **THEN** they SHALL find view logic in `view.go`
- **AND** update logic SHALL be in `update.go`
- **AND** message types SHALL be in `messages.go`
```
