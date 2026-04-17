## 1. OpenSpec

- [x] 1.1 Add CLI spec deltas for guarded creation, quieter runtime failures, and richer list/view output
- [x] 1.2 Add TUI spec deltas for richer search and explicit bulk-operation feedback

## 2. CLI behavior

- [x] 2.1 Block `workspace new` when no repositories resolve in the normal create path
- [x] 2.2 Make `init` output a concrete next-step checklist and ship example-oriented starter config
- [x] 2.3 Suppress Cobra usage output for runtime/domain failures after parsing
- [x] 2.4 Improve `status`, `workspace list --status`, and `workspace view` human-readable output

## 3. TUI behavior

- [x] 3.1 Expand TUI search to repo names, branch names, and status keywords
- [x] 3.2 Report partial-success bulk action summaries with failed workspace IDs
- [x] 3.3 Clear stale error banners after later successful reloads and operations
- [x] 3.4 Make zero-repo detail states actionable

## 4. Validation

- [x] 4.1 Add or update Go tests for the changed CLI/TUI behavior
- [x] 4.2 Update user-facing docs and examples to match the new flow
- [x] 4.3 Validate the OpenSpec change
