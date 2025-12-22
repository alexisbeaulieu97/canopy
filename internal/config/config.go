// Package config provides configuration loading and management for Canopy.
//
// # Configuration Loading Priority
//
// Configuration is loaded with the following priority (highest to lowest):
//  1. Explicit --config flag path
//  2. CANOPY_CONFIG environment variable
//  3. Default search paths (in order):
//     - ./config.yaml (current directory)
//     - ~/.canopy/config.yaml
//     - ~/.config/canopy/config.yaml
//
// When an explicit config path is provided via --config flag or CANOPY_CONFIG
// environment variable, the file must exist or loading will fail. Default search
// paths are optional - if no config file is found, defaults are used.
//
// Paths support tilde (~) expansion to the user's home directory.
//
// Environment variables with the CANOPY_ prefix can override configuration values.
//
// # Configuration Options
//
// Key configuration options include:
//   - projects_root: Directory for canonical bare repositories
//   - workspaces_root: Directory for active workspaces
//   - closed_root: Directory for archived workspace metadata
//   - workspace_close_default: Default close behavior ("delete" or "archive")
//   - stale_threshold_days: Days before a workspace is considered stale
//
// # Workspace Patterns
//
// Workspace patterns allow automatic repository assignment based on workspace ID:
//
//	defaults:
//	  workspace_patterns:
//	    - pattern: "^PROJ-"
//	      repos: ["backend", "frontend"]
//
// # Lifecycle Hooks
//
// Hooks execute commands at workspace lifecycle events:
//
//	hooks:
//	  post_create:
//	    - command: "npm install"
//	      repos: ["frontend"]
//	  pre_close:
//	    - command: "git stash"
//
// See the configuration documentation for complete reference.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
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
	UseEmoji    *bool       `mapstructure:"use_emoji"` // nil means default (true)
}

// GetUseEmoji returns whether emoji should be used in the TUI.
// Defaults to true for backward compatibility.
func (t TUIConfig) GetUseEmoji() bool {
	if t.UseEmoji == nil {
		return true
	}

	return *t.UseEmoji
}

// GitRetrySettings holds YAML configuration for git network operation retry behavior.
// This is the config-file representation; use ParsedRetryConfig for runtime use.
type GitRetrySettings struct {
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

// Parse converts the string-based GitRetrySettings to ParsedRetryConfig with proper duration types.
func (r GitRetrySettings) Parse() (ParsedRetryConfig, error) {
	initialDelay, err := time.ParseDuration(r.InitialDelay)
	if err != nil {
		return ParsedRetryConfig{}, cerrors.NewConfigValidation("git.retry.initial_delay", fmt.Sprintf("invalid duration %q: %v", r.InitialDelay, err))
	}

	maxDelay, err := time.ParseDuration(r.MaxDelay)
	if err != nil {
		return ParsedRetryConfig{}, cerrors.NewConfigValidation("git.retry.max_delay", fmt.Sprintf("invalid duration %q: %v", r.MaxDelay, err))
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
	Retry GitRetrySettings `mapstructure:"retry"`
}

// Config holds the global configuration
type Config struct {
	ProjectsRoot       string              `mapstructure:"projects_root"`
	WorkspacesRoot     string              `mapstructure:"workspaces_root"`
	ClosedRoot         string              `mapstructure:"closed_root"`
	CloseDefault       string              `mapstructure:"workspace_close_default"`
	WorkspaceNaming    string              `mapstructure:"workspace_naming"`
	StaleThresholdDays int                 `mapstructure:"stale_threshold_days"`
	ParallelWorkers    int                 `mapstructure:"parallel_workers"`
	LockTimeout        string              `mapstructure:"lock_timeout"`
	LockStaleThreshold string              `mapstructure:"lock_stale_threshold"`
	Defaults           Defaults            `mapstructure:"defaults"`
	Templates          map[string]Template `mapstructure:"templates"`
	Hooks              Hooks               `mapstructure:"hooks"`
	TUI                TUIConfig           `mapstructure:"tui"`
	Git                GitConfig           `mapstructure:"git"`
	Registry           *RepoRegistry       `mapstructure:"-"`
	Warnings           []string            `mapstructure:"-"` // Warnings collected during loading (e.g., deprecated keys)
}

// WorkspaceNamingTemplateData defines the data available to workspace naming templates.
type WorkspaceNamingTemplateData struct {
	ID string
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

// Template defines reusable workspace defaults.
type Template struct {
	Name          string   `mapstructure:"-"`
	Repos         []string `mapstructure:"repos"`
	DefaultBranch string   `mapstructure:"default_branch"`
	Description   string   `mapstructure:"description"`
	SetupCommands []string `mapstructure:"setup_commands"`
}

// knownConfigFields contains all valid top-level and nested config field names
// for providing suggestions when unknown fields are detected.
var knownConfigFields = []string{
	"projects_root",
	"workspaces_root",
	"closed_root",
	"workspace_close_default",
	"workspace_naming",
	"stale_threshold_days",
	"parallel_workers",
	"lock_timeout",
	"lock_stale_threshold",
	"defaults",
	"defaults.workspace_patterns",
	"templates",
	"templates.repos",
	"templates.default_branch",
	"templates.description",
	"templates.setup_commands",
	"hooks",
	"hooks.post_create",
	"hooks.pre_close",
	"tui",
	"tui.keybindings",
	"tui.use_emoji",
	"git",
	"git.retry",
	"git.retry.max_attempts",
	"git.retry.initial_delay",
	"git.retry.max_delay",
	"git.retry.multiplier",
	"git.retry.jitter_factor",
	// Hook fields
	"command",
	"repos",
	"shell",
	"timeout",
	"continue_on_error",
	// Keybinding fields
	"quit",
	"search",
	"push",
	"close",
	"open_editor",
	"toggle_stale",
	"details",
	"confirm",
	"cancel",
	// Pattern fields
	"pattern",
}

// DeprecatedKey represents a deprecated configuration key with migration guidance.
type DeprecatedKey struct {
	OldKey    string // The deprecated key name
	NewKey    string // The replacement key name (empty if removed entirely)
	Message   string // Migration guidance message
	RemovedIn string // Version when the key will be removed (empty if just deprecated)
}

// deprecatedKeys maps deprecated config field names to their migration information.
// Add entries here when deprecating config keys to provide helpful warnings to users.
var deprecatedKeys = map[string]DeprecatedKey{
	// Example (uncomment when deprecating a key):
	// "old_field_name": {
	//     OldKey:    "old_field_name",
	//     NewKey:    "new_field_name",
	//     Message:   "Use 'new_field_name' instead",
	//     RemovedIn: "v2.0.0",
	// },
}

// checkDeprecatedKeys checks for deprecated keys in the raw config map
// and returns warnings for any that are found.
func checkDeprecatedKeys(allSettings map[string]interface{}) []string {
	var warnings []string

	for oldKey, info := range deprecatedKeys {
		if _, exists := allSettings[oldKey]; exists {
			var warning string

			if info.NewKey != "" {
				warning = fmt.Sprintf("config key %q is deprecated, use %q instead", oldKey, info.NewKey)
			} else {
				warning = fmt.Sprintf("config key %q is deprecated: %s", oldKey, info.Message)
			}

			if info.RemovedIn != "" {
				warning += fmt.Sprintf(" (will be removed in %s)", info.RemovedIn)
			}

			warnings = append(warnings, warning)
		}
	}

	return warnings
}

// findSimilarField finds the most similar known field name using Levenshtein distance.
// Returns empty string if no similar field is found (distance > 3).
func findSimilarField(unknown string) string {
	bestMatch := ""
	bestDistance := 4 // Only suggest if distance is 3 or less

	for _, known := range knownConfigFields {
		// Extract just the field name for nested fields
		parts := strings.Split(known, ".")
		fieldName := parts[len(parts)-1]

		dist := levenshteinDistance(strings.ToLower(unknown), strings.ToLower(fieldName))
		if dist < bestDistance {
			bestDistance = dist
			bestMatch = fieldName
		}

		// Also check full path for nested fields
		if len(parts) > 1 {
			dist = levenshteinDistance(strings.ToLower(unknown), strings.ToLower(known))
			if dist < bestDistance {
				bestDistance = dist
				bestMatch = known
			}
		}
	}

	return bestMatch
}

// levenshteinDistance calculates the edit distance between two strings.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}

	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)

	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}

	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// formatUnknownFieldError creates a user-friendly error message for unknown config fields.
func formatUnknownFieldError(unknownFields []string) string {
	var msgs []string

	for _, field := range unknownFields {
		similar := findSimilarField(field)
		if similar != "" {
			msgs = append(msgs, fmt.Sprintf("unknown config field %q, did you mean %q?", field, similar))
		} else {
			msgs = append(msgs, fmt.Sprintf("unknown config field %q", field))
		}
	}

	return strings.Join(msgs, "; ")
}

// extractUnknownFields parses mapstructure error messages to extract unknown field names.
// The error format is: "” has invalid keys: field1, field2" or similar.
func extractUnknownFields(errMsg string) []string {
	// Look for "invalid keys:" pattern
	idx := strings.Index(errMsg, "invalid keys:")
	if idx == -1 {
		return nil
	}

	// Extract everything after "invalid keys:"
	keysStr := strings.TrimSpace(errMsg[idx+len("invalid keys:"):])

	// Split by comma and clean up each field name
	var fields []string

	for _, field := range strings.Split(keysStr, ",") {
		field = strings.TrimSpace(field)
		if field != "" {
			fields = append(fields, field)
		}
	}

	return fields
}

// Load initializes and loads the configuration.
// If configPath is provided (non-empty), it takes precedence over all other config locations.
// Otherwise, CANOPY_CONFIG environment variable is checked, then default locations.
// Priority order: configPath parameter > CANOPY_CONFIG env > default locations.
func Load(configPath string) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, cerrors.NewIOFailed("get user home dir", err)
	}

	viper.SetConfigType("yaml")

	// Track whether an explicit config path was provided (flag or env var)
	explicitConfigPath := false

	// Determine config path with priority: parameter > env > default search paths
	if configPath != "" {
		// Explicit path provided via flag - expand tilde if present
		expandedPath := expandPath(configPath, home)
		viper.SetConfigFile(expandedPath)

		explicitConfigPath = true
	} else if envPath := os.Getenv("CANOPY_CONFIG"); envPath != "" {
		// Environment variable specified - expand tilde if present
		expandedPath := expandPath(envPath, home)
		viper.SetConfigFile(expandedPath)

		explicitConfigPath = true
	} else {
		// Use default search paths
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".canopy"))
		viper.AddConfigPath(filepath.Join(home, ".config", "canopy"))
	}

	viper.SetDefault("projects_root", filepath.Join(home, ".canopy", "projects"))
	viper.SetDefault("workspaces_root", filepath.Join(home, ".canopy", "workspaces"))
	viper.SetDefault("closed_root", filepath.Join(home, ".canopy", "closed"))
	viper.SetDefault("workspace_close_default", CloseDefaultDelete)
	viper.SetDefault("workspace_naming", "{{.ID}}")
	viper.SetDefault("stale_threshold_days", 14)
	viper.SetDefault("lock_timeout", DefaultLockTimeout.String())
	viper.SetDefault("lock_stale_threshold", DefaultLockStaleThreshold.String())

	// Parallel workers default
	viper.SetDefault("parallel_workers", DefaultParallelWorkers)

	// Git retry defaults
	viper.SetDefault("git.retry.max_attempts", 3)
	viper.SetDefault("git.retry.initial_delay", "1s")
	viper.SetDefault("git.retry.max_delay", "30s")
	viper.SetDefault("git.retry.multiplier", 2.0)
	viper.SetDefault("git.retry.jitter_factor", 0.25)

	viper.SetEnvPrefix("CANOPY")
	// Replace dots with underscores for nested keys (e.g., CANOPY_GIT_RETRY_MAX_ATTEMPTS)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found: fail fast if explicit path was provided,
			// otherwise use defaults for search paths
			if explicitConfigPath {
				return nil, cerrors.NewIOFailed("read config file", fmt.Errorf("config file not found: %s", viper.ConfigFileUsed()))
			}
			// Default search paths: config file not found is okay, use defaults
		} else {
			return nil, cerrors.NewIOFailed("read config file", err)
		}
	}

	var cfg Config
	// Use strict unmarshaling to detect unknown config fields (typos, etc.)
	if err := viper.Unmarshal(&cfg, func(config *mapstructure.DecoderConfig) {
		config.ErrorUnused = true
	}); err != nil {
		return nil, handleUnmarshalError(err)
	}

	// Expand tilde
	cfg.ProjectsRoot = expandPath(cfg.ProjectsRoot, home)
	cfg.WorkspacesRoot = expandPath(cfg.WorkspacesRoot, home)
	cfg.ClosedRoot = expandPath(cfg.ClosedRoot, home)
	cfg.CloseDefault = strings.ToLower(cfg.CloseDefault)

	// Check for deprecated keys and collect warnings
	cfg.Warnings = checkDeprecatedKeys(viper.AllSettings())

	registry, err := LoadRepoRegistry("")
	if err != nil {
		return nil, cerrors.NewRegistryError("load", "repository registry", err)
	}

	cfg.Registry = registry

	return &cfg, nil
}

// handleUnmarshalError processes viper unmarshal errors and provides helpful suggestions
// for unknown fields (typos, etc.).
func handleUnmarshalError(err error) error {
	errMsg := err.Error()
	if strings.Contains(errMsg, "has invalid keys") {
		// Extract the unknown field names from the error
		unknownFields := extractUnknownFields(errMsg)
		if len(unknownFields) > 0 {
			return cerrors.NewConfigValidation("config", formatUnknownFieldError(unknownFields))
		}
	}

	return cerrors.NewConfigInvalid(fmt.Sprintf("failed to unmarshal: %v", err))
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

// GetTemplates returns a copy of the configured templates keyed by name.
func (c *Config) GetTemplates() map[string]Template {
	if len(c.Templates) == 0 {
		return map[string]Template{}
	}

	templates := make(map[string]Template, len(c.Templates))
	for name, tmpl := range c.Templates {
		tmpl.Name = name
		templates[name] = tmpl
	}

	return templates
}

// ResolveTemplate returns a template by name with helpful errors if missing.
func (c *Config) ResolveTemplate(name string) (Template, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Template{}, cerrors.NewInvalidArgument("template", "template name is required")
	}

	templates := c.GetTemplates()
	if tmpl, ok := templates[name]; ok {
		return tmpl, nil
	}

	if len(templates) == 0 {
		return Template{}, cerrors.NewInvalidArgument("template", "no templates are defined")
	}

	names := make([]string, 0, len(templates))
	for tmplName := range templates {
		names = append(names, tmplName)
	}

	sort.Strings(names)

	return Template{}, cerrors.NewInvalidArgument("template", fmt.Sprintf("unknown template %q (available: %s)", name, strings.Join(names, ", ")))
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
	if err := c.validateWorkspaceSettings(); err != nil {
		return err
	}

	if err := c.validateRuntimeSettings(); err != nil {
		return err
	}

	return c.validateKeybindings()
}

// ValidateTemplates validates template definitions without performing filesystem checks.
func (c *Config) ValidateTemplates() error {
	return c.validateTemplates()
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

// validatePatterns checks that all workspace regex patterns are valid.
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

// validateStaleThreshold checks that the stale threshold is non-negative.
func (c *Config) validateStaleThreshold() error {
	if c.StaleThresholdDays < 0 {
		return cerrors.NewConfigValidation("stale_threshold_days", fmt.Sprintf("must be zero or positive, got %d", c.StaleThresholdDays))
	}

	return nil
}

// validateParallelWorkers checks that parallel_workers is within valid range (1-10).
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

// Close behavior constants
const (
	CloseDefaultDelete  = "delete"
	CloseDefaultArchive = "archive"
)

// Parallel workers constants
const (
	// DefaultParallelWorkers is the default number of parallel workers for repository operations.
	DefaultParallelWorkers = 4
	// DefaultLockTimeout is the default time to wait for a workspace lock.
	DefaultLockTimeout = 30 * time.Second
	// DefaultLockStaleThreshold is the default age before a lock is considered stale.
	DefaultLockStaleThreshold = 5 * time.Minute
	// MinParallelWorkers is the minimum allowed value for parallel workers.
	MinParallelWorkers = 1
	// MaxParallelWorkers is the maximum allowed value for parallel workers.
	MaxParallelWorkers = 10
)

// maxRetryAttempts is the maximum allowed value for retry attempts to prevent misconfiguration.
const maxRetryAttempts = 10

// validateGitRetry validates the git retry configuration.
//
//nolint:gocyclo // Sequential validation checks are simpler to read than refactoring for lower complexity
func (c *Config) validateGitRetry() error {
	retry := c.Git.Retry

	if retry.MaxAttempts < 1 {
		return cerrors.NewConfigValidation("git.retry.max_attempts", fmt.Sprintf("must be at least 1, got %d", retry.MaxAttempts))
	}

	if retry.MaxAttempts > maxRetryAttempts {
		return cerrors.NewConfigValidation("git.retry.max_attempts", fmt.Sprintf("must not exceed %d, got %d", maxRetryAttempts, retry.MaxAttempts))
	}

	initialDelay, err := time.ParseDuration(retry.InitialDelay)
	if err != nil {
		return cerrors.NewConfigValidation("git.retry.initial_delay", fmt.Sprintf("invalid: %v", err))
	}

	if initialDelay <= 0 {
		return cerrors.NewConfigValidation("git.retry.initial_delay", fmt.Sprintf("must be positive, got %s", retry.InitialDelay))
	}

	maxDelay, err := time.ParseDuration(retry.MaxDelay)
	if err != nil {
		return cerrors.NewConfigValidation("git.retry.max_delay", fmt.Sprintf("invalid: %v", err))
	}

	if maxDelay <= 0 {
		return cerrors.NewConfigValidation("git.retry.max_delay", fmt.Sprintf("must be positive, got %s", retry.MaxDelay))
	}

	if initialDelay > maxDelay {
		return cerrors.NewConfigValidation("git.retry.initial_delay", fmt.Sprintf("(%s) must not exceed max_delay (%s)", retry.InitialDelay, retry.MaxDelay))
	}

	if retry.Multiplier < 1.0 {
		return cerrors.NewConfigValidation("git.retry.multiplier", fmt.Sprintf("must be at least 1.0, got %f", retry.Multiplier))
	}

	if retry.JitterFactor < 0 || retry.JitterFactor > 1 {
		return cerrors.NewConfigValidation("git.retry.jitter_factor", fmt.Sprintf("must be between 0 and 1, got %f", retry.JitterFactor))
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

	// Validate shell is non-empty if specified
	if h.Shell != "" && strings.TrimSpace(h.Shell) == "" {
		return cerrors.NewConfigValidation(field, "shell cannot be empty or whitespace-only when specified")
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
		return cerrors.NewConfigValidation(label, "is required")
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

// ComputeWorkspaceDir computes the workspace directory name for a given ID.
func (c *Config) ComputeWorkspaceDir(id string) (string, error) {
	if err := validation.ValidateWorkspaceID(id); err != nil {
		return "", err
	}

	return c.computeWorkspaceDir(id)
}

func (c *Config) computeWorkspaceDir(id string) (string, error) {
	if strings.TrimSpace(c.WorkspaceNaming) == "" {
		c.WorkspaceNaming = "{{.ID}}"
	}

	tmpl, err := template.New("workspace_naming").Option("missingkey=error").Parse(c.WorkspaceNaming)
	if err != nil {
		return "", cerrors.NewConfigValidation("workspace_naming", fmt.Sprintf("invalid template: %v", err))
	}

	var rendered strings.Builder
	if err := tmpl.Execute(&rendered, WorkspaceNamingTemplateData{ID: id}); err != nil {
		return "", cerrors.NewConfigValidation("workspace_naming", fmt.Sprintf("template execution failed: %v", err))
	}

	rawDir := rendered.String()

	dirName, err := validation.NormalizeWorkspaceDirName(rawDir)
	if err != nil {
		return "", cerrors.NewConfigValidation("workspace_naming", fmt.Sprintf("template output %q is invalid: %v", rawDir, err))
	}

	return dirName, nil
}

// GetStaleThresholdDays returns the stale threshold in days.
func (c *Config) GetStaleThresholdDays() int {
	return c.StaleThresholdDays
}

// GetParallelWorkers returns the number of parallel workers for repository operations.
func (c *Config) GetParallelWorkers() int {
	return c.ParallelWorkers
}

// GetLockTimeout returns the workspace lock timeout.
func (c *Config) GetLockTimeout() time.Duration {
	if c.LockTimeout == "" {
		return DefaultLockTimeout
	}

	parsed, err := time.ParseDuration(c.LockTimeout)
	if err != nil {
		return DefaultLockTimeout
	}

	return parsed
}

// GetLockStaleThreshold returns the stale lock threshold.
func (c *Config) GetLockStaleThreshold() time.Duration {
	if c.LockStaleThreshold == "" {
		return DefaultLockStaleThreshold
	}

	parsed, err := time.ParseDuration(c.LockStaleThreshold)
	if err != nil {
		return DefaultLockStaleThreshold
	}

	return parsed
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

// GetUseEmoji returns whether emoji should be used in the TUI.
func (c *Config) GetUseEmoji() bool {
	return c.TUI.GetUseEmoji()
}

// GetGitRetryConfig returns the parsed git retry configuration.
// Since validation has already run, we can safely ignore the error.
func (c *Config) GetGitRetryConfig() ParsedRetryConfig {
	parsed, _ := c.Git.Retry.Parse()
	return parsed
}

// GetWarnings returns any warnings collected during config loading.
// These may include deprecation warnings or other non-fatal issues.
func (c *Config) GetWarnings() []string {
	return c.Warnings
}

// HasWarnings returns true if there are any warnings from config loading.
func (c *Config) HasWarnings() bool {
	return len(c.Warnings) > 0
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
		return cerrors.NewConfigValidation("tui.keybindings", strings.Join(errors, "; "))
	}

	return nil
}
