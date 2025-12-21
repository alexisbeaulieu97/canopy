# Change: Standardize CLI output formatting

## Why
Multiple issues with current CLI output:
1. ANSI color codes are scattered across files (repo.go, doctor.go, presenters.go)
2. JSON error output is inconsistent (PrintErrorJSON used in some commands, formatErrorJSON unused)
3. No TTY detection for color output
4. Magic numbers for separator widths (50, 20, 10)

## What Changes
- Create centralized color/style utilities with TTY detection
- Standardize JSON error output with shared presenter
- Consolidate formatting constants (separator widths, etc.)
- Replace raw ANSI codes with lipgloss styles

## Impact
- Affected specs: cli
- Affected code:
  - `internal/output/cli.go` - Add color utilities
  - `internal/output/json.go` - Standardize JSON error format
  - `cmd/canopy/presenters.go` - Use centralized styles
  - `cmd/canopy/repo.go` - Remove inline ANSI codes
  - `cmd/canopy/doctor.go` - Remove inline ANSI codes
