```markdown
# Implementation Tasks

## 1. Data Model
- [ ] 1.1 Create `CanonicalRepoStatus` struct (Name, Path, LastFetch, Size, Branches, UsedBy)
- [ ] 1.2 Add method to determine last fetch time from git metadata

## 2. Git Engine
- [ ] 2.1 Add `GetRepoInfo(name string)` method to gitx
- [ ] 2.2 Read last fetch time from `.git/FETCH_HEAD` mtime
- [ ] 2.3 List remote branches
- [ ] 2.4 Calculate repo size on disk

## 3. Service Layer
- [ ] 3.1 Add `GetCanonicalRepoStatus(name string) (*CanonicalRepoStatus, error)`
- [ ] 3.2 Add `ListCanonicalRepoStatuses() ([]CanonicalRepoStatus, error)`
- [ ] 3.3 Cross-reference with workspaces to find "used by"

## 4. CLI Command
- [ ] 4.1 Create `repoStatusCmd` cobra command
- [ ] 4.2 Handle optional NAME argument
- [ ] 4.3 Add `--stale` flag with configurable days threshold
- [ ] 4.4 Add `--json` flag for machine-readable output
- [ ] 4.5 Format human-readable output with colors

## 5. Testing
- [ ] 5.1 Unit test for GetCanonicalRepoStatus
- [ ] 5.2 Manual test CLI output
- [ ] 5.3 Test --json output format
```
