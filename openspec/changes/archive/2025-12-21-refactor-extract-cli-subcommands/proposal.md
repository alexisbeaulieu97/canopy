# Change: Extract CLI Business Logic to Service Layer

## Why
The `cmd/canopy/workspace.go` file is 1,176 lines containing significant business orchestration logic mixed with CLI concerns. This violates separation of concerns:
- Flag processing and validation mixed with business logic
- Result formatting interleaved with domain operations
- Business-critical logic in `RunE` blocks is not reusable by other interfaces (TUI, future API)
- Testing requires CLI infrastructure rather than pure unit tests

## What Changes
- Extract workspace subcommands into separate files (`workspace_new.go`, `workspace_list.go`, etc.)
- Move business orchestration from CLI layer to service layer
- Create "presenter" helpers for output formatting
- Keep CLI layer focused on: flag parsing, user I/O, calling services, formatting output

**Proposed Structure:**
```text
cmd/canopy/
├── workspace.go           # Parent command, shared flags (~50 lines)
├── workspace_new.go       # new subcommand (~100 lines)
├── workspace_list.go      # list subcommand (~150 lines)
├── workspace_close.go     # close subcommand (~100 lines)
├── workspace_view.go      # view subcommand (~80 lines)
├── workspace_rename.go    # rename subcommand (~80 lines)
├── workspace_export.go    # export/import subcommands (~100 lines)
├── workspace_hooks.go     # hooks-related subcommands (~60 lines)
└── presenters.go          # Output formatting helpers (~100 lines)
```

## Impact
- Affected specs: `specs/core-architecture/spec.md` (clarify layer responsibilities)
- Affected code:
  - `cmd/canopy/workspace.go` (split into multiple files)
  - New files as listed above
- No breaking changes - all CLI behavior preserved
- Better testability and maintainability
