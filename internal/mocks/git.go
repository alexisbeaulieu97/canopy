// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"

	"github.com/go-git/go-git/v5"

	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockGitOperations implements ports.GitOperations.
var _ ports.GitOperations = (*MockGitOperations)(nil)

// MockGitOperations is a mock implementation of ports.GitOperations for testing.
type MockGitOperations struct {
	EnsureCanonicalFunc func(ctx context.Context, repoURL, repoName string) (*git.Repository, error)
	CreateWorktreeFunc  func(ctx context.Context, repoName, worktreePath, branchName string) error
	StatusFunc          func(ctx context.Context, path string) (bool, int, int, string, error)
	CloneFunc           func(ctx context.Context, url, name string) error
	FetchFunc           func(ctx context.Context, name string) error
	PullFunc            func(ctx context.Context, path string) error
	PushFunc            func(ctx context.Context, path, branch string) error
	ListFunc            func(ctx context.Context) ([]string, error)
	CheckoutFunc        func(ctx context.Context, path, branchName string, create bool) error
	RenameBranchFunc    func(ctx context.Context, repoPath, oldName, newName string) error
	RunCommandFunc      func(ctx context.Context, repoPath string, args ...string) (*ports.CommandResult, error)
	GetUpstreamURLFunc  func(repoName string) (string, error)
	RemoveWorktreeFunc  func(ctx context.Context, repoName, worktreePath string) error
	PruneWorktreesFunc  func(ctx context.Context, repoName string) error
}

// NewMockGitOperations creates a new MockGitOperations with default no-op behavior.
func NewMockGitOperations() *MockGitOperations {
	return &MockGitOperations{}
}

// EnsureCanonical calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) EnsureCanonical(ctx context.Context, repoURL, repoName string) (*git.Repository, error) {
	if m.EnsureCanonicalFunc != nil {
		return m.EnsureCanonicalFunc(ctx, repoURL, repoName)
	}

	return nil, nil
}

// CreateWorktree calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) CreateWorktree(ctx context.Context, repoName, worktreePath, branchName string) error {
	if m.CreateWorktreeFunc != nil {
		return m.CreateWorktreeFunc(ctx, repoName, worktreePath, branchName)
	}

	return nil
}

// Status calls the mock function if set, otherwise returns default values.
func (m *MockGitOperations) Status(ctx context.Context, path string) (bool, int, int, string, error) {
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx, path)
	}

	return false, 0, 0, "main", nil
}

// Clone calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Clone(ctx context.Context, url, name string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(ctx, url, name)
	}

	return nil
}

// Fetch calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Fetch(ctx context.Context, name string) error {
	if m.FetchFunc != nil {
		return m.FetchFunc(ctx, name)
	}

	return nil
}

// Pull calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Pull(ctx context.Context, path string) error {
	if m.PullFunc != nil {
		return m.PullFunc(ctx, path)
	}

	return nil
}

// Push calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Push(ctx context.Context, path, branch string) error {
	if m.PushFunc != nil {
		return m.PushFunc(ctx, path, branch)
	}

	return nil
}

// List calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) List(ctx context.Context) ([]string, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}

	return nil, nil
}

// Checkout calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Checkout(ctx context.Context, path, branchName string, create bool) error {
	if m.CheckoutFunc != nil {
		return m.CheckoutFunc(ctx, path, branchName, create)
	}

	return nil
}

// RenameBranch calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) RenameBranch(ctx context.Context, repoPath, oldName, newName string) error {
	if m.RenameBranchFunc != nil {
		return m.RenameBranchFunc(ctx, repoPath, oldName, newName)
	}

	return nil
}

// RunCommand calls the mock function if set, otherwise returns an empty result.
func (m *MockGitOperations) RunCommand(ctx context.Context, repoPath string, args ...string) (*ports.CommandResult, error) {
	if m.RunCommandFunc != nil {
		return m.RunCommandFunc(ctx, repoPath, args...)
	}

	return &ports.CommandResult{}, nil
}

// GetUpstreamURL calls the mock function if set, otherwise returns empty string.
func (m *MockGitOperations) GetUpstreamURL(repoName string) (string, error) {
	if m.GetUpstreamURLFunc != nil {
		return m.GetUpstreamURLFunc(repoName)
	}

	return "", nil
}

// RemoveWorktree calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) RemoveWorktree(ctx context.Context, repoName, worktreePath string) error {
	if m.RemoveWorktreeFunc != nil {
		return m.RemoveWorktreeFunc(ctx, repoName, worktreePath)
	}

	return nil
}

// PruneWorktrees calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) PruneWorktrees(ctx context.Context, repoName string) error {
	if m.PruneWorktreesFunc != nil {
		return m.PruneWorktreesFunc(ctx, repoName)
	}

	return nil
}
