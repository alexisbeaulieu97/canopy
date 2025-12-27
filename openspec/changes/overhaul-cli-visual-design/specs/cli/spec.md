## ADDED Requirements

### Requirement: Table-Formatted List Output

The `workspace list` command SHALL display workspaces in a formatted table with aligned columns and box-drawing borders.

#### Scenario: Basic workspace list

- **WHEN** user runs `canopy workspace list`
- **THEN** output SHALL be a formatted table:
  ```
  ┌─ Workspaces ──────────────────────────────────────────────────────┐
  │                                                                   │
  │  WORKSPACE      REPOS   SIZE      MODIFIED     STATUS             │
  │  ───────────────────────────────────────────────────────────────  │
  │  PROJ-123       3       45.2 MB   2 hours ago  ✓ clean            │
  │  PROJ-456       5       128.3 MB  14 days ago  ● 2 dirty          │
  │  feature-auth   2       23.1 MB   1 day ago    ✓ clean            │
  │                                                                   │
  └───────────────────────────────────────────────────────────────────┘

  3 workspaces • 196.6 MB total
  ```

#### Scenario: Workspace list with status details

- **WHEN** user runs `canopy workspace list --status`
- **THEN** status column SHALL show detailed status:
  ```
  │  WORKSPACE      REPOS   SIZE      MODIFIED     STATUS             │
  │  ───────────────────────────────────────────────────────────────  │
  │  PROJ-123       3       45.2 MB   2 hours ago  ✓ clean            │
  │  PROJ-456       5       128.3 MB  14 days ago  ● 2 dirty ↑ 1 unpush │
  │  feature-auth   2       23.1 MB   1 day ago    ⚠ 3 behind         │
  ```

#### Scenario: Empty workspace list

- **WHEN** user runs `canopy workspace list` with no workspaces
- **THEN** output SHALL show:
  ```
  No workspaces found.

  Create one with: canopy workspace new <name> --repos <repo1,repo2>
  ```

#### Scenario: List with NO_COLOR

- **GIVEN** `NO_COLOR=1` environment variable is set
- **WHEN** user runs `canopy workspace list`
- **THEN** output SHALL use ASCII box characters and no ANSI colors:
  ```
  +- Workspaces -----------------------------------------+
  |                                                      |
  |  WORKSPACE      REPOS   SIZE      MODIFIED  STATUS   |
  |  --------------------------------------------------- |
  |  PROJ-123       3       45.2 MB   2h ago    [ok]     |
  |  PROJ-456       5       128.3 MB  14d ago   * dirty  |
  |                                                      |
  +------------------------------------------------------+
  ```

### Requirement: Workspace View Sectioned Output

The `workspace view` command SHALL display workspace details in visually distinct sections.

#### Scenario: View workspace details

- **WHEN** user runs `canopy workspace view PROJ-123`
- **THEN** output SHALL be sectioned:
  ```
  ┌─ Workspace: PROJ-123 ─────────────────────────────────────────────┐
  │                                                                   │
  │  Branch       feature/user-auth                                   │
  │  Path         /Users/dev/workspaces/PROJ-123                      │
  │  Disk Size    45.2 MB                                             │
  │  Modified     2 hours ago (Dec 24, 2025 10:30 AM)                 │
  │  Created      Dec 20, 2025                                        │
  │                                                                   │
  ├─ Repositories (3) ────────────────────────────────────────────────┤
  │                                                                   │
  │  NAME          BRANCH              STATUS                         │
  │  ─────────────────────────────────────────────────────────────    │
  │  api           feature/user-auth   ● 2 modified  ↑ 1 unpushed     │
  │  frontend      feature/user-auth   ✓ clean                        │
  │  worker        feature/user-auth   ✓ clean       ↓ 3 behind       │
  │                                                                   │
  └───────────────────────────────────────────────────────────────────┘
  ```

#### Scenario: View workspace with orphaned worktrees

- **GIVEN** workspace has orphaned worktrees
- **WHEN** user runs `canopy workspace view PROJ-123`
- **THEN** a warning section SHALL appear:
  ```
  ┌─ ⚠ Warning ───────────────────────────────────────────────────────┐
  │                                                                   │
  │  2 orphaned worktrees found:                                      │
  │    • old-feature: branch 'old-feature' no longer exists           │
  │    • experiment: remote repository was removed                    │
  │                                                                   │
  │  Run 'canopy workspace cleanup PROJ-123' to remove them.          │
  │                                                                   │
  └───────────────────────────────────────────────────────────────────┘
  ```

#### Scenario: View workspace with errors

- **GIVEN** workspace has repos with errors
- **WHEN** user runs `canopy workspace view PROJ-123`
- **THEN** errors SHALL be shown inline:
  ```
  │  NAME          BRANCH              STATUS                         │
  │  ─────────────────────────────────────────────────────────────    │
  │  api           feature/user-auth   ✗ error: permission denied     │
  │  frontend      feature/user-auth   ✓ clean                        │
  ```

### Requirement: Progress Indicators for Operations

Long-running CLI operations SHALL display progress feedback.

#### Scenario: Sync operation progress

- **WHEN** user runs `canopy workspace sync PROJ-123`
- **THEN** output SHALL show real-time progress:
  ```
  Syncing PROJ-123...

  ⠋ Syncing api...
  ```
  (spinner animates)
  ```
  ✓ api               pulled 3 commits
  ⠋ Syncing frontend...
  ```
  (operation completes)
  ```
  ✓ api               pulled 3 commits
  ✓ frontend          already up-to-date
  ✓ worker            pulled 1 commit

  ────────────────────────────────────────────────
  Sync complete: 3 repos synced, 4 commits pulled
  ```

#### Scenario: Push operation progress

- **WHEN** user runs `canopy workspace push PROJ-123`
- **THEN** output SHALL show:
  ```
  Pushing PROJ-123...

  ✓ api               pushed 2 commits
  ✓ frontend          nothing to push
  ✗ worker            error: remote rejected (protected branch)

  ────────────────────────────────────────────────
  Push complete: 1 pushed, 1 skipped, 1 failed
  ```

#### Scenario: Progress with --quiet flag

- **WHEN** user runs `canopy workspace sync PROJ-123 --quiet`
- **THEN** only the final summary SHALL be shown (no spinners, no per-repo lines)

### Requirement: Styled Error Output

CLI errors SHALL be displayed in a styled format with context and suggestions.

#### Scenario: Workspace not found error

- **WHEN** user runs `canopy workspace view NONEXISTENT`
- **THEN** error output SHALL be:
  ```
  ┌─ Error ────────────────────────────────────────────────────────────┐
  │                                                                    │
  │  ✗ Workspace not found: NONEXISTENT                                │
  │                                                                    │
  │  No workspace with ID 'NONEXISTENT' exists.                        │
  │                                                                    │
  │  Similar workspaces:                                               │
  │    • PROJ-123                                                      │
  │    • feature-auth                                                  │
  │                                                                    │
  │  List all workspaces: canopy workspace list                        │
  │                                                                    │
  └────────────────────────────────────────────────────────────────────┘
  ```
- **AND** exit code SHALL be 1
- **AND** output SHALL go to stderr

#### Scenario: Configuration error

- **WHEN** config file has invalid YAML
- **THEN** error output SHALL be:
  ```
  ┌─ Configuration Error ──────────────────────────────────────────────┐
  │                                                                    │
  │  ✗ Failed to parse config file                                     │
  │                                                                    │
  │  File: /Users/dev/.canopy/config.yaml                              │
  │  Line: 12                                                          │
  │  Error: unexpected character ':'                                   │
  │                                                                    │
  │  Run 'canopy doctor' to validate your configuration.               │
  │                                                                    │
  └────────────────────────────────────────────────────────────────────┘
  ```

#### Scenario: Operation error with recovery suggestion

- **WHEN** push fails due to uncommitted changes
- **THEN** error output SHALL suggest resolution:
  ```
  ┌─ Push Failed ──────────────────────────────────────────────────────┐
  │                                                                    │
  │  ✗ Cannot push: uncommitted changes in 'api'                       │
  │                                                                    │
  │  The repository 'api' has 3 modified files that must be            │
  │  committed or stashed before pushing.                              │
  │                                                                    │
  │  Options:                                                          │
  │    1. Commit changes: cd api && git add . && git commit            │
  │    2. Stash changes:  cd api && git stash                          │
  │    3. Force push:     canopy workspace push PROJ-123 --force       │
  │                                                                    │
  └────────────────────────────────────────────────────────────────────┘
  ```

### Requirement: Styled Success Output

Success messages SHALL use consistent formatting with icons and colors.

#### Scenario: Workspace created success

- **WHEN** user runs `canopy workspace new PROJ-789 --repos api,frontend`
- **THEN** success output SHALL be:
  ```
  ✓ Created workspace PROJ-789

    Path:  /Users/dev/workspaces/PROJ-789
    Repos: api, frontend

  Open in editor: canopy workspace open PROJ-789
  View details:   canopy workspace view PROJ-789
  ```

#### Scenario: Workspace closed success

- **WHEN** user runs `canopy workspace close PROJ-123` and confirms
- **THEN** success output SHALL be:
  ```
  ✓ Closed workspace PROJ-123

    Removed: 3 worktrees
    Freed:   45.2 MB
  ```

#### Scenario: Multiple workspaces synced

- **WHEN** user runs `canopy workspace sync --all`
- **THEN** success output SHALL summarize:
  ```
  ✓ Synced 5 workspaces

    Repos updated:  12
    Commits pulled: 47
    Already synced: 3

  1 workspace had errors (use --verbose for details)
  ```

### Requirement: Color and Symbol Detection

CLI output SHALL adapt to terminal capabilities and user preferences.

#### Scenario: Color detection precedence

- **WHEN** determining whether to use colors
- **THEN** the following precedence SHALL apply:
  1. `NO_COLOR` environment variable set → no colors
  2. `CANOPY_COLOR=always` → force colors
  3. `CANOPY_COLOR=never` → no colors
  4. stdout is not a TTY → no colors
  5. `TERM=dumb` → no colors
  6. Otherwise → colors enabled

#### Scenario: Piped output

- **WHEN** output is piped to another command (e.g., `canopy workspace list | grep PROJ`)
- **THEN** ANSI color codes SHALL NOT be included
- **AND** Unicode box characters SHALL be replaced with ASCII

#### Scenario: Force colors in pipe

- **WHEN** `CANOPY_COLOR=always` is set
- **AND** output is piped
- **THEN** colors SHALL be included (user explicitly requested)

### Requirement: Consistent Status Icons

Status indicators SHALL use a consistent icon set across all CLI output.

#### Scenario: Status icon definitions

- **WHEN** displaying status in any CLI command
- **THEN** icons SHALL be consistent:

| Status     | Unicode | ASCII    | Color  |
|------------|---------|----------|--------|
| Success    | `✓`     | `[ok]`   | Green  |
| Warning    | `⚠`     | `[!]`    | Amber  |
| Error      | `✗`     | `[X]`    | Red    |
| Info       | `ℹ`     | `[i]`    | Blue   |
| Dirty      | `●`     | `*`      | Red    |
| Clean      | `○`     | `-`      | Green  |
| Unpushed   | `↑`     | `^`      | Red    |
| Behind     | `↓`     | `v`      | Amber  |
| Loading    | `⠋⠙⠹⠸` | `...`    | Cyan   |

#### Scenario: Icon usage consistency

- **WHEN** displaying a "success" status
- **THEN** the `✓` icon (or `[ok]` in ASCII) SHALL be used
- **AND** the icon SHALL be colored green
- **AND** the same icon/color SHALL be used in all commands (list, view, sync, etc.)

### Requirement: Summary Footers

Commands that list or process multiple items SHALL include summary footers.

#### Scenario: List command footer

- **WHEN** `canopy workspace list` completes
- **THEN** a summary line SHALL appear after the table:
  ```
  5 workspaces • 234.7 MB total
  ```

#### Scenario: Status list footer

- **WHEN** `canopy workspace list --status` completes
- **THEN** summary SHALL include status counts:
  ```
  5 workspaces • 234.7 MB total • 2 dirty • 1 needs sync
  ```

#### Scenario: Sync command footer

- **WHEN** `canopy workspace sync PROJ-123` completes
- **THEN** a summary line SHALL appear:
  ```
  ────────────────────────────────────────────────
  Sync complete: 3 repos synced, 4 commits pulled
  ```

#### Scenario: Bulk operation footer

- **WHEN** `canopy workspace sync --all` completes
- **THEN** summary SHALL show aggregate results:
  ```
  ────────────────────────────────────────────────
  5 workspaces synced • 12 repos updated • 47 commits pulled
  1 error (PROJ-456: network timeout)
  ```
