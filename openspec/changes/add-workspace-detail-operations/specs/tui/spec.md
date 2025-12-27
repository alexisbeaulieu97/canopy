## ADDED Requirements

### Requirement: Detail View Operations

The TUI detail view SHALL provide keyboard shortcuts for common workspace operations, allowing users to act on the currently viewed workspace without returning to the list view.

The following operations SHALL be available in detail view:
- Push all repos (configurable key, default `p`)
- Sync all repos (configurable key, default `S`)
- Open in editor (configurable key, default `o`)
- Close workspace (configurable key, default `c`)

#### Scenario: Push from detail view

- **GIVEN** I am viewing workspace `PROJ-123` in detail view
- **WHEN** I press `p`
- **THEN** a confirmation dialog SHALL appear: "Confirm push PROJ-123?"
- **AND** pressing `y` SHALL push all repos in the workspace
- **AND** after push completes, I SHALL remain in the detail view
- **AND** the detail view SHALL refresh to show updated status

Example interaction:
```
┌─ Workspace: PROJ-123 ─────────────────────────────────────┐
│ Branch: feature/auth                                      │
│ Disk:   45.2 MB                                          │
│ Modified: 2 hours ago                                     │
│                                                           │
│ Repositories (3):                                         │
│   api        ● 2 unpushed                                │
│   frontend   ✓ clean                                      │
│   worker     ● 1 unpushed                                │
└───────────────────────────────────────────────────────────┘

⚠️ Confirm push PROJ-123? [y/n]
```

#### Scenario: Sync from detail view

- **GIVEN** I am viewing workspace `PROJ-123` in detail view
- **WHEN** I press `S` (shift+s)
- **THEN** a confirmation dialog SHALL appear: "Confirm sync PROJ-123?"
- **AND** pressing `y` SHALL sync (pull) all repos in the workspace
- **AND** after sync completes, I SHALL remain in the detail view
- **AND** the detail view SHALL refresh to show updated status

#### Scenario: Open editor from detail view

- **GIVEN** I am viewing workspace `PROJ-123` in detail view
- **WHEN** I press `o`
- **THEN** the workspace SHALL open in `$VISUAL` or `$EDITOR`
- **AND** no confirmation dialog SHALL appear (non-destructive action)
- **AND** I SHALL remain in the detail view after the editor launches

#### Scenario: Close from detail view

- **GIVEN** I am viewing workspace `PROJ-123` in detail view
- **WHEN** I press `c`
- **THEN** a confirmation dialog SHALL appear: "Confirm close PROJ-123?"
- **AND** pressing `y` SHALL close the workspace
- **AND** after close completes, I SHALL return to the list view (workspace no longer exists)

Example interaction:
```
⚠️ Confirm close PROJ-123? [y/n]

(user presses y)

✓ Closed workspace PROJ-123
(returns to list view)
```

#### Scenario: Cancel operation returns to detail view

- **GIVEN** I am viewing workspace `PROJ-123` in detail view
- **AND** a confirmation dialog is active for push/sync/close
- **WHEN** I press `n` or `Esc`
- **THEN** the confirmation dialog SHALL close
- **AND** I SHALL remain in the detail view for `PROJ-123`

#### Scenario: Detail view footer shows available shortcuts

- **GIVEN** I am viewing workspace details
- **THEN** the footer SHALL display available shortcuts:
  ```
  p push • S sync • o open • c close • esc back • q quit
  ```
- **AND** keys SHALL respect user's keybinding configuration
