# Tasks: Add Lifecycle Hooks Support

## Implementation Checklist

### Phase 1: Configuration Schema
- [ ] Define `Hook` struct in `internal/config/`:
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
- [ ] Add `Hooks` field to `Config` struct
- [ ] Update config loading to parse hooks

### Phase 2: Hook Execution Engine
- [ ] Create `internal/hooks/executor.go`
- [ ] Implement `ExecuteHooks(hooks []Hook, context HookContext) error`
- [ ] Define `HookContext` with workspace path, repo names, etc.
- [ ] Run commands in appropriate working directory
- [ ] Capture and log stdout/stderr
- [ ] Handle errors (fail fast vs continue)

### Phase 3: Integrate with Workspace Lifecycle
- [ ] Update `CreateWorkspace()` to run post_create hooks
- [ ] Update `CloseWorkspace()` to run pre_close hooks
- [ ] Pass hook context with workspace info

### Phase 4: CLI Flags
- [ ] Add `--no-hooks` flag to `workspace new`
- [ ] Add `--no-hooks` flag to `workspace close`
- [ ] Add `--hooks-only` flag to run hooks on existing workspace

### Phase 5: Security Considerations
- [ ] Document that hooks run arbitrary commands
- [ ] Consider adding `--dry-run` to show hooks without running
- [ ] Validate hook commands don't contain dangerous patterns (optional)

### Phase 6: Testing
- [ ] Add unit tests for hook execution
- [ ] Add integration test for post_create hook
- [ ] Add integration test for pre_close hook
- [ ] Test `--no-hooks` flag
