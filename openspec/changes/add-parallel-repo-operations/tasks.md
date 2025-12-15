## 1. Configuration
- [ ] 1.1 Add `parallel_workers` config option with default of 4
- [ ] 1.2 Add validation for parallel_workers (1-10 range)

## 2. Parallel Execution Framework
- [ ] 2.1 Create worker pool helper for bounded concurrency
- [ ] 2.2 Implement error collection from parallel operations
- [ ] 2.3 Add context cancellation support for fail-fast behavior

## 3. Service Integration
- [ ] 3.1 Update CreateWorkspace to parallelize EnsureCanonical calls
- [ ] 3.2 Maintain sequential worktree creation (depends on canonical)
- [ ] 3.3 Implement atomic rollback on failure

## 4. Testing
- [ ] 4.1 Add tests for parallel execution
- [ ] 4.2 Add tests for error handling in parallel mode
- [ ] 4.3 Add tests for context cancellation during parallel ops

