## ADDED Requirements

### Requirement: Panel-Based Layout System

The TUI SHALL use a panel-based layout system with distinct visual sections for header, content, and footer areas.

Each panel SHALL have:
- Box borders using Unicode box-drawing characters (or ASCII fallback)
- Consistent internal padding (1 character horizontal, 0 vertical)
- Clear visual separation from adjacent panels

#### Scenario: Main layout structure

- **WHEN** the TUI is rendered
- **THEN** the layout SHALL consist of three vertical sections:
  ```
  ┌─ Header ──────────────────────────────────────────────────┐
  │  Canopy                    3 workspaces • 156.4 MB total │
  └───────────────────────────────────────────────────────────┘
  ┌─ Content ─────────────────────────────────────────────────┐
  │                                                           │
  │  (workspace list or detail view)                          │
  │                                                           │
  └───────────────────────────────────────────────────────────┘
  ┌─ Footer ──────────────────────────────────────────────────┐
  │ ↑↓ navigate • enter details • / search • q quit          │
  └───────────────────────────────────────────────────────────┘
  ```

#### Scenario: ASCII fallback for box characters

- **GIVEN** `tui.use_emoji` is false OR terminal does not support Unicode
- **WHEN** rendering panel borders
- **THEN** ASCII characters SHALL be used:
  ```
  +- Header ---------------------------------------------+
  |  Canopy                    3 workspaces • 156.4 MB  |
  +------------------------------------------------------+
  ```

### Requirement: Header Component

The TUI header SHALL display a logo/title, summary statistics, and contextual information.

#### Scenario: Default header display

- **WHEN** viewing the workspace list
- **THEN** the header SHALL display:
  - Left: Application name "Canopy" with tree icon (, or `[W]` in ASCII)
  - Right: Summary stats "N workspaces • X.X MB total"

Example:
```
┌───────────────────────────────────────────────────────────────┐
│  Canopy                              5 workspaces • 234.7 MB │
└───────────────────────────────────────────────────────────────┘
```

#### Scenario: Header with active filters

- **WHEN** a search filter is active
- **THEN** the header SHALL show the filter indicator:
  ```
  │  Canopy    🔍 "auth"                  2/5 shown • 234.7 MB │
  ```

- **WHEN** the stale filter is active
- **THEN** the header SHALL show:
  ```
  │  Canopy    ⏰ stale only                1/5 shown • 45.2 MB │
  ```

#### Scenario: Header in detail view

- **WHEN** viewing workspace details
- **THEN** the header SHALL show breadcrumb navigation:
  ```
  │  Canopy › PROJ-123                        45.2 MB • 3 repos │
  ```

### Requirement: Two-Line Workspace List Items

Workspace list items SHALL use a compact two-line layout with inline status indicators.

#### Scenario: Clean workspace item

- **GIVEN** workspace `PROJ-123` has no dirty repos, no unpushed commits
- **WHEN** rendered in the list
- **THEN** it SHALL display as:
  ```
  ❯   PROJ-123                                             clean
       3 repos • 45.2 MB • 2d ago
  ```

#### Scenario: Workspace with issues

- **GIVEN** workspace `PROJ-456` has 2 dirty repos, 1 repo behind remote, and is stale
- **WHEN** rendered in the list
- **THEN** it SHALL display as:
  ```
     PROJ-456                         ● dirty(2) ↓ behind(1) ⏰ stale
       5 repos • 128.3 MB • 14d ago
  ```

Status pills SHALL be color-coded:
- `● dirty(N)` in danger color (#EF4444)
- `↑ unpush(N)` in danger color (#EF4444)
- `↓ behind(N)` in warning color (#F59E0B)
- `⏰ stale` in warning color (#F59E0B)
- ` clean` in success color (#22C55E)

#### Scenario: Selected workspace highlighting

- **GIVEN** workspace `PROJ-123` is under the cursor
- **WHEN** rendered
- **THEN** the row SHALL have:
  - Cursor indicator (`❯`) in accent color (#8B5CF6)
  - Row background in subtle color (#374151)
  - Text in primary color (#F9FAFB)

#### Scenario: Multi-selected workspace

- **GIVEN** workspace `PROJ-123` is in the selection set
- **WHEN** rendered
- **THEN** a selection checkbox SHALL appear:
  ```
  ❯ [x] PROJ-123                                          clean
  ```
- **AND** `[x]` SHALL be in accent color
- **AND** unselected items show `[ ]` in muted color

### Requirement: Detail View Card Layout

The detail view SHALL use a card-based layout with visually distinct sections.

#### Scenario: Detail view structure

- **WHEN** viewing workspace `PROJ-123` details
- **THEN** the layout SHALL be:
  ```
  ┌─ Workspace: PROJ-123 ────────────────────────────────────────┐
  │                                                              │
  │   Branch      feature/user-auth                             │
  │   Disk Size   45.2 MB                                        │
  │   Modified    2 hours ago                                    │
  │   Repos       3                                              │
  │                                                              │
  ├─ Repositories ───────────────────────────────────────────────┤
  │                                                              │
  │   api         ● 2 modified  ↑ 1 unpushed                    │
  │   frontend     clean                                        │
  │   worker       clean        ↓ 3 behind                      │
  │                                                              │
  └──────────────────────────────────────────────────────────────┘
  ```

#### Scenario: Detail view with orphaned worktrees

- **GIVEN** workspace has orphaned worktrees
- **WHEN** viewing details
- **THEN** a warning banner SHALL appear:
  ```
  ┌─ ⚠ Orphaned Worktrees ───────────────────────────────────────┐
  │  2 worktrees reference missing branches                      │
  │   • old-feature (branch deleted)                             │
  │   • experiment (remote gone)                                 │
  └──────────────────────────────────────────────────────────────┘
  ```
- **AND** the banner SHALL use warning color scheme

#### Scenario: Repository status columns

- **WHEN** displaying repository list in detail view
- **THEN** columns SHALL be aligned:
  ```
  │ NAME          STATUS          SYNC STATUS              │
  │ api           ● 2 modified    ↑ 1 unpushed             │
  │ frontend       clean         (up to date)             │
  │ worker         clean         ↓ 3 behind               │
  ```

### Requirement: Modal Confirmation Dialog

Confirmation dialogs SHALL appear as centered modal overlays.

#### Scenario: Destructive action confirmation

- **WHEN** user initiates a destructive action (close, push)
- **THEN** a modal dialog SHALL appear centered on screen:
  ```
  ┌─ Confirm ─────────────────────────┐
  │                                   │
  │  ⚠ Close workspace PROJ-123?      │
  │                                   │
  │  This will delete 3 worktrees     │
  │  and free 45.2 MB of disk space.  │
  │                                   │
  │    [Y] Confirm    [N] Cancel      │
  │                                   │
  └───────────────────────────────────┘
  ```

#### Scenario: Modal styling by action type

- **WHEN** confirming a destructive action (close)
- **THEN** the modal SHALL use danger color for icon and border

- **WHEN** confirming a sync/push action
- **THEN** the modal SHALL use accent color for icon and border

#### Scenario: Modal backdrop

- **WHEN** modal is displayed
- **THEN** the background content SHALL be dimmed (reduced opacity/muted colors)
- **AND** the modal SHALL be visually prominent

### Requirement: Context-Aware Footer Help Bar

The footer SHALL display context-sensitive keyboard shortcuts.

#### Scenario: List view footer

- **WHEN** in list view with no selection
- **THEN** footer SHALL show:
  ```
  │ ↑↓ navigate • ⏎ details • / search • s stale • q quit       │
  ```

#### Scenario: List view footer with selection

- **WHEN** in list view with 3 items selected
- **THEN** footer SHALL show:
  ```
  │ ␣ select • a all • d none • P push(3) • S sync(3) • C close │
  ```

#### Scenario: Detail view footer

- **WHEN** in detail view
- **THEN** footer SHALL show:
  ```
  │ p push • S sync • o open • c close • esc back • q quit      │
  ```

#### Scenario: Footer key formatting

- **WHEN** displaying keyboard shortcuts
- **THEN** keys SHALL be formatted with visual distinction:
  - Key character in bold or accent color
  - Action name in normal text
  - Separator: " • " (bullet with spaces)

### Requirement: Semantic Color System

The TUI SHALL use a semantic color palette with consistent meaning across all views.

#### Scenario: Color definitions

- **WHEN** styling UI elements
- **THEN** colors SHALL be applied by semantic meaning:

| Semantic | Hex       | Usage Examples                          |
|----------|-----------|----------------------------------------|
| accent   | `#8B5CF6` | Cursor, selection, primary actions     |
| success  | `#22C55E` | Clean status, operation success        |
| warning  | `#F59E0B` | Stale, behind remote, needs attention  |
| danger   | `#EF4444` | Dirty, errors, destructive actions     |
| muted    | `#6B7280` | Secondary text, disabled items         |
| subtle   | `#374151` | Borders, dividers, backgrounds         |
| surface  | `#1F2937` | Panel backgrounds, cards               |
| text     | `#F9FAFB` | Primary text content                   |

#### Scenario: Color consistency

- **WHEN** displaying a "dirty" status anywhere in the UI
- **THEN** it SHALL always use the danger color (#EF4444)
- **AND** the same color SHALL be used in list items, detail view, and badges

### Requirement: Enhanced Icon Set

The TUI SHALL use Nerd Font icons with ASCII fallback for visual richness.

#### Scenario: Icon rendering with Nerd Fonts

- **GIVEN** `tui.use_emoji` is true (default)
- **WHEN** rendering icons
- **THEN** Nerd Font glyphs SHALL be used:

| Purpose     | Glyph | Description      |
|-------------|-------|------------------|
| Workspace   | ``   | Tree/folder      |
| Repository  | ``   | Git branch icon  |
| Branch      | ``   | Branch icon      |
| Dirty       | ``   | Modified circle  |
| Clean       | ``   | Checkmark        |
| Warning     | ``   | Warning triangle |
| Error       | ``   | Error X          |
| Unpushed    | ``   | Arrow up         |
| Behind      | ``   | Arrow down       |
| Stale       | ``   | Clock            |
| Disk        | ``   | Hard drive       |
| Time        | ``   | Calendar         |
| Search      | ``   | Magnifying glass |

#### Scenario: Icon rendering in ASCII mode

- **GIVEN** `tui.use_emoji` is false
- **WHEN** rendering icons
- **THEN** ASCII equivalents SHALL be used:

| Purpose     | ASCII |
|-------------|-------|
| Workspace   | `[W]` |
| Repository  | `[R]` |
| Branch      | `[B]` |
| Dirty       | `*`   |
| Clean       | `ok`  |
| Warning     | `!`   |
| Error       | `X`   |
| Unpushed    | `^`   |
| Behind      | `v`   |
| Stale       | `~`   |
| Disk        | `[D]` |
| Time        | `@`   |
| Search      | `?`   |

### Requirement: Loading and Progress States

The TUI SHALL provide clear visual feedback during async operations.

#### Scenario: Initial loading state

- **WHEN** the TUI is loading workspace list
- **THEN** a loading indicator SHALL display:
  ```
  ┌─ Content ─────────────────────────────────────────────────────┐
  │                                                               │
  │                    ⏳ Loading workspaces...                    │
  │                                                               │
  └───────────────────────────────────────────────────────────────┘
  ```

#### Scenario: Operation in progress

- **WHEN** a push/sync operation is running
- **THEN** the header SHALL show progress:
  ```
  │  Canopy    ⟳ Pushing PROJ-123...                              │
  ```
- **AND** a spinner animation SHALL indicate activity

#### Scenario: Per-item loading state

- **WHEN** fetching status for individual workspaces
- **THEN** items without status SHALL show:
  ```
     PROJ-789                                          ⏳ loading...
       - repos • - MB • -
  ```
