# Change: Refactor Service Initialization with Dependency Injection

## Why
Every command handler duplicates the same 7-line service initialization pattern across 15+ commands. This violates DRY principles, makes testing harder (can't inject mocks), and creates maintenance burden when adding new dependencies. As identified in the code review, this is a blocking architectural issue that will compound as the project grows.

## What Changes
- Create a centralized `App` struct that initializes all services once
- Refactor all command handlers to receive pre-initialized dependencies
- Move global variables (`cfg`, `logger`) from `main.go` into the `App` context
- Update all 15+ command files to use dependency injection pattern
- **BREAKING**: Internal command handler signatures change (no user-facing impact)

## Impact
- Affected specs: `specs/core-architecture/spec.md`
- Affected code:
  - `cmd/yard/main.go` - Add App struct and initialization
  - `cmd/yard/workspace.go` - Remove duplicated initialization (7 commands)
  - `cmd/yard/repo.go` - Remove duplicated initialization (5 commands)
  - `cmd/yard/tui.go` - Remove duplicated initialization
  - `cmd/yard/status.go` - Remove duplicated initialization
  - `cmd/yard/check.go` - Remove duplicated initialization
  - All command tests - Update to use App for easier mocking
