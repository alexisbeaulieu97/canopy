// Package mocks provides mock implementations for testing.
package mocks

import (
	"github.com/go-git/go-git/v5"

	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockGitOperations implements ports.GitOperations.
var _ ports.GitOperations = (*MockGitOperations)(nil)

// MockGitOperations is a mock implementation of ports.GitOperations for testing.
type MockGitOperations struct {
	EnsureCanonicalFunc  func(repoURL, repoName string) (*git.Repository, error)
	CreateWorktreeFunc   func(repoName, worktreePath, branchName string) error
	StatusFunc           func(path string) (bool, int, int, string, error)
	CloneFunc            func(url, name string) error
	FetchFunc            func(name string) error
	PullFunc             func(path string) error
	PushFunc             func(path, branch string) error
	ListFunc             func() ([]string, error)
	CheckoutFunc         func(path, branchName string, create bool) error
}

// NewMockGitOperations creates a new MockGitOperations with default no-op behavior.
func NewMockGitOperations() *MockGitOperations {
	return &MockGitOperations{}
}

// EnsureCanonical calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) EnsureCanonical(repoURL, repoName string) (*git.Repository, error) {
	if m.EnsureCanonicalFunc != nil {
		return m.EnsureCanonicalFunc(repoURL, repoName)
	}
	return nil, nil
}

// CreateWorktree calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) CreateWorktree(repoName, worktreePath, branchName string) error {
	if m.CreateWorktreeFunc != nil {
		return m.CreateWorktreeFunc(repoName, worktreePath, branchName)
	}
	return nil
}

// Status calls the mock function if set, otherwise returns default values.
func (m *MockGitOperations) Status(path string) (bool, int, int, string, error) {
	if m.StatusFunc != nil {
		return m.StatusFunc(path)
	}
	return false, 0, 0, "main", nil
}

// Clone calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Clone(url, name string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(url, name)
	}
	return nil
}

// Fetch calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Fetch(name string) error {
	if m.FetchFunc != nil {
		return m.FetchFunc(name)
	}
	return nil
}

// Pull calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Pull(path string) error {
	if m.PullFunc != nil {
		return m.PullFunc(path)
	}
	return nil
}

// Push calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Push(path, branch string) error {
	if m.PushFunc != nil {
		return m.PushFunc(path, branch)
	}
	return nil
}

// List calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) List() ([]string, error) {
	if m.ListFunc != nil {
		return m.ListFunc()
	}
	return nil, nil
}

// Checkout calls the mock function if set, otherwise returns nil.
func (m *MockGitOperations) Checkout(path, branchName string, create bool) error {
	if m.CheckoutFunc != nil {
		return m.CheckoutFunc(path, branchName, create)
	}
	return nil
}
