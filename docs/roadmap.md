# Canopy Roadmap

This roadmap organizes approved OpenSpec changes into implementation phases based on dependencies, risk, and value.

---

## Phase 1: Foundation Refactoring
**Goal**: Improve code quality and testability before adding features.

**Rationale**: These refactors reduce technical debt and make subsequent feature work cleaner. They have no user-facing changes but improve maintainability.

### 1.1 Dependency Injection (`refactor-inject-dependencies`)
- **Effort**: Small (1-2 days)
- **Risk**: Low
- **Why First**: Enables proper unit testing for all subsequent work. Other refactors and features benefit from mockable dependencies.

### 1.2 Config Validation (`refactor-config-validation`)
- **Effort**: Small (1 day)
- **Risk**: Low
- **Why Now**: Quick win that improves testability. No dependencies on other work.

### 1.3 Error Consistency (`refactor-error-consistency`)
- **Effort**: Medium (2-3 days)
- **Risk**: Low
- **Why Now**: Establishes error patterns needed by features (dry-run, orphan detection). Better to standardize before adding more commands.

---

## Phase 2: Core Infrastructure
**Goal**: Build foundational capabilities that multiple features depend on.

### 2.1 JSON Output Everywhere (`add-json-output-everywhere`)
- **Effort**: Medium (2-3 days)
- **Risk**: Low
- **Why Now**: Dry-run mode depends on JSON output schema. Shell completions and scripting benefit from consistent JSON. Do this before features that need `--json`.
- **Depends on**: Phase 1 error consistency (JSON errors should be typed)

### 2.2 Orphan Detection (`add-orphan-detection`)
- **Effort**: Medium (2-3 days)
- **Risk**: Low
- **Why Now**: Dry-run mode references orphan detection for `repo remove` warnings. Detection logic is also needed by TUI indicators.
- **Depends on**: Error consistency (orphan errors should be typed)

### 2.3 Dry Run Mode (`add-dry-run-mode`)
- **Effort**: Small (1-2 days)
- **Risk**: Very Low
- **Why Now**: Safety feature that should exist before users do destructive operations. Uses JSON output and orphan detection.
- **Depends on**: JSON output, Orphan detection

---

## Phase 3: Service Layer Improvements
**Goal**: Split large services to improve maintainability for upcoming features.

### 3.1 Service Splitting (`refactor-service-splitting`)
- **Effort**: Medium (3-4 days)
- **Risk**: Medium
- **Why Now**: Workspace export/import, hooks, and rename all touch the service layer. Splitting first makes those features cleaner to implement.
- **Depends on**: Dependency injection (new services need proper wiring)

### 3.2 Git Abstraction (`refactor-git-abstraction`)
- **Effort**: Medium (3-4 days)  
- **Risk**: Medium
- **Why Now**: Eliminates CLI shelling. Hooks and export/import will add more git operations; consolidate patterns first.
- **Depends on**: None, but benefits from error consistency

---

## Phase 4: Workspace Features
**Goal**: Add high-value workspace management capabilities.

### 4.1 Workspace Rename (`add-workspace-rename`)
- **Effort**: Small (1-2 days)
- **Risk**: Low
- **Why First in Phase**: Simple, isolated feature. Good warmup for workspace service changes.
- **Depends on**: Service splitting (cleaner to add to focused service)

### 4.2 Workspace Templates (`add-workspace-templates`)
- **Effort**: Medium (2-3 days)
- **Risk**: Low
- **Why Now**: Frequently requested feature. Config parsing patterns established.
- **Depends on**: Config validation (template config should be validated)

### 4.3 Workspace Export/Import (`add-workspace-export`)
- **Effort**: Medium (3-4 days)
- **Risk**: Low
- **Why Now**: Enables multi-machine workflows. JSON output patterns established.
- **Depends on**: JSON output, Service splitting

### 4.4 Lifecycle Hooks (`add-hooks-support`)
- **Effort**: Medium (3-4 days)
- **Risk**: Medium (executes arbitrary commands)
- **Why Last in Phase**: Most complex workspace feature. Service layer and error handling should be solid.
- **Depends on**: Service splitting, Error consistency, Config validation

---

## Phase 5: Repository Features
**Goal**: Improve canonical repository management.

### 5.1 Repo Status Command (`add-repo-status-command`)
- **Effort**: Small (1-2 days)
- **Risk**: Low
- **Why Now**: Provides visibility into canonical repos. Uses established JSON output patterns.
- **Depends on**: JSON output, Git abstraction (LastFetchTime)

---

## Phase 6: User Experience
**Goal**: Polish CLI and TUI experience.

### 6.1 Shell Completions (`add-shell-completions`)
- **Effort**: Small (1 day)
- **Risk**: Very Low
- **Why Now**: Quick win using Cobra built-ins. All commands should be stable before adding completions.
- **Depends on**: All CLI commands finalized

### 6.2 TUI Component Extraction (`refactor-tui-extract-reusable-components`)
- **Effort**: Medium (2-3 days)
- **Risk**: Medium (UI changes)
- **Why Now**: Prepares TUI for keyboard customization. Orphan indicators already added.
- **Depends on**: Orphan detection (TUI indicators)

### 6.3 TUI Keyboard Customization (`add-tui-keyboard-customization`)
- **Effort**: Small (1-2 days)
- **Risk**: Low
- **Why Last**: Requires stable TUI components and keybinding patterns.
- **Depends on**: TUI component extraction, Config validation

---

## Summary Timeline

| Phase | Changes | Est. Effort | Cumulative |
|-------|---------|-------------|------------|
| 1. Foundation | 3 refactors | 4-6 days | Week 1 |
| 2. Infrastructure | 3 features | 5-8 days | Week 2-3 |
| 3. Service Layer | 2 refactors | 6-8 days | Week 3-4 |
| 4. Workspace | 4 features | 9-13 days | Week 5-7 |
| 5. Repository | 1 feature | 1-2 days | Week 7 |
| 6. UX Polish | 3 changes | 4-6 days | Week 8 |

**Total**: ~16 changes over 8-10 weeks

---

## Dependency Graph

```
Phase 1 (Foundation)
├── refactor-inject-dependencies
├── refactor-config-validation  
└── refactor-error-consistency
         │
         ▼
Phase 2 (Infrastructure)
├── add-json-output-everywhere ◄── (error consistency)
├── add-orphan-detection ◄── (error consistency)
└── add-dry-run-mode ◄── (json output, orphan detection)
         │
         ▼
Phase 3 (Service Layer)
├── refactor-service-splitting ◄── (dependency injection)
└── refactor-git-abstraction
         │
         ▼
Phase 4 (Workspace Features)
├── add-workspace-rename ◄── (service splitting)
├── add-workspace-templates ◄── (config validation)
├── add-workspace-export ◄── (json output, service splitting)
└── add-hooks-support ◄── (service splitting, error consistency, config validation)
         │
         ▼
Phase 5 (Repository)
└── add-repo-status-command ◄── (json output, git abstraction)
         │
         ▼
Phase 6 (UX Polish)
├── add-shell-completions ◄── (all CLI commands)
├── refactor-tui-extract-reusable-components ◄── (orphan detection)
└── add-tui-keyboard-customization ◄── (tui components, config validation)
```

---

## Implementation Notes

### Parallel Work Opportunities
- Phase 1 changes are independent; can be done in parallel
- `refactor-git-abstraction` can run parallel to Phase 2
- `add-repo-status-command` can start once git abstraction is done

### Risk Mitigation
- **Medium-risk items** (service splitting, git abstraction, hooks): Add extra testing, review before merge
- **Hooks security**: Follow design.md threat model; test timeout and failure modes

### Testing Strategy
- Each phase should have full test coverage before proceeding
- Integration tests especially important for Phases 3-4
- Visual testing for Phase 6 TUI changes
