package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/yard/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal user interface",
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		printPath, _ := cmd.Flags().GetBool("print-path")

		p := tea.NewProgram(tui.NewModel(app.Service, app.Config.WorkspacesRoot, printPath))
		m, err := p.Run()
		if err != nil {
			return err
		}

		if printPath {
			if model, ok := m.(tui.Model); ok {
				if model.SelectedPath != "" {
					fmt.Println(model.SelectedPath) //nolint:forbidigo // user-facing CLI output
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
