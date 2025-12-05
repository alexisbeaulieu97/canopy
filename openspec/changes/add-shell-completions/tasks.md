# Tasks: Add Shell Completions

## Implementation Checklist

### Phase 1: Basic Completion Command
- [ ] Create `cmd/canopy/completion.go`
- [ ] Add completion command using Cobra's built-in generator:
  ```go
  var completionCmd = &cobra.Command{
      Use:   "completion [bash|zsh|fish|powershell]",
      Short: "Generate shell completion script",
  }
  ```
- [ ] Add subcommands for each shell type
- [ ] Register with root command

### Phase 2: Dynamic Workspace Completion
- [ ] Create completion function for workspace IDs:
  ```go
  func workspaceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
  ```
- [ ] List active workspaces matching prefix
- [ ] Register with workspace subcommands that take ID

### Phase 3: Dynamic Repo Completion
- [ ] Create completion function for repo names
- [ ] Complete from canonical repo list
- [ ] Complete from registry aliases
- [ ] Register with repo subcommands

### Phase 4: Flag Completion
- [ ] Add completion for `--repos` flag (comma-separated)
- [ ] Add completion for `--template` flag (when templates exist)

### Phase 5: Documentation
- [ ] Add shell-specific installation instructions to README:
  - Bash: `source <(canopy completion bash)`
  - Zsh: `canopy completion zsh > "${fpath[1]}/_canopy"`
  - Fish: `canopy completion fish | source`
- [ ] Add to `canopy completion --help` output

### Phase 6: Testing
- [ ] Test completion output for each shell
- [ ] Verify dynamic completions return expected values
