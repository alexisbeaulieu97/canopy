## 1. Domain Types
- [x] 1.1 Add HealthCheck and HealthStatus types to domain
- [x] 1.2 Define health check categories (worktree, remote, metadata)

## 2. Health Check Implementation
- [x] 2.1 Create health check service in `internal/workspaces/`
- [x] 2.2 Implement worktree integrity check
- [x] 2.3 Implement git config validity check
- [x] 2.4 Implement remote connectivity check (optional, slow)
- [x] 2.5 Implement metadata consistency check

## 3. CLI Command
- [x] 3.1 Add `doctor workspace` subcommand
- [x] 3.2 Add `--workspace` flag for specific workspace
- [x] 3.3 Add `--fix` flag for auto-remediation
- [x] 3.4 Add `--json` output format

## 4. Remediation
- [x] 4.1 Implement fix for common issues
- [x] 4.2 Provide suggestions for unfixable issues
- [x] 4.3 Log all remediation actions

## 5. Testing
- [x] 5.1 Add tests for health check detection
- [x] 5.2 Add tests for remediation actions
- [x] 5.3 Add integration tests

## 6. Documentation
- [x] 6.1 Document health check command in usage.md
- [x] 6.2 Document health check scenarios
