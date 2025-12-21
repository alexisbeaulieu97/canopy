# Change: Rename internal/workspace package to internal/storage

## Why
The current naming `internal/workspace` (singular) vs `internal/workspaces` (plural) is confusing. It's unclear which package contains the entity, the service, or the storage. Renaming `workspace` to `storage` clarifies that this package implements the `WorkspaceStorage` interface (filesystem persistence adapter).

## What Changes
- Rename `internal/workspace/` directory to `internal/storage/`
- Update all import statements referencing the old package path
- Update `project.md` architecture documentation to reflect new package name

## Impact
- Affected specs: core-architecture (documentation update)
- Affected code:
  - `internal/workspace/*.go` (4 files to rename/move)
  - `internal/app/app.go` (import update)
  - `internal/workspaces/service_test.go` (import update)
  - `openspec/project.md` (documentation update)
- No breaking changes to public API or behavior
