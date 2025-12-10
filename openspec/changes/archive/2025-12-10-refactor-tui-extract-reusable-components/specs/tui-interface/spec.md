# tui-interface Specification Delta

## ADDED Requirements

### Requirement: Reusable TUI Components
The TUI SHALL provide reusable UI components that can be shared across views.

#### Scenario: StatusBadge renders workspace state
- **WHEN** rendering a workspace item
- **THEN** the StatusBadge component renders the appropriate state (dirty, clean, stale, error)
- **AND** styling is consistent across all views using the component

#### Scenario: ConfirmDialog handles user confirmation
- **WHEN** a destructive action requires confirmation
- **THEN** the ConfirmDialog component displays the prompt
- **AND** handles yes/no response with callbacks

#### Scenario: WorkspaceListItem renders workspace entry
- **WHEN** displaying a workspace in a list
- **THEN** the WorkspaceListItem component renders name, status, and metadata
- **AND** styling is consistent with the delegate pattern
