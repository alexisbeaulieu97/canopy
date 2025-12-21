## 1. Create Unified Executor
- [ ] 1.1 Design `ParallelExecutor` interface
- [ ] 1.2 Implement with configurable worker limits
- [ ] 1.3 Add proper context cancellation propagation
- [ ] 1.4 Add result aggregation helpers

## 2. Migrate Existing Code
- [ ] 2.1 Update `sync.go` to use unified executor
- [ ] 2.2 Update `status_batch.go` to use unified executor
- [ ] 2.3 Update `canonical.go` parallel operations
- [ ] 2.4 Update `git_service.go` parallel operations

## 3. Configuration
- [ ] 3.1 Move `defaultMaxParallel` to config
- [ ] 3.2 Add `parallel_workers` to config.yaml
- [ ] 3.3 Document configuration option

## 4. Context Propagation
- [ ] 4.1 Audit context.Background() usage
- [ ] 4.2 Replace with caller context where appropriate
- [ ] 4.3 Ensure hook execution respects caller context

## 5. Testing
- [ ] 5.1 Add unit tests for ParallelExecutor
- [ ] 5.2 Test cancellation behavior
- [ ] 5.3 Test error aggregation

## 6. Documentation
- [ ] 6.1 Update docs/architecture.md with parallel execution pattern
