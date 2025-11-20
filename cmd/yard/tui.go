package main

import (
	"fmt"

	"github.com/alexisbeaulieu97/yard/internal/config"
	"github.com/alexisbeaulieu97/yard/internal/gitx"
	"github.com/alexisbeaulieu97/yard/internal/tui"
	"github.com/alexisbeaulieu97/yard/internal/workspace"
	"github.com/alexisbeaulieu97/yard/internal/workspaces"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal user interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		gitEngine := gitx.New(cfg.ProjectsRoot)
		wsEngine := workspace.New(cfg.WorkspacesRoot)
		svc := workspaces.NewService(cfg, gitEngine, wsEngine, logger)

		printPath, _ := cmd.Flags().GetBool("print-path")

		p := tea.NewProgram(tui.NewModel(svc, cfg.WorkspacesRoot, printPath))
		m, err := p.Run()
		if err != nil {
			return err
		}

		if printPath {
			if model, ok := m.(tui.Model); ok {
				if model.SelectedPath != "" {
					fmt.Println(model.SelectedPath)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().Bool("print-path", false, "Print the selected workspace path to stdout on exit")
}
