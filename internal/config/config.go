package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Default keybindings for TUI actions.
var (
	DefaultQuitKeys        = []string{"q", "ctrl+c"}
	DefaultSearchKeys      = []string{"/"}
	DefaultPushKeys        = []string{"p"}
	DefaultCloseKeys       = []string{"c"}
	DefaultOpenEditorKeys  = []string{"o"}
	DefaultToggleStaleKeys = []string{"s"}
	DefaultDetailsKeys     = []string{"enter"}
	DefaultConfirmKeys     = []string{"y", "Y"}
	DefaultCancelKeys      = []string{"n", "N", "esc"}
)

// Hook defines a single lifecycle hook command.
type Hook struct {
	Command         string   `mapstructure:"command"`
	Repos           []string `mapstructure:"repos,omitempty"`             // filter to specific repos
	Shell           string   `mapstructure:"shell,omitempty"`             // default: sh -c
	Timeout         int      `mapstructure:"timeout,omitempty"`           // default: 30 seconds
	ContinueOnError bool     `mapstructure:"continue_on_error,omitempty"` // don't fail workspace operation
}

// Hooks holds lifecycle hook configurations.
type Hooks struct {
	PostCreate []Hook `mapstructure:"post_create"`
	PreClose   []Hook `mapstructure:"pre_close"`
}

// Keybindings holds TUI keybinding configurations.
type Keybindings struct {
	Quit        []string `mapstructure:"quit"`
	Search      []string `mapstructure:"search"`
	Push        []string `mapstructure:"push"`
	Close       []string `mapstructure:"close"`
	OpenEditor  []string `mapstructure:"open_editor"`
	ToggleStale []string `mapstructure:"toggle_stale"`
	Details     []string `mapstructure:"details"`
	Confirm     []string `mapstructure:"confirm"`
	Cancel      []string `mapstructure:"cancel"`
}

// TUIConfig holds TUI-specific configuration.
type TUIConfig struct {
	Keybindings Keybindings `mapstructure:"keybindings"`
}

// RetryConfig holds retry configuration for network operations.
type RetryConfig struct {
	MaxAttempts  int     `mapstructure:"max_attempts"`
	InitialDelay string  `mapstructure:"initial_delay"` // Duration string, e.g. "1s"
	MaxDelay     string  `mapstructure:"max_delay"`     // Duration string, e.g. "30s"
	Multiplier   float64 `mapstructure:"multiplier"`
	JitterFactor float64 `mapstructure:"jitter_factor"`
}

// ParsedRetryConfig holds the parsed retry configuration with proper Go types.
type ParsedRetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	JitterFactor float64
}

// Parse converts the string-based RetryConfig to ParsedRetryConfig with proper duration types.
func (r RetryConfig) Parse() (ParsedRetryConfig, error) {
	initialDelay, err := time.ParseDuration(r.InitialDelay)
	if err != nil {
		return ParsedRetryConfig{}, fmt.Errorf("invalid initial_delay %q: %w", r.InitialDelay, err)
	}

	maxDelay, err := time.ParseDuration(r.MaxDelay)
	if err != nil {
		return ParsedRetryConfig{}, fmt.Errorf("invalid max_delay %q: %w", r.MaxDelay, err)
	}

	return ParsedRetryConfig{
		MaxAttempts:  r.MaxAttempts,
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   r.Multiplier,
		JitterFactor: r.JitterFactor,
	}, nil
}

// GitConfig holds git-related configuration.
type GitConfig struct {
	Retry RetryConfig `mapstructure:"retry"`
}

// Config holds the global configuration
type Config struct {
	ProjectsRoot       string        `mapstructure:"projects_root"`
	WorkspacesRoot     string        `mapstructure:"workspaces_root"`
	ClosedRoot         string        `mapstructure:"closed_root"`
	CloseDefault       string        `mapstructure:"workspace_close_default"`
	WorkspaceNaming    string        `mapstructure:"workspace_naming"`
	StaleThresholdDays int           `mapstructure:"stale_threshold_days"`
	Defaults           Defaults      `mapstructure:"defaults"`
	Hooks              Hooks         `mapstructure:"hooks"`
	TUI                TUIConfig     `mapstructure:"tui"`
	Git                GitConfig     `mapstructure:"git"`
	Registry           *RepoRegistry `mapstructure:"-"`
}

// WorkspacePattern defines a regex pattern and default repos
type WorkspacePattern struct {
	Pattern string   `mapstructure:"pattern"`
	Repos   []string `mapstructure:"repos"`
}

// Defaults holds default configurations
type Defaults struct {
	WorkspacePatterns []WorkspacePattern `mapstructure:"workspace_patterns"`
}

// Load initializes and loads the configuration
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(home, ".canopy"))
	viper.AddConfigPath(filepath.Join(home, ".config", "canopy"))

	viper.SetDefault("projects_root", filepath.Join(home, ".canopy", "projects"))
	viper.SetDefault("workspaces_root", filepath.Join(home, ".canopy", "workspaces"))
	viper.SetDefault("closed_root", filepath.Join(home, ".canopy", "closed"))
	viper.SetDefault("workspace_close_default", "delete")
	viper.SetDefault("workspace_naming", "{{.ID}}")
	viper.SetDefault("stale_threshold_days", 14)

	// Git retry defaults
	viper.SetDefault("git.retry.max_attempts", 3)
	viper.SetDefault("git.retry.initial_delay", "1s")
	viper.SetDefault("git.retry.max_delay", "30s")
	viper.SetDefault("git.retry.multiplier", 2.0)
	viper.SetDefault("git.retry.jitter_factor", 0.25)

	viper.SetEnvPrefix("CANOPY")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is okay, use defaults
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand tilde
	cfg.ProjectsRoot = expandPath(cfg.ProjectsRoot, home)
	cfg.WorkspacesRoot = expandPath(cfg.WorkspacesRoot, home)
	cfg.ClosedRoot = expandPath(cfg.ClosedRoot, home)
	cfg.CloseDefault = strings.ToLower(cfg.CloseDefault)

	registry, err := LoadRepoRegistry("")
	if err != nil {
		return nil, fmt.Errorf("failed to load repository registry: %w", err)
	}

	cfg.Registry = registry

	return &cfg, nil
}

func expandPath(path, home string) string {
	if path == "~" {
		return home
	}

	if len(path) > 1 && path[:2] == "~/" {
		return filepath.Join(home, path[2:])
	}

	return path
}

// GetReposForWorkspace returns default repos for a given workspace ID based on patterns
func (c *Config) GetReposForWorkspace(workspaceID string) []string {
	for _, p := range c.Defaults.WorkspacePatterns {
		matched, err := regexp.MatchString(p.Pattern, workspaceID)
		if err == nil && matched {
			return p.Repos
		}
	}

	return nil
}

// Validate performs complete configuration validation by first checking values
// (pure validation) and then verifying the environment (filesystem checks).
// This is the main validation entry point that maintains backward compatibility.
func (c *Config) Validate() error {
	if err := c.ValidateValues(); err != nil {
		return err
	}

	return c.ValidateEnvironment()
}

// ValidateValues performs pure validation of configuration values without any
// filesystem or I/O operations. This includes checking required fields, enum
// values, regex patterns, and numeric constraints. Use this method when you
// need fast validation that doesn't depend on the environment.
func (c *Config) ValidateValues() error {
	if err := c.validateRequiredFields(); err != nil {
		return err
	}

	if err := c.validateCloseDefault(); err != nil {
		return err
	}

	if err := c.validatePatterns(); err != nil {
		return err
	}

	if err := c.validateStaleThreshold(); err != nil {
		return err
	}

	if err := c.validateHooks(); err != nil {
		return err
	}

	if err := c.validateGitRetry(); err != nil {
		return err
	}

	return c.validateKeybindings()
}

// validateKeybindings validates the TUI keybindings configuration.
func (c *Config) validateKeybindings() error {
	// Apply defaults first, then validate for conflicts
	kb := c.TUI.Keybindings.WithDefaults()
	return kb.ValidateKeybindings()
}

// validateRequiredFields checks that all required configuration fields are set.
func (c *Config) validateRequiredFields() error {
	if err := validateRequiredField("projects_root", c.ProjectsRoot); err != nil {
		return err
	}

	if err := validateRequiredField("workspaces_root", c.WorkspacesRoot); err != nil {
		return err
	}

	return validateRequiredField("closed_root", c.ClosedRoot)
}

// validateCloseDefault validates and applies default for the close behavior.
func (c *Config) validateCloseDefault() error {
	if c.CloseDefault == "" {
		c.CloseDefault = "delete"
	}

	if c.CloseDefault != "delete" && c.CloseDefault != "archive" {
		return fmt.Errorf("workspace_close_default must be either 'delete' or 'archive', got %q", c.CloseDefault)
	}

	return nil
}

// validatePatterns checks that all workspace regex patterns are valid.
func (c *Config) validatePatterns() error {
	for _, p := range c.Defaults.WorkspacePatterns {
		if _, err := regexp.Compile(p.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", p.Pattern, err)
		}
	}

	return nil
}

// validateStaleThreshold checks that the stale threshold is non-negative.
func (c *Config) validateStaleThreshold() error {
	if c.StaleThresholdDays < 0 {
		return fmt.Errorf("stale_threshold_days must be zero or positive, got %d", c.StaleThresholdDays)
	}

	return nil
}

// validateGitRetry validates the git retry configuration.
func (c *Config) validateGitRetry() error {
	retry := c.Git.Retry

	if retry.MaxAttempts < 1 {
		return fmt.Errorf("git.retry.max_attempts must be at least 1, got %d", retry.MaxAttempts)
	}

	if _, err := time.ParseDuration(retry.InitialDelay); err != nil {
		return fmt.Errorf("git.retry.initial_delay is invalid: %w", err)
	}

	if _, err := time.ParseDuration(retry.MaxDelay); err != nil {
		return fmt.Errorf("git.retry.max_delay is invalid: %w", err)
	}

	if retry.Multiplier < 1.0 {
		return fmt.Errorf("git.retry.multiplier must be at least 1.0, got %f", retry.Multiplier)
	}

	if retry.JitterFactor < 0 || retry.JitterFactor > 1 {
		return fmt.Errorf("git.retry.jitter_factor must be between 0 and 1, got %f", retry.JitterFactor)
	}

	return nil
}

// validateHooks validates all hook configurations.
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

// validateHook checks that a hook has valid configuration.
func validateHook(h Hook, hookType string, index int) error {
	if strings.TrimSpace(h.Command) == "" {
		return fmt.Errorf("%s hook[%d] command cannot be empty", hookType, index)
	}

	if strings.Contains(h.Command, "\x00") {
		return fmt.Errorf("%s hook[%d] command contains invalid null byte", hookType, index)
	}

	if strings.ContainsAny(h.Command, "\n\r") {
		return fmt.Errorf("%s hook[%d] command cannot contain newlines", hookType, index)
	}

	if h.Timeout < 0 {
		return fmt.Errorf("%s hook[%d] timeout must be non-negative, got %d", hookType, index, h.Timeout)
	}

	return nil
}

// ValidateEnvironment verifies that the configuration's filesystem paths exist
// and are directories. This method performs I/O operations and should be called
// after ValidateValues() when you need to ensure the environment is ready.
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

// validateRequiredField checks that a field value is non-empty.
func validateRequiredField(label, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", label)
	}

	return nil
}

// validateRootPath checks that a path exists and is a directory.
// If the path doesn't exist but is absolute, it's considered valid
// (it will be created later). Non-absolute paths that don't exist are invalid.
func validateRootPath(label, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if !filepath.IsAbs(path) {
				return fmt.Errorf("%s must be an absolute path: %s", label, path)
			}

			return nil
		}

		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory: %s", label, path)
	}

	return nil
}

// GetProjectsRoot returns the projects root directory.
func (c *Config) GetProjectsRoot() string {
	return c.ProjectsRoot
}

// GetWorkspacesRoot returns the workspaces root directory.
func (c *Config) GetWorkspacesRoot() string {
	return c.WorkspacesRoot
}

// GetClosedRoot returns the closed workspaces root directory.
func (c *Config) GetClosedRoot() string {
	return c.ClosedRoot
}

// GetCloseDefault returns the default close behavior.
func (c *Config) GetCloseDefault() string {
	return c.CloseDefault
}

// GetWorkspaceNaming returns the workspace naming pattern.
func (c *Config) GetWorkspaceNaming() string {
	return c.WorkspaceNaming
}

// GetStaleThresholdDays returns the stale threshold in days.
func (c *Config) GetStaleThresholdDays() int {
	return c.StaleThresholdDays
}

// GetRegistry returns the repository registry.
func (c *Config) GetRegistry() *RepoRegistry {
	return c.Registry
}

// GetHooks returns the lifecycle hooks configuration.
func (c *Config) GetHooks() Hooks {
	return c.Hooks
}

// GetTUI returns the TUI configuration.
func (c *Config) GetTUI() TUIConfig {
	return c.TUI
}

// GetKeybindings returns the TUI keybindings with defaults applied.
func (c *Config) GetKeybindings() Keybindings {
	return c.TUI.Keybindings.WithDefaults()
}

// copyKeys creates a copy of a string slice to avoid sharing references.
func copyKeys(keys []string) []string {
	if keys == nil {
		return nil
	}

	result := make([]string, len(keys))
	copy(result, keys)

	return result
}

// WithDefaults returns a copy of Keybindings with defaults applied for empty fields.
// Returned slices are copies to prevent mutation of global defaults.
func (k Keybindings) WithDefaults() Keybindings {
	result := k

	if len(result.Quit) == 0 {
		result.Quit = copyKeys(DefaultQuitKeys)
	}

	if len(result.Search) == 0 {
		result.Search = copyKeys(DefaultSearchKeys)
	}

	if len(result.Push) == 0 {
		result.Push = copyKeys(DefaultPushKeys)
	}

	if len(result.Close) == 0 {
		result.Close = copyKeys(DefaultCloseKeys)
	}

	if len(result.OpenEditor) == 0 {
		result.OpenEditor = copyKeys(DefaultOpenEditorKeys)
	}

	if len(result.ToggleStale) == 0 {
		result.ToggleStale = copyKeys(DefaultToggleStaleKeys)
	}

	if len(result.Details) == 0 {
		result.Details = copyKeys(DefaultDetailsKeys)
	}

	if len(result.Confirm) == 0 {
		result.Confirm = copyKeys(DefaultConfirmKeys)
	}

	if len(result.Cancel) == 0 {
		result.Cancel = copyKeys(DefaultCancelKeys)
	}

	return result
}

// validKeys is the set of recognized key names for TUI keybindings.
var validKeys = map[string]bool{
	// Letters
	"a": true, "b": true, "c": true, "d": true, "e": true, "f": true, "g": true, "h": true,
	"i": true, "j": true, "k": true, "l": true, "m": true, "n": true, "o": true, "p": true,
	"q": true, "r": true, "s": true, "t": true, "u": true, "v": true, "w": true, "x": true,
	"y": true, "z": true,
	// Uppercase letters (for confirm/cancel dialogs)
	"A": true, "B": true, "C": true, "D": true, "E": true, "F": true, "G": true, "H": true,
	"I": true, "J": true, "K": true, "L": true, "M": true, "N": true, "O": true, "P": true,
	"Q": true, "R": true, "S": true, "T": true, "U": true, "V": true, "W": true, "X": true,
	"Y": true, "Z": true,
	// Numbers
	"0": true, "1": true, "2": true, "3": true, "4": true,
	"5": true, "6": true, "7": true, "8": true, "9": true,
	// Special keys
	"enter": true, "esc": true, "tab": true, "backspace": true, "delete": true,
	"up": true, "down": true, "left": true, "right": true,
	"home": true, "end": true, "pgup": true, "pgdown": true,
	"space": true,
	// Function keys
	"f1": true, "f2": true, "f3": true, "f4": true, "f5": true, "f6": true,
	"f7": true, "f8": true, "f9": true, "f10": true, "f11": true, "f12": true,
	// Symbols
	"/": true, "\\": true, ".": true, ",": true, ";": true, "'": true, "`": true,
	"[": true, "]": true, "-": true, "=": true,
}

// isValidKey checks if a key string is a recognized keybinding.
func isValidKey(key string) bool {
	if key == "" {
		return false
	}

	// Check for modifier combinations (ctrl+x, alt+x, shift+x)
	for _, prefix := range []string{"ctrl+", "alt+", "shift+"} {
		if strings.HasPrefix(key, prefix) {
			base := strings.TrimPrefix(key, prefix)
			return isValidKey(base)
		}
	}

	return validKeys[key]
}

// ValidateKeybindings checks for invalid and conflicting keybindings.
// Returns an error listing all issues found.
func (k Keybindings) ValidateKeybindings() error {
	var errors []string

	// Validate each key is a recognized format
	validateKeys := func(keys []string, action string) {
		for _, key := range keys {
			if !isValidKey(key) {
				errors = append(errors, fmt.Sprintf("invalid key %q for action %q", key, action))
			}
		}
	}

	validateKeys(k.Quit, "quit")
	validateKeys(k.Search, "search")
	validateKeys(k.Push, "push")
	validateKeys(k.Close, "close")
	validateKeys(k.OpenEditor, "open_editor")
	validateKeys(k.ToggleStale, "toggle_stale")
	validateKeys(k.Details, "details")
	validateKeys(k.Confirm, "confirm")
	validateKeys(k.Cancel, "cancel")

	// Map key -> list of actions using that key
	keyUsage := make(map[string][]string)

	addKeys := func(keys []string, action string) {
		for _, key := range keys {
			keyUsage[key] = append(keyUsage[key], action)
		}
	}

	addKeys(k.Quit, "quit")
	addKeys(k.Search, "search")
	addKeys(k.Push, "push")
	addKeys(k.Close, "close")
	addKeys(k.OpenEditor, "open_editor")
	addKeys(k.ToggleStale, "toggle_stale")
	addKeys(k.Details, "details")
	addKeys(k.Confirm, "confirm")
	addKeys(k.Cancel, "cancel")

	// Find conflicts (sort keys for deterministic output)
	var conflictKeys []string

	for key, actions := range keyUsage {
		if len(actions) > 1 {
			conflictKeys = append(conflictKeys, key)
		}
	}

	sort.Strings(conflictKeys)

	for _, key := range conflictKeys {
		sort.Strings(keyUsage[key])
		errors = append(errors, fmt.Sprintf("key %q is assigned to multiple actions: %s",
			key, strings.Join(keyUsage[key], ", ")))
	}

	if len(errors) > 0 {
		return fmt.Errorf("keybinding validation errors:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}
