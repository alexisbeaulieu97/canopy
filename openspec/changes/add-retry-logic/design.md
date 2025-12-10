## Context

Git network operations can fail for transient reasons:
- Network timeouts
- DNS resolution failures
- Server rate limiting (GitHub, GitLab)
- Temporary server unavailability
- SSH connection drops

Currently, users must manually retry failed operations, which is frustrating during workspace creation with multiple repos.

## Goals / Non-Goals

**Goals:**
- Automatically retry transient network failures
- Use exponential backoff to avoid overwhelming servers
- Make retry behavior configurable
- Log retry attempts for debugging
- Preserve final error with full context

**Non-Goals:**
- Retry non-network errors (auth failures, permission denied)
- Infinite retries
- Retry local git operations

## Decisions

### Decision: Exponential Backoff with Jitter
Use exponential backoff with random jitter to prevent thundering herd:
- Initial delay: 1 second
- Max delay: 30 seconds
- Multiplier: 2x
- Jitter: ±25%

**Rationale:** Industry standard approach, prevents synchronized retries.

### Decision: Retry Only Specific Errors
Only retry errors that indicate transient failures:
- Network timeouts
- Connection refused
- DNS errors
- 5xx HTTP errors
- SSH connection reset

Do NOT retry:
- Authentication failures (401, 403)
- Repository not found (404)
- Permission denied

**Rationale:** Retrying permanent failures wastes time and can trigger rate limits.

### Decision: Default Configuration
- Max attempts: 3
- Initial delay: 1s
- Configurable via config file (optional)

**Rationale:** Sensible defaults that work for most cases without configuration.

## Risks / Trade-offs

- **Risk**: Increased latency on permanent failures
  - **Mitigation**: Only retry transient errors, cap max attempts
  
- **Risk**: Hidden failures (user doesn't know retries happened)
  - **Mitigation**: Log retry attempts at Info level

## Migration Plan

1. Implement retry wrapper in `internal/gitx/retry.go`
2. Add error classification for retryable errors
3. Wrap Clone, Fetch, Push, Pull operations
4. Add configuration options
5. Update tests

**Rollback:** Disable retry wrapper - no data changes

## Open Questions

- Should retry configuration be per-operation or global?
- Should we expose retry count in error messages?

