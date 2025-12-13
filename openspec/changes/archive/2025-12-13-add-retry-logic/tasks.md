# Tasks: Add Retry Logic for Git Network Operations

## Implementation Checklist

### 1. Retry Infrastructure
- [x] 1.1 Create `internal/gitx/retry.go`
- [x] 1.2 Implement `RetryConfig` struct:
  ```go
  type RetryConfig struct {
      MaxAttempts  int
      InitialDelay time.Duration
      MaxDelay     time.Duration
      Multiplier   float64
      JitterFactor float64
  }
  ```
- [x] 1.3 Implement `DefaultRetryConfig()` function
- [x] 1.4 Implement backoff calculation with jitter

### 2. Error Classification
- [x] 2.1 Create `isRetryableError(err error) bool` function
- [x] 2.2 Classify go-git errors:
  - Retryable: transport errors, timeouts
  - Not retryable: auth errors, not found
- [x] 2.3 Classify exec.Command errors for RunCommand

### 3. Retry Wrapper
- [x] 3.1 Implement generic retry wrapper:
  ```go
  func WithRetry[T any](ctx context.Context, cfg RetryConfig, op func() (T, error)) (T, error)
  ```
- [x] 3.2 Add context cancellation support
- [x] 3.3 Add logging for retry attempts

### 4. Integrate with Git Operations
- [x] 4.1 Wrap `Clone` with retry
- [x] 4.2 Wrap `Fetch` with retry
- [x] 4.3 Wrap `Push` with retry
- [x] 4.4 Wrap `Pull` with retry
- [x] 4.5 Wrap `EnsureCanonical` with retry (for initial clone)

### 5. Configuration (Optional)
- [x] 5.1 Add retry config to `config.Config`
- [x] 5.2 Add config file options:
  ```yaml
  git:
    retry:
      max_attempts: 3
      initial_delay: 1s
  ```
- [x] 5.3 Pass config to GitEngine

### 6. Testing
- [x] 6.1 Unit test backoff calculation
- [x] 6.2 Unit test error classification
- [x] 6.3 Test retry wrapper with mock operations
- [x] 6.4 Test context cancellation during retry
- [ ] 6.5 Integration test with network simulation (optional)
