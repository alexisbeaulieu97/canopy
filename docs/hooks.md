# Lifecycle Hooks

Canopy supports lifecycle hooks that run shell commands at specific points during workspace operations. Use hooks to automate setup tasks, notifications, or cleanup.

## Hook Types

| Hook | Trigger | Use Cases |
|------|---------|-----------|
| `post_create` | After workspace creation | Install dependencies, start services, send notifications |
| `pre_close` | Before workspace closure | Backup data, stop services, cleanup temp files |

## Configuration

Define hooks in your `~/.canopy/config.yaml`:

```yaml
hooks:
  post_create:
    - command: "echo 'Workspace {{.WorkspaceID}} created at {{.WorkspacePath}}'"
    - command: "cd {{.WorkspacePath}}/backend && npm install"
      description: "Install backend dependencies"

  pre_close:
    - command: "echo 'Closing workspace {{.WorkspaceID}}'"
```

## Hook Structure

Each hook has the following properties:

| Property | Required | Description |
|----------|----------|-------------|
| `command` | Yes | Shell command to execute |
| `description` | No | Human-readable description (shown in logs) |

## Template Variables

Hooks support Go template variables for dynamic command construction:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `{{.WorkspaceID}}` | Workspace identifier | `PROJ-123` |
| `{{.WorkspacePath}}` | Absolute path to workspace | `/home/user/workspaces/PROJ-123` |
| `{{.BranchName}}` | Git branch name | `PROJ-123` |
| `{{.Repos}}` | List of repository objects | See below |

### Working with Repos

The `{{.Repos}}` variable is a list of repository objects. Each repo has:
- `.Name` - Repository name
- `.URL` - Repository URL

Example iterating over repos:
```yaml
hooks:
  post_create:
    - command: "for repo in {{range .Repos}}{{.Name}} {{end}}; do echo \"Repo: $repo\"; done"
```

## Execution

### Working Directory

Hooks execute with the **workspace directory** as the working directory.

### Error Handling

By default, hook failures stop execution:
- If a `post_create` hook fails, the error is reported but the workspace remains created
- If a `pre_close` hook fails, the workspace is not closed

### Running Hooks Independently

Run hooks without performing the associated workspace operation:

```bash
# Run post_create hooks for an existing workspace
canopy workspace new PROJ-123 --hooks-only

# Run pre_close hooks without closing
canopy workspace close PROJ-123 --hooks-only
```

## Security

Canopy validates hook commands to prevent injection attacks:

- **Null bytes rejected** — Prevents path traversal and injection attacks
- **Newlines rejected** — Commands must be single-line; use `&&` or `;` for chaining within the command
- **Empty commands rejected** — Whitespace-only commands are not allowed

These validations prevent shell injection attacks where untrusted input could execute arbitrary commands.

> **Note:** Multi-line YAML syntax (`|`) is not supported for hook commands. All commands must be on a single line.

## Examples

### Install Dependencies

```yaml
hooks:
  post_create:
    - command: "cd {{.WorkspacePath}}/backend && go mod download"
      description: "Download Go dependencies"
    - command: "cd {{.WorkspacePath}}/frontend && npm ci"
      description: "Install npm packages"
```

### Send Slack Notification

```yaml
hooks:
  post_create:
    - command: "curl -X POST -H 'Content-type: application/json' --data '{\"text\":\"Started work on {{.WorkspaceID}}\"}' $SLACK_WEBHOOK_URL"
      description: "Notify team"
```

### Start Development Services

```yaml
hooks:
  post_create:
    - command: "cd {{.WorkspacePath}} && docker-compose up -d"
      description: "Start local services"

  pre_close:
    - command: "cd {{.WorkspacePath}} && docker-compose down"
      description: "Stop local services"
```

### Open in Editor

```yaml
hooks:
  post_create:
    - command: "code {{.WorkspacePath}}"
      description: "Open in VS Code"
```
