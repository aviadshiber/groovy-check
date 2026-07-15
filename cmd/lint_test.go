package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aviadshiber/groovy-check/internal/lintconfig"
)

func TestBuildLintCommand_FilePathPassedPositionally(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "Sample.groovy")
	if err := os.WriteFile(filePath, []byte("def foo() { return 1 }\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	c, cleanup, err := buildLintCommand(filePath, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}
	t.Cleanup(cleanup)

	// npm-groovy-lint's --path/--files flags are deprecated and, verified
	// empirically, do not reliably scope to a single named file (they still
	// lint every default-pattern-matching file in --path's directory). The
	// only reliable single-file scoping is a bare positional argument.
	if containsArg(c.Args, "--path") || containsArg(c.Args, "--files") {
		t.Errorf("expected no --path/--files flags (deprecated, unreliable for scoping), got args %v", c.Args)
	}
	if c.Args[len(c.Args)-1] != filePath {
		t.Errorf("expected the absolute file path as the last (positional) arg, got args %v", c.Args)
	}
}

func TestBuildLintCommand_DirectoryPathPassedPositionally(t *testing.T) {
	dir := t.TempDir()

	c, cleanup, err := buildLintCommand(dir, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}
	t.Cleanup(cleanup)

	if containsArg(c.Args, "--path") || containsArg(c.Args, "--files") {
		t.Errorf("expected no --path/--files flags, got args %v", c.Args)
	}
	if c.Args[len(c.Args)-1] != dir {
		t.Errorf("expected the absolute directory path as the last (positional) arg, got args %v", c.Args)
	}
}

func TestBuildLintCommand_RelativePathResolvedToAbsolute(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "Sample.groovy")
	if err := os.WriteFile(filePath, []byte("def foo() { return 1 }\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	wantPath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", filePath, err)
	}

	chdir(t, dir)

	c, cleanup, err := buildLintCommand("Sample.groovy", false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}
	t.Cleanup(cleanup)

	gotPath := c.Args[len(c.Args)-1]
	if gotPath == "" || !filepath.IsAbs(gotPath) {
		t.Fatalf("expected the positional arg to be resolved to an absolute path, got %q", gotPath)
	}
	gotPathResolved, err := filepath.EvalSymlinks(gotPath)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", gotPath, err)
	}
	if gotPathResolved != wantPath {
		t.Errorf("expected the relative input to resolve against the cwd to %q, got %q", wantPath, gotPathResolved)
	}
}

// TestBuildLintCommand_SecurityInvariantsHoldForBothInputShapes proves the
// three security-critical properties (neutral+unique cwd, locked-down
// config scoped inside that cwd, pinned linter version) hold regardless of
// whether the caller passes a file or a directory.
func TestBuildLintCommand_SecurityInvariantsHoldForBothInputShapes(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "Sample.groovy")
	if err := os.WriteFile(filePath, []byte("def foo() { return 1 }\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cases := []struct {
		name string
		path string
	}{
		{"file", filePath},
		{"directory", dir},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, cleanup, err := buildLintCommand(tc.path, false, false)
			if err != nil {
				t.Fatalf("buildLintCommand: %v", err)
			}
			t.Cleanup(cleanup)

			// cwd must be a fresh scratch dir, never the linted tree itself —
			// this is the direct guard against the npx local
			// ./node_modules/.bin resolution vector.
			if c.Dir == "" || c.Dir == dir || c.Dir == filepath.Dir(filePath) {
				t.Fatalf("cmd.Dir must be a neutral scratch dir, never the linted tree; got %q", c.Dir)
			}
			if !strings.HasPrefix(filepath.Base(c.Dir), "groovy-check-npx-") {
				t.Errorf("expected cmd.Dir to be a groovy-check-npx-* scratch dir, got %q", c.Dir)
			}
			if info, statErr := os.Stat(c.Dir); statErr != nil || !info.IsDir() {
				t.Errorf("expected cmd.Dir %q to exist as a directory", c.Dir)
			}

			// --config must live INSIDE that same scratch dir (exact
			// containment, not a string prefix — a sibling directory like
			// <tmp>/groovy-check-npx-evil could satisfy a bare prefix check).
			configPath := findArgValue(c.Args, "--config")
			if configPath == "" {
				t.Fatal("expected --config to always be present")
			}
			if filepath.Dir(configPath) != c.Dir {
				t.Errorf("expected --config to live directly inside cmd.Dir %q, got %q", c.Dir, configPath)
			}

			// The config file's actual content must be the locked-down
			// embedded default — not just present, but correct — since an
			// empty or wrong file would silently fall back to
			// npm-groovy-lint's own auto-discovery.
			got, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("reading written config %q: %v", configPath, err)
			}
			if string(got) != string(lintconfig.DefaultConfig) {
				t.Errorf("config file content = %q, want embedded default %q", got, lintconfig.DefaultConfig)
			}

			if !containsArg(c.Args, "npm-groovy-lint@"+pinnedNpmGroovyLintVersion) {
				t.Errorf("expected pinned package spec npm-groovy-lint@%s in args %v", pinnedNpmGroovyLintVersion, c.Args)
			}
			for _, a := range c.Args {
				if a == "npm-groovy-lint" || a == "npm-groovy-lint@latest" {
					t.Errorf("expected a pinned npm-groovy-lint version, found unpinned arg %q", a)
				}
			}
		})
	}
}

func TestBuildLintCommand_EachCallGetsADistinctScratchDir(t *testing.T) {
	dir := t.TempDir()

	c1, cleanup1, err := buildLintCommand(dir, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand (1): %v", err)
	}
	t.Cleanup(cleanup1)

	c2, cleanup2, err := buildLintCommand(dir, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand (2): %v", err)
	}
	t.Cleanup(cleanup2)

	if c1.Dir == c2.Dir {
		t.Errorf("expected each invocation to get its own unique scratch dir, both got %q", c1.Dir)
	}
}

func TestBuildLintCommand_CleanupRemovesScratchDir(t *testing.T) {
	dir := t.TempDir()

	c, cleanup, err := buildLintCommand(dir, false, false)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}
	scratch := c.Dir

	cleanup()

	if _, statErr := os.Stat(scratch); !os.IsNotExist(statErr) {
		t.Errorf("expected cleanup to remove the scratch dir %q, stat err: %v", scratch, statErr)
	}
}

func TestBuildLintCommand_JSONAndFixFlags(t *testing.T) {
	dir := t.TempDir()

	c, cleanup, err := buildLintCommand(dir, true, true)
	if err != nil {
		t.Fatalf("buildLintCommand: %v", err)
	}
	t.Cleanup(cleanup)

	assertArgsContainInOrder(t, c.Args, "--output", "json")
	if !containsArg(c.Args, "--fix") {
		t.Errorf("expected --fix flag in args %v", c.Args)
	}
}

// chdir changes the working directory for the duration of the test and
// restores it afterward. (Not using testing.T.Chdir: CI pins the Go
// toolchain to 1.23 for golangci-lint compatibility, and T.Chdir requires
// Go 1.24+.)
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q): %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})
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
