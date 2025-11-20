package main

import (
	"fmt"

	"github.com/alexisbeaulieu97/yard/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal user interface",
	RunE: func(cmd *cobra.Command, args []string) error {
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
