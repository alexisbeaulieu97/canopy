# Workspace Management Spec (Delta)

## Purpose
Workspace lifecycle: create, delete, list, and rename workspaces.

## New Capabilities

### Rename Workspace
- Command: `canopy workspace rename <old-id> <new-id>`
- Validates new ID format and uniqueness
- Atomically renames workspace directory
- Updates workspace.yaml with new ID
- Optional: rename branches in all repos with `--rename-branches`

### Behavior
- Fails if old workspace doesn't exist
- Fails if new ID conflicts with existing workspace
- Shows confirmation before renaming (bypass with `--force`)
- Rolls back on failure
