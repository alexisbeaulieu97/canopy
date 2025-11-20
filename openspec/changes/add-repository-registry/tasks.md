# Implementation Tasks

## 1. Create Registry Infrastructure
- [ ] 1.1 Create `internal/config/repo_registry.go` with RepoRegistry struct
- [ ] 1.2 Implement YAML marshaling/unmarshaling for registry format
- [ ] 1.3 Add registry file path resolution (~/.yard/repos.yaml)
- [ ] 1.4 Implement Load() and Save() methods for registry persistence
- [ ] 1.5 Write unit tests for registry operations

## 2. Define Registry Data Model
- [ ] 2.1 Define RegistryEntry struct (alias, url, default_branch, description, tags)
- [ ] 2.2 Add validation for unique aliases
- [ ] 2.3 Add validation for valid Git URLs
- [ ] 2.4 Implement registry search/filter by tags

## 3. Integrate Registry with Config
- [ ] 3.1 Add Registry field to Config struct
- [ ] 3.2 Load registry during config.Load()
- [ ] 3.3 Handle missing registry file gracefully (create empty)
- [ ] 3.4 Add registry accessor methods to Config

## 4. Update Repository Resolution
- [ ] 4.1 Modify ResolveRepos() to check registry first
- [ ] 4.2 Add fallback to URL pattern matching if not in registry
- [ ] 4.3 Support mixing registry aliases and URLs in --repos flag
- [ ] 4.4 Update error messages to suggest "yard repo register" for unknown aliases

## 5. Add Registry Commands
- [ ] 5.1 Implement `yard repo register <alias> <url>` command
- [ ] 5.2 Add --branch, --description, --tags flags to register command
- [ ] 5.3 Implement `yard repo unregister <alias>` command
- [ ] 5.4 Implement `yard repo list-registry` command with formatting
- [ ] 5.5 Add `yard repo show <alias>` to display registry entry details

## 6. Auto-Registration Feature
- [ ] 6.1 Update `yard repo add <url>` to auto-register with derived alias
- [ ] 6.2 Add --alias flag to override default alias
- [ ] 6.3 Add --no-register flag to skip auto-registration
- [ ] 6.4 Prompt user for alias if auto-derived name conflicts

## 7. Documentation & UX
- [ ] 7.1 Add registry examples to README.md
- [ ] 7.2 Update `yard repo --help` with registry commands
- [ ] 7.3 Add colored output for registry list command
- [ ] 7.4 Include registry file location in `yard check` output

## 8. Testing & Validation
- [ ] 8.1 Write unit tests for RepoRegistry
- [ ] 8.2 Write integration tests for registry commands
- [ ] 8.3 Test registry with workspace creation workflow
- [ ] 8.4 Test migration from non-registry to registry usage
