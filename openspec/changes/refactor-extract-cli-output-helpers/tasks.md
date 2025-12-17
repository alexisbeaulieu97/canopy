## 1. Create Output Helpers

- [x] 1.1 Create `internal/output/cli.go` with output helper functions
- [x] 1.2 Add `Success(action, target string)` for success messages
- [x] 1.3 Add `SuccessWithPath(action, target, path string)` for path-aware messages
- [x] 1.4 Add `Info(message string)` for neutral information
- [x] 1.5 Add `Warn(message string)` for warning messages
- [x] 1.6 Add unit tests for output helpers

## 2. Update Workspace Commands

- [x] 2.1 Update `workspaceNewCmd` to use output helpers
- [x] 2.2 Update `workspaceCloseCmd` to use output helpers
- [x] 2.3 Update `workspaceReopenCmd` to use output helpers
- [x] 2.4 Update `workspaceRenameCmd` to use output helpers
- [x] 2.5 Update `workspaceListCmd` to use output helpers
- [x] 2.6 Update `workspaceViewCmd` to use output helpers
- [x] 2.7 Update `workspaceBranchCmd` to use output helpers
- [x] 2.8 Update `workspaceGitCmd` to use output helpers
- [x] 2.9 Update `workspaceExportCmd` and `workspaceImportCmd` to use output helpers
- [x] 2.10 Update `workspaceRepoAddCmd` and `workspaceRepoRemoveCmd` to use output helpers

## 3. Update Repo Commands

- [x] 3.1 Update `repoAddCmd` to use output helpers
- [x] 3.2 Update `repoRemoveCmd` to use output helpers
- [x] 3.3 Update `repoSyncCmd` to use output helpers
- [x] 3.4 Update `repoRegisterCmd` and `repoUnregisterCmd` to use output helpers
- [x] 3.5 Update `repoListCmd` and `repoListRegistryCmd` to use output helpers
- [x] 3.6 Update `repoShowCmd` to use output helpers

## 4. Verification

- [x] 4.1 Run existing integration tests to verify output compatibility
- [x] 4.2 Manual verification of CLI output consistency
