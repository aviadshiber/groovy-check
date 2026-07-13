package cmd

import (
	"github.com/spf13/cobra"
)

var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "groovy-check",
	Short: "Lint and validate Groovy / Jenkinsfile code",
	Long: `groovy-check wraps two independent checks for Groovy and Jenkinsfile code:

  lint      local static analysis via npm-groovy-lint (CodeNarc rules), no network needed
  validate  authoritative declarative-pipeline syntax check against a live Jenkins server

Neither check understands the Jenkins CPS transform (workflow-cps runtime closure
semantics) used by Jenkins Shared Libraries. A closure that captures an
enclosing-class field inside steps like withCredentials, a custom retry helper,
or similar constructs can pass both of these checks and still fail at runtime
under real Jenkins. Treat lint/validate as a fast pre-check, not a substitute
for testing shared-library changes against a real Jenkins controller.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")
}
