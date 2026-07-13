package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

type doctorReport struct {
	NpmGroovyLint    string `json:"npm_groovy_lint"`
	NpxAvailable     bool   `json:"npx_available"`
	JenkinsURL       string `json:"jenkins_url"`
	JenkinsUser      string `json:"jenkins_user"`
	TokenSource      string `json:"token_source"` // env | missing
	JenkinsReachable bool   `json:"jenkins_reachable"`
	ConfigPath       string `json:"config_path"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local tooling, config, and Jenkins connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		report := doctorReport{}

		if _, err := exec.LookPath("npx"); err == nil {
			report.NpxAvailable = true
			if out, err := exec.Command("npx", "--no-install", "npm-groovy-lint", "--version").Output(); err == nil {
				report.NpmGroovyLint = "installed (" + trimNL(string(out)) + ")"
			} else {
				report.NpmGroovyLint = "not installed locally; `npx npm-groovy-lint` will fetch it on first run"
			}
		} else {
			report.NpmGroovyLint = "npx not found — install Node.js"
		}

		jc := loadJenkinsConfig()
		report.JenkinsURL = jc.URL
		report.JenkinsUser = jc.User
		report.ConfigPath = jc.ConfigPath
		report.TokenSource = jc.TokenFrom

		if jc.URL != "" {
			client := http.Client{Timeout: 3 * time.Second}
			if resp, err := client.Head(jc.URL); err == nil {
				resp.Body.Close()
				report.JenkinsReachable = true
			}
		}

		if jsonOutput {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(report)
		}

		fmt.Printf("npm-groovy-lint : %s\n", report.NpmGroovyLint)
		fmt.Printf("jenkins url     : %s (reachable: %v)\n", formatJenkinsURLForDisplay(report.JenkinsURL), report.JenkinsReachable)
		fmt.Printf("jenkins user    : %s\n", report.JenkinsUser)
		fmt.Printf("jenkins token   : %s\n", report.TokenSource)
		fmt.Printf("config path     : %s\n", report.ConfigPath)
		if report.TokenSource == "missing" {
			fmt.Println("\nNo Jenkins token set. `validate`/`request` need one — see the error they print for setup steps.")
		}
		return nil
	},
}

func formatJenkinsURLForDisplay(url string) string {
	if url == "" {
		return "(not configured — set $JENKINS_URL or jenkins_url in ~/.groovy-check/config.toml)"
	}
	return url
}

func trimNL(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
