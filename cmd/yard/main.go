package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexisbeaulieu97/yard/internal/app"
	"github.com/spf13/cobra"
)

type contextKey string

const appContextKey contextKey = "app"

var (
	debug   bool
	rootCmd = &cobra.Command{
		Use:   "yard",
		Short: "Ticket-centric workspaces",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			appInstance, err := app.New(debug)
			if err != nil {
				return err
			}

			ctx := context.WithValue(cmd.Context(), appContextKey, appInstance)
			cmd.SetContext(ctx)
			cmd.Root().SetContext(ctx)
			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getApp(cmd *cobra.Command) (*app.App, error) {
	value := cmd.Context().Value(appContextKey)
	if value == nil {
		return nil, fmt.Errorf("app not initialized")
	}

	appInstance, ok := value.(*app.App)
	if !ok {
		return nil, fmt.Errorf("invalid app in context")
	}

	return appInstance, nil
}
