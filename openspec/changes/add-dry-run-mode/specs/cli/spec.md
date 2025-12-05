# cli Specification Delta

## Dependencies

> **Note**: This change has runtime dependencies on other changes:
> - **add-orphan-detection**: The `--dry-run` output for `repo remove` references orphan detection. If orphan detection is not available, the "workspaces that would become orphaned" output should gracefully degrade to "workspaces using this repo" without orphan terminology.
> - **add-json-output-everywhere**: The `--json` output for dry-run commands SHALL follow the JSON output schema defined in that change. Specifically, dry-run JSON output SHALL include the standard envelope (`success`, `data`, `error` fields) with dry-run-specific data nested under `data.preview`.

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
- **THEN** the preview is output as JSON following the standard envelope schema
- **AND** preview data is nested under `data.preview`
- **AND** includes all affected paths and sizes

#### Scenario: Dry run with force flag
- **WHEN** user runs `canopy workspace close PROJ-123 --dry-run --force`
- **THEN** the system still only previews
- **AND** force flag does not bypass dry-run
