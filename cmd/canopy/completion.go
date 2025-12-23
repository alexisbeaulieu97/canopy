package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion scripts for Canopy.

To load completions:

Bash:
  $ source <(canopy completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ canopy completion bash > /etc/bash_completion.d/canopy
  # macOS:
  $ canopy completion bash > $(brew --prefix)/etc/bash_completion.d/canopy

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ canopy completion zsh > "${fpath[1]}/_canopy"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ canopy completion fish | source

  # To load completions for each session, execute once:
  $ canopy completion fish > ~/.config/fish/completions/canopy.fish

PowerShell:
  PS> canopy completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> canopy completion powershell > canopy.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
