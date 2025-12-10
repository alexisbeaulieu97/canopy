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

## TUI Keybindings

Customize TUI keyboard shortcuts to match your preferences or resolve terminal conflicts:

```yaml
tui:
  keybindings:
    quit: ["q", "ctrl+c"]
    search: ["/"]
    push: ["p"]
    close: ["c"]
    open_editor: ["o"]
    toggle_stale: ["s"]
    details: ["enter"]
    confirm: ["y", "Y"]
    cancel: ["n", "N", "esc"]
```

### Available Actions

| Action | Default Keys | Description |
|--------|-------------|-------------|
| `quit` | `q`, `ctrl+c` | Exit the TUI |
| `search` | `/` | Start workspace search/filter |
| `push` | `p` | Push selected workspace |
| `close` | `c` | Close selected workspace |
| `open_editor` | `o` | Open workspace in editor |
| `toggle_stale` | `s` | Toggle stale workspace filter |
| `details` | `enter` | View workspace details |
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

