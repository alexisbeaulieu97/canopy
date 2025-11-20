# Implementation Tasks

## 1. Archive Infrastructure
- [ ] 1.1 Add ArchivesRoot field to Config (default ~/.yard/archives)
- [ ] 1.2 Create archive directory on first archive operation
- [ ] 1.3 Implement ArchiveWorkspace() in service layer
- [ ] 1.4 Implement RestoreWorkspace() in service layer

## 2. Archive Operation
- [ ] 2.1 Copy workspace.yaml to archive directory with timestamp
- [ ] 2.2 Remove workspace worktrees from workspaces_root
- [ ] 2.3 Update workspace metadata to include archive date
- [ ] 2.4 Add validation to prevent archiving non-existent workspaces

## 3. Restore Operation
- [ ] 3.1 Read archived workspace metadata
- [ ] 3.2 Recreate workspace directory
- [ ] 3.3 Recreate worktrees for all repos in metadata
- [ ] 3.4 Optionally remove from archive or mark as restored

## 4. CLI Commands
- [ ] 4.1 Implement `yard workspace archive <ID>` command
- [ ] 4.2 Implement `yard workspace restore <ID>` command
- [ ] 4.3 Add --archived flag to `yard workspace list`
- [ ] 4.4 Update `yard workspace close` to suggest archiving

## 5. Testing
- [ ] 5.1 Unit tests for archive/restore operations
- [ ] 5.2 Integration test for full archive→restore cycle
- [ ] 5.3 Test error handling (archive nonexistent, restore conflicting)
