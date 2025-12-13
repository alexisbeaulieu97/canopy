package gitx

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	assert.Equal(t, 3, cfg.MaxAttempts)
	assert.Equal(t, 1*time.Second, cfg.InitialDelay)
	assert.Equal(t, 30*time.Second, cfg.MaxDelay)
	assert.Equal(t, 2.0, cfg.Multiplier)
	assert.Equal(t, 0.25, cfg.JitterFactor)
}

func TestCalculateBackoff(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0, // No jitter for deterministic testing
	}

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"first attempt (no delay)", 0, 0},
		{"second attempt", 1, 1 * time.Second},
		{"third attempt", 2, 2 * time.Second},
		{"fourth attempt", 3, 4 * time.Second},
		{"fifth attempt", 4, 8 * time.Second},
		{"sixth attempt (capped)", 5, 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := cfg.calculateBackoff(tt.attempt)
			assert.Equal(t, tt.expected, delay)
		})
	}
}

func TestCalculateBackoff_WithJitter(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.25,
	}

	// Run multiple times to verify jitter creates variance
	delays := make(map[time.Duration]bool)

	for i := 0; i < 100; i++ {
		delay := cfg.calculateBackoff(1)
		delays[delay] = true
	}

	// Should have multiple different delay values due to jitter
	assert.Greater(t, len(delays), 1, "jitter should create variance in delays")

	// All delays should be within expected range (1s ± 25%)
	for delay := range delays {
		assert.GreaterOrEqual(t, delay, 750*time.Millisecond, "delay should be >= 750ms")
		assert.LessOrEqual(t, delay, 1250*time.Millisecond, "delay should be <= 1250ms")
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil error", nil, false},
		{"context canceled", context.Canceled, false},
		{"context deadline exceeded", context.DeadlineExceeded, false},
		{"auth required", transport.ErrAuthenticationRequired, false},
		{"auth failed", transport.ErrAuthorizationFailed, false},
		{"repo not found", transport.ErrRepositoryNotFound, false},
		{"empty remote", transport.ErrEmptyRemoteRepository, false},
		{"connection reset", syscall.ECONNRESET, true},
		{"connection refused", syscall.ECONNREFUSED, true},
		{"broken pipe", syscall.EPIPE, true},
		{"timeout syscall", syscall.ETIMEDOUT, true},
		{"connection reset message", errors.New("connection reset by peer"), true},
		{"connection refused message", errors.New("connection refused"), true},
		{"dns error message", errors.New("dns lookup failed"), true},
		{"i/o timeout message", errors.New("i/o timeout"), true},
		{"502 error", errors.New("HTTP 502 Bad Gateway"), true},
		{"503 error", errors.New("HTTP 503 Service Unavailable"), true},
		{"504 error", errors.New("504 Gateway Timeout"), true},
		{"429 rate limited", errors.New("429 Too Many Requests"), true},
		{"authentication error message", errors.New("authentication failed"), false},
		{"permission denied message", errors.New("permission denied"), false},
		{"404 not found", errors.New("repository not found 404"), false},
		{"401 unauthorized", errors.New("401 unauthorized"), false},
		{"403 forbidden", errors.New("403 forbidden"), false},
		{"random error", errors.New("some unknown error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.retryable, result, "error: %v", tt.err)
		})
	}
}

func TestIsRetryableError_NetError(t *testing.T) {
	// Create a mock net.Error that is a timeout
	timeoutErr := &mockNetError{timeout: true}
	assert.True(t, isRetryableError(timeoutErr))

	// Non-timeout net error
	nonTimeoutErr := &mockNetError{timeout: false}
	assert.False(t, isRetryableError(nonTimeoutErr))
}

func TestIsRetryableError_OpError(t *testing.T) {
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: errors.New("connection refused"),
	}
	assert.True(t, isRetryableError(opErr))
}

type mockNetError struct {
	timeout bool
}

func (e *mockNetError) Error() string   { return "mock net error" }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return false }

func TestWithRetry_Success(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	callCount := 0
	result, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 1, callCount, "should succeed on first attempt")
}

func TestWithRetry_RetryableError_EventualSuccess(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	callCount := 0
	result, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", errors.New("connection reset by peer")
		}

		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, callCount, "should succeed on third attempt")
}

func TestWithRetry_RetryableError_AllFail(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	callCount := 0
	result, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		return "", errors.New("connection reset by peer")
	})

	require.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, 3, callCount, "should attempt 3 times")
	assert.Contains(t, err.Error(), "connection reset")
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	callCount := 0
	result, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		return "", transport.ErrAuthenticationRequired
	})

	require.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, 1, callCount, "should not retry on auth error")
	assert.True(t, errors.Is(err, transport.ErrAuthenticationRequired))
}

func TestWithRetry_ContextCanceled(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := WithRetry(ctx, cfg, func() (string, error) {
		callCount++
		return "", errors.New("connection reset by peer")
	})

	require.Error(t, err)
	assert.Equal(t, "", result)
	// Should have made 1-2 attempts before context was canceled
	assert.LessOrEqual(t, callCount, 2, "should stop retrying after context canceled")
}

func TestWithRetry_ContextCanceledBeforeStart(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	callCount := 0
	result, err := WithRetry(ctx, cfg, func() (string, error) {
		callCount++
		return "success", nil
	})

	require.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, 0, callCount, "should not attempt if context already canceled")
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestWithRetryNoResult_Success(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	callCount := 0
	err := WithRetryNoResult(context.Background(), cfg, func() error {
		callCount++
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestWithRetryNoResult_RetryAndSucceed(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	callCount := 0
	err := WithRetryNoResult(context.Background(), cfg, func() error {
		callCount++
		if callCount < 2 {
			return errors.New("connection reset by peer")
		}

		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestWithRetry_SingleAttempt(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	callCount := 0
	result, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		return "", errors.New("connection reset by peer")
	})

	require.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, 1, callCount, "should only attempt once when MaxAttempts=1")
}
