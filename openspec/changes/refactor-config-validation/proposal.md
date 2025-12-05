# Change: Decouple Config Validation

## Why
`config.Validate()` mixes pure validation (checking values) with filesystem checks (verifying paths exist), making unit tests require real filesystem setup. Separating these concerns enables fast validation tests and clear distinction between "valid config" and "ready config".

## What Changes
- Split `Validate()` into `ValidateValues()` (pure) and `ValidateEnvironment()` (filesystem)
- `ValidateValues()` checks: non-empty fields, valid regex, valid enum values
- `ValidateEnvironment()` checks: paths exist, are directories, are writable
- Existing `Validate()` calls both for backward compatibility

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/config/config.go` - Split validation methods
  - `internal/config/config_test.go` - Add pure validation tests
  - `internal/app/app.go` - No changes (uses existing Validate)
- **Risk**: Low - Internal refactoring, API unchanged
