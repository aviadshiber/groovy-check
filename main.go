package main

import (
	"fmt"
	"os"

	"github.com/aviadshiber/groovy-check/cmd"
)

// Build-time variables injected via ldflags (see .goreleaser.yml / Makefile).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
