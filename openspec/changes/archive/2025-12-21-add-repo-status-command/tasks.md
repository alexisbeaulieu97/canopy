# Tasks: Add Repository Status Command

## Implementation Checklist

### 1. Domain Model
- [x] 1.1 Define `CanonicalRepoStatus` struct in `internal/domain/`:
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
- [x] 2.1 Add `LastFetchTime(repoName string) (*time.Time, error)` to `ports.GitOperations`
- [x] 2.2 Implement in `gitx/git.go` by reading `.git/FETCH_HEAD` mtime or refs
- [x] 2.3 Add `GetRepoSize(repoName string) (int64, error)` method

### 3. Service Layer
- [x] 3.1 Add `GetCanonicalRepoStatus(name string) (*domain.CanonicalRepoStatus, error)`
- [x] 3.2 Add `GetAllCanonicalRepoStatuses() ([]domain.CanonicalRepoStatus, error)`
- [x] 3.3 Calculate workspace usage by scanning workspace metadata

### 4. CLI Command
- [x] 4.1 Add `repoStatusCmd` to `cmd/canopy/repo.go`
- [x] 4.2 Handle single repo: `canopy repo status backend`
- [x] 4.3 Handle all repos: `canopy repo status`
- [x] 4.4 Add `--json` flag for machine-readable output
- [x] 4.5 Format output with table-like display

### 5. Testing
- [x] 5.1 Add unit tests for `LastFetchTime()`
- [x] 5.2 Add unit tests for `GetCanonicalRepoStatus()`
- [x] 5.3 Add integration test for `canopy repo status`
- [x] 5.4 Test `--json` output format
