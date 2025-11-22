# Implementation Tasks

## 1. Archive Infrastructure
- [x] 1.1 Add ArchivesRoot field to Config (default ~/.canopy/archives)
- [x] 1.2 Create archive directory on first archive operation
- [x] 1.3 Implement ArchiveWorkspace() in service layer
- [x] 1.4 Implement RestoreWorkspace() in service layer
- [x] 1.5 Add configurable default for workspace close behavior (archive vs delete)

## 2. Archive Operation
- [x] 2.1 Copy workspace.yaml to archive directory with timestamp
- [x] 2.2 Remove workspace worktrees from workspaces_root
- [x] 2.3 Update workspace metadata to include archive date
- [x] 2.4 Add validation to prevent archiving non-existent workspaces

## 3. Restore Operation
- [x] 3.1 Read archived workspace metadata
- [x] 3.2 Recreate workspace directory
- [x] 3.3 Recreate worktrees for all repos in metadata
- [x] 3.4 Optionally remove from archive or mark as restored

## 4. CLI Commands
- [x] 4.1 Implement `canopy workspace archive <ID>` command
- [x] 4.2 Implement `canopy workspace restore <ID>` command
- [x] 4.3 Add --archived flag to `canopy workspace list`
- [x] 4.4 Update `canopy workspace close` to suggest archiving

## 5. Testing
- [x] 5.1 Unit tests for archive/restore operations
- [x] 5.2 Integration test for full archive→restore cycle
- [x] 5.3 Test error handling (archive nonexistent, restore conflicting)
