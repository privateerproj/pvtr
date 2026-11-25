package cmd

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestGeneratePluginCmd_Flags(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addGenPluginCmd()

	cmd, _, err := c.rootCmd.Find([]string{"generate-plugin"})
	if err != nil {
		t.Fatalf("expected to find generate-plugin command: %v", err)
	}

	tests := []struct {
		name         string
		shorthand    string
		defaultValue string
	}{
		{"source-path", "p", ""},
		{"local-templates", "", ""},
		{"service-name", "n", ""},
		{"organization", "g", ""},
		{"output-dir", "o", "generated-plugin/"},
	}

	for _, tc := range tests {
		flag := cmd.Flags().Lookup(tc.name)
		if flag == nil {
			t.Errorf("expected flag %q to be registered", tc.name)
			continue
		}
		if flag.Shorthand != tc.shorthand {
			t.Errorf("flag %q: expected shorthand %q, got %q", tc.name, tc.shorthand, flag.Shorthand)
		}
		if flag.DefValue != tc.defaultValue {
			t.Errorf("flag %q: expected default value %q, got %q", tc.name, tc.defaultValue, flag.DefValue)
		}
	}
}

func TestGeneratePluginCmd_ViperBindings(t *testing.T) {
	t.Cleanup(viper.Reset)
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addGenPluginCmd()

	cmd, _, err := c.rootCmd.Find([]string{"generate-plugin"})
	if err != nil {
		t.Fatalf("expected to find generate-plugin command: %v", err)
	}

	if err := cmd.Flags().Set("output-dir", "custom-dir/"); err != nil {
		t.Fatalf("unexpected error setting flag: %v", err)
	}
	if val := viper.GetString("output-dir"); val != "custom-dir/" {
		t.Errorf("expected viper to reflect flag value 'custom-dir/', got %q", val)
	}
}

func TestGeneratePluginCmd_RunSuccess(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("PVTR_LOGLEVEL", "")
	t.Cleanup(viper.Reset)
	origLogWriter := log.Writer()
	t.Cleanup(func() { log.SetOutput(origLogWriter) })

	origGeneratePluginFn := generatePluginFn
	defer func() { generatePluginFn = origGeneratePluginFn }()

	called := false
	generatePluginFn = func(c *CLI) int {
		called = true
		return 0
	}

	c := NewCLI("1.0.0", "hash", "date")
	c.rootCmd.SetArgs([]string{"generate-plugin"})
	if err := c.rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing generate-plugin: %v", err)
	}

	if !called {
		t.Error("expected generatePluginFn to be called")
	}
}

func TestGeneratePluginCmd_NoArgsRejectsPositionalArgs(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("PVTR_LOGLEVEL", "")
	t.Cleanup(viper.Reset)
	origLogWriter := log.Writer()
	t.Cleanup(func() { log.SetOutput(origLogWriter) })

	origGeneratePluginFn := generatePluginFn
	defer func() { generatePluginFn = origGeneratePluginFn }()
	generatePluginFn = func(c *CLI) int {
		t.Error("generatePluginFn should not be called when positional args are rejected")
		return 0
	}

	c := NewCLI("1.0.0", "hash", "date")
	c.rootCmd.SetArgs([]string{"generate-plugin", "unexpected-arg"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected generate-plugin command with extra positional arguments to fail")
	}
	if !strings.Contains(err.Error(), `unknown command "unexpected-arg" for "pvtr generate-plugin"`) {
		t.Errorf("expected error message to indicate unknown command violation, got: %v", err)
	}
}

