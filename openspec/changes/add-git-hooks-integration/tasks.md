# Implementation Tasks

## 1. Configuration Schema
- [ ] 1.1 Add GitHooks map to Config struct
- [ ] 1.2 Support inline script format in YAML
- [ ] 1.3 Support file path references in YAML
- [ ] 1.4 Add validation for valid hook names
- [ ] 1.5 Write tests for config parsing

## 2. Hook Installation Logic
- [ ] 2.1 Create internal/gitx/hooks.go with InstallHook function
- [ ] 2.2 Implement script writing to .git/hooks/
- [ ] 2.3 Set executable permissions on hook files
- [ ] 2.4 Handle file path resolution for external scripts
- [ ] 2.5 Add error handling for hook installation failures

## 3. Integrate with Workspace Creation
- [ ] 3.1 Call hook installation after each worktree is created
- [ ] 3.2 Install hooks for all repos in workspace
- [ ] 3.3 Log hook installation activity
- [ ] 3.4 Continue workspace creation if hook installation fails (warn only)

## 4. Sync Hooks Command
- [ ] 4.1 Implement `yard workspace sync-hooks <ID>` command
- [ ] 4.2 Re-install all hooks for workspace repos
- [ ] 4.3 Show summary of hooks installed per repo
- [ ] 4.4 Add --dry-run flag to preview changes

## 5. Documentation
- [ ] 5.1 Add git_hooks examples to config.yaml
- [ ] 5.2 Document hook format in README
- [ ] 5.3 Provide example hooks (go fmt, golangci-lint)

## 6. Testing
- [ ] 6.1 Unit tests for hook installation
- [ ] 6.2 Integration test verifying hooks execute
- [ ] 6.3 Test with both inline and file-based hooks
