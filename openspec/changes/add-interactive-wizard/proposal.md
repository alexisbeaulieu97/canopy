# Change: Add Interactive Workspace Creation Wizard

## Why
The `yard workspace new` command requires users to remember flag names and syntax (--repos, --branch, --template, etc.). For new users or infrequent workspace creators, an interactive wizard provides a friendlier experience with prompts, validation, and contextual help. This leverages the existing Bubble Tea TUI infrastructure for consistency.

## What Changes
- Add `yard workspace create` command (alias for interactive mode, distinct from `new`)
- Implement multi-step wizard using Bubble Tea:
  1. Prompt for workspace ID
  2. Show available templates (optional selection)
  3. Show repository list with multi-select (from registry + patterns)
  4. Prompt for custom branch name (default: ID)
  5. Confirm and create
- Add `--wizard` flag to `yard workspace new` to enter wizard mode
- Provide keyboard shortcuts (arrows, space, enter) for navigation
- Show real-time validation (e.g., "Workspace ID already exists")

## Impact
- Affected specs: `specs/tui-interface/spec.md` (new)
- Affected code:
  - `internal/tui/wizard.go` (new) - Wizard UI component
  - `cmd/yard/workspace.go` - Add create command and --wizard flag
  - `internal/tui/tui.go` - Reuse existing TUI styles and components
