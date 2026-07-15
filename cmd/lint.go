package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aviadshiber/groovy-check/internal/lintconfig"
	"github.com/spf13/cobra"
)

var lintFix bool

// pinnedNpmGroovyLintVersion is pinned (rather than "latest") so npx's
// resolution is deterministic and doesn't silently drift onto a newer,
// unvetted release on every invocation.
const pinnedNpmGroovyLintVersion = "18.0.0"

var lintCmd = &cobra.Command{
	Use:   "lint [path]",
	Short: "Run local CodeNarc diagnostics via npm-groovy-lint (no network required)",
	Long: `Runs npx npm-groovy-lint against a Groovy file, Jenkinsfile, or directory
(e.g. a Jenkins Shared Library's vars/ or src/ directory). Does not require
Jenkins access and does not understand the CPS transform — it only catches
static Groovy issues.

The linter always runs from a fixed scratch directory with a locked-down
config, regardless of the target path. This is deliberate: npx resolves a
project-local ./node_modules/.bin/npm-groovy-lint binary before PATH, and
npm-groovy-lint auto-discovers .groovylintrc.js/.json/.yml by walking up
from the linted path — a .groovylintrc.js is a Node module that gets
executed on load. Running from a fixed cwd with an explicit --config means
linting an untrusted tree never picks up an attacker-controlled binary or
config file from that tree.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) == 1 {
			path = args[0]
		}

		c, err := buildLintCommand(path, jsonOutput, lintFix)
		if err != nil {
			return err
		}

		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		// npm-groovy-lint exits non-zero when it finds lint issues — that's a
		// normal result, not a CLI failure, so don't wrap/propagate exec errors here.
		_ = c.Run()
		return nil
	},
}

// buildLintCommand constructs the npx invocation for linting path. It always
// resolves path to an absolute value and runs the process from a fixed
// scratch directory with a locked-down --config, so the process's cwd and
// config resolution never depend on (or are influenced by) the tree being
// linted — see the RCE note in the command's Long help text.
func buildLintCommand(path string, jsonOut, fix bool) (*exec.Cmd, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving path %q: %w", path, err)
	}

	scratch, err := scratchDir()
	if err != nil {
		return nil, err
	}
	configPath, err := writeDefaultConfig(scratch)
	if err != nil {
		return nil, err
	}

	lintArgs := []string{
		"npm-groovy-lint@" + pinnedNpmGroovyLintVersion,
		"--noserver",
		"--config", configPath,
	}

	if info, statErr := os.Stat(absPath); statErr == nil && !info.IsDir() {
		lintArgs = append(lintArgs, "--path", filepath.Dir(absPath), "--files", filepath.Base(absPath))
	} else {
		lintArgs = append(lintArgs, "--path", absPath)
	}

	if jsonOut {
		lintArgs = append(lintArgs, "--output", "json")
	}
	if fix {
		lintArgs = append(lintArgs, "--fix")
	}

	c := exec.Command("npx", lintArgs...)
	c.Dir = scratch
	return c, nil
}

func scratchDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "groovy-check-npx")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating scratch dir %s: %w", dir, err)
	}
	return dir, nil
}

func writeDefaultConfig(scratch string) (string, error) {
	configPath := filepath.Join(scratch, "default-groovylintrc.json")
	if err := os.WriteFile(configPath, lintconfig.DefaultConfig, 0o644); err != nil {
		return "", fmt.Errorf("writing default lint config to %s: %w", configPath, err)
	}
	return configPath, nil
}

func init() {
	lintCmd.Flags().BoolVar(&lintFix, "fix", false, "apply auto-fixable rule fixes in place (modifies files on disk)")
	rootCmd.AddCommand(lintCmd)
}
