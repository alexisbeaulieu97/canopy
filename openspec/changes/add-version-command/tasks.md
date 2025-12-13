# Tasks: Add Version Command

## Implementation Checklist

### 1. Version Variables
- [x] 1.1 Create `cmd/canopy/version.go` with version variables:
  ```go
  var (
      version   = "dev"
      commit    = "unknown"
      buildDate = "unknown"
  )
  ```
- [x] 1.2 Add Go version detection using `runtime.Version()`

### 2. Version Command
- [x] 2.1 Create `versionCmd` cobra command
- [x] 2.2 Implement text output format:
  ```
  canopy version v1.2.3
  commit: abc1234
  built: 2025-01-15T10:30:00Z
  go: go1.24.0
  ```
- [x] 2.3 Implement `--json` output format
- [x] 2.4 Register command with root

### 3. Version Flag
- [ ] 3.1 Add `--version` persistent flag to root command
- [x] 3.2 Print version and exit when flag is set

### 4. Build Integration
- [x] 4.1 Create/update Makefile with ldflags:
  ```makefile
  VERSION := $(shell git describe --tags --always --dirty)
  COMMIT := $(shell git rev-parse --short HEAD)
  BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
  LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)
  ```
- [x] 4.2 Update `go install` instructions in README
- [ ] 4.3 Add goreleaser config (optional, for releases)

### 5. Documentation
- [x] 5.1 Add `canopy version` to README command list
- [x] 5.2 Add version output example

### 6. Testing
- [ ] 6.1 Test version command output format
- [ ] 6.2 Test `--json` output structure
- [ ] 6.3 Test `--version` flag behavior
