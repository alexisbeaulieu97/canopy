# Error Codes Reference

Canopy uses typed error codes for predictable scripting and automation. This document lists all error codes, their meanings, and how to handle them.

## Exit Codes

Canopy uses the following exit codes:

| Exit Code | Meaning | Mapped Error Codes |
|-----------|---------|--------------------|
| `0` | Success | N/A |
| `1` | General error | Unclassified error without an error code |
| `2` | Workspace/resource not found | `WORKSPACE_NOT_FOUND`, `REPO_NOT_FOUND` |
| `3` | Resource already exists | `WORKSPACE_EXISTS`, `REPO_ALREADY_EXISTS` |
| `4` | Dirty workspace (uncommitted changes) | `REPO_NOT_CLEAN` |
| `5` | Configuration error | `CONFIG_INVALID`, `CONFIG_VALIDATION` |
| `6` | Git operation failed | `GIT_OPERATION_FAILED` |
| `7` | Unknown resource | `UNKNOWN_REPOSITORY` |
| `8` | Not in workspace | `NOT_IN_WORKSPACE` |
| `9` | Invalid argument | `INVALID_ARGUMENT` |
| `10` | I/O error | `IO_FAILED` |
| `11` | Registry error | `REGISTRY_ERROR` |
| `12` | Command execution failed | `COMMAND_FAILED` |
| `13` | Internal error | `INTERNAL_ERROR` |
| `14` | Repository in use | `REPO_IN_USE` |
| `15` | Metadata error | `WORKSPACE_METADATA_ERROR` |
| `16` | No repositories configured | `NO_REPOS_CONFIGURED` |
| `17` | Missing branch configuration | `MISSING_BRANCH_CONFIG` |
| `18` | Operation aborted/cancelled | `OPERATION_CANCELLED` |
| `19` | Workspace locked | `WORKSPACE_LOCKED` |
| `20` | Unpushed commits present | `REPO_HAS_UNPUSHED_COMMITS` |
| `21` | Operation timeout | `OPERATION_TIMEOUT`, `HOOK_TIMEOUT` |
| `22` | Hook failed | `HOOK_FAILED` |
| `23` | Path error | `PATH_INVALID`, `PATH_NOT_DIRECTORY` |

Note: Multiple error codes can map to the same exit code. Use the error code in JSON output for exact diagnosis.

## Error Code Reference

### Workspace Errors

| Code | Exit Code | Description | Common Causes |
|------|-----------|-------------|---------------|
| `WORKSPACE_NOT_FOUND` | `2` | The requested workspace does not exist | Typo in workspace ID, workspace was deleted |
| `WORKSPACE_EXISTS` | `3` | A workspace with this ID already exists | Attempting to create duplicate workspace |
| `WORKSPACE_LOCKED` | `19` | Workspace is locked by another operation | Concurrent workspace operation, stale lock |
| `WORKSPACE_METADATA_ERROR` | `15` | Failed to read or write workspace metadata | Corrupted metadata file, permissions issue |
| `NO_REPOS_CONFIGURED` | `16` | No repositories specified and no patterns matched | Missing `--repos` flag or no matching workspace patterns |
| `MISSING_BRANCH_CONFIG` | `17` | Workspace has no branch set in metadata | Corrupted workspace metadata |

### Repository Errors

| Code | Exit Code | Description | Common Causes |
|------|-----------|-------------|---------------|
| `REPO_NOT_FOUND` | `2` | The requested repository is not found | Repository not cloned, typo in name |
| `REPO_NOT_CLEAN` | `4` | Repository has uncommitted changes | Uncommitted local changes blocking operation |
| `REPO_ALREADY_EXISTS` | `3` | Repository already exists in workspace | Attempting to add duplicate repo |
| `REPO_IN_USE` | `14` | Repository is used by one or more workspaces | Attempting to remove repo with active worktrees |
| `REPO_HAS_UNPUSHED_COMMITS` | `20` | Repository has unpushed commits | Local commits not pushed to remote |
| `UNKNOWN_REPOSITORY` | `7` | Cannot resolve repository identifier | Not a URL and not in registry |
| `REGISTRY_ERROR` | `11` | Registry operation failed | Invalid registry file, permissions issue |

### Git Errors

| Code | Exit Code | Description | Common Causes |
|------|-----------|-------------|---------------|
| `GIT_OPERATION_FAILED` | `6` | A git operation failed | Network issues, authentication, missing refs |
| `OPERATION_CANCELLED` | `18` | Operation was cancelled by user | Ctrl+C pressed, context cancelled |
| `OPERATION_TIMEOUT` | `21` | Operation timed out | Network timeout, slow server |

### Configuration Errors

| Code | Exit Code | Description | Common Causes |
|------|-----------|-------------|---------------|
| `CONFIG_INVALID` | `5` | Configuration file is invalid | Syntax error in config.yaml |
| `CONFIG_VALIDATION` | `5` | Configuration validation failed | Invalid values in configuration |

### Input Errors

| Code | Exit Code | Description | Common Causes |
|------|-----------|-------------|---------------|
| `INVALID_ARGUMENT` | `9` | Invalid argument value | Invalid workspace ID characters, empty value |
| `PATH_INVALID` | `23` | Invalid path value | Path traversal attempt, invalid characters |
| `PATH_NOT_DIRECTORY` | `23` | Expected directory but got file | Wrong path type |
| `NOT_IN_WORKSPACE` | `8` | Command requires being inside a workspace | Running workspace-specific command outside workspace |

### Hook Errors

| Code | Exit Code | Description | Common Causes |
|------|-----------|-------------|---------------|
| `HOOK_FAILED` | `22` | Hook command exited with non-zero status | Script error, missing dependencies |
| `HOOK_TIMEOUT` | `21` | Hook execution timed out | Hook running too long |

### System Errors

| Code | Exit Code | Description | Common Causes |
|------|-----------|-------------|---------------|
| `IO_FAILED` | `10` | I/O operation failed | Disk full, permissions denied |
| `COMMAND_FAILED` | `12` | External command execution failed | Missing binary, permission denied |
| `INTERNAL_ERROR` | `13` | Unexpected internal error | Bug in canopy |

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
