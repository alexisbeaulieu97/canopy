# Tasks: Add Lifecycle Hooks Support

## Implementation Checklist

### 1. Configuration Schema
- [ ] 1.1 Define `Hook` struct in `internal/config/`:
  ```go
  type Hook struct {
      Command string   `mapstructure:"command"`
      Repos   []string `mapstructure:"repos,omitempty"`  // filter to specific repos
      Shell   string   `mapstructure:"shell,omitempty"`  // default: sh -c
  }
  
  type Hooks struct {
      PostCreate []Hook `mapstructure:"post_create"`
      PreClose   []Hook `mapstructure:"pre_close"`
  }
  ```
- [ ] 1.2 Add `Hooks` field to `Config` struct
- [ ] 1.3 Update config loading to parse hooks

### 2. Hook Execution Engine
- [ ] 2.1 Create `internal/hooks/executor.go`
- [ ] 2.2 Implement `ExecuteHooks(hooks []Hook, context HookContext) error`
- [ ] 2.3 Define `HookContext` with workspace path, repo names, etc.
- [ ] 2.4 Run commands in appropriate working directory
- [ ] 2.5 Capture and log stdout/stderr
- [ ] 2.6 Handle errors (fail fast vs continue)

### 3. Integrate with Workspace Lifecycle
- [ ] 3.1 Update `CreateWorkspace()` to run post_create hooks
- [ ] 3.2 Update `CloseWorkspace()` to run pre_close hooks
- [ ] 3.3 Pass hook context with workspace info

### 4. CLI Flags
- [ ] 4.1 Add `--no-hooks` flag to `workspace new`
- [ ] 4.2 Add `--no-hooks` flag to `workspace close`
- [ ] 4.3 Add `--hooks-only` flag to run hooks on existing workspace

### 5. Security Considerations
- [ ] 5.1 Document that hooks run arbitrary commands
- [ ] 5.2 Consider adding `--dry-run` to show hooks without running
- [ ] 5.3 Validate hook commands don't contain dangerous patterns (optional)

### 6. Testing
- [ ] 6.1 Add unit tests for hook execution
- [ ] 6.2 Add integration test for post_create hook
- [ ] 6.3 Add integration test for pre_close hook
- [ ] 6.4 Test `--no-hooks` flag
