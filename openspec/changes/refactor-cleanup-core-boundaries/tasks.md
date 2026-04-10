## 1. Ports And Boundary Cleanup
- [x] 1.1 Add port-owned config/registry/template DTOs and interfaces under `internal/ports`
- [x] 1.2 Update `internal/config` to convert outward-facing values to the port-owned types
- [x] 1.3 Update `internal/workspaces`, `internal/hooks`, `internal/tui`, and `cmd/canopy` to consume the new port-owned types
- [x] 1.4 Move hook-adapter construction to `internal/app`
- [x] 1.5 Route template setup commands through the injected hook executor instead of direct shell execution in `internal/workspaces`

## 2. Dead Code Removal
- [x] 2.1 Remove unused output abstractions from `internal/output`
- [x] 2.2 Remove unused TUI component abstractions from `internal/tui/components`
- [x] 2.3 Remove service bulk APIs that are only retained by tests
- [x] 2.4 Delete or update tests that only cover removed dead code

## 3. Deduplication And Flattening
- [x] 3.1 Extract shared internal formatting helpers for bytes and relative time
- [x] 3.2 Replace duplicated formatting logic in CLI, output, and TUI packages
- [x] 3.3 Remove thin TUI wrapper layers that only alias `components`
- [x] 3.4 Keep only actively used TUI component primitives

## 4. File Decomposition
- [x] 4.1 Split `internal/config/config.go` by concern
- [x] 4.2 Split `internal/output/progress.go` by concern
- [x] 4.3 Move shared bulk command helpers and presenters into dedicated `cmd/canopy` helpers
- [x] 4.4 Keep CLI behavior, JSON shapes, prompts, and exit handling unchanged

## 5. Verification
- [x] 5.1 Add focused tests for shared formatting helpers
- [x] 5.2 Run `make test` after each concern-level slice and once at the end
- [x] 5.3 Run `golangci-lint run ./...` with a repo-compatible toolchain or document the environment blocker
