# Change: Decouple Config Validation

## Why
`internal/config/config.go:Validate()` mixes two concerns:
1. **Pure validation**: Checking values are correct (non-empty, valid regex, etc.)
2. **Filesystem checks**: Verifying paths exist and are directories

This coupling makes testing harder:
- Unit tests need real filesystem setup
- Can't test validation logic in isolation
- `validateRoot()` has side effects (checks `os.Stat`)

Separating these concerns enables:
- Fast unit tests for validation rules
- Clear distinction between "valid config" and "ready config"
- Reuse of validation in different contexts

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
