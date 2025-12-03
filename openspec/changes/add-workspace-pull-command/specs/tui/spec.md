```markdown
## ADDED Requirements

### Requirement: Pull Shortcut
The TUI SHALL provide a keyboard shortcut to pull the selected workspace.

#### Scenario: Pull selected workspace
- **GIVEN** I have selected workspace `PROJ-123` in the list
- **WHEN** I press `l`
- **THEN** the TUI asks for confirmation
- **AND** confirming pulls all repos for the selected workspace
- **AND** a spinner displays during the operation
```
