package workspaces

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// GetCanonicalRepoStatus returns detailed status for a single canonical repository.
func (s *Service) GetCanonicalRepoStatus(ctx context.Context, name string) (*domain.CanonicalRepoStatus, error) {
	if s.gitEngine == nil {
		return nil, cerrors.NewInternalError("git engine not initialized", nil)
	}

	usageMap, err := s.buildRepoUsageMap(ctx)
	if err != nil {
		return nil, err
	}

	return s.getCanonicalRepoStatus(name, usageMap)
}

// GetAllCanonicalRepoStatuses returns status for all canonical repositories.
func (s *Service) GetAllCanonicalRepoStatuses(ctx context.Context) ([]domain.CanonicalRepoStatus, error) {
	if s.gitEngine == nil {
		return nil, cerrors.NewInternalError("git engine not initialized", nil)
	}

	repoNames, err := s.gitEngine.List(ctx)
	if err != nil {
		return nil, cerrors.WrapGitError(err, "list canonical repos")
	}

	usageMap, err := s.buildRepoUsageMap(ctx)
	if err != nil {
		return nil, err
	}

	executor := NewParallelExecutor(s.config.GetParallelWorkers())

	results, err := ParallelMap(ctx, executor, len(repoNames), func(runCtx context.Context, index int) (*domain.CanonicalRepoStatus, error) {
		if runCtx.Err() != nil {
			return nil, runCtx.Err()
		}

		return s.getCanonicalRepoStatus(repoNames[index], usageMap)
	}, ParallelOptions{ContinueOnError: true})
	if err != nil {
		return nil, err
	}

	statuses := make([]domain.CanonicalRepoStatus, 0, len(repoNames))

	for i, result := range results {
		if result.Err != nil {
			if s.logger != nil {
				s.logger.Warn("failed to get canonical repo status", "repo", repoNames[i], "error", result.Err)
			}

			continue
		}

		if result.Value != nil {
			statuses = append(statuses, *result.Value)
		}
	}

	return statuses, nil
}

func (s *Service) getCanonicalRepoStatus(name string, usageMap map[string][]string) (*domain.CanonicalRepoStatus, error) {
	path := filepath.Join(s.config.GetProjectsRoot(), name)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, cerrors.NewRepoNotFound(name)
		}

		return nil, cerrors.NewIOFailed(fmt.Sprintf("stat canonical repo %s", name), err)
	}

	size, err := s.gitEngine.GetRepoSize(name)
	if err != nil {
		return nil, cerrors.NewIOFailed(fmt.Sprintf("get repo size for %s", name), err)
	}

	lastFetch, err := s.gitEngine.LastFetchTime(name)
	if err != nil {
		return nil, cerrors.WrapGitError(err, fmt.Sprintf("get last fetch time for %s", name))
	}

	usedBy := usageMap[name]

	return &domain.CanonicalRepoStatus{
		Name:           name,
		Path:           path,
		DiskUsageBytes: size,
		LastFetchTime:  lastFetch,
		UsedByCount:    len(usedBy),
		UsedBy:         usedBy,
	}, nil
}

func (s *Service) buildRepoUsageMap(ctx context.Context) (map[string][]string, error) {
	workspaces, err := s.wsEngine.List(ctx)
	if err != nil {
		return nil, cerrors.NewIOFailed("list workspaces", err)
	}

	usageMap := make(map[string][]string)

	for _, ws := range workspaces {
		for _, repo := range ws.Repos {
			usageMap[repo.Name] = append(usageMap[repo.Name], ws.ID)
		}
	}

	return usageMap, nil
}
