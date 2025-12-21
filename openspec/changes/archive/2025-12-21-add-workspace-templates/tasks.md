# Implementation Tasks

## 1. Define Template Data Model
- [x] 1.1 Create Template struct in internal/config/config.go
- [x] 1.2 Add fields: Name, Repos, DefaultBranch, Description, SetupCommands
- [x] 1.3 Add Templates map[string]Template to Config struct
- [x] 1.4 Implement YAML unmarshaling for templates section
- [x] 1.5 Write unit tests for template parsing

## 2. Template Resolution Logic
- [x] 2.1 Create ResolveTemplate(name string) method on Config
- [x] 2.2 Implement template lookup with clear error messages
- [x] 2.3 Support merging template repos with explicit --repos flag
- [x] 2.4 Add validation for template references (repos must be valid)

## 3. Integrate Templates with Workspace Creation
- [x] 3.1 Update CreateWorkspace to accept optional template parameter
- [x] 3.2 Apply template repos if specified
- [x] 3.3 Apply template default branch if no explicit branch given
- [x] 3.4 Execute template setup commands after worktree creation
- [x] 3.5 Handle setup command failures gracefully

## 4. Add Template CLI Commands
- [x] 4.1 Add --template flag to `canopy workspace new` command
- [x] 4.2 Implement `canopy template list` command
- [x] 4.3 Add colorized output showing template name, description, repos
- [x] 4.4 Implement `canopy template show <name>` for detailed view
- [x] 4.5 Add `canopy template validate` to check template definitions

## 5. Documentation & Examples
- [x] 5.1 Add templates section to example config.yaml
- [x] 5.2 Update README.md with template usage examples
- [x] 5.3 Document template format in configuration guide
- [x] 5.4 Add common templates to docs (fullstack, backend-only, frontend-only)

## 6. Testing & Validation
- [x] 6.1 Write unit tests for template resolution
- [x] 6.2 Write integration test for workspace creation with template
- [x] 6.3 Test template + explicit repos combination
- [x] 6.4 Test template with setup commands
- [x] 6.5 Test error handling for invalid templates
