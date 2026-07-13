package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestLoadJenkinsConfig_NoHardcodedDefaultURL(t *testing.T) {
	overrideConfigPath = "/nonexistent/config.toml"
	defer func() { overrideConfigPath = "" }()
	os.Unsetenv("JENKINS_URL")
	os.Unsetenv("JENKINS_USER")
	os.Unsetenv("JENKINS_TOKEN")

	cfg := loadJenkinsConfig()
	if cfg.URL != "" {
		t.Errorf("expected no default Jenkins URL when unconfigured, got %q", cfg.URL)
	}
}

func TestLoadJenkinsConfig_EnvOverride(t *testing.T) {
	overrideConfigPath = "/nonexistent/config.toml"
	defer func() { overrideConfigPath = "" }()
	t.Setenv("JENKINS_URL", "https://jenkins.example.com")
	t.Setenv("JENKINS_TOKEN", "abc123")

	cfg := loadJenkinsConfig()
	if cfg.URL != "https://jenkins.example.com" {
		t.Errorf("expected env JENKINS_URL to be used, got %q", cfg.URL)
	}
	if cfg.TokenFrom != "env" {
		t.Errorf("expected token source 'env', got %q", cfg.TokenFrom)
	}
}

func TestRequireToken_GenericMessage(t *testing.T) {
	cfg := jenkinsConfig{Token: "", ConfigPath: "/tmp/x/config.toml", User: "alice"}
	err := cfg.requireToken()
	if err == nil {
		t.Fatal("expected error when token is empty")
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "taboola") {
		t.Errorf("error message must not mention Taboola-internal tooling: %q", err.Error())
	}
	if !strings.Contains(msg, "/me/configure") {
		t.Errorf("expected generic Jenkins token instructions mentioning /me/configure, got %q", err.Error())
	}
}

func TestRequireToken_NoErrorWhenTokenPresent(t *testing.T) {
	cfg := jenkinsConfig{Token: "present"}
	if err := cfg.requireToken(); err != nil {
		t.Errorf("expected no error when token is set, got %v", err)
	}
}
