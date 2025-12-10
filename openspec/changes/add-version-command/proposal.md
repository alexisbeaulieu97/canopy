# Change: Add Version Command

## Why
Users have no way to check which version of Canopy they're running. This makes bug reports difficult to triage and prevents users from knowing if they need to upgrade. The version command is a standard CLI feature.

## What Changes
- Add `canopy version` command
- Add `--version` flag to root command
- Embed version information from git tags at build time
- Include Go version and build date in output
- Support `--json` output for scripting

## Impact
- **Affected specs**: `specs/cli/spec.md`
- **Affected code**:
  - `cmd/canopy/version.go` - New file
  - `cmd/canopy/main.go` - Add version flag
  - `Makefile` or build script - Add ldflags for version embedding
- **Risk**: Very Low - Additive feature, no behavioral changes

