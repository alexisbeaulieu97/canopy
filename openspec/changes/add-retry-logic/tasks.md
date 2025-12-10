# Tasks: Add Retry Logic for Git Network Operations

## Implementation Checklist

### 1. Retry Infrastructure
- [ ] 1.1 Create `internal/gitx/retry.go`
- [ ] 1.2 Implement `RetryConfig` struct:
  ```go
  type RetryConfig struct {
      MaxAttempts  int
      InitialDelay time.Duration
      MaxDelay     time.Duration
      Multiplier   float64
      JitterFactor float64
  }
  ```
- [ ] 1.3 Implement `DefaultRetryConfig()` function
- [ ] 1.4 Implement backoff calculation with jitter

### 2. Error Classification
- [ ] 2.1 Create `isRetryableError(err error) bool` function
- [ ] 2.2 Classify go-git errors:
  - Retryable: transport errors, timeouts
  - Not retryable: auth errors, not found
- [ ] 2.3 Classify exec.Command errors for RunCommand

### 3. Retry Wrapper
- [ ] 3.1 Implement generic retry wrapper:
  ```go
  func WithRetry[T any](ctx context.Context, cfg RetryConfig, op func() (T, error)) (T, error)
  ```
- [ ] 3.2 Add context cancellation support
- [ ] 3.3 Add logging for retry attempts

### 4. Integrate with Git Operations
- [ ] 4.1 Wrap `Clone` with retry
- [ ] 4.2 Wrap `Fetch` with retry
- [ ] 4.3 Wrap `Push` with retry
- [ ] 4.4 Wrap `Pull` with retry
- [ ] 4.5 Wrap `EnsureCanonical` with retry (for initial clone)

### 5. Configuration (Optional)
- [ ] 5.1 Add retry config to `config.Config`
- [ ] 5.2 Add config file options:
  ```yaml
  git:
    retry:
      max_attempts: 3
      initial_delay: 1s
  ```
- [ ] 5.3 Pass config to GitEngine

### 6. Testing
- [ ] 6.1 Unit test backoff calculation
- [ ] 6.2 Unit test error classification
- [ ] 6.3 Test retry wrapper with mock operations
- [ ] 6.4 Test context cancellation during retry
- [ ] 6.5 Integration test with network simulation (optional)

