package workspaces

import (
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// GitHubShorthandStrategy resolves GitHub shorthand (owner/repo) to full URLs.
type GitHubShorthandStrategy struct{}

// NewGitHubShorthandStrategy creates a new GitHub shorthand resolution strategy.
func NewGitHubShorthandStrategy() *GitHubShorthandStrategy {
	return &GitHubShorthandStrategy{}
}

// Name returns the strategy name for debugging.
func (s *GitHubShorthandStrategy) Name() string {
	return "github-shorthand"
}

// Resolve attempts to resolve a GitHub shorthand (owner/repo) to a repository.
func (s *GitHubShorthandStrategy) Resolve(input string) (domain.Repo, bool) {
	// Must have exactly one slash
	if strings.Count(input, "/") != 1 {
		return domain.Repo{}, false
	}

	parts := strings.Split(input, "/")
	owner := strings.TrimSpace(parts[0])
	repo := strings.TrimSpace(parts[1])

	// Both owner and repo must be non-empty for valid GitHub shorthand
	if owner == "" || repo == "" {
		return domain.Repo{}, false
	}

	url := "https://github.com/" + owner + "/" + repo

	return domain.Repo{Name: repo, URL: url}, true
}
