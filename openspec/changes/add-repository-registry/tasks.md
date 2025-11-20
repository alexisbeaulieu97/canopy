# Implementation Tasks

## 1. Create Registry Infrastructure
- [x] 1.1 Create `internal/config/repo_registry.go` with RepoRegistry struct
- [x] 1.2 Implement YAML marshaling/unmarshaling for registry format
- [x] 1.3 Add registry file path resolution (~/.yard/repos.yaml)
- [x] 1.4 Implement Load() and Save() methods for registry persistence
- [x] 1.5 Write unit tests for registry operations

## 2. Define Registry Data Model
- [x] 2.1 Define RegistryEntry struct (alias, url, default_branch, description, tags)
- [x] 2.2 Add validation for unique aliases
- [x] 2.3 Add validation for valid Git URLs
- [x] 2.4 Implement registry search/filter by tags

## 3. Integrate Registry with Config
- [x] 3.1 Add Registry field to Config struct
- [x] 3.2 Load registry during config.Load()
- [x] 3.3 Handle missing registry file gracefully (create empty)
- [x] 3.4 Add registry accessor methods to Config

## 4. Update Repository Resolution
- [x] 4.1 Modify ResolveRepos() to check registry first
- [x] 4.2 Add fallback to URL pattern matching if not in registry
- [x] 4.3 Support mixing registry aliases and URLs in --repos flag
- [x] 4.4 Update error messages to suggest "yard repo register" for unknown aliases

## 5. Add Registry Commands
- [x] 5.1 Implement `yard repo register <alias> <url>` command
- [x] 5.2 Add --branch, --description, --tags flags to register command
- [x] 5.3 Implement `yard repo unregister <alias>` command
- [x] 5.4 Implement `yard repo list-registry` command with formatting
- [x] 5.5 Add `yard repo show <alias>` to display registry entry details

## 6. Auto-Registration Feature
- [x] 6.1 Update `yard repo add <url>` to auto-register with derived alias
- [x] 6.2 Add --alias flag to override default alias
- [x] 6.3 Add --no-register flag to skip auto-registration
- [x] 6.4 Prompt user for alias if auto-derived name conflicts

## 7. Documentation & UX
- [x] 7.1 Add registry examples to README.md
- [x] 7.2 Update `yard repo --help` with registry commands
- [x] 7.3 Add colored output for registry list command
- [x] 7.4 Include registry file location in `yard check` output

## 8. Testing & Validation
- [x] 8.1 Write unit tests for RepoRegistry
- [x] 8.2 Write integration tests for registry commands
- [x] 8.3 Test registry with workspace creation workflow
- [x] 8.4 Test migration from non-registry to registry usage
