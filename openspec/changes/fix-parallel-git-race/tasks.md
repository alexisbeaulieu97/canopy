# Tasks: Fix Parallel Git Race Condition

## Implementation Checklist

### 1. Add Cancellation Support
- [ ] 1.1 Create cancellable context in `runGitParallel`:
  ```go
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  ```
- [ ] 1.2 Pass context to goroutines
- [ ] 1.3 Check context before acquiring semaphore

### 2. Implement Early Termination
- [ ] 2.1 Add error channel for first error reporting
- [ ] 2.2 Cancel context on first error when `continueOnError=false`
- [ ] 2.3 Ensure goroutines exit promptly on cancellation

### 3. Synchronize Error Collection
- [ ] 3.1 Use atomic operations or mutex for error tracking
- [ ] 3.2 Collect all errors that occurred before cancellation
- [ ] 3.3 Return meaningful error message with context

### 4. Testing
- [ ] 4.1 Add test for early termination on first error
- [ ] 4.2 Add test for all operations completing when `continueOnError=true`
- [ ] 4.3 Add race detector test (`go test -race`)
- [ ] 4.4 Add benchmark for parallel vs sequential performance

