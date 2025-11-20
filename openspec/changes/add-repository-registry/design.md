# Design: Repository Registry System

## Context
Currently, users must provide full Git URLs or configure regex patterns in config.yaml to use repositories with workspaces. The system has a hardcoded fallback to "github.com/example/" which doesn't work in practice. This creates friction in daily workflows where users repeatedly type long URLs.

## Goals / Non-Goals

### Goals
- Provide simple alias system for frequently-used repositories
- Reduce typing burden for workspace creation
- Support repository metadata (default branch, tags, description)
- Work seamlessly with existing URL-based workflows
- Auto-register repositories when cloned via `yard repo add`

### Non-Goals
- Auto-discover repos from Git hosting APIs (future enhancement)
- Support complex query languages for registry
- Replace existing workspace pattern matching (they complement each other)
- Sync registry across machines (future enhancement)

## Decisions

### Decision 1: Registry File Format
Use YAML format at `~/.yard/repos.yaml` for human readability and editability.

```yaml
repos:
  backend:
    url: https://github.com/myorg/backend-api
    default_branch: develop
    description: Backend API service
    tags: [backend, api, golang]

  frontend:
    url: https://github.com/myorg/frontend-web
    default_branch: main
    description: React frontend application
    tags: [frontend, react]
```

**Rationale:**
- Consistent with config.yaml format
- Easy for users to manually edit
- Supports comments for documentation
- go-yaml makes parsing trivial

**Alternative considered:** JSON
- **Rejected:** Less human-friendly, no comments, harder to hand-edit

### Decision 2: Resolution Priority
Resolution order for repository names:
1. Check if it's a full URL (starts with http://, https://,  git@) → use directly
2. Check if it's in registry → use registered URL
3. Check if it matches "owner/repo" format → construct GitHub URL
4. Check workspace patterns in config.yaml → use pattern repos
5. Error: unknown repository

**Rationale:**
- Backwards compatible with existing URL-based workflows
- Registry is opt-in enhancement, not mandatory
- Explicit URLs always take precedence (principle of least surprise)
- Patterns remain useful for org-wide conventions

### Decision 3: Auto-Registration
When `yard repo add <url>` clones a repository, automatically create a registry entry.

**Behavior:**
```bash
yard repo add https://github.com/myorg/backend-api
# Output:
# Cloning repository...
# Registered as 'backend-api' in registry
# Use: yard workspace new PROJ-123 --repos backend-api
```

**Rationale:**
- Reduces friction: users don't think about registry, it just works
- One-time setup: clone creates both canonical repo and alias
- Can be disabled with `--no-register` flag

**Alternative considered:** Require explicit registration
- **Rejected:** Extra step users will forget; less ergonomic

### Decision 4: Alias Derivation Rules
Derive alias from URL using these rules:
1. Take last path component (e.g., "backend-api.git" → "backend-api")
2. Remove ".git" suffix
3. Convert to lowercase
4. If collision, append "-2", "-3", etc.

**Rationale:**
- Predictable and intuitive
- Matches most developers' mental model
- Handles collisions gracefully

### Decision 5: Registry Struct Design
```go
type RepoRegistry struct {
    path    string
    entries map[string]RegistryEntry
}

type RegistryEntry struct {
    Alias         string   `yaml:"-"`  // Key in map, not in file
    URL           string   `yaml:"url"`
    DefaultBranch string   `yaml:"default_branch,omitempty"`
    Description   string   `yaml:"description,omitempty"`
    Tags          []string `yaml:"tags,omitempty"`
}

func (r *RepoRegistry) Resolve(name string) (url string, found bool)
func (r *RepoRegistry) Register(alias, url string, opts ...Option) error
func (r *RepoRegistry) Unregister(alias string) error
func (r *RepoRegistry) List(filter ...TagFilter) []RegistryEntry
```

**Rationale:**
- Simple in-memory map for fast lookups
- Lazy loading: only load when needed
- Options pattern for metadata (WithBranch(), WithTags(), etc.)
- Tag filtering supports future enhancements (e.g., "show me all backend repos")

## Implementation Approach

### Phase 1: Core Registry
```go
// internal/config/repo_registry.go
package config

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

type RepoRegistry struct {
    path    string
    Repos   map[string]RegistryEntry `yaml:"repos"`
}

func LoadRegistry() (*RepoRegistry, error) {
    home, _ := os.UserHomeDir()
    path := filepath.Join(home, ".yard", "repos.yaml")

    r := &RepoRegistry{path: path, Repos: make(map[string]RegistryEntry)}

    data, err := os.ReadFile(path)
    if os.IsNotExist(err) {
        return r, nil // Empty registry is valid
    }
    if err != nil {
        return nil, err
    }

    if err := yaml.Unmarshal(data, r); err != nil {
        return nil, err
    }

    return r, nil
}
```

### Phase 2: Update ResolveRepos
```go
// internal/workspaces/service.go
func (s *Service) ResolveRepos(workspaceID string, requestedRepos []string) ([]domain.Repo, error) {
    for _, val := range requestedRepos {
        // 1. Full URL?
        if isURL(val) {
            repos = append(repos, domain.Repo{Name: deriveName(val), URL: val})
            continue
        }

        // 2. Registry alias?
        if entry, found := s.registry.Resolve(val); found {
            repos = append(repos, domain.Repo{Name: val, URL: entry.URL})
            continue
        }

        // 3. owner/repo format?
        if strings.Contains(val, "/") && strings.Count(val, "/") == 1 {
            url := "https://github.com/" + val
            repos = append(repos, domain.Repo{Name: strings.Split(val, "/")[1], URL: url})
            continue
        }

        // 4. Unknown
        return nil, fmt.Errorf("unknown repository '%s'. Use 'yard repo register %s <url>' to add it", val, val)
    }
}
```

### Phase 3: Commands
```go
// cmd/yard/repo.go
var repoRegisterCmd = &cobra.Command{
    Use:   "register <alias> <url>",
    Short: "Register a repository alias",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        alias, url := args[0], args[1]

        registry, _ := config.LoadRegistry()
        if err := registry.Register(alias, url); err != nil {
            return err
        }
        if err := registry.Save(); err != nil {
            return err
        }

        fmt.Printf("Registered '%s' → %s\n", alias, url)
        return nil
    },
}
```

## Risks / Trade-offs

### Risk: Registry File Corruption
**Mitigation:**
- Validate YAML before save
- Create backup before writes (.repos.yaml.bak)
- Clear error messages for parse failures

### Risk: Alias Collisions
**Mitigation:**
- Auto-append suffix (-2, -3) when auto-registering
- Show error on manual registration with existing alias
- Add --force flag to overwrite

### Trade-off: Another File to Manage
Users now have config.yaml and repos.yaml. This is acceptable because:
- repos.yaml is optional (fallback to URLs works)
- Auto-registration reduces manual management
- Separation of concerns: config vs data

## Open Questions

**Q: Should registry support remote sync (e.g., shared team registry)?**
A: Not in v1. Add as future enhancement if users request it. Implementation would use git-based sync similar to dotfiles.

**Q: Should we support importing from .gitconfig?**
A: Future enhancement. Would scan `[url "..."]` sections for common patterns. Not MVP.

**Q: How to handle registry entry removal when repo is deleted?**
A: `yard repo remove` should ask "Remove from registry too? [Y/n]". Keep separate commands for flexibility.
