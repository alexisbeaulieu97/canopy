# Design: Lifecycle Hooks

## Overview
This document covers the execution model, failure handling, and security constraints for workspace lifecycle hooks.

## Execution Model

### Hook Types
| Hook | Trigger Point | Working Directory | Context Available |
|------|--------------|-------------------|-------------------|
| `post_create` | After workspace directory and worktrees are created | Workspace root or specific repo | Workspace ID, branch, repo list |
| `pre_close` | Before workspace deletion begins | Workspace root or specific repo | Workspace ID, branch, repo list |

### Execution Order
1. Hooks execute sequentially in the order defined in config
2. If `repos` filter is specified, hook runs once per matching repo
3. If no `repos` filter, hook runs once in workspace root directory

### Environment Variables
Hooks receive context via environment variables:
- `CANOPY_WORKSPACE_ID` - Workspace identifier
- `CANOPY_WORKSPACE_PATH` - Absolute path to workspace directory
- `CANOPY_BRANCH` - Branch name for the workspace
- `CANOPY_REPO_NAME` - Current repo name (when running per-repo)
- `CANOPY_REPO_PATH` - Current repo path (when running per-repo)

## Failure Handling

### Failure Modes
| Scenario | Behavior | Rationale |
|----------|----------|-----------|
| Hook command not found | Fail with clear error | User config error, should be fixed |
| Hook exits non-zero | Fail by default, continue with `--continue-on-hook-error` | Respect hook's exit status |
| Hook timeout (30s default) | Kill process, fail | Prevent hanging operations |
| Hook writes to stderr | Log as warning, continue | Stderr is informational |

### Rollback Strategy
- **post_create failure**: Workspace is already created; log error but keep workspace
- **pre_close failure**: Abort close operation, workspace remains intact
- No automatic rollback of partial hook execution

### Error Reporting
```text
Hook failed: post_create[0] in repo 'backend'
  Command: npm install
  Exit code: 1
  Stderr: npm ERR! code ENETUNREACH
  
Use --force to close anyway, or fix the issue and retry.
```

## Security Constraints

### Threat Model
| Threat | Mitigation |
|--------|------------|
| Malicious config injection | Config file is user-controlled; trust boundary is user's machine |
| Command injection via workspace ID | Workspace IDs are validated; IDs passed as env vars, not shell-interpolated |
| Runaway resource consumption | Timeout enforcement (configurable, default 30s) |
| Privilege escalation | Hooks run with same privileges as canopy process |

### Execution Environment
- Hooks run via user's default shell (`$SHELL -c "command"`)
- No sandboxing - hooks have full user permissions
- Working directory is explicitly set before execution
- No network restrictions

### Validation Rules
1. Config file must be owned by current user (or root)
2. Hook commands are logged before execution when `--debug` is set
3. `--no-hooks` flag available to skip all hooks
4. `--dry-run` shows hooks that would execute without running them

### Audit Logging
When hooks execute:
```text
[DEBUG] Executing hook: post_create[0]
[DEBUG]   Command: npm install
[DEBUG]   Working dir: /Users/alex/workspaces/PROJ-123/frontend
[DEBUG]   Timeout: 30s
[INFO]  Hook post_create[0] completed (exit 0, 2.3s)
```

## Configuration Schema

```yaml
hooks:
  post_create:
    - command: "npm install"
      repos: ["frontend"]      # Optional: filter to specific repos
      timeout: 60              # Optional: override default 30s
      continue_on_error: false # Optional: don't fail workspace creation
    - command: "go mod download"
      repos: ["backend"]
  pre_close:
    - command: "git stash"     # Runs in each repo
```

## Deployment Considerations

### Migration
- Existing configs without `hooks` section continue to work unchanged
- No database migrations required
- Feature is purely additive

### Backward Compatibility
- `--no-hooks` ensures scripts that don't expect hooks can opt out
- Default timeout prevents breaking existing workflows with slow hooks
