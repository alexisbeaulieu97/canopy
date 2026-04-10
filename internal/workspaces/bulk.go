package workspaces

import (
	"context"
	"regexp"
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// ListWorkspacesMatching returns workspaces with IDs that match the regex pattern.
func (s *Service) ListWorkspacesMatching(ctx context.Context, pattern string) ([]domain.Workspace, error) {
	re, err := compileWorkspacePattern(pattern)
	if err != nil {
		return nil, err
	}

	workspaces, err := s.wsEngine.List(ctx)
	if err != nil {
		return nil, err
	}

	var matched []domain.Workspace

	for _, ws := range workspaces {
		if re.MatchString(ws.ID) {
			matched = append(matched, ws)
		}
	}

	return matched, nil
}

func compileWorkspacePattern(pattern string) (*regexp.Regexp, error) {
	if strings.TrimSpace(pattern) == "" {
		return nil, cerrors.NewInvalidArgument("pattern", "pattern is required")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, cerrors.NewInvalidArgument("pattern", err.Error())
	}

	return re, nil
}
