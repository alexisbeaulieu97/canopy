package main

import (
	"encoding/json"
	"fmt"
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

		if jsonOutput {
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data)) //nolint:forbidigo // version output
			return nil
		}

		fmt.Printf("canopy version %s\n", info.Version)   //nolint:forbidigo // version output
		fmt.Printf("commit: %s\n", info.Commit)           //nolint:forbidigo // version output
		fmt.Printf("built: %s\n", info.BuildDate)         //nolint:forbidigo // version output
		fmt.Printf("go: %s\n", info.GoVersion)            //nolint:forbidigo // version output

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().Bool("json", false, "Output version information as JSON")
}

// printVersion prints a short version string and returns true if the version flag was set.
func printVersion() bool {
	fmt.Printf("canopy version %s\n", version) //nolint:forbidigo // version output
	return true
}

