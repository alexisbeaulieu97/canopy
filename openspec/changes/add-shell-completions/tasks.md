# Tasks: Add Shell Completions

## Implementation Checklist

### 1. Basic Completion Command
- [ ] 1.1 Create `cmd/canopy/completion.go`
- [ ] 1.2 Add completion command using Cobra's built-in generator:
  ```go
  var completionCmd = &cobra.Command{
      Use:   "completion [bash|zsh|fish|powershell]",
      Short: "Generate shell completion script",
  }
  ```
- [ ] 1.3 Add subcommands for each shell type
- [ ] 1.4 Register with root command

### 2. Dynamic Workspace Completion
- [ ] 2.1 Create completion function for workspace IDs:
  ```go
  func workspaceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
  ```
- [ ] 2.2 List active workspaces matching prefix
- [ ] 2.3 Register with workspace subcommands that take ID

### 3. Dynamic Repo Completion
- [ ] 3.1 Create completion function for repo names
- [ ] 3.2 Complete from canonical repo list
- [ ] 3.3 Complete from registry aliases
- [ ] 3.4 Register with repo subcommands

### 4. Flag Completion
- [ ] 4.1 Add completion for `--repos` flag (comma-separated)
- [ ] 4.2 Add completion for `--template` flag (when templates exist)

### 5. Documentation
- [ ] 5.1 Add shell-specific installation instructions to README:
  - Bash: `source <(canopy completion bash)`
  - Zsh: `canopy completion zsh > "${fpath[1]}/_canopy"`
  - Fish: `canopy completion fish | source`
- [ ] 5.2 Add to `canopy completion --help` output

### 6. Testing
- [ ] 6.1 Test completion output for each shell
- [ ] 6.2 Verify dynamic completions return expected values
