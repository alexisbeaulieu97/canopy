# Error Codes Reference

Canopy uses typed error codes for predictable scripting and automation. This document lists all error codes, their meanings, and how to handle them.

## Exit Codes

Canopy uses the following exit codes:

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Configuration error |
| `64` | Command line usage error |
| `65` | Data format error |
| `73` | Cannot create output file |
| `74` | I/O error |

## Error Code Reference

### Workspace Errors

| Code | Description | Common Causes |
|------|-------------|---------------|
| `WORKSPACE_NOT_FOUND` | The requested workspace does not exist | Typo in workspace ID, workspace was deleted |
| `WORKSPACE_EXISTS` | A workspace with this ID already exists | Attempting to create duplicate workspace |
| `WORKSPACE_METADATA_ERROR` | Failed to read or write workspace metadata | Corrupted metadata file, permissions issue |
| `NO_REPOS_CONFIGURED` | No repositories specified and no patterns matched | Missing `--repos` flag or no matching workspace patterns |
| `MISSING_BRANCH_CONFIG` | Workspace has no branch set in metadata | Corrupted workspace metadata |

### Repository Errors

| Code | Description | Common Causes |
|------|-------------|---------------|
| `REPO_NOT_FOUND` | The requested repository is not found | Repository not cloned, typo in name |
| `REPO_NOT_CLEAN` | Repository has uncommitted changes | Uncommitted local changes blocking operation |
| `REPO_ALREADY_EXISTS` | Repository already exists in workspace | Attempting to add duplicate repo |
| `REPO_IN_USE` | Repository is used by one or more workspaces | Attempting to remove repo with active worktrees |
| `UNKNOWN_REPOSITORY` | Cannot resolve repository identifier | Not a URL and not in registry |
| `REGISTRY_ERROR` | Registry operation failed | Invalid registry file, permissions issue |

### Git Errors

| Code | Description | Common Causes |
|------|-------------|---------------|
| `GIT_OPERATION_FAILED` | A git operation failed | Network issues, authentication, missing refs |
| `OPERATION_CANCELLED` | Operation was cancelled by user | Ctrl+C pressed, context cancelled |
| `OPERATION_TIMEOUT` | Operation timed out | Network timeout, slow server |

### Configuration Errors

| Code | Description | Common Causes |
|------|-------------|---------------|
| `CONFIG_INVALID` | Configuration file is invalid | Syntax error in config.yaml |
| `CONFIG_VALIDATION` | Configuration validation failed | Invalid values in configuration |

### Input Errors

| Code | Description | Common Causes |
|------|-------------|---------------|
| `INVALID_ARGUMENT` | Invalid argument value | Invalid workspace ID characters, empty value |
| `PATH_INVALID` | Invalid path value | Path traversal attempt, invalid characters |
| `PATH_NOT_DIRECTORY` | Expected directory but got file | Wrong path type |
| `NOT_IN_WORKSPACE` | Command requires being inside a workspace | Running workspace-specific command outside workspace |

### Hook Errors

| Code | Description | Common Causes |
|------|-------------|---------------|
| `HOOK_FAILED` | Hook command exited with non-zero status | Script error, missing dependencies |
| `HOOK_TIMEOUT` | Hook execution timed out | Hook running too long |

### System Errors

| Code | Description | Common Causes |
|------|-------------|---------------|
| `IO_FAILED` | I/O operation failed | Disk full, permissions denied |
| `COMMAND_FAILED` | External command execution failed | Missing binary, permission denied |
| `INTERNAL_ERROR` | Unexpected internal error | Bug in canopy |

## JSON Output Format

When using `--json` flag, errors are returned in this format:

```json
{
  "success": false,
  "error": {
    "code": "WORKSPACE_NOT_FOUND",
    "message": "workspace my-workspace not found",
    "context": {
      "workspace_id": "my-workspace"
    }
  }
}
```

## Scripting Examples

### Bash Error Handling

```bash
#!/bin/bash
set -e

# Create workspace and capture output
if output=$(canopy workspace new TICKET-123 --repos backend --json 2>&1); then
    echo "Workspace created successfully"
    path=$(echo "$output" | jq -r '.data.path')
    cd "$path"
else
    error_code=$(echo "$output" | jq -r '.error.code')
    case "$error_code" in
        "WORKSPACE_EXISTS")
            echo "Workspace already exists, using existing"
            cd "$(canopy workspace path TICKET-123)"
            ;;
        "UNKNOWN_REPOSITORY")
            echo "Repository not found in registry"
            exit 1
            ;;
        *)
            echo "Failed: $error_code"
            exit 1
            ;;
    esac
fi
```

### Checking Specific Error Codes

```bash
# Check if workspace exists before creating
if canopy workspace view my-workspace --json 2>/dev/null | jq -e '.success' >/dev/null; then
    echo "Workspace exists"
else
    canopy workspace new my-workspace --repos backend
fi
```

### Handling Repository Errors

```bash
# Force remove a repository even if used by workspaces
if ! canopy repo remove my-repo --json 2>&1 | jq -e '.success' >/dev/null; then
    canopy repo remove my-repo --force
fi
```

## Error Context

Many errors include additional context in the `context` field:

```json
{
  "error": {
    "code": "REPO_IN_USE",
    "message": "repository backend is used by workspaces: TICKET-123, TICKET-456",
    "context": {
      "repo_name": "backend"
    }
  }
}
```

Common context fields:

| Field | Description |
|-------|-------------|
| `workspace_id` | The workspace ID involved |
| `repo_name` | The repository name involved |
| `path` | File or directory path |
| `operation` | The operation that failed |
| `detail` | Additional error details |

## Debugging Tips

1. **Use `--json` flag** for machine-readable errors with context
2. **Check `canopy check`** to validate configuration
3. **Enable debug logging** with `--debug` flag for verbose output
4. **Check permissions** on `~/.canopy/` directories
