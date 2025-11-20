# TUI Interface Specification Deltas

## ADDED Requirements

### Requirement: Interactive Workspace Creation Wizard
The system SHALL provide an interactive wizard for creating workspaces with step-by-step prompts and validation.

#### Scenario: Launch wizard
- **WHEN** user runs `yard workspace create` or `yard workspace new --wizard`
- **THEN** interactive TUI wizard launches
- **AND** first step (workspace ID prompt) is displayed
- **AND** keyboard shortcuts help is shown at bottom

#### Scenario: Enter workspace ID with validation
- **WHEN** user types workspace ID in wizard
- **THEN** system validates ID in real-time
- **AND** shows error if workspace already exists
- **AND** shows error if ID contains invalid characters
- **AND** allows proceeding only with valid ID

#### Scenario: Select template (optional step)
- **WHEN** user reaches template selection step
- **THEN** list of available templates is shown
- **AND** user can select one template or skip
- **AND** selecting template pre-populates repository list

#### Scenario: Multi-select repositories
- **WHEN** user reaches repository selection step
- **THEN** list of available repos from registry and patterns is shown
- **AND** repos from template are pre-selected if template was chosen
- **AND** user can toggle repos with spacebar
- **AND** selected count is shown in header

#### Scenario: Confirm and create
- **WHEN** user reaches confirmation step
- **THEN** summary is shown with ID, repos, branch
- **AND** user can confirm or go back to edit
- **AND** confirming triggers workspace creation
- **AND** progress is shown during creation

#### Scenario: Cancel wizard
- **WHEN** user presses Esc or Ctrl+C during wizard
- **THEN** wizard exits without creating workspace
- **AND** confirmation prompt asks "Cancel? [y/N]"

### Requirement: Wizard Navigation
The wizard SHALL support intuitive keyboard navigation between steps with ability to go back and edit.

#### Scenario: Navigate forward through steps
- **WHEN** user completes current step and presses Enter
- **THEN** wizard advances to next step
- **AND** previous input is preserved

#### Scenario: Navigate backward through steps
- **WHEN** user presses designated back key (e.g., Left arrow or b)
- **THEN** wizard returns to previous step
- **AND** previous input is editable
- **AND** can navigate forward again

### Requirement: Wizard Visual Feedback
The wizard SHALL provide clear visual indicators for validation states and current progress.

#### Scenario: Invalid input indication
- **WHEN** user enters invalid data in any field
- **THEN** error message is shown in red
- **AND** field is highlighted
- **AND** user cannot proceed until fixed

#### Scenario: Valid input indication
- **WHEN** user enters valid data in any field
- **THEN** checkmark or green indicator is shown
- **AND** user can proceed to next step

#### Scenario: Progress indication
- **WHEN** user is on any wizard step
- **THEN** step number and total steps are shown (e.g., "Step 2 of 5")
- **AND** completed steps are visually distinct
