// Package main implements the canopy CLI.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/app"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

type contextKey string

const appContextKey contextKey = "app"

var (
	debug       bool
	showVersion bool
	configPath  string
	rootCmd     = &cobra.Command{
		Use:   "canopy",
		Short: "Workspace-centric development",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Skip app initialization for version command or --version flag
			if cmd.Name() == "version" || showVersion {
				return nil
			}

			appInstance, err := app.New(debug, app.WithConfigPath(configPath))
			if err != nil {
				return err
			}

			ctx := context.WithValue(cmd.Context(), appContextKey, appInstance)
			cmd.SetContext(ctx)
			cmd.Root().SetContext(ctx)
			return nil
		},
		Run: func(cmd *cobra.Command, _ []string) {
			// Handle --version flag on root command
			if showVersion {
				printVersion(cmd.OutOrStdout())
				return
			}
			// Show help when no subcommand is provided
			_ = cmd.Help()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file (overrides CANOPY_CONFIG and default locations)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "V", false, "Print version information and exit")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Check for ExitCodeError to use specific exit code
		var exitErr *ExitCodeError
		if errors.As(err, &exitErr) {
			// ExitCodeError may have an empty message (e.g., doctor command)
			// Only print if there's a message
			if exitErr.Message != "" {
				fmt.Fprintln(os.Stderr, exitErr.Message)
			}

			os.Exit(exitErr.Code)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getApp(cmd *cobra.Command) (*app.App, error) {
	value := cmd.Context().Value(appContextKey)
	if value == nil {
		return nil, cerrors.NewInternalError("app not initialized", nil)
	}

	appInstance, ok := value.(*app.App)
	if !ok {
		return nil, cerrors.NewInternalError("invalid app in context", nil)
	}

	return appInstance, nil
}
