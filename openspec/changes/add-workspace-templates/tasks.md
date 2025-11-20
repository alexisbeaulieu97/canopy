# Implementation Tasks

## 1. Define Template Data Model
- [ ] 1.1 Create Template struct in internal/config/config.go
- [ ] 1.2 Add fields: Name, Repos, DefaultBranch, Description, SetupCommands
- [ ] 1.3 Add Templates map[string]Template to Config struct
- [ ] 1.4 Implement YAML unmarshaling for templates section
- [ ] 1.5 Write unit tests for template parsing

## 2. Template Resolution Logic
- [ ] 2.1 Create ResolveTemplate(name string) method on Config
- [ ] 2.2 Implement template lookup with clear error messages
- [ ] 2.3 Support merging template repos with explicit --repos flag
- [ ] 2.4 Add validation for template references (repos must be valid)

## 3. Integrate Templates with Workspace Creation
- [ ] 3.1 Update CreateWorkspace to accept optional template parameter
- [ ] 3.2 Apply template repos if specified
- [ ] 3.3 Apply template default branch if no explicit branch given
- [ ] 3.4 Execute template setup commands after worktree creation
- [ ] 3.5 Handle setup command failures gracefully

## 4. Add Template CLI Commands
- [ ] 4.1 Add --template flag to `yard workspace new` command
- [ ] 4.2 Implement `yard template list` command
- [ ] 4.3 Add colorized output showing template name, description, repos
- [ ] 4.4 Implement `yard template show <name>` for detailed view
- [ ] 4.5 Add `yard template validate` to check template definitions

## 5. Documentation & Examples
- [ ] 5.1 Add templates section to example config.yaml
- [ ] 5.2 Update README.md with template usage examples
- [ ] 5.3 Document template format in configuration guide
- [ ] 5.4 Add common templates to docs (fullstack, backend-only, frontend-only)

## 6. Testing & Validation
- [ ] 6.1 Write unit tests for template resolution
- [ ] 6.2 Write integration test for workspace creation with template
- [ ] 6.3 Test template + explicit repos combination
- [ ] 6.4 Test template with setup commands
- [ ] 6.5 Test error handling for invalid templates
