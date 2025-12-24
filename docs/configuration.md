# Configuration Reference

Canopy uses a YAML configuration file for all settings.

## Table of Contents

- [Configuration Reference](#configuration-reference)
  - [Table of Contents](#table-of-contents)
  - [File Locations](#file-locations)
  - [Configuration Validation](#configuration-validation)
    - [Strict Field Validation](#strict-field-validation)
    - [Config Validate Command](#config-validate-command)
    - [Common Configuration Mistakes](#common-configuration-mistakes)
  - [Core Settings](#core-settings)
  - [Workspace Naming Template](#workspace-naming-template)
  - [Git Retry Settings](#git-retry-settings)
  - [Workspace Patterns](#workspace-patterns)
  - [Workspace Templates](#workspace-templates)
    - [Common Templates](#common-templates)
  - [Environment Variables](#environment-variables)
  - [Hooks](#hooks)
  - [Full Example](#full-example)
  - [TUI Keybindings](#tui-keybindings)
    - [Available Actions](#available-actions)
    - [Key Name Format](#key-name-format)
    - [Multiple Keys Per Action](#multiple-keys-per-action)
    - [Conflict Detection](#conflict-detection)

## File Locations

Configuration can be specified in multiple ways (listed in priority order):

1. **`--config` flag**: Explicit path to config file
2. **`CANOPY_CONFIG` environment variable**: Path to config file
3. **Default locations** (first found wins):
   - `./config.yaml` (current directory)
   - `~/.canopy/config.yaml`
   - `~/.config/canopy/config.yaml`

If no config file exists and no override is specified, Canopy uses sensible defaults.

### Config Override Examples

```bash
# Use a specific config file
canopy --config /path/to/config.yaml workspace list

# Use environment variable
export CANOPY_CONFIG=/path/to/config.yaml
canopy workspace list

# Per-project config (useful in CI/CD)
CANOPY_CONFIG=./ci-config.yaml canopy workspace new PROJ-123
```

## Configuration Validation

Canopy performs strict validation on configuration files to catch errors early. This includes detecting unknown fields, typos, and invalid values at startup.

### Strict Field Validation

Canopy rejects configuration files containing unknown or misspelled fields. When a typo is detected, Canopy suggests the correct field name:

```bash
$ canopy workspace list
Error: configuration validation failed for config: unknown config field "parrallel_workers", did you mean "parallel_workers"?
```

This helps catch common mistakes like:
- Typos in field names (`parrallel_workers` → `parallel_workers`)
- Obsolete field names from older versions
- Made-up fields that have no effect

### Config Validate Command

Use `canopy config validate` to check your configuration file without running other commands:

```bash
# Validate current config
canopy config validate

# Validate a specific config file
canopy config validate --config /path/to/config.yaml

# JSON output for scripting
canopy config validate --json
```

Exit codes:
- `0` - Configuration is valid
- `1` - Configuration has errors

Example output:
```
Configuration is valid.
  Projects root:   /home/user/.canopy/projects
  Workspaces root: /home/user/.canopy/workspaces
  Closed root:     /home/user/.canopy/closed
  Workspace naming: {{.ID}}
  Registry file:   /home/user/.canopy/repos.yaml
```

### Common Configuration Mistakes

| Mistake | Error | Fix |
|---------|-------|-----|
| `parrallel_workers: 4` | unknown config field "parrallel_workers" | Use `parallel_workers` |
| `stale_treshold_days: 14` | unknown config field "stale_treshold_days" | Use `stale_threshold_days` |
| `project_root: ~/projects` | unknown config field "project_root" | Use `projects_root` (plural) |
| `workspace_root: ~/ws` | unknown config field "workspace_root" | Use `workspaces_root` (plural) |
| Hook `timeout: -5` | timeout must be non-negative | Use positive timeout value |
| Hook `shell: "   "` | shell cannot be empty or whitespace-only | Provide valid shell path or omit |

## Core Settings

| Key | Default | Description |
|-----|---------|-------------|
| `projects_root` | `~/.canopy/projects` | Directory for bare git repositories |
| `workspaces_root` | `~/.canopy/workspaces` | Directory for active worktrees |
| `closed_root` | `~/.canopy/closed` | Directory for archived workspace metadata (used only when `workspace_close_default` is `archive` or `--keep` flag is passed to `workspace close`) |
| `workspace_close_default` | `delete` | Default behavior for `workspace close`. Set to `archive` to archive by default |
| `workspace_naming` | `{{.ID}}` | Template for workspace directory names |
| `parallel_workers` | `4` | Maximum number of parallel operations for workspace and repo tasks |
| `lock_timeout` | `30s` | Time to wait when acquiring a workspace lock. Uses Go duration format (e.g., `30s`, `1m`) |
| `lock_stale_threshold` | `5m` | Age after which a lock is considered stale and can be forcibly acquired. Uses Go duration format |

All paths support `~` expansion and must be absolute (after expansion).

## Workspace Naming Template

The `workspace_naming` setting uses Go templates:

| Variable | Description |
|----------|-------------|
| `{{.ID}}` | Workspace identifier |

Examples:
- `{{.ID}}` → `PROJ-123`
- `ws-{{.ID}}` → `ws-PROJ-123`

The rendered name must be a valid directory name (no path separators or traversal sequences).

## Git Retry Settings

Configure retry behavior for transient network failures during git operations:

```yaml
git:
  retry:
    max_attempts: 3       # Number of retry attempts (1-10)
    initial_delay: "1s"   # Delay before first retry
    max_delay: "30s"      # Maximum delay between retries
    multiplier: 2.0       # Exponential backoff multiplier (≥1.0)
    jitter_factor: 0.25   # Random jitter factor (0-1) to prevent thundering herd
```

| Key | Default | Description |
|-----|---------|-------------|
| `git.retry.max_attempts` | `3` | Maximum number of retry attempts (1-10) |
| `git.retry.initial_delay` | `1s` | Initial delay before retrying |
| `git.retry.max_delay` | `30s` | Maximum delay between retries |
| `git.retry.multiplier` | `2.0` | Multiplier for exponential backoff |
| `git.retry.jitter_factor` | `0.25` | Random jitter factor to prevent synchronized retries |

**When to tweak these settings:**
- Slow/unreliable network: Increase `max_attempts` and `max_delay`
- CI/CD environments: Lower `max_attempts` to fail faster
- Rate-limited APIs: Increase `initial_delay` and reduce `jitter_factor`

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

## Workspace Templates

Templates provide reusable repo sets and defaults for workspace creation:

```yaml
templates:
  backend:
    description: "Backend workspace defaults"
    repos: ["backend", "common"]
    default_branch: "main"
  frontend:
    description: "Frontend workspace defaults"
    repos: ["frontend", "ui-kit", "design-system"]
    setup_commands:
      - "npm install"
      - "npm run build"
  fullstack:
    description: "Fullstack workspace defaults"
    repos: ["backend", "frontend", "common", "ui-kit"]
```

Create a workspace using a template:

```bash
canopy workspace new PROJ-123 --template backend
canopy workspace new PROJ-456 --template frontend --repos extra-lib
```

### Common Templates

- `backend`: Backend services + shared libraries
- `frontend`: UI apps + design system
- `fullstack`: Backend + frontend + shared dependencies

## Environment Variables

All settings can be overridden via environment variables with the `CANOPY_` prefix:

```bash
export CANOPY_PROJECTS_ROOT=/custom/path
export CANOPY_WORKSPACE_CLOSE_DEFAULT=archive
```

## Hooks

Configure lifecycle hooks to run commands on workspace creation and closure:

```yaml
hooks:
  post_create:
    - command: "echo 'Created {{.WorkspaceID}}'"
    - command: "cd {{.WorkspacePath}}/backend && npm install"
  pre_close:
    - command: "docker-compose down"
```

See [Hooks Documentation](hooks.md) for complete details on template variables and configuration options.

## Full Example

```yaml
projects_root: ~/projects
workspaces_root: ~/workspaces
closed_root: ~/.canopy/closed
workspace_close_default: delete  # default; set to "archive" to keep metadata
workspace_naming: "{{.ID}}"
parallel_workers: 4
lock_timeout: "30s"
lock_stale_threshold: "5m"

git:
  retry:
    max_attempts: 3
    initial_delay: "1s"
    max_delay: "30s"
    multiplier: 2.0
    jitter_factor: 0.25

defaults:
  workspace_patterns:
    - pattern: "^PROJ-"
      repos: ["backend", "frontend", "shared"]
    - pattern: "^DOCS-"
      repos: ["documentation"]

templates:
  backend:
    description: "Backend workspace defaults"
    repos: ["backend", "common"]
    default_branch: "main"
  frontend:
    description: "Frontend workspace defaults"
    repos: ["frontend", "ui-kit", "design-system"]
    setup_commands:
      - "npm install"
      - "npm run build"
  fullstack:
    description: "Fullstack workspace defaults"
    repos: ["backend", "frontend", "common", "ui-kit"]

hooks:
  post_create:
    - command: "echo 'Workspace ready'"
      description: "Notify workspace creation"
  pre_close:
    - command: "echo 'Closing workspace'"

tui:
  use_emoji: true  # Set to false for ASCII-only output
  keybindings:
    quit: ["q", "ctrl+c"]
    open_editor: ["o", "e"]
```

## TUI Emoji Configuration

Control whether the TUI uses emoji or ASCII characters:

```yaml
tui:
  use_emoji: true  # default: true (emoji enabled)
```

When `use_emoji: false`, emoji are replaced with ASCII fallbacks for better compatibility with terminals that don't support Unicode:

| Emoji | ASCII | Usage |
|-------|-------|-------|
| 🌲 | `[W]` | Workspaces header |
| 💾 | `[D]` | Disk usage |
| 📂 | `[>]` | Workspace detail |
| ⚠ | `[!]` | Warnings |
| ✓ | `[*]` | Success indicators |
| 🔍 | `[?]` | Search filter |
| ⏳ | `[...]` | Loading indicator |
| 📁 | `[-]` | Repository |

## TUI Keybindings

Customize TUI keyboard shortcuts to match your preferences or resolve terminal conflicts:

```yaml
tui:
  keybindings:
    quit: ["q", "ctrl+c"]
    search: ["/"]
    sync: ["s"]
    push: ["p"]
    close: ["c"]
    open_editor: ["o"]
    toggle_stale: ["t"]
    details: ["enter"]
    select: ["space"]
    select_all: ["a"]
    deselect_all: ["A"]
    confirm: ["y", "Y"]
    cancel: ["n", "N", "esc"]
```

### Available Actions

| Action | Default Keys | Description |
|--------|-------------|-------------|
| `quit` | `q`, `ctrl+c` | Exit the TUI |
| `search` | `/` | Start workspace search/filter |
| `sync` | `s` | Sync selected workspaces |
| `push` | `p` | Push selected workspaces |
| `close` | `c` | Close selected workspaces |
| `open_editor` | `o` | Open workspace in editor |
| `toggle_stale` | `t` | Toggle stale workspace filter |
| `details` | `enter` | View workspace details |
| `select` | `space` | Toggle workspace selection |
| `select_all` | `a` | Select all visible workspaces |
| `deselect_all` | `A` | Deselect all workspaces |
| `confirm` | `y`, `Y` | Confirm action in dialogs |
| `cancel` | `n`, `N`, `esc` | Cancel/go back |

### Key Name Format

Keys are specified as strings:
- **Regular keys**: Lowercase letters (`a`, `b`, `q`), numbers (`1`, `2`), and symbols (`/`, `.`, `-`)
- **Uppercase letters**: Use uppercase directly (`Y`, `N`) for case-sensitive matching in dialogs
- **Modifier keys**: `ctrl+<key>`, `alt+<key>`, `shift+<key>` (e.g., `ctrl+c`, `alt+x`)
- **Special keys**: `enter`, `esc`, `tab`, `backspace`, `delete`, `space`
- **Arrow keys**: `up`, `down`, `left`, `right`, `home`, `end`, `pgup`, `pgdown`
- **Function keys**: `f1` through `f12`

Examples: `ctrl+c`, `shift+a`, `alt+x`, `enter`, `esc`, `Y`

**Note**: Keys are case-sensitive. `y` and `Y` are different keybindings.

### Multiple Keys Per Action

Each action can have multiple keys assigned:

```yaml
tui:
  keybindings:
    open_editor: ["o", "e"]  # Both 'o' and 'e' open in editor
    quit: ["q", "ctrl+c", "ctrl+q"]
```

### Conflict Detection

Canopy validates keybindings at startup. If the same key is assigned to multiple actions, an error is returned listing all conflicts.
