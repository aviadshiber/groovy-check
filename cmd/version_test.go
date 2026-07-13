package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionCmd_TextOutput(t *testing.T) {
	SetVersionInfo("1.2.3", "abc123", "2026-01-01")
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "1.2.3") {
		t.Errorf("expected output to contain version 1.2.3, got %q", buf.String())
	}
}

func TestVersionCmd_JSONOutput(t *testing.T) {
	SetVersionInfo("1.2.3", "abc123", "2026-01-01")
	jsonOutput = true
	defer func() { jsonOutput = false }()

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("expected valid JSON output, got %q: %v", buf.String(), err)
	}
	if got["version"] != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %q", got["version"])
	}
	if !strings.Contains(got["commit"], "abc123") {
		t.Errorf("expected commit abc123, got %q", got["commit"])
	}
}
