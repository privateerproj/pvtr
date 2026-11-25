package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/tabwriter"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newTestCLI(buf *bytes.Buffer, version, commit, buildTime string) *CLI {
	c := &CLI{
		buildVersion:       version,
		buildGitCommitHash: commit,
		buildTime:          buildTime,
		writer:             tabwriter.NewWriter(buf, 1, 1, 1, ' ', 0),
	}
	c.rootCmd = &cobra.Command{Use: "pvtr"}
	return c
}

func TestEnvCmd_ContainsExpectedFields(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.2.3", "a1b2c3d", "2026-01-01T00:00:00Z")
	c.addEnvCmd()

	envCmd, _, _ := c.rootCmd.Find([]string{"env"})
	envCmd.Run(envCmd, []string{})

	output := buf.String()

	expectedFields := []string{
		"Binary:",
		"Config:",
		"Plugins Dir:",
		"Plugins:",
		"Version:",
		"Commit:",
		"Build Time:",
		"Go Version:",
		"OS/Arch:",
	}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("expected output to contain %q, got:\n%s", field, output)
		}
	}
}

func TestEnvCmd_DisplaysBuildInfo(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "test-version", "e4f5a6b", "2026-02-15T00:00:00Z")
	c.addEnvCmd()

	envCmd, _, _ := c.rootCmd.Find([]string{"env"})
	envCmd.Run(envCmd, []string{})

	output := buf.String()

	if !strings.Contains(output, "test-version") {
		t.Errorf("expected output to contain version 'test-version', got:\n%s", output)
	}
	if !strings.Contains(output, "e4f5a6b") {
		t.Errorf("expected output to contain commit 'e4f5a6b', got:\n%s", output)
	}
	if !strings.Contains(output, "2026-02-15T00:00:00Z") {
		t.Errorf("expected output to contain build time, got:\n%s", output)
	}
	if !strings.Contains(output, runtime.Version()) {
		t.Errorf("expected output to contain Go version %q, got:\n%s", runtime.Version(), output)
	}
	expectedArch := runtime.GOOS + "/" + runtime.GOARCH
	if !strings.Contains(output, expectedArch) {
		t.Errorf("expected output to contain OS/Arch %q, got:\n%s", expectedArch, output)
	}
}

func TestEnvCmd_ShowsBinaryPath(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "", "", "")
	c.addEnvCmd()

	envCmd, _, _ := c.rootCmd.Find([]string{"env"})
	envCmd.Run(envCmd, []string{})

	output := buf.String()

	execPath, err := os.Executable()
	if err == nil && !strings.Contains(output, execPath) {
		t.Errorf("expected output to contain executable path %q, got:\n%s", execPath, output)
	}
}

func TestDiscoverPluginNames_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pvtr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	result := discoverPluginNames(hclog.NewNullLogger(), tmpDir)
	if result != "none" {
		t.Errorf("expected 'none' for empty dir, got: %s", result)
	}
}

func TestDiscoverPluginNames_NonexistentDir(t *testing.T) {
	result := discoverPluginNames(hclog.NewNullLogger(), "/nonexistent/path")
	if result != "none" {
		t.Errorf("expected 'none' for nonexistent dir, got: %s", result)
	}
}

func writeManifest(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "plugins.json"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
}

func TestDiscoverPluginNames_ReadsManifest(t *testing.T) {
	tmpDir := t.TempDir()
	// A grc.store install lives at a nested path; top level holds only the
	// coordinate folder and the manifest (the layout from issue #348).
	nested := filepath.Join(tmpDir, "local", "myplugin", "local")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "myplugin"), []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	writeManifest(t, tmpDir, `{"plugins":{
		"local/myplugin@local":{"name":"local/myplugin","version":"local","binaryPath":"local/myplugin/local/myplugin"},
		"ossf/pvtr-github-repo-scanner@v0.2.0":{"name":"ossf/pvtr-github-repo-scanner","version":"v0.2.0","binaryPath":"github-repo"}
	}}`)

	result := discoverPluginNames(hclog.NewNullLogger(), tmpDir)
	want := "local/myplugin@local, ossf/pvtr-github-repo-scanner@v0.2.0"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
	if strings.Contains(result, "plugins.json") {
		t.Errorf("manifest file must not be listed as a plugin, got: %s", result)
	}
}

func TestDiscoverPluginNames_LegacyArrayManifest(t *testing.T) {
	tmpDir := t.TempDir()
	writeManifest(t, tmpDir, `{"plugins":[{"name":"b/two","version":"v2"},{"name":"a/one","version":"v1"}]}`)

	result := discoverPluginNames(hclog.NewNullLogger(), tmpDir)
	if want := "a/one@v1, b/two@v2"; result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestDiscoverPluginNames_EmptyManifest(t *testing.T) {
	tmpDir := t.TempDir()
	writeManifest(t, tmpDir, `{"plugins":{}}`)

	if result := discoverPluginNames(hclog.NewNullLogger(), tmpDir); result != "none" {
		t.Errorf("expected 'none' for empty manifest, got: %s", result)
	}
}

func TestDiscoverPluginNames_IgnoresLooseFilesWithoutManifest(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"pvtr", "my-plugin"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
	}

	if result := discoverPluginNames(hclog.NewNullLogger(), tmpDir); result != "none" {
		t.Errorf("expected 'none' without a manifest, got: %s", result)
	}
}

func TestDiscoverPluginNames_MalformedManifest(t *testing.T) {
	tmpDir := t.TempDir()
	writeManifest(t, tmpDir, `not json`)

	if result := discoverPluginNames(hclog.NewNullLogger(), tmpDir); result != "none" {
		t.Errorf("expected 'none' for malformed manifest, got: %s", result)
	}
}

func TestEnvCmd_ConfigStatusFound(t *testing.T) {
	t.Cleanup(viper.Reset)
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(cfgPath, []byte("key: value\n"), 0644); err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	viper.SetConfigFile(cfgPath)

	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addEnvCmd()

	envCmd, _, _ := c.rootCmd.Find([]string{"env"})
	envCmd.Run(envCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "(found)") {
		t.Errorf("expected output to contain '(found)', got:\n%s", output)
	}
	if !strings.Contains(output, cfgPath) {
		t.Errorf("expected output to contain config path %q, got:\n%s", cfgPath, output)
	}
}

func TestEnvCmd_ConfigStatusNotFound(t *testing.T) {
	t.Cleanup(viper.Reset)
	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent.yml")
	viper.SetConfigFile(nonExistentPath)

	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addEnvCmd()

	envCmd, _, _ := c.rootCmd.Find([]string{"env"})
	envCmd.Run(envCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "(not found)") {
		t.Errorf("expected output to contain '(not found)', got:\n%s", output)
	}
	if !strings.Contains(output, nonExistentPath) {
		t.Errorf("expected output to contain config path %q, got:\n%s", nonExistentPath, output)
	}
}

func TestDiscoverPluginNames_EntryWithoutVersion(t *testing.T) {
	tmpDir := t.TempDir()
	writeManifest(t, tmpDir, `{"plugins":{"a/one@":{"name":"a/one"},"b/two@v2":{"name":"b/two","version":"v2"}}}`)

	result := discoverPluginNames(hclog.NewNullLogger(), tmpDir)
	if want := "a/one, b/two@v2"; result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestDiscoverPluginNames_SkipsEmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	writeManifest(t, tmpDir, `{"plugins":[{"name":"","version":"v1"},{"name":"a/one","version":"v1"}]}`)

	result := discoverPluginNames(hclog.NewNullLogger(), tmpDir)
	if want := "a/one@v1"; result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestDiscoverPluginNames_DeDuplicates(t *testing.T) {
	tmpDir := t.TempDir()
	// The legacy array format can carry the same name+version twice.
	writeManifest(t, tmpDir, `{"plugins":[{"name":"a/one","version":"v1"},{"name":"a/one","version":"v1"},{"name":"a/one","version":"v2"}]}`)

	result := discoverPluginNames(hclog.NewNullLogger(), tmpDir)
	if want := "a/one@v1, a/one@v2"; result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestDiscoverPluginNames_UnreadableManifest(t *testing.T) {
	tmpDir := t.TempDir()
	// A directory where the manifest should be makes ReadFile fail with an
	// error other than "not exist", regardless of the user running the test.
	if err := os.Mkdir(filepath.Join(tmpDir, "plugins.json"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	logger := &errorCountLogger{}
	if result := discoverPluginNames(logger, tmpDir); result != "none" {
		t.Errorf("expected 'none' for unreadable manifest, got: %s", result)
	}
	if logger.errors != 1 {
		t.Errorf("expected 1 Error call for an unreadable manifest, got %d", logger.errors)
	}
}

func TestDiscoverPluginNames_MissingManifestLogsNothing(t *testing.T) {
	logger := &errorCountLogger{}
	if result := discoverPluginNames(logger, t.TempDir()); result != "none" {
		t.Errorf("expected 'none' without a manifest, got: %s", result)
	}
	if logger.errors != 0 {
		t.Errorf("expected no Error calls for a missing manifest, got %d", logger.errors)
	}
}

func TestDiscoverPluginNames_WrongFieldTypeMatchesSDK(t *testing.T) {
	tmpDir := t.TempDir()
	// The SDK types binaryPath as a string, so it refuses this manifest; env
	// must not list a plugin that list and run cannot see.
	writeManifest(t, tmpDir, `{"plugins":{"org/plugin@1.0.0":{"name":"org/plugin","version":"1.0.0","binaryPath":123}}}`)

	logger := &errorCountLogger{}
	if result := discoverPluginNames(logger, tmpDir); result != "none" {
		t.Errorf("expected 'none' for a manifest the SDK rejects, got: %s", result)
	}
	if logger.errors != 1 {
		t.Errorf("expected 1 Error call, got %d", logger.errors)
	}
}

// errorCountLogger counts Error calls so tests can tell a reported read error
// from a silently missing manifest.
type errorCountLogger struct{ errors int }

func (l *errorCountLogger) Error(string, ...interface{}) { l.errors++ }

func TestEnvCmd_NoArgsRejectsPositionalArgs(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.2.3", "a1b2c3d", "2026-01-01T00:00:00Z")
	c.addEnvCmd()

	c.rootCmd.SetArgs([]string{"env", "unexpected-arg"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected env command with extra positional arguments to fail")
	}
	if !strings.Contains(err.Error(), `unknown command "unexpected-arg" for "pvtr env"`) {
		t.Errorf("expected error message to indicate unknown command violation, got: %v", err)
	}
}

