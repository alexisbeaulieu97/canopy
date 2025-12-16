# Workspace Schema Reference

This document describes the `workspace.yaml` schema used by Canopy for workspace metadata.

## Schema Versioning

Workspace metadata files include a version field for forward compatibility. This allows Canopy to:

- Detect and migrate older workspace formats automatically
- Warn about workspaces created by newer versions of Canopy
- Maintain backward compatibility with existing workspaces

## Current Schema (Version 1)

```yaml
version: 1
id: "PROJ-123"
branch_name: "feature/PROJ-123"
repos:
  - name: "backend"
    url: "https://github.com/org/backend.git"
  - name: "frontend"
    url: "https://github.com/org/frontend.git"
closed_at: null  # Only present for archived workspaces
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | integer | Yes | Schema version (currently `1`) |
| `id` | string | Yes | Unique workspace identifier |
| `branch_name` | string | No | Git branch name for worktrees |
| `repos` | array | Yes | List of repositories in the workspace |
| `repos[].name` | string | Yes | Repository display name |
| `repos[].url` | string | Yes | Git clone URL |
| `closed_at` | timestamp | No | When the workspace was archived (ISO 8601) |

## Version History

### Version 1 (Current)

- Added `version` field for schema versioning
- No structural changes from version 0

### Version 0 (Legacy)

- Original schema without version field
- Workspaces without a version field are treated as version 0
- Automatically migrated to version 1 on next save

## Migration Behavior

When Canopy loads a workspace:

1. **Missing version**: Treated as version 0 (legacy workspace)
2. **Version 0-1**: Automatically upgraded to current version on save
3. **Future versions**: Warning logged, workspace loaded as-is

### Example: Legacy Workspace (No Version)

```yaml
# Old format (version 0)
id: "PROJ-123"
branch_name: "main"
repos:
  - name: "backend"
    url: "https://github.com/org/backend.git"
```

After any modification, this becomes:

```yaml
# Migrated format (version 1)
version: 1
id: "PROJ-123"
branch_name: "main"
repos:
  - name: "backend"
    url: "https://github.com/org/backend.git"
```

## Export Format

When exporting workspaces, the export file includes both:

- `version`: Export format version (string `"1"`)
- `workspace_version`: The workspace schema version (integer)

```yaml
version: "1"
workspace_version: 1
id: "PROJ-123"
branch: "main"
exported_at: "2024-01-15T10:30:00Z"
repos:
  - name: "backend"
    url: "https://github.com/org/backend.git"
    alias: "org/backend"  # Registry alias if available
```

## Compatibility Notes

- **Forward compatibility**: Canopy can read workspaces from older versions
- **Backward compatibility**: Workspaces saved with newer Canopy versions include the version field
- **Import validation**: Importing workspaces with unsupported future versions is rejected with an error
