package main

import (
	"fmt"
	"os"

	"github.com/alexisbeaulieu97/yard/internal/config"
	"github.com/alexisbeaulieu97/yard/internal/logging"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	logger *logging.Logger
	debug  bool
)

var rootCmd = &cobra.Command{
	Use:   "yard",
	Short: "Ticket-centric workspaces",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}
		logger = logging.New(debug)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
