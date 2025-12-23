package tui

import (
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

type tuiServiceDeps struct {
	svc     *workspaces.Service
	config  *mocks.MockConfigProvider
	git     *mocks.MockGitOperations
	storage *mocks.MockWorkspaceStorage
	disk    *mocks.MockDiskUsage
	cache   *mocks.MockWorkspaceCache
}

func newTUITestService(t *testing.T) tuiServiceDeps {
	t.Helper()

	config := mocks.NewMockConfigProvider()
	config.WorkspacesRoot = t.TempDir()
	config.ClosedRoot = t.TempDir()

	git := mocks.NewMockGitOperations()
	storage := mocks.NewMockWorkspaceStorage()
	disk := mocks.NewMockDiskUsage()
	cache := mocks.NewMockWorkspaceCache()

	svc := workspaces.NewService(
		config,
		git,
		storage,
		nil,
		workspaces.WithDiskUsage(disk),
		workspaces.WithCache(cache),
	)

	return tuiServiceDeps{
		svc:     svc,
		config:  config,
		git:     git,
		storage: storage,
		disk:    disk,
		cache:   cache,
	}
}

func newTUITestModel(t *testing.T) (Model, tuiServiceDeps) {
	t.Helper()

	deps := newTUITestService(t)

	return NewModel(deps.svc, false), deps
}

func addTUIWorkspace(storage *mocks.MockWorkspaceStorage, ws domain.Workspace) {
	storage.Workspaces[ws.ID] = ws
}
