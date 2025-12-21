# Change: Add Configuration Override Flag and Environment Variable

## Why
Currently, tests and users must manipulate HOME or XDG directories to use alternative configurations. This makes:
1. **Testing difficult**: Integration tests must hijack HOME to swap configs
2. **Automation brittle**: CI/CD pipelines can't easily use per-job configs
3. **Per-project configs impossible**: Users can't have project-specific configurations

## What Changes
- Add `--config <path>` global flag to all commands
- Add `CANOPY_CONFIG` environment variable support
- Priority order: `--config` flag > `CANOPY_CONFIG` env > default `~/.canopy/config.yaml`
- Update App initialization to respect config override
- Document the new configuration options

## Impact
- Affected specs: `cli`
- Affected code:
  - `cmd/canopy/main.go` - Add persistent global flag
  - `internal/config/config.go` - Accept config path parameter
  - `internal/app/app.go` - Pass config path to loader
  - `docs/configuration.md` - Document new options
- **No breaking changes** - Existing behavior unchanged when flag/env not set

