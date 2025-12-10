# Tasks: Add Lifecycle Hooks Support

## Implementation Checklist

### 1. Configuration Schema
- [x] 1.1 Define `Hook` struct in `internal/config/`:
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
- [x] 1.2 Add `Hooks` field to `Config` struct
- [x] 1.3 Update config loading to parse hooks

### 2. Hook Execution Engine
- [x] 2.1 Create `internal/hooks/executor.go`
- [x] 2.2 Implement `ExecuteHooks(hooks []Hook, context HookContext) error`
- [x] 2.3 Define `HookContext` with workspace path, repo names, etc.
- [x] 2.4 Run commands in appropriate working directory
- [x] 2.5 Capture and log stdout/stderr
- [x] 2.6 Handle errors (fail fast vs continue)

### 3. Integrate with Workspace Lifecycle
- [x] 3.1 Update `CreateWorkspace()` to run post_create hooks
- [x] 3.2 Update `CloseWorkspace()` to run pre_close hooks
- [x] 3.3 Pass hook context with workspace info

### 4. CLI Flags
- [x] 4.1 Add `--no-hooks` flag to `workspace new`
- [x] 4.2 Add `--no-hooks` flag to `workspace close`
- [x] 4.3 Add `--hooks-only` flag to run hooks on existing workspace

### 5. Security Considerations
- [x] 5.1 Document that hooks run arbitrary commands
- [x] 5.2 Consider adding `--dry-run` to show hooks without running
- [x] 5.3 Validate hook commands don't contain dangerous patterns (optional)

### 6. Testing
- [x] 6.1 Add unit tests for hook execution
- [x] 6.2 Add integration test for post_create hook
- [x] 6.3 Add integration test for pre_close hook
- [x] 6.4 Test `--no-hooks` flag
