package workspaces

import (
	"context"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

type mockServiceDeps struct {
	svc     *Service
	config  *mocks.MockConfigProvider
	git     *mocks.MockGitOperations
	storage *mocks.MockWorkspaceStorage
	cache   *mocks.MockWorkspaceCache
	disk    *mocks.MockDiskUsage
}

func newMockService(t *testing.T) mockServiceDeps {
	t.Helper()

	config := mocks.NewMockConfigProvider()
	config.WorkspacesRoot = t.TempDir()
	config.ClosedRoot = t.TempDir()

	git := mocks.NewMockGitOperations()
	storage := mocks.NewMockWorkspaceStorage()
	storage.LoadFunc = func(_ context.Context, id string) (*domain.Workspace, error) {
		ws, ok := storage.Workspaces[id]
		if !ok {
			return nil, cerrors.NewWorkspaceNotFound(id)
		}

		return &ws, nil
	}
	cache := mocks.NewMockWorkspaceCache()
	disk := mocks.NewMockDiskUsage()

	svc := NewService(
		config,
		git,
		storage,
		nil,
		WithCache(cache),
		WithDiskUsage(disk),
	)

	return mockServiceDeps{
		svc:     svc,
		config:  config,
		git:     git,
		storage: storage,
		cache:   cache,
		disk:    disk,
	}
}

func addWorkspaceFixture(storage *mocks.MockWorkspaceStorage, ws domain.Workspace) {
	storage.Workspaces[ws.ID] = ws
}
