package cmd

import (
	"strings"
	"testing"
)

func TestFormatJenkinsURL_Empty(t *testing.T) {
	got := formatJenkinsURLForDisplay("")
	if !strings.Contains(got, "not configured") {
		t.Errorf("expected 'not configured' placeholder for empty URL, got %q", got)
	}
}

func TestFormatJenkinsURL_Set(t *testing.T) {
	got := formatJenkinsURLForDisplay("https://jenkins.example.com")
	if got != "https://jenkins.example.com" {
		t.Errorf("expected URL passed through unchanged, got %q", got)
	}
}
