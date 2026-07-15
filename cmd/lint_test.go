package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildLintCommand_FilePath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "Sample.groovy")
	if err := os.WriteFile(filePath, []byte("def foo() { return 1 }\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	c, err := buildLintCommand(filePath, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}

	assertArgsContainInOrder(t, c.Args, "--path", dir)
	assertArgsContainInOrder(t, c.Args, "--files", "Sample.groovy")
}

func TestBuildLintCommand_DirectoryPath(t *testing.T) {
	dir := t.TempDir()

	c, err := buildLintCommand(dir, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}

	assertArgsContainInOrder(t, c.Args, "--path", dir)
	if containsArg(c.Args, "--files") {
		t.Errorf("expected no --files flag for a directory path, got args %v", c.Args)
	}
}

func TestBuildLintCommand_CwdIsAlwaysScratchDir(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "Sample.groovy")
	if err := os.WriteFile(filePath, []byte("def foo() { return 1 }\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	c, err := buildLintCommand(filePath, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}

	scratch, err := scratchDir()
	if err != nil {
		t.Fatalf("scratchDir: %v", err)
	}

	if c.Dir != scratch {
		t.Errorf("expected cmd.Dir to be the fixed scratch dir %q, got %q", scratch, c.Dir)
	}
	if c.Dir == dir {
		t.Errorf("cmd.Dir must never equal the linted file's own directory (this is the RCE vector this test guards against)")
	}
}

func TestBuildLintCommand_AlwaysPassesConfigInsideScratchDir(t *testing.T) {
	dir := t.TempDir()

	c, err := buildLintCommand(dir, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}

	scratch, err := scratchDir()
	if err != nil {
		t.Fatalf("scratchDir: %v", err)
	}

	configPath := findArgValue(c.Args, "--config")
	if configPath == "" {
		t.Fatal("expected --config flag to always be present")
	}
	if !strings.HasPrefix(configPath, scratch) {
		t.Errorf("expected --config path to live inside the scratch dir %q, got %q", scratch, configPath)
	}
}

func TestBuildLintCommand_PinnedVersion(t *testing.T) {
	dir := t.TempDir()

	c, err := buildLintCommand(dir, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}

	found := false
	for _, a := range c.Args {
		if a == "npm-groovy-lint@"+pinnedNpmGroovyLintVersion {
			found = true
		}
		if a == "npm-groovy-lint" || a == "npm-groovy-lint@latest" {
			t.Errorf("expected a pinned npm-groovy-lint version, found unpinned arg %q", a)
		}
	}
	if !found {
		t.Errorf("expected pinned package spec npm-groovy-lint@%s in args %v", pinnedNpmGroovyLintVersion, c.Args)
	}
}

func TestBuildLintCommand_JSONAndFixFlags(t *testing.T) {
	dir := t.TempDir()

	c, err := buildLintCommand(dir, true, true)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}

	assertArgsContainInOrder(t, c.Args, "--output", "json")
	if !containsArg(c.Args, "--fix") {
		t.Errorf("expected --fix flag in args %v", c.Args)
	}
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

func findArgValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func assertArgsContainInOrder(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, a := range args {
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return
		}
	}
	t.Errorf("expected args to contain %q followed by %q, got %v", flag, value, args)
}
