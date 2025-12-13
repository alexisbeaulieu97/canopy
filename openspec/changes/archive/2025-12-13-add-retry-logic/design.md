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

**Alternatives Considered:**
- *Linear backoff* — Rejected: slower recovery under light contention, doesn't scale well for burst failures.
- *Fixed delay* — Rejected: causes thundering herd when multiple operations retry simultaneously.
- *Jittered fixed delay* — Rejected: better than fixed but lacks progressive back-off, leading to higher contention under sustained failures.

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

**Alternatives Considered:**
- *Retry-all errors* — Rejected: wastes time on permanent failures (auth, not found), risks rate-limit escalation.
- *Retry on 4xx client errors* — Rejected: most 4xx errors are permanent (auth failures, bad requests); retrying escalates rate limits.
- *Exponential backoff for client errors* — Rejected: auth failures won't resolve with time; delays user feedback unnecessarily.

### Decision: Default Configuration
- Max attempts: 3
- Initial delay: 1s
- Configurable via config file (optional)

**Rationale:** Sensible defaults that work for most cases without configuration.

**Alternatives Considered:**
- *2 attempts* — Rejected: insufficient for transient network issues; often fails on the first retry.
- *5 attempts* — Rejected: adds ~30s latency on permanent failures; diminishing returns after 3 attempts.
- *10 attempts* — Rejected: excessive latency (minutes); operational overhead outweighs marginal success gains.

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

