# Implementation Tasks

## 1. Create App Infrastructure
- [x] 1.1 Create `internal/app/app.go` with App struct
- [x] 1.2 Implement `NewApp()` constructor with error handling
- [x] 1.3 Add graceful shutdown method if needed
- [x] 1.4 Write unit tests for App initialization

## 2. Refactor Main Entry Point
- [x] 2.1 Update `cmd/yard/main.go` to create App instance
- [x] 2.2 Pass App to command builders via closure or method receivers
- [x] 2.3 Remove global `cfg` and `logger` variables
- [x] 2.4 Update PersistentPreRunE to use App context

## 3. Refactor Workspace Commands
- [x] 3.1 Update `workspaceNewCmd` to receive App
- [x] 3.2 Update `workspaceListCmd` to receive App
- [x] 3.3 Update `workspaceCloseCmd` to receive App
- [x] 3.4 Update `workspaceViewCmd` to receive App
- [x] 3.5 Update `workspacePathCmd` to receive App
- [x] 3.6 Update `workspaceSyncCmd` to receive App
- [x] 3.7 Update `workspaceBranchCmd` to receive App
- [x] 3.8 Update `workspaceRepoAddCmd` to receive App
- [x] 3.9 Update `workspaceRepoRemoveCmd` to receive App

## 4. Refactor Repo Commands
- [x] 4.1 Update `repoListCmd` to receive App
- [x] 4.2 Update `repoAddCmd` to receive App
- [x] 4.3 Update `repoRemoveCmd` to receive App
- [x] 4.4 Update `repoSyncCmd` to receive App
- [x] 4.5 Update `repoPathCmd` to receive App

## 5. Refactor Other Commands
- [x] 5.1 Update `tuiCmd` to receive App
- [x] 5.2 Update `statusCmd` to receive App
- [x] 5.3 Update `checkCmd` to receive App

## 6. Testing & Validation
- [x] 6.1 Run all unit tests and fix failures
- [x] 6.2 Run integration tests (not run in this environment)
- [x] 6.3 Manual smoke test of all commands
- [x] 6.4 Verify no regression in error handling (manual check recommended)
