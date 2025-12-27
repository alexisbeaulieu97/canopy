## Context

The TUI is the primary interface for daily workspace management. Users spend significant time navigating workspaces, checking status, and performing operations. A polished visual design reduces friction and cognitive load.

## Goals

- Create a visually cohesive, modern terminal UI
- Improve information density without sacrificing readability
- Establish clear visual hierarchy
- Ensure accessibility (color contrast, ASCII fallback)

## Non-Goals

- Mouse interaction support
- Theming/customizable colors (future consideration)
- Animation effects beyond spinners

## Decisions

### Decision: Panel-Based Layout

Use lipgloss box rendering for distinct visual sections. This provides clear boundaries and allows consistent padding/margins.

**Alternatives considered:**
- Flat layout with divider lines: Less visually distinct, harder to scan
- Tab-based views: Adds complexity, not needed for current feature set

### Decision: Two-Line List Items

Reduce from three-line to two-line items for higher density. Status information moves to inline pills rather than dedicated line.

```
Current (3 lines):
❯ [x] ● PROJ-123                          STALE  2 dirty
      3 repos • 45.2 MB • 2 days ago
      ● 2 dirty  ● 1 unpushed

New (2 lines):
❯ [x] PROJ-123               ● dirty(2) ↑ unpush(1) ⏰ stale
      3 repos • 45.2 MB • 2d ago
```

### Decision: Semantic Color Palette

Define colors by semantic meaning, not visual preference:

| Semantic Name | Hex       | Usage                        |
|---------------|-----------|------------------------------|
| `accent`      | `#8B5CF6` | Selection, primary actions   |
| `success`     | `#22C55E` | Clean status, confirmations  |
| `warning`     | `#F59E0B` | Stale, behind, needs sync    |
| `danger`      | `#EF4444` | Dirty, errors, destructive   |
| `muted`       | `#6B7280` | Secondary text, disabled     |
| `subtle`      | `#374151` | Borders, dividers            |
| `surface`     | `#1F2937` | Panel backgrounds            |
| `text`        | `#F9FAFB` | Primary text                 |

### Decision: Nerd Font Icons with Fallback

Use Nerd Font icons for rich terminals, ASCII for basic terminals:

| Icon Purpose    | Nerd Font | ASCII |
|-----------------|-----------|-------|
| Workspace       | ``       | `[W]` |
| Repository      | ``       | `[R]` |
| Branch          | ``       | `[B]` |
| Dirty           | ``       | `*`   |
| Clean           | ``       | `ok`  |
| Warning         | ``       | `!`   |
| Error           | ``       | `X`   |
| Unpushed        | ``       | `^`   |
| Behind          | ``       | `v`   |
| Stale           | ``       | `~`   |
| Disk            | ``       | `[D]` |
| Time            | ``       | `@`   |

## Risks / Trade-offs

| Risk                           | Mitigation                              |
|--------------------------------|-----------------------------------------|
| Nerd Font not installed        | ASCII fallback always available         |
| Box chars broken in some terms | Test across iTerm2, Terminal.app, Alacritty |
| Higher render complexity       | Benchmark rendering, optimize if needed |

## Open Questions

- Should we support 256-color vs true-color detection?
- Should selection use background color or border accent?
