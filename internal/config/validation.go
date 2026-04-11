package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// Validate validates configuration values and environment prerequisites and returns an error when invalid.
func (c *Config) Validate() error {
	if err := c.ValidateValues(); err != nil {
		return err
	}

	return c.ValidateEnvironment()
}

// ValidateValues validates configuration values without performing filesystem checks and returns an error when invalid.
func (c *Config) ValidateValues() error {
	if err := c.validateWorkspaceSettings(); err != nil {
		return err
	}

	if err := c.validateRuntimeSettings(); err != nil {
		return err
	}

	return c.validateKeybindings()
}

// ValidateTemplates validates configured workspace templates and returns an error when any template is invalid.
func (c *Config) ValidateTemplates() error {
	return c.validateTemplates()
}

func (c *Config) validateKeybindings() error {
	return c.TUI.Keybindings.WithDefaults().ValidateKeybindings()
}

func (c *Config) validateRequiredFields() error {
	if err := validateRequiredField("projects_root", c.ProjectsRoot); err != nil {
		return err
	}

	if err := validateRequiredField("workspaces_root", c.WorkspacesRoot); err != nil {
		return err
	}

	return validateRequiredField("closed_root", c.ClosedRoot)
}

func (c *Config) validateCloseDefault() error {
	if c.CloseDefault == "" {
		c.CloseDefault = CloseDefaultDelete
	}

	if c.CloseDefault != CloseDefaultDelete && c.CloseDefault != CloseDefaultArchive {
		return cerrors.NewConfigValidation("workspace_close_default", fmt.Sprintf("must be either '%s' or '%s', got %q", CloseDefaultDelete, CloseDefaultArchive, c.CloseDefault))
	}

	return nil
}

func (c *Config) validateWorkspaceSettings() error {
	if err := c.validateRequiredFields(); err != nil {
		return err
	}

	if err := c.validateCloseDefault(); err != nil {
		return err
	}

	if err := c.validateWorkspaceNaming(); err != nil {
		return err
	}

	if err := c.validatePatterns(); err != nil {
		return err
	}

	return c.validateTemplates()
}

func (c *Config) validateRuntimeSettings() error {
	if err := c.validateStaleThreshold(); err != nil {
		return err
	}

	if err := c.validateHooks(); err != nil {
		return err
	}

	if err := c.validateGitRetry(); err != nil {
		return err
	}

	if err := c.validateParallelWorkers(); err != nil {
		return err
	}

	return c.validateLockSettings()
}

func (c *Config) validateWorkspaceNaming() error {
	if strings.TrimSpace(c.WorkspaceNaming) == "" {
		c.WorkspaceNaming = "{{.ID}}"
	}

	_, err := c.computeWorkspaceDir("EXAMPLE-123")

	return err
}

func (c *Config) validatePatterns() error {
	for _, p := range c.Defaults.WorkspacePatterns {
		if _, err := regexp.Compile(p.Pattern); err != nil {
			return cerrors.NewConfigValidation("workspace_patterns", fmt.Sprintf("invalid regex pattern '%s': %v", p.Pattern, err))
		}
	}

	return nil
}

func (c *Config) validateTemplates() error {
	for name, tmpl := range c.Templates {
		if err := validateTemplate(name, tmpl); err != nil {
			return err
		}
	}

	return nil
}

func validateTemplate(name string, tmpl Template) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return cerrors.NewConfigValidation("templates", "template name cannot be empty")
	}

	if trimmedName != name {
		return cerrors.NewConfigValidation("templates", fmt.Sprintf("template name %q must not contain leading or trailing whitespace", name))
	}

	if len(tmpl.Repos) == 0 {
		return cerrors.NewConfigValidation(fmt.Sprintf("templates.%s.repos", name), "must define at least one repo")
	}

	for i, repo := range tmpl.Repos {
		if strings.TrimSpace(repo) == "" {
			return cerrors.NewConfigValidation(fmt.Sprintf("templates.%s.repos", name), fmt.Sprintf("repo at index %d is empty", i))
		}
	}

	if tmpl.DefaultBranch != "" {
		if err := validation.ValidateBranchName(tmpl.DefaultBranch); err != nil {
			return cerrors.NewConfigValidation(fmt.Sprintf("templates.%s.default_branch", name), err.Error())
		}
	}

	for i, cmd := range tmpl.SetupCommands {
		if strings.TrimSpace(cmd) == "" {
			return cerrors.NewConfigValidation(fmt.Sprintf("templates.%s.setup_commands", name), fmt.Sprintf("command at index %d is empty", i))
		}
	}

	return nil
}

func (c *Config) validateStaleThreshold() error {
	if c.StaleThresholdDays < 0 {
		return cerrors.NewConfigValidation("stale_threshold_days", fmt.Sprintf("must be zero or positive, got %d", c.StaleThresholdDays))
	}

	return nil
}

func (c *Config) validateParallelWorkers() error {
	if c.ParallelWorkers < MinParallelWorkers {
		return cerrors.NewConfigValidation("parallel_workers", fmt.Sprintf("must be at least %d, got %d", MinParallelWorkers, c.ParallelWorkers))
	}

	if c.ParallelWorkers > MaxParallelWorkers {
		return cerrors.NewConfigValidation("parallel_workers", fmt.Sprintf("must not exceed %d, got %d", MaxParallelWorkers, c.ParallelWorkers))
	}

	return nil
}

func (c *Config) validateLockSettings() error {
	if c.LockTimeout == "" {
		c.LockTimeout = DefaultLockTimeout.String()
	}

	if c.LockStaleThreshold == "" {
		c.LockStaleThreshold = DefaultLockStaleThreshold.String()
	}

	lockTimeout, err := time.ParseDuration(c.LockTimeout)
	if err != nil {
		return cerrors.NewConfigValidation("lock_timeout", fmt.Sprintf("invalid duration %q: %v", c.LockTimeout, err))
	}

	if lockTimeout <= 0 {
		return cerrors.NewConfigValidation("lock_timeout", fmt.Sprintf("must be positive, got %s", lockTimeout))
	}

	lockStale, err := time.ParseDuration(c.LockStaleThreshold)
	if err != nil {
		return cerrors.NewConfigValidation("lock_stale_threshold", fmt.Sprintf("invalid duration %q: %v", c.LockStaleThreshold, err))
	}

	if lockStale <= 0 {
		return cerrors.NewConfigValidation("lock_stale_threshold", fmt.Sprintf("must be positive, got %s", lockStale))
	}

	return nil
}

// Configuration defaults and bounds for workspace closing, worker counts, and lock timings.
const (
	CloseDefaultDelete        = "delete"
	CloseDefaultArchive       = "archive"
	DefaultParallelWorkers    = 4
	DefaultLockTimeout        = 30 * time.Second
	DefaultLockStaleThreshold = 5 * time.Minute
	MinParallelWorkers        = 1
	MaxParallelWorkers        = 10
)

const maxRetryAttempts = 10

func (c *Config) validateGitRetry() error {
	retry := c.Git.Retry
	if err := validateRetryAttempts(retry.MaxAttempts); err != nil {
		return err
	}

	initialDelay, err := validatePositiveDuration("git.retry.initial_delay", retry.InitialDelay)
	if err != nil {
		return err
	}

	maxDelay, err := validatePositiveDuration("git.retry.max_delay", retry.MaxDelay)
	if err != nil {
		return err
	}

	if initialDelay > maxDelay {
		return cerrors.NewConfigValidation("git.retry.initial_delay", fmt.Sprintf("(%s) must not exceed max_delay (%s)", retry.InitialDelay, retry.MaxDelay))
	}

	if err := validateRetryMultiplier(retry.Multiplier); err != nil {
		return err
	}

	return validateJitterFactor(retry.JitterFactor)
}

func validateRetryAttempts(maxAttempts int) error {
	if maxAttempts < 1 {
		return cerrors.NewConfigValidation("git.retry.max_attempts", fmt.Sprintf("must be at least 1, got %d", maxAttempts))
	}

	if maxAttempts > maxRetryAttempts {
		return cerrors.NewConfigValidation("git.retry.max_attempts", fmt.Sprintf("must not exceed %d, got %d", maxRetryAttempts, maxAttempts))
	}

	return nil
}

func validatePositiveDuration(field, raw string) (time.Duration, error) {
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, cerrors.NewConfigValidation(field, fmt.Sprintf("invalid: %v", err))
	}

	if duration <= 0 {
		return 0, cerrors.NewConfigValidation(field, fmt.Sprintf("must be positive, got %s", raw))
	}

	return duration, nil
}

func validateRetryMultiplier(multiplier float64) error {
	if multiplier < 1.0 {
		return cerrors.NewConfigValidation("git.retry.multiplier", fmt.Sprintf("must be at least 1.0, got %f", multiplier))
	}

	return nil
}

func validateJitterFactor(jitterFactor float64) error {
	if jitterFactor < 0 || jitterFactor > 1 {
		return cerrors.NewConfigValidation("git.retry.jitter_factor", fmt.Sprintf("must be between 0 and 1, got %f", jitterFactor))
	}

	return nil
}

func (c *Config) validateHooks() error {
	for i, h := range c.Hooks.PostCreate {
		if err := validateHook(h, "post_create", i); err != nil {
			return err
		}
	}

	for i, h := range c.Hooks.PreClose {
		if err := validateHook(h, "pre_close", i); err != nil {
			return err
		}
	}

	return nil
}

func validateHook(h Hook, hookType string, index int) error {
	field := fmt.Sprintf("%s hook[%d]", hookType, index)
	if strings.TrimSpace(h.Command) == "" {
		return cerrors.NewConfigValidation(field, "command cannot be empty")
	}

	if strings.Contains(h.Command, "\x00") {
		return cerrors.NewConfigValidation(field, "command contains invalid null byte")
	}

	if strings.ContainsAny(h.Command, "\n\r") {
		return cerrors.NewConfigValidation(field, "command cannot contain newlines")
	}

	if h.Timeout < 0 {
		return cerrors.NewConfigValidation(field, fmt.Sprintf("timeout must be non-negative, got %d", h.Timeout))
	}

	if h.Shell != "" && strings.TrimSpace(h.Shell) == "" {
		return cerrors.NewConfigValidation(field, "shell cannot be empty or whitespace-only when specified")
	}

	return nil
}

// ValidateEnvironment validates configured root paths against the local filesystem.
func (c *Config) ValidateEnvironment() error {
	if err := validateRootPath("projects_root", c.ProjectsRoot); err != nil {
		return err
	}

	if err := validateRootPath("workspaces_root", c.WorkspacesRoot); err != nil {
		return err
	}

	if err := validateRootPath("closed_root", c.ClosedRoot); err != nil {
		return err
	}

	return nil
}

func validateRequiredField(label, value string) error {
	if value == "" {
		return cerrors.NewConfigValidation(label, "is required")
	}

	return nil
}

func validateRootPath(label, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if !filepath.IsAbs(path) {
				return cerrors.NewPathInvalid(path, fmt.Sprintf("%s must be an absolute path", label))
			}

			return nil
		}

		return err
	}

	if !info.IsDir() {
		return cerrors.NewPathNotDirectory(path)
	}

	return nil
}
