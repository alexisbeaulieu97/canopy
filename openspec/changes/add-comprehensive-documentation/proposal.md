# Change: Add Comprehensive Documentation

## Why
Documentation is incomplete for v1.0 release. Users and developers need clear guides, API documentation, and error code references.

## What Changes
- Add godoc comments to all exported functions in key packages
- Create `CHANGELOG.md` documenting changes across versions
- Create `CONTRIBUTING.md` with development setup and guidelines
- Document all error codes for scripting/automation use
- Add architecture overview for developers
- Update README with complete feature documentation

## Impact
- Affected specs: cli (error codes documentation)
- Affected code:
  - All exported functions in internal packages (godoc)
  - New files: `CHANGELOG.md`, `CONTRIBUTING.md`
  - `docs/error-codes.md` (new)
  - `docs/architecture.md` (new)
  - `README.md` (updates)

