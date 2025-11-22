# Configuration Reference

Canopy uses a YAML configuration file. Files are loaded in this order (first found wins):

1. `./config.yaml` (current directory)
2. `~/.canopy/config.yaml`
3. `~/.config/canopy/config.yaml`

If no config file exists, Canopy uses sensible defaults.

## Settings

| Key | Default | Description |
|-----|---------|-------------|
| `projects_root` | `~/.canopy/projects` | Directory for bare git repositories |
| `workspaces_root` | `~/.canopy/workspaces` | Directory for active worktrees |
| `archives_root` | `~/.canopy/archives` | Directory for archived workspace metadata |
| `workspace_close_default` | `delete` | Behavior when `workspace close` is called without flags. Must be `delete` or `archive`. Override per-command with `--archive` or `--no-archive` |
| `workspace_naming` | `{{.ID}}` | Template for workspace directory names |

All paths support `~` expansion and must be absolute (after expansion).

## Workspace Patterns

Auto-assign repositories to workspaces based on ID patterns:

```yaml
defaults:
  workspace_patterns:
    - pattern: "^PROJ-"
      repos: ["backend", "frontend"]
    - pattern: "^INFRA-"
      repos: ["infrastructure"]
```

When creating a workspace with an ID matching a pattern, the configured repos are used automatically if `--repos` is not specified.

## Environment Variables

All settings can be overridden via environment variables with the `CANOPY_` prefix:

```bash
export CANOPY_PROJECTS_ROOT=/custom/path
export CANOPY_WORKSPACE_CLOSE_DEFAULT=archive
```

## Full Example

```yaml
projects_root: ~/projects
workspaces_root: ~/workspaces
archives_root: ~/.canopy/archives
workspace_close_default: archive
workspace_naming: "{{.ID}}"

defaults:
  workspace_patterns:
    - pattern: "^PROJ-"
      repos: ["backend", "frontend", "shared"]
    - pattern: "^DOCS-"
      repos: ["documentation"]
```
