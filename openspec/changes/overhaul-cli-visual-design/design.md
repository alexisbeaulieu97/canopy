## Context

CLI output is consumed by humans (interactive use) and machines (scripts with `--json`). The visual overhaul targets human-readable output while preserving JSON for automation.

## Goals

- Make CLI output visually appealing and easy to scan
- Establish consistent formatting across all commands
- Provide clear feedback for operations in progress
- Format errors with actionable context

## Non-Goals

- Changing JSON output structure (already standardized)
- Adding interactive prompts (TUI handles that)
- Supporting Windows cmd.exe special formatting

## Decisions

### Decision: Box-Drawn Tables

Use Unicode box-drawing characters for tables, with detection for terminals that don't support them.

**Implementation:**
```go
type TableStyle struct {
    TopLeft, TopRight, BottomLeft, BottomRight string
    Horizontal, Vertical                        string
    LeftT, RightT, TopT, BottomT, Cross        string
}

var UnicodeTable = TableStyle{
    TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
    Horizontal: "─", Vertical: "│",
    LeftT: "├", RightT: "┤", TopT: "┬", BottomT: "┴", Cross: "┼",
}

var ASCIITable = TableStyle{
    TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+",
    Horizontal: "-", Vertical: "|",
    LeftT: "+", RightT: "+", TopT: "+", BottomT: "+", Cross: "+",
}
```

**Alternatives considered:**
- Plain text columns with spacing: Less visually distinct
- Markdown tables: Render poorly without viewer

### Decision: Semantic Status Icons

Use consistent icons across all commands:

| Status     | Icon | Color   | Fallback |
|------------|------|---------|----------|
| Success    | `✓`  | Green   | `[ok]`   |
| Warning    | `⚠`  | Amber   | `[!]`    |
| Error      | `✗`  | Red     | `[X]`    |
| Info       | `ℹ`  | Blue    | `[i]`    |
| Loading    | `⠋`  | Cyan    | `...`    |
| Dirty      | `●`  | Red     | `*`      |
| Clean      | `○`  | Green   | `-`      |

### Decision: Color Detection

Check for color support in order:
1. `NO_COLOR` env var set → no colors
2. `CANOPY_COLOR=always` → force colors
3. `CANOPY_COLOR=never` → no colors
4. stdout is not TTY → no colors
5. `TERM=dumb` → no colors
6. Otherwise → colors enabled

### Decision: Error Box Format

Errors display in a bordered box with:
- Error type header
- Error message
- Context (workspace ID, path, etc.)
- Suggestion (if applicable)

```
┌─ Error ──────────────────────────────────────────────────────┐
│                                                              │
│  Workspace not found: PROJ-999                               │
│                                                              │
│  The workspace 'PROJ-999' does not exist.                    │
│                                                              │
│  Available workspaces:                                       │
│    • PROJ-123                                                │
│    • PROJ-456                                                │
│    • PROJ-789                                                │
│                                                              │
│  Did you mean: PROJ-999 → PROJ-789?                          │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## Risks / Trade-offs

| Risk                          | Mitigation                               |
|-------------------------------|------------------------------------------|
| Box chars broken on Windows   | ASCII fallback, test on Windows Terminal |
| Colors in piped output        | Detect TTY, respect NO_COLOR             |
| Slower output rendering       | Benchmark, buffer output                 |
| Breaking scripts parsing text | Document that text output may change     |

## Open Questions

- Should we add `--no-table` flag for plain output?
- Should progress spinners be configurable (on/off)?
