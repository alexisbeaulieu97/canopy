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

## Risks / Trade-offs

- Moving runtime-facing types into `internal/ports` improves dependency direction, but it adds a conversion seam in `internal/config` that must stay aligned with config parsing.
- Reusing the hook executor for template setup removes duplicated shell policy, but it also makes template setup behavior depend on hook execution semantics and context propagation.
- Deleting thin wrapper layers simplifies maintenance, but any overlooked dependency would surface only through broader CLI and TUI verification.
- The current local lint toolchain may lag the repo Go version, so verification can be partially blocked even when code and tests are correct.

## Migration Plan

1. Fix the ports/config boundary so core packages depend on `internal/ports` contracts instead of concrete config types.
2. Remove dead code and obsolete tests once the new boundaries compile cleanly.
3. Deduplicate formatting helpers and update CLI/TUI callers to share the same implementation while preserving output strings.
4. Flatten TUI wrappers and keep only the component primitives used by the production TUI flow.
5. Decompose bulky command and output files, gating each slice with `make test` and a lint pass when a compatible linter binary is available.

## Open Questions

- Whether hook timeout semantics should eventually move from integer seconds to `time.Duration` in the runtime port types.
- Whether the project should explicitly standardize on `gofumpt` in addition to `gofmt`, since lint currently enforces `gofumpt` formatting.
- Whether the no-op default hook executor should remain the long-term testability strategy or be made explicit at every `Service` construction site.

## Alternatives Considered

### Keep runtime DTOs in `internal/config`

Rejected because it preserves the original boundary leak: core packages still need concrete adapter types to operate.

### Continue shelling out directly from `internal/workspaces`

Rejected because it duplicates command execution policy and keeps shell selection/timeouts inside the core package.

### Introduce a brand-new shared orchestration package for bulk commands

Rejected because the cleanup goal is to reduce concepts, not add another abstraction layer beyond the existing command package.
