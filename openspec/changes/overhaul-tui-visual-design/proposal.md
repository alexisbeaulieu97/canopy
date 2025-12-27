# Change: Overhaul TUI Visual Design

## Why

The current TUI is functional but visually basic. A refined visual design improves usability, reduces cognitive load, and makes the tool more pleasant to use daily. Modern terminal applications (lazygit, k9s, bottom) demonstrate that TUIs can be both functional and beautiful.

## What Changes

- **Layout**: Introduce panel-based design with box borders and consistent spacing
- **Header**: Redesigned header with logo, status summary, and breadcrumb navigation
- **List Items**: Two-line compact layout with inline status indicators
- **Color Palette**: Refined colors with semantic meaning and proper contrast
- **Typography**: Consistent use of bold, dim, and color for visual hierarchy
- **Detail View**: Card-based layout with clear sections and visual groupings
- **Confirmations**: Modal-style dialogs with proper styling
- **Footer**: Context-aware help bar with key legends

## Impact

- Affected specs: `tui`
- Affected code:
  - `internal/tui/view.go` (all rendering functions)
  - `internal/tui/components/` (all component files)
  - `internal/tui/styles.go` (style definitions)
  - `internal/tui/symbols.go` (icon updates)
