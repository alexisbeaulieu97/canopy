## 1. Create Unified Executor
- [x] 1.1 Design `ParallelExecutor` interface
- [x] 1.2 Implement with configurable worker limits
- [x] 1.3 Add proper context cancellation propagation
- [x] 1.4 Add result aggregation helpers

## 2. Migrate Existing Code
- [x] 2.1 Update `sync.go` to use unified executor
- [x] 2.2 Update `status_batch.go` to use unified executor
- [x] 2.3 Update `canonical.go` parallel operations
- [x] 2.4 Update `git_service.go` parallel operations

## 3. Configuration
- [x] 3.1 Move `defaultMaxParallel` to config
- [x] 3.2 Add `parallel_workers` to config.yaml
- [x] 3.3 Document configuration option

## 4. Context Propagation
- [x] 4.1 Audit context.Background() usage
- [x] 4.2 Replace with caller context where appropriate
- [x] 4.3 Ensure hook execution respects caller context

## 5. Testing
- [x] 5.1 Add unit tests for ParallelExecutor
- [x] 5.2 Test cancellation behavior
- [x] 5.3 Test error aggregation

## 6. Documentation
- [x] 6.1 Update docs/architecture.md with parallel execution pattern
