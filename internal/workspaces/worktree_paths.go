package workspaces

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

var errInvalidGitWorktreeLink = errors.New("invalid git worktree link format")

type invalidGitWorktreeLinkError struct {
	line string
}

func (e invalidGitWorktreeLinkError) Error() string {
	return errInvalidGitWorktreeLink.Error()
}

func (e invalidGitWorktreeLinkError) Unwrap() error {
	return errInvalidGitWorktreeLink
}

func (e invalidGitWorktreeLinkError) Line() string {
	return e.line
}

func workspacePath(root, dirName string) string {
	return filepath.Join(root, dirName)
}

func workspaceRepoPath(root, dirName, repoName string) string {
	return filepath.Join(workspacePath(root, dirName), repoName)
}

func canonicalRepoPath(root, repoName string) string {
	return filepath.Join(root, repoName)
}

func worktreeGitPath(worktreePath string) string {
	return filepath.Join(worktreePath, ".git")
}

func statWorktreeGitPath(worktreePath string) (string, os.FileInfo, error) {
	gitPath := worktreeGitPath(worktreePath)

	info, err := os.Stat(gitPath)
	if err != nil {
		return gitPath, nil, err
	}

	return gitPath, info, nil
}

func resolveWorktreeGitDir(worktreePath string) (string, error) {
	gitPath, info, err := statWorktreeGitPath(worktreePath)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return gitPath, nil
	}

	content, err := os.ReadFile(gitPath) //nolint:gosec // G304: gitPath is constructed from workspace/repo paths, not user input
	if err != nil {
		return "", err
	}

	gitdirLine := strings.TrimSpace(string(content))
	if !strings.HasPrefix(gitdirLine, "gitdir:") {
		return "", invalidGitWorktreeLinkError{line: gitdirLine}
	}

	gitdirPath := strings.TrimSpace(strings.TrimPrefix(gitdirLine, "gitdir:"))
	if gitdirPath == "" {
		return "", invalidGitWorktreeLinkError{line: gitdirLine}
	}

	return resolveGitdirPath(gitdirPath, worktreePath), nil
}

func resolveWorktreeGitConfigPath(worktreePath string) (string, error) {
	gitdirPath, err := resolveWorktreeGitDir(worktreePath)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(gitdirPath); err != nil {
		return "", fmt.Errorf("resolved gitdir does not exist: %w", err)
	}

	return resolveGitConfigPath(gitdirPath), nil
}

func resolveGitdirPath(gitdirPath, worktreePath string) string {
	if filepath.IsAbs(gitdirPath) {
		return gitdirPath
	}

	return filepath.Join(worktreePath, gitdirPath)
}

func workspaceHookContext(root, workspaceID, dirName, branchName string, repos []domain.Repo) domain.HookContext {
	return domain.HookContext{
		WorkspaceID:   workspaceID,
		WorkspacePath: workspacePath(root, dirName),
		BranchName:    branchName,
		Repos:         repos,
	}
}
