# CLI Spec (Delta)

## ADDED Requirements

### Requirement: Global JSON Output Flag
All CLI commands SHALL support a `--json` flag for machine-parseable output.

#### Scenario: Enable JSON output
- **WHEN** user runs any command with `--json` flag
- **THEN** output SHALL be valid JSON
- **AND** output SHALL use consistent structure across all commands

#### Scenario: JSON output for scripting
- **WHEN** `--json` flag is provided
- **THEN** output SHALL be parseable by tools like `jq`
- **AND** no human-readable decorations SHALL be included

### Requirement: JSON Error Handling
Errors SHALL be output as structured JSON when `--json` flag is set.

#### Scenario: Error in JSON mode
- **WHEN** command fails with `--json` flag
- **THEN** output SHALL include `"success": false`
- **AND** `"error"` object SHALL contain `code`, `message`, and `context` fields

#### Scenario: Success in JSON mode
- **WHEN** command succeeds with `--json` flag
- **THEN** output SHALL include `"success": true`
- **AND** `"data"` field SHALL contain command-specific results

### Requirement: Consistent JSON Structure
All JSON output SHALL follow a standard envelope format.

#### Scenario: Standard envelope format
- **WHEN** any command outputs JSON
- **THEN** response SHALL have top-level `success`, `data`, and `error` fields
- **AND** `data` SHALL be null on error, `error` SHALL be null on success

## Reference

### Output Structure
```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

### Error Structure
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "WORKSPACE_NOT_FOUND",
    "message": "Workspace 'foo' not found",
    "context": { "workspace_id": "foo" }
  }
}
```

### Example: `canopy workspace list --json`
```json
{
  "success": true,
  "data": {
    "workspaces": [
      {
        "id": "feature-x",
        "path": "/path/to/ws",
        "repos": ["main", "lib"],
        "branch": "feature/x"
      }
    ]
  }
}
```

### Example: `canopy status --json`
```json
{
  "success": true,
  "data": {
    "workspace": "feature-x",
    "repos": [
      {
        "name": "main",
        "branch": "feature/x",
        "clean": false,
        "ahead": 2,
        "behind": 0,
        "modified": 3
      }
    ]
  }
}
```
