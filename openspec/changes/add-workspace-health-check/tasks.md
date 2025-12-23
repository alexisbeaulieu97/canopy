## 1. Domain Types
- [ ] 1.1 Add HealthCheck and HealthStatus types to domain
- [ ] 1.2 Define health check categories (worktree, remote, metadata)

## 2. Health Check Implementation
- [ ] 2.1 Create health check service in `internal/workspaces/`
- [ ] 2.2 Implement worktree integrity check
- [ ] 2.3 Implement git config validity check
- [ ] 2.4 Implement remote connectivity check (optional, slow)
- [ ] 2.5 Implement metadata consistency check

## 3. CLI Command
- [ ] 3.1 Add `doctor workspace` subcommand
- [ ] 3.2 Add `--workspace` flag for specific workspace
- [ ] 3.3 Add `--fix` flag for auto-remediation
- [ ] 3.4 Add `--json` output format

## 4. Remediation
- [ ] 4.1 Implement fix for common issues
- [ ] 4.2 Provide suggestions for unfixable issues
- [ ] 4.3 Log all remediation actions

## 5. Testing
- [ ] 5.1 Add tests for health check detection
- [ ] 5.2 Add tests for remediation actions
- [ ] 5.3 Add integration tests

## 6. Documentation
- [ ] 6.1 Document health check command in usage.md
- [ ] 6.2 Document health check scenarios
