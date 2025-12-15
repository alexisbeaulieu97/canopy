// Package config provides configuration and registry utilities.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/giturl"
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
	path  string                   `yaml:"-"`
	Repos map[string]RegistryEntry `yaml:"repos"`
}

func (r *RepoRegistry) ensureMap() {
	if r.Repos == nil {
		r.Repos = make(map[string]RegistryEntry)
	}
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
	r.ensureMap()

	if r.path == "" {
		var err error

		r.path, err = defaultRegistryPath()
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(r.path), 0o750); err != nil {
		return cerrors.NewRegistryError("save", "create registry directory", err)
	}

	data, err := yaml.Marshal(r)
	if err != nil {
		return cerrors.NewRegistryError("save", "marshal registry", err)
	}

	if err := os.WriteFile(r.path, data, 0o600); err != nil {
		return cerrors.NewRegistryError("save", "write registry", err)
	}

	return nil
}

// Resolve returns a registry entry by alias if present.
func (r *RepoRegistry) Resolve(alias string) (RegistryEntry, bool) {
	r.ensureMap()

	entry, ok := r.Repos[alias]
	if !ok {
		return RegistryEntry{}, false
	}

	entry.Alias = alias

	return entry, true
}

// ResolveByURL returns a registry entry whose URL matches exactly.
func (r *RepoRegistry) ResolveByURL(url string) (RegistryEntry, bool) {
	r.ensureMap()

	for alias, entry := range r.Repos {
		if entry.URL == strings.TrimSpace(url) {
			entry.Alias = alias
			return entry, true
		}
	}

	return RegistryEntry{}, false
}

// Register adds an entry under the provided alias. Errors if alias exists unless force is true.
func (r *RepoRegistry) Register(alias string, entry RegistryEntry, force bool) error {
	r.ensureMap()

	alias = strings.TrimSpace(alias)
	if alias == "" {
		return cerrors.NewInvalidArgument("alias", "is required")
	}

	entry.URL = strings.TrimSpace(entry.URL)

	sanitizedURL := giturl.Sanitize(entry.URL)
	if !giturl.IsURL(entry.URL) {
		return cerrors.NewInvalidArgument("url", fmt.Sprintf("invalid repository URL: %s", sanitizedURL))
	}

	if _, exists := r.Repos[alias]; exists && !force {
		existing := r.Repos[alias]
		return cerrors.NewRegistryError("register", fmt.Sprintf("alias '%s' already exists for %s", alias, giturl.Sanitize(existing.URL)), nil)
	}

	entry.Alias = alias
	r.Repos[alias] = stripAlias(entry)

	return nil
}

// RegisterWithSuffix registers an entry, appending "-2" style suffixes until unique.
func (r *RepoRegistry) RegisterWithSuffix(alias string, entry RegistryEntry) (string, error) {
	r.ensureMap()

	alias = strings.TrimSpace(alias)
	if alias == "" {
		return "", cerrors.NewInvalidArgument("alias", "is required")
	}

	entry.URL = strings.TrimSpace(entry.URL)

	sanitizedURL := giturl.Sanitize(entry.URL)
	if !giturl.IsURL(entry.URL) {
		return "", cerrors.NewInvalidArgument("url", fmt.Sprintf("invalid repository URL: %s", sanitizedURL))
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
	r.ensureMap()

	if _, exists := r.Repos[alias]; !exists {
		return cerrors.NewRepoNotFound(alias)
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
//
// Deprecated: Use giturl.DeriveAlias instead.
func DeriveAliasFromURL(url string) string {
	return giturl.DeriveAlias(url)
}

func defaultRegistryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", cerrors.NewIOFailed("get user home dir", err)
	}

	return filepath.Join(home, ".canopy", "repos.yaml"), nil
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
