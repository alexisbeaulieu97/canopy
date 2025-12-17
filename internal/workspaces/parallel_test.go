package workspaces

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestRunParallelCanonical_Success(t *testing.T) {
	t.Parallel()

	mockGit := mocks.NewMockGitOperations()
	mockCfg := mocks.NewMockConfigProvider()
	mockCfg.ParallelWorkers = 4

	var callCount atomic.Int32

	mockGit.EnsureCanonicalFunc = func(_ context.Context, _, _ string) (*git.Repository, error) {
		callCount.Add(1)
		return nil, nil
	}

	svc := &Service{
		config:    mockCfg,
		gitEngine: mockGit,
	}

	repos := []domain.Repo{
		{Name: "repo1", URL: "https://github.com/org/repo1.git"},
		{Name: "repo2", URL: "https://github.com/org/repo2.git"},
		{Name: "repo3", URL: "https://github.com/org/repo3.git"},
	}

	err := svc.runParallelCanonical(context.Background(), repos, parallelCanonicalOptions{workers: 4})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if callCount.Load() != 3 {
		t.Errorf("expected 3 EnsureCanonical calls, got %d", callCount.Load())
	}
}

func TestRunParallelCanonical_SingleRepo(t *testing.T) {
	t.Parallel()

	mockGit := mocks.NewMockGitOperations()
	mockCfg := mocks.NewMockConfigProvider()

	var callCount atomic.Int32

	mockGit.EnsureCanonicalFunc = func(_ context.Context, _, _ string) (*git.Repository, error) {
		callCount.Add(1)
		return nil, nil
	}

	svc := &Service{
		config:    mockCfg,
		gitEngine: mockGit,
	}

	repos := []domain.Repo{
		{Name: "repo1", URL: "https://github.com/org/repo1.git"},
	}

	err := svc.runParallelCanonical(context.Background(), repos, parallelCanonicalOptions{workers: 4})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if callCount.Load() != 1 {
		t.Errorf("expected 1 EnsureCanonical call, got %d", callCount.Load())
	}
}

func TestRunParallelCanonical_EmptyRepos(t *testing.T) {
	t.Parallel()

	mockGit := mocks.NewMockGitOperations()
	mockCfg := mocks.NewMockConfigProvider()

	svc := &Service{
		config:    mockCfg,
		gitEngine: mockGit,
	}

	err := svc.runParallelCanonical(context.Background(), nil, parallelCanonicalOptions{workers: 4})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRunParallelCanonical_ErrorFailsFast(t *testing.T) {
	t.Parallel()

	mockGit := mocks.NewMockGitOperations()
	mockCfg := mocks.NewMockConfigProvider()
	mockCfg.ParallelWorkers = 4

	expectedErr := errors.New("clone failed")

	var callCount atomic.Int32

	mockGit.EnsureCanonicalFunc = func(_ context.Context, _, repoName string) (*git.Repository, error) {
		callCount.Add(1)
		// First repo fails
		if repoName == "repo1" {
			return nil, expectedErr
		}
		// Other repos take time so they can be cancelled
		time.Sleep(50 * time.Millisecond)

		return nil, nil
	}

	svc := &Service{
		config:    mockCfg,
		gitEngine: mockGit,
	}

	repos := []domain.Repo{
		{Name: "repo1", URL: "https://github.com/org/repo1.git"},
		{Name: "repo2", URL: "https://github.com/org/repo2.git"},
		{Name: "repo3", URL: "https://github.com/org/repo3.git"},
	}

	err := svc.runParallelCanonical(context.Background(), repos, parallelCanonicalOptions{workers: 4})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		// The error should be wrapped, so check the message contains the original error
		if !strings.Contains(err.Error(), expectedErr.Error()) {
			t.Errorf("expected wrapped error containing %q, got %v", expectedErr.Error(), err)
		}
	}
}

func TestRunParallelCanonical_ContextCancellation(t *testing.T) {
	t.Parallel()

	mockGit := mocks.NewMockGitOperations()
	mockCfg := mocks.NewMockConfigProvider()
	mockCfg.ParallelWorkers = 4

	var startedCount atomic.Int32

	mockGit.EnsureCanonicalFunc = func(ctx context.Context, _, _ string) (*git.Repository, error) {
		startedCount.Add(1)
		// Simulate slow operation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
			return nil, nil
		}
	}

	svc := &Service{
		config:    mockCfg,
		gitEngine: mockGit,
	}

	repos := []domain.Repo{
		{Name: "repo1", URL: "https://github.com/org/repo1.git"},
		{Name: "repo2", URL: "https://github.com/org/repo2.git"},
		{Name: "repo3", URL: "https://github.com/org/repo3.git"},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := svc.runParallelCanonical(ctx, repos, parallelCanonicalOptions{workers: 4})
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestRunParallelCanonical_BoundedConcurrency(t *testing.T) {
	t.Parallel()

	mockGit := mocks.NewMockGitOperations()
	mockCfg := mocks.NewMockConfigProvider()
	mockCfg.ParallelWorkers = 2

	var (
		concurrentCount atomic.Int32
		maxConcurrent   atomic.Int32
	)

	mockGit.EnsureCanonicalFunc = func(_ context.Context, _, _ string) (*git.Repository, error) {
		current := concurrentCount.Add(1)
		// Track max concurrent
		for {
			maxVal := maxConcurrent.Load()
			if current <= maxVal || maxConcurrent.CompareAndSwap(maxVal, current) {
				break
			}
		}

		time.Sleep(20 * time.Millisecond)
		concurrentCount.Add(-1)

		return nil, nil
	}

	svc := &Service{
		config:    mockCfg,
		gitEngine: mockGit,
	}

	repos := []domain.Repo{
		{Name: "repo1", URL: "https://github.com/org/repo1.git"},
		{Name: "repo2", URL: "https://github.com/org/repo2.git"},
		{Name: "repo3", URL: "https://github.com/org/repo3.git"},
		{Name: "repo4", URL: "https://github.com/org/repo4.git"},
	}

	err := svc.runParallelCanonical(context.Background(), repos, parallelCanonicalOptions{workers: 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if maxConcurrent.Load() > 2 {
		t.Errorf("expected max concurrent <= 2, got %d", maxConcurrent.Load())
	}
}

func TestRunParallelCanonical_SingleWorkerRunsSequentially(t *testing.T) {
	t.Parallel()

	mockGit := mocks.NewMockGitOperations()
	mockCfg := mocks.NewMockConfigProvider()
	mockCfg.ParallelWorkers = 1

	var callOrder []string

	mockGit.EnsureCanonicalFunc = func(_ context.Context, _, repoName string) (*git.Repository, error) {
		callOrder = append(callOrder, repoName)
		return nil, nil
	}

	svc := &Service{
		config:    mockCfg,
		gitEngine: mockGit,
	}

	repos := []domain.Repo{
		{Name: "repo1", URL: "https://github.com/org/repo1.git"},
		{Name: "repo2", URL: "https://github.com/org/repo2.git"},
	}

	err := svc.runParallelCanonical(context.Background(), repos, parallelCanonicalOptions{workers: 1})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// With single worker, should run sequentially in order
	if len(callOrder) != 2 {
		t.Errorf("expected 2 calls, got %d", len(callOrder))
	}

	if callOrder[0] != "repo1" || callOrder[1] != "repo2" {
		t.Errorf("expected sequential order [repo1, repo2], got %v", callOrder)
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
