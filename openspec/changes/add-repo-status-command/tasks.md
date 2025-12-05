# Tasks: Add Repository Status Command

## Implementation Checklist

### Phase 1: Domain Model
- [ ] Define `CanonicalRepoStatus` struct in `internal/domain/`:
  ```go
  type CanonicalRepoStatus struct {
      Name           string
      Path           string
      DiskUsageBytes int64
      LastFetchTime  *time.Time
      UsedByCount    int
      UsedBy         []string
  }
  ```

### Phase 2: Git Engine Extension
- [ ] Add `LastFetchTime(repoName string) (*time.Time, error)` to `ports.GitOperations`
- [ ] Implement in `gitx/git.go` by reading `.git/FETCH_HEAD` mtime or refs
- [ ] Add `GetRepoSize(repoName string) (int64, error)` method

### Phase 3: Service Layer
- [ ] Add `GetCanonicalRepoStatus(name string) (*domain.CanonicalRepoStatus, error)`
- [ ] Add `GetAllCanonicalRepoStatuses() ([]domain.CanonicalRepoStatus, error)`
- [ ] Calculate workspace usage by scanning workspace metadata

### Phase 4: CLI Command
- [ ] Add `repoStatusCmd` to `cmd/canopy/repo.go`
- [ ] Handle single repo: `canopy repo status backend`
- [ ] Handle all repos: `canopy repo status`
- [ ] Add `--json` flag for machine-readable output
- [ ] Format output with table-like display

### Phase 5: Testing
- [ ] Add unit tests for `LastFetchTime()`
- [ ] Add unit tests for `GetCanonicalRepoStatus()`
- [ ] Add integration test for `canopy repo status`
- [ ] Test `--json` output format
