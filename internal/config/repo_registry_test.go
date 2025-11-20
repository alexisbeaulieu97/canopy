package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRegistryMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repos.yaml")

	registry, err := LoadRepoRegistry(path)
	if err != nil {
		t.Fatalf("expected no error loading missing file, got %v", err)
	}

	if registry.Path() != path {
		t.Fatalf("expected path %s, got %s", path, registry.Path())
	}

	if len(registry.Repos) != 0 {
		t.Fatalf("expected empty registry")
	}
}

func TestRegisterAndResolve(t *testing.T) {
	registry := &RepoRegistry{path: filepath.Join(t.TempDir(), "repos.yaml"), Repos: map[string]RegistryEntry{}}

	entry := RegistryEntry{URL: "https://github.com/example/api.git", Description: "API"}
	if err := registry.Register("api", entry, false); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	resolved, ok := registry.Resolve("api")
	if !ok {
		t.Fatalf("expected to resolve alias")
	}

	if resolved.URL != entry.URL {
		t.Fatalf("unexpected url: %s", resolved.URL)
	}

	if err := registry.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Reload and ensure persistence
	reloaded, err := LoadRepoRegistry(registry.path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	if _, ok := reloaded.Resolve("api"); !ok {
		t.Fatalf("expected alias after reload")
	}
}

func TestRegisterCollisionAndForce(t *testing.T) {
	registry := &RepoRegistry{path: filepath.Join(t.TempDir(), "repos.yaml"), Repos: map[string]RegistryEntry{}}

	entry := RegistryEntry{URL: "https://github.com/example/api.git"}
	if err := registry.Register("api", entry, false); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	if err := registry.Register("api", entry, false); err == nil {
		t.Fatalf("expected collision error")
	}

	if err := registry.Register("api", RegistryEntry{URL: "https://github.com/example/other.git"}, true); err != nil {
		t.Fatalf("force register failed: %v", err)
	}
}

func TestRegisterWithSuffix(t *testing.T) {
	registry := &RepoRegistry{path: filepath.Join(t.TempDir(), "repos.yaml"), Repos: map[string]RegistryEntry{}}

	entry := RegistryEntry{URL: "https://github.com/example/api.git"}
	if _, err := registry.RegisterWithSuffix("api", entry); err != nil {
		t.Fatalf("register with suffix failed: %v", err)
	}

	if alias, err := registry.RegisterWithSuffix("api", entry); err != nil || alias != "api-2" {
		t.Fatalf("expected alias api-2, got %s (err=%v)", alias, err)
	}
}

func TestDeriveAliasFromURL(t *testing.T) {
	if got := DeriveAliasFromURL("https://github.com/org/backend-api.git"); got != "backend-api" {
		t.Fatalf("unexpected alias: %s", got)
	}

	if got := DeriveAliasFromURL("git@github.com:org/Service.git"); got != "service" {
		t.Fatalf("unexpected alias: %s", got)
	}
}

func TestListFiltersTags(t *testing.T) {
	registry := &RepoRegistry{path: filepath.Join(t.TempDir(), "repos.yaml"), Repos: map[string]RegistryEntry{
		"api":  {URL: "https://github.com/example/api", Tags: []string{"backend", "go"}},
		"web":  {URL: "https://github.com/example/web", Tags: []string{"frontend"}},
		"util": {URL: "https://github.com/example/util", Tags: []string{}},
	}}

	entries := registry.List([]string{"backend"})
	if len(entries) != 1 || entries[0].Alias != "api" {
		t.Fatalf("expected only api entry, got %v", entries)
	}
}

func TestUnregister(t *testing.T) {
	registry := &RepoRegistry{path: filepath.Join(t.TempDir(), "repos.yaml"), Repos: map[string]RegistryEntry{"api": {URL: "https://github.com/example/api"}}}

	if err := registry.Unregister("api"); err != nil {
		t.Fatalf("unregister failed: %v", err)
	}

	if err := registry.Unregister("api"); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "nested", "repos.yaml")
	registry := &RepoRegistry{path: path, Repos: map[string]RegistryEntry{"api": {URL: "https://github.com/example/api"}}}

	if err := registry.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file saved: %v", err)
	}
}
