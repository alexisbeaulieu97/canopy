## Context

This cleanup is intentionally internal-only. Public CLI commands, flags, JSON envelopes, exit codes, and TUI workflows stay the same. The design work is focused on removing structural friction that has accumulated from additive development.

The main architectural issue is that the ports layer leaks concrete `internal/config` types into the core. Secondary issues are abandoned abstractions and duplicated formatting/orchestration logic.

## Goals

- Restore the intended dependency direction around `internal/ports`
- Remove production-unused code that increases maintenance cost
- Consolidate duplicated logic without changing behavior
- Keep each cleanup slice small enough to verify independently

## Non-Goals

- No feature work
- No config schema changes
- No user-visible CLI or TUI redesign
- No new dependencies

## Decisions

### Port-owned config DTOs

`internal/ports` will own the lightweight types and interfaces consumed by the service/core layer:

- registry entry and registry interface
- hooks config and hook spec
- keybindings
- workspace template
- parsed git retry config

`internal/config` remains the YAML loader and validator, but callers outside that package consume only the port-owned types.

### Template setup execution

Template setup commands are operationally similar to post-create hooks. Instead of letting `internal/workspaces` shell out directly, workspace creation will translate template commands into hook specs and execute them through the injected hook executor. This keeps command execution policy in the adapter layer.

### Cleanup priority

1. boundary fixes
2. dead-code deletion
3. formatting deduplication
4. TUI flattening
5. command decomposition

This order reduces the risk of reworking the same files twice.

### Verification strategy

Each concern-level change must keep `make test` green before moving to the next slice. Linting is still required, but the current environment may need a newer linter build than the one available in the workspace.
