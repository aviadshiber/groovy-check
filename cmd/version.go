package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var versionInfo struct {
	version string
	commit  string
	date    string
}

// SetVersionInfo stores build metadata for the version command. Called from
// main before Execute, with values injected via -ldflags at build/release time.
func SetVersionInfo(version, commit, date string) {
	versionInfo.version = version
	versionInfo.commit = commit
	versionInfo.date = date
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the groovy-check version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]string{
				"version": versionInfo.version,
				"commit":  versionInfo.commit,
				"date":    versionInfo.date,
			})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "groovy-check version %s (commit: %s, built: %s)\n",
			versionInfo.version, versionInfo.commit, versionInfo.date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
