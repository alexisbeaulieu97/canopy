## ADDED Requirements

### Requirement: Close Safety Checks
The system SHALL verify workspaces are safe to close before removing worktrees.

#### Scenario: Block close with uncommitted changes
- **GIVEN** workspace `PROJ-123` has a repo with uncommitted changes
- **WHEN** user runs `canopy workspace close PROJ-123`
- **THEN** system returns error indicating which repo has uncommitted changes
- **AND** no worktrees are removed

#### Scenario: Block close with unpushed commits
- **GIVEN** workspace `PROJ-123` has a repo with committed but unpushed changes
- **WHEN** user runs `canopy workspace close PROJ-123`
- **THEN** system returns error indicating which repo has unpushed commits
- **AND** no worktrees are removed

#### Scenario: Force close bypasses safety checks
- **GIVEN** workspace `PROJ-123` has repos with uncommitted or unpushed changes
- **WHEN** user runs `canopy workspace close PROJ-123 --force`
- **THEN** workspace is closed despite safety warnings
- **AND** worktrees are removed

#### Scenario: Preview shows safety status
- **GIVEN** workspace `PROJ-123` has repos with various states
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run`
- **THEN** preview shows which repos have uncommitted changes
- **AND** preview shows which repos have unpushed commits
- **AND** preview indicates whether close would be blocked
