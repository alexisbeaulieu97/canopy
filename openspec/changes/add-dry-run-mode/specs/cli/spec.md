# cli Specification Delta

## ADDED Requirements

### Requirement: Dry Run Mode for Destructive Commands
Destructive commands SHALL support `--dry-run` flag to preview changes without executing them.

#### Scenario: Dry run workspace close
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run`
- **THEN** the system displays what would be deleted
- **AND** shows workspace directory, affected repos, and total size
- **AND** no filesystem changes occur

#### Scenario: Dry run repo remove
- **WHEN** user runs `canopy repo remove backend --dry-run`
- **THEN** the system displays what would be removed
- **AND** shows directory path, size, and workspaces that would become orphaned
- **AND** no filesystem changes occur

#### Scenario: Dry run with JSON output
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run --json`
- **THEN** the preview is output as JSON
- **AND** includes all affected paths and sizes

#### Scenario: Dry run with force flag
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run --force`
- **THEN** the system still only previews
- **AND** force flag does not bypass dry-run
