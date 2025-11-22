# Implementation Tasks

## 1. Configuration
- [ ] 1.1 Add `workspace.list.include` config field (default: `["*"]`)
- [ ] 1.2 Add `workspace.list.exclude` config field (default: `[]`)
- [ ] 1.3 Document config options in config struct

## 2. Status Detection
- [ ] 2.1 Add `IsStale()` method to workspace (requires last-modified tracking)
- [ ] 2.2 Add `IsBehindRemote()` method to workspace status
- [ ] 2.3 Add `IsClean()` helper (not dirty and not behind)

## 3. CLI Flags
- [ ] 3.1 Add `--include` flag to `workspace list` command
- [ ] 3.2 Add `--exclude` flag to `workspace list` command
- [ ] 3.3 Parse comma-separated values into slice
- [ ] 3.4 Validate filter values (error on unknown values)

## 4. Filtering Logic
- [ ] 4.1 Implement `FilterWorkspaces(workspaces, include, exclude)` function
- [ ] 4.2 Apply include filter first (keep only matching)
- [ ] 4.3 Apply exclude filter second (remove matching)
- [ ] 4.4 Handle `*` wildcard for include-all

## 5. Integration
- [ ] 5.1 Merge CLI flags with config defaults (CLI takes precedence)
- [ ] 5.2 Apply filtering in `workspace list` command
- [ ] 5.3 Display active filters in output header (when filtering)

## 6. Testing
- [ ] 6.1 Test include single filter
- [ ] 6.2 Test exclude single filter
- [ ] 6.3 Test combined include + exclude
- [ ] 6.4 Test config defaults
- [ ] 6.5 Test CLI override of config
