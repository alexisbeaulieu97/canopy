# CLI Spec (Delta)

## Purpose
CLI interface conventions and output formatting.

## JSON Output Support

### Global Flag
- `--json` flag enables JSON output for all commands
- Errors also output as JSON when flag is set
- Machine-parseable for scripting and CI

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

### Command-Specific Data

#### `canopy workspace list --json`
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

#### `canopy status --json`
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
