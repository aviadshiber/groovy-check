package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var lintFix bool

var lintCmd = &cobra.Command{
	Use:   "lint [path]",
	Short: "Run local CodeNarc diagnostics via npm-groovy-lint (no network required)",
	Long: `Runs npx npm-groovy-lint against a Groovy file, Jenkinsfile, or directory
(e.g. a Jenkins Shared Library's vars/ or src/ directory). Does not require
Jenkins access and does not understand the CPS transform — it only catches
static Groovy issues.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) == 1 {
			path = args[0]
		}

		lintArgs := []string{"npm-groovy-lint", "--path", path, "--noserver"}
		if jsonOutput {
			lintArgs = append(lintArgs, "--output", "json")
		}
		if lintFix {
			lintArgs = append(lintArgs, "--fix")
		}

		c := exec.Command("npx", lintArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		// npm-groovy-lint exits non-zero when it finds lint issues — that's a
		// normal result, not a CLI failure, so don't wrap/propagate exec errors here.
		_ = c.Run()
		return nil
	},
}

func init() {
	lintCmd.Flags().BoolVar(&lintFix, "fix", false, "apply auto-fixable rule fixes in place (modifies files on disk)")
	rootCmd.AddCommand(lintCmd)
}
