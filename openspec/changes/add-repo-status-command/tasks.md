# Tasks: Add Repository Status Command

## Implementation Checklist

### 1. Domain Model
- [ ] 1.1 Define `CanonicalRepoStatus` struct in `internal/domain/`:
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

### 2. Git Engine Extension
- [ ] 2.1 Add `LastFetchTime(repoName string) (*time.Time, error)` to `ports.GitOperations`
- [ ] 2.2 Implement in `gitx/git.go` by reading `.git/FETCH_HEAD` mtime or refs
- [ ] 2.3 Add `GetRepoSize(repoName string) (int64, error)` method

### 3. Service Layer
- [ ] 3.1 Add `GetCanonicalRepoStatus(name string) (*domain.CanonicalRepoStatus, error)`
- [ ] 3.2 Add `GetAllCanonicalRepoStatuses() ([]domain.CanonicalRepoStatus, error)`
- [ ] 3.3 Calculate workspace usage by scanning workspace metadata

### 4. CLI Command
- [ ] 4.1 Add `repoStatusCmd` to `cmd/canopy/repo.go`
- [ ] 4.2 Handle single repo: `canopy repo status backend`
- [ ] 4.3 Handle all repos: `canopy repo status`
- [ ] 4.4 Add `--json` flag for machine-readable output
- [ ] 4.5 Format output with table-like display

### 5. Testing
- [ ] 5.1 Add unit tests for `LastFetchTime()`
- [ ] 5.2 Add unit tests for `GetCanonicalRepoStatus()`
- [ ] 5.3 Add integration test for `canopy repo status`
- [ ] 5.4 Test `--json` output format
