// Package gitx wraps git operations used by canopy.
package gitx

import (
	"context"
	"errors"
	"math/rand/v2"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts (including the first one).
	// Set to 1 to disable retries.
	MaxAttempts int

	// InitialDelay is the initial backoff delay between retries.
	InitialDelay time.Duration

	// MaxDelay is the maximum backoff delay cap.
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry.
	Multiplier float64

	// JitterFactor adds randomness to delays (0.25 = ±25%).
	JitterFactor float64
}

// DefaultRetryConfig returns sensible default retry configuration.
// - 3 max attempts
// - 1s initial delay
// - 30s max delay
// - 2x multiplier
// - 25% jitter
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.25,
	}
}

// calculateBackoff computes the backoff delay for a given attempt (0-indexed).
// Returns the delay with jitter applied.
func (cfg RetryConfig) calculateBackoff(attempt int) time.Duration {
	if attempt == 0 {
		return 0
	}

	// Calculate base delay: initialDelay * multiplier^(attempt-1)
	delay := float64(cfg.InitialDelay)
	for i := 1; i < attempt; i++ {
		delay *= cfg.Multiplier
	}

	// Cap at max delay
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	// Apply jitter: delay * (1 ± jitterFactor)
	// rand.Float64() returns [0.0, 1.0), so we map it to [-jitter, +jitter]
	// G404: Using math/rand is fine for non-security jitter calculation
	jitter := (rand.Float64()*2 - 1) * cfg.JitterFactor //nolint:gosec
	delay *= (1 + jitter)

	return time.Duration(delay)
}

// isRetryableError determines if an error is transient and worth retrying.
// Returns true for network timeouts, connection errors, and server errors.
// Returns false for auth failures, not found, and other permanent errors.
//
//nolint:gocyclo // Complexity is expected for comprehensive error classification
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation - not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for network errors (DNS, connection refused, timeout)
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Timeout errors are retryable
		if netErr.Timeout() {
			return true
		}
	}

	// Check for connection refused (transient)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Check for syscall errors (connection reset, pipe broken, etc.)
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	// Check for go-git transport errors
	// Authentication errors are NOT retryable
	if errors.Is(err, transport.ErrAuthenticationRequired) ||
		errors.Is(err, transport.ErrAuthorizationFailed) {
		return false
	}

	// Repository not found is NOT retryable
	if errors.Is(err, transport.ErrRepositoryNotFound) {
		return false
	}

	// Empty remote repository is NOT retryable
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return false
	}

	// Check error message for common transient patterns
	errStr := strings.ToLower(err.Error())

	// Retryable patterns
	retryablePatterns := []string{
		"connection reset",
		"connection refused",
		"connection timed out",
		"network is unreachable",
		"no route to host",
		"temporary failure",
		"dns",
		"lookup",
		"i/o timeout",
		"eof",
		"broken pipe",
		"502",
		"503",
		"504",
		"429", // rate limited
		"too many requests",
		"internal server error",
		"service unavailable",
		"gateway timeout",
		"bad gateway",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Non-retryable patterns
	nonRetryablePatterns := []string{
		"authentication",
		"permission denied",
		"not found",
		"404",
		"401",
		"403",
		"invalid",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	return false
}

// WithRetry executes the operation with retry logic based on the configuration.
// It returns the result of the first successful attempt or the final error after
// all retries are exhausted. Context cancellation is respected between attempts.
// If MaxAttempts is <= 0, the operation is executed exactly once.
//
//nolint:gocyclo // Complexity is expected for retry logic with context handling
func WithRetry[T any](ctx context.Context, cfg RetryConfig, op func() (T, error)) (T, error) {
	var (
		zero    T
		lastErr error
	)

	// Ensure at least one attempt is made
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check context before attempting
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return zero, lastErr
			}

			return zero, err
		}

		// Wait before retry (no wait on first attempt)
		if attempt > 0 {
			delay := cfg.calculateBackoff(attempt)
			log.Info("retrying operation",
				"attempt", attempt+1,
				"max_attempts", maxAttempts,
				"delay", delay.Round(time.Millisecond))

			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Execute the operation
		result, err := op()
		if err == nil {
			if attempt > 0 {
				log.Info("operation succeeded after retry", "attempts", attempt+1)
			}

			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			log.Debug("error is not retryable", "error", err)
			return zero, err
		}

		if attempt < maxAttempts-1 {
			log.Warn("operation failed, will retry",
				"attempt", attempt+1,
				"error", err)
		}
	}

	log.Error("operation failed after all retries",
		"attempts", maxAttempts,
		"error", lastErr)

	return zero, lastErr
}

// WithRetryNoResult executes an operation that returns only an error.
// This is a convenience wrapper around WithRetry for void operations.
func WithRetryNoResult(ctx context.Context, cfg RetryConfig, op func() error) error {
	_, err := WithRetry(ctx, cfg, func() (struct{}, error) {
		return struct{}{}, op()
	})

	return err
}
