# Change: Add Doctor Command

## Why
Users have no way to diagnose environment issues or validate their Canopy setup. When things go wrong (git auth fails, directories have wrong permissions, config is invalid), users must manually investigate. A `canopy doctor` command would proactively identify common issues and provide actionable guidance.

## What Changes
- Add `canopy doctor` command that validates the environment and configuration
- Check git installation and version compatibility
- Validate configuration files (`~/.canopy/config.yaml`, `registry.yaml`)
- Verify directory permissions for `projects_root`, `workspaces_root`, `closed_root`
- Check health of canonical repositories (existence, fetchability)
- Report issues with severity levels (error, warning, info)
- Support `--json` flag for scripted health checks
- Support `--fix` flag to auto-remediate simple issues (create missing directories)

## Impact
- Affected specs: `specs/cli/spec.md`
- Affected code:
  - `cmd/canopy/doctor.go` (new) - Command implementation
  - `internal/workspaces/service.go` - Add health check methods
  - `internal/config/config.go` - Add validation helpers for doctor
- New dependencies: None
