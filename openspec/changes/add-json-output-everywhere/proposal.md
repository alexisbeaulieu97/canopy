```markdown
# Change: Add JSON Output Everywhere

## Why
Currently, only `canopy workspace list` supports `--json` output. For scripting, automation, and integration with other tools, all commands that produce output should support machine-readable JSON format. This enables users to pipe canopy output to `jq`, build scripts, and integrate with CI/CD pipelines.

## What Changes
- Add `--json` flag to all commands that produce output:
  - `canopy workspace status <ID>`
  - `canopy workspace path <ID>`
  - `canopy repo list`
  - `canopy repo status`
  - `canopy template list`
  - `canopy check`
- Standardize JSON output format with consistent field naming
- Add `--json` to root command as global flag option
- Document JSON schemas in help text

## Impact
- Affected specs: `specs/cli/spec.md`
- Affected code:
  - `cmd/canopy/workspace.go` - Add --json to status, path commands
  - `cmd/canopy/repo.go` - Add --json to list command
  - `cmd/canopy/check.go` - Add --json output
  - Create shared JSON output helpers
```
