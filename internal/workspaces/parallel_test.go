package workspaces

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
)

func TestParallelExecutor_RunSuccess(t *testing.T) {
	t.Parallel()

	executor := NewParallelExecutor(4)

	var callCount atomic.Int32

	err := executor.Run(context.Background(), 3, func(_ context.Context, _ int) error {
		callCount.Add(1)
		return nil
	}, ParallelOptions{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", callCount.Load())
	}
}

func TestParallelExecutor_RunStopsOnError(t *testing.T) {
	t.Parallel()

	executor := NewParallelExecutor(4)
	expectedErr := errors.New("boom")

	err := executor.Run(context.Background(), 3, func(_ context.Context, index int) error {
		if index == 0 {
			return expectedErr
		}

		time.Sleep(5 * time.Millisecond)

		return nil
	}, ParallelOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestParallelExecutor_RunContinueOnError(t *testing.T) {
	t.Parallel()

	executor := NewParallelExecutor(2)

	err := executor.Run(context.Background(), 3, func(_ context.Context, index int) error {
		if index == 1 {
			return errors.New("failure")
		}

		return nil
	}, ParallelOptions{ContinueOnError: true})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestParallelExecutor_MapCollectsResults(t *testing.T) {
	t.Parallel()

	executor := NewParallelExecutor(1)

	results, err := ParallelMap(context.Background(), executor, 3, func(_ context.Context, index int) (string, error) {
		return "value-" + strconv.Itoa(index), nil
	}, ParallelOptions{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	values := ExtractValues(results)
	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}

	if values[0] != "value-0" || values[1] != "value-1" || values[2] != "value-2" {
		t.Fatalf("unexpected values: %v", values)
	}
}

func TestParallelExecutor_BoundedConcurrency(t *testing.T) {
	t.Parallel()

	executor := NewParallelExecutor(2)

	var (
		concurrentCount atomic.Int32
		maxConcurrent   atomic.Int32
	)

	err := executor.Run(context.Background(), 4, func(_ context.Context, _ int) error {
		current := concurrentCount.Add(1)

		for {
			maxVal := maxConcurrent.Load()
			if current <= maxVal || maxConcurrent.CompareAndSwap(maxVal, current) {
				break
			}
		}

		time.Sleep(10 * time.Millisecond)
		concurrentCount.Add(-1)

		return nil
	}, ParallelOptions{Workers: 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if maxConcurrent.Load() > 2 {
		t.Errorf("expected max concurrent <= 2, got %d", maxConcurrent.Load())
	}
}

func TestParallelExecutor_ContextCancellation(t *testing.T) {
	t.Parallel()

	executor := NewParallelExecutor(4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerStarted := make(chan struct{})

	var startedOnce sync.Once

	go func() {
		<-workerStarted
		cancel()
	}()

	err := executor.Run(ctx, 3, func(runCtx context.Context, _ int) error {
		startedOnce.Do(func() { close(workerStarted) })
		<-runCtx.Done()

		return runCtx.Err()
	}, ParallelOptions{})
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
}

func TestParallelExecutor_ErrorHelpers(t *testing.T) {
	t.Parallel()

	results := []ParallelResult[string]{
		{Value: "a"},
		{Value: "b", Err: errors.New("fail")},
		{Value: "c"},
	}

	if CountErrors(results) != 1 {
		t.Fatalf("expected 1 error, got %d", CountErrors(results))
	}

	if FirstError(results) == nil {
		t.Fatal("expected first error, got nil")
	}

	if AggregateErrors(results) == nil {
		t.Fatal("expected aggregate error, got nil")
	}
}

func TestConfigParallelWorkersValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		workers   int
		expectErr bool
	}{
		{"valid minimum", config.MinParallelWorkers, false},
		{"valid maximum", config.MaxParallelWorkers, false},
		{"valid middle", 5, false},
		{"invalid zero", 0, true},
		{"invalid negative", -1, true},
		{"invalid too high", config.MaxParallelWorkers + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				ProjectsRoot:    "/projects",
				WorkspacesRoot:  "/workspaces",
				ClosedRoot:      "/closed",
				CloseDefault:    "delete",
				ParallelWorkers: tt.workers,
				Git: config.GitConfig{
					Retry: config.GitRetrySettings{
						MaxAttempts:  3,
						InitialDelay: "1s",
						MaxDelay:     "30s",
						Multiplier:   2.0,
						JitterFactor: 0.25,
					},
				},
			}

			err := cfg.ValidateValues()

			if tt.expectErr && err == nil {
				t.Errorf("expected error for workers=%d, got nil", tt.workers)
			}

			if !tt.expectErr && err != nil {
				t.Errorf("expected no error for workers=%d, got %v", tt.workers, err)
			}
		})
	}
}

func TestConfigParallelWorkersDefault(t *testing.T) {
	t.Parallel()

	if config.DefaultParallelWorkers != 4 {
		t.Errorf("expected default parallel workers to be 4, got %d", config.DefaultParallelWorkers)
	}

	if config.MinParallelWorkers != 1 {
		t.Errorf("expected min parallel workers to be 1, got %d", config.MinParallelWorkers)
	}

	if config.MaxParallelWorkers != 10 {
		t.Errorf("expected max parallel workers to be 10, got %d", config.MaxParallelWorkers)
	}
}

// BenchmarkParallelExecutor benchmarks parallel execution performance.
func BenchmarkParallelExecutor(b *testing.B) {
	executor := NewParallelExecutor(4)

	b.Run("parallel-4-workers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = executor.Run(context.Background(), 10, func(_ context.Context, _ int) error {
				time.Sleep(time.Microsecond)
				return nil
			}, ParallelOptions{})
		}
	})

	b.Run("sequential-1-worker", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = executor.Run(context.Background(), 10, func(_ context.Context, _ int) error {
				time.Sleep(time.Microsecond)
				return nil
			}, ParallelOptions{Workers: 1})
		}
	})
}
