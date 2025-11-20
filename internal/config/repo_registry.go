// Package config provides configuration and registry utilities.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// RegistryEntry represents a single repository alias entry.
type RegistryEntry struct {
	Alias         string   `yaml:"-"`   // populated in-memory for convenience
	URL           string   `yaml:"url"` // required
	DefaultBranch string   `yaml:"default_branch,omitempty"`
	Description   string   `yaml:"description,omitempty"`
	Tags          []string `yaml:"tags,omitempty"`
}

// RepoRegistry stores repository aliases and metadata.
type RepoRegistry struct {
	path  string
	Repos map[string]RegistryEntry `yaml:"repos"`
}

// LoadRepoRegistry loads the registry from disk. Missing files are treated as empty registries.
func LoadRepoRegistry(path string) (*RepoRegistry, error) {
	if path == "" {
		var err error

		path, err = defaultRegistryPath()
		if err != nil {
			return nil, err
		}
	}

	registry := &RepoRegistry{
		path:  path,
		Repos: make(map[string]RegistryEntry),
	}

	data, err := os.ReadFile(path) //nolint:gosec // registry path is constructed internally
	if err != nil {
		if os.IsNotExist(err) {
			return registry, nil
		}

		return nil, err
	}

	if err := yaml.Unmarshal(data, registry); err != nil {
		return nil, err
	}

	if registry.Repos == nil {
		registry.Repos = make(map[string]RegistryEntry)
	}

	return registry, nil
}

// Save persists the registry to disk.
func (r *RepoRegistry) Save() error {
	if r.path == "" {
		var err error

		r.path, err = defaultRegistryPath()
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(r.path), 0o750); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(r.path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// Resolve returns a registry entry by alias if present.
func (r *RepoRegistry) Resolve(alias string) (RegistryEntry, bool) {
	entry, ok := r.Repos[alias]
	if !ok {
		return RegistryEntry{}, false
	}

	entry.Alias = alias

	return entry, true
}

// Register adds an entry under the provided alias. Errors if alias exists unless force is true.
func (r *RepoRegistry) Register(alias string, entry RegistryEntry, force bool) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return fmt.Errorf("alias is required")
	}

	if !isLikelyURL(entry.URL) {
		return fmt.Errorf("invalid repository URL: %s", entry.URL)
	}

	if _, exists := r.Repos[alias]; exists && !force {
		existing := r.Repos[alias]
		return fmt.Errorf("alias '%s' already exists for %s", alias, existing.URL)
	}

	entry.Alias = alias
	r.Repos[alias] = stripAlias(entry)

	return nil
}

// RegisterWithSuffix registers an entry, appending "-2" style suffixes until unique.
func (r *RepoRegistry) RegisterWithSuffix(alias string, entry RegistryEntry) (string, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return "", fmt.Errorf("alias is required")
	}

	if !isLikelyURL(entry.URL) {
		return "", fmt.Errorf("invalid repository URL: %s", entry.URL)
	}

	target := alias
	for idx := 2; ; idx++ {
		if _, exists := r.Repos[target]; !exists {
			break
		}

		target = fmt.Sprintf("%s-%d", alias, idx)
	}

	entry.Alias = target
	r.Repos[target] = stripAlias(entry)

	return target, nil
}

// Unregister removes an alias from the registry.
func (r *RepoRegistry) Unregister(alias string) error {
	if _, exists := r.Repos[alias]; !exists {
		return fmt.Errorf("alias '%s' not found", alias)
	}

	delete(r.Repos, alias)

	return nil
}

// List returns all entries, optionally filtered by tags. Results are sorted by alias.
func (r *RepoRegistry) List(tags []string) []RegistryEntry {
	var entries []RegistryEntry

	for alias, entry := range r.Repos {
		entry.Alias = alias
		if len(tags) == 0 || hasAllTags(entry.Tags, tags) {
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Alias < entries[j].Alias
	})

	return entries
}

// Path returns the registry file path.
func (r *RepoRegistry) Path() string {
	if r.path == "" {
		path, _ := defaultRegistryPath()
		return path
	}

	return r.path
}

// DeriveAliasFromURL returns a sensible alias from a Git URL.
func DeriveAliasFromURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}

	// Handle scp-like git@host:owner/repo.git
	if strings.Contains(url, ":") && !strings.HasPrefix(url, "http") {
		parts := strings.Split(url, ":")
		if len(parts) > 1 {
			url = parts[len(parts)-1]
		}
	}

	parts := strings.Split(url, "/")
	last := parts[len(parts)-1]
	last = strings.TrimSuffix(last, ".git")
	last = strings.ToLower(last)

	return last
}

func defaultRegistryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	return filepath.Join(home, ".yard", "repos.yaml"), nil
}

func stripAlias(entry RegistryEntry) RegistryEntry {
	entry.Alias = ""
	return entry
}

func hasAllTags(entryTags, required []string) bool {
	if len(required) == 0 {
		return true
	}

	tagSet := make(map[string]struct{}, len(entryTags))
	for _, t := range entryTags {
		tagSet[strings.ToLower(t)] = struct{}{}
	}

	for _, r := range required {
		if _, ok := tagSet[strings.ToLower(r)]; !ok {
			return false
		}
	}

	return true
}

func isLikelyURL(val string) bool {
	return strings.HasPrefix(val, "http://") ||
		strings.HasPrefix(val, "https://") ||
		strings.HasPrefix(val, "git@") ||
		strings.HasPrefix(val, "file://")
}
