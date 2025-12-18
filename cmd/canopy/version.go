package main

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/cobra"
)

// Version variables are set via ldflags at build time.
// Example: go build -ldflags "-X main.version=v1.0.0 -X main.commit=abc123 -X main.buildDate=2025-01-01T00:00:00Z"
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

// VersionInfo holds version information for JSON output.
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the version, commit hash, build date, and Go version used to build canopy.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		info := VersionInfo{
			Version:   version,
			Commit:    commit,
			BuildDate: buildDate,
			GoVersion: runtime.Version(),
		}

		out := cmd.OutOrStdout()

		if jsonOutput {
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(out, string(data))

			return err
		}

		_, err := fmt.Fprintf(out, "canopy version %s\ncommit: %s\nbuilt: %s\ngo: %s\n",
			info.Version, info.Commit, info.BuildDate, info.GoVersion)

		return err
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().Bool("json", false, "Output version information as JSON")
}

// printVersion prints a short version string to the given writer.
// Write errors are intentionally ignored as this is CLI output with no recovery path.
func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "canopy version %s\n", version)
}
