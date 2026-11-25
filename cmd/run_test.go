package cmd

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

type runContextKey struct{}

func TestRunCmd_Flags(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addRunCmd()

	cmd, _, err := c.rootCmd.Find([]string{"run"})
	if err != nil {
		t.Fatalf("expected to find run command: %v", err)
	}

	tests := []struct {
		name         string
		shorthand    string
		defaultValue string
	}{
		{"write-directory", "w", "evaluation_results"},
		{"service", "s", ""},
		{"target", "", ""},
		{"test-suites", "t", "default"},
		{"silent", "", "false"},
		{"write", "", "true"},
		{"output", "o", "yaml"},
		{"include-payload", "", "false"},
		{"concurrency", "", "1"},
	}

	for _, tc := range tests {
		flag := cmd.PersistentFlags().Lookup(tc.name)
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

func TestRunCmd_ViperBindings(t *testing.T) {
	t.Cleanup(viper.Reset)
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addRunCmd()

	cmd, _, err := c.rootCmd.Find([]string{"run"})
	if err != nil {
		t.Fatalf("expected to find run command: %v", err)
	}

	if err := cmd.PersistentFlags().Set("output", "json"); err != nil {
		t.Fatalf("unexpected error setting flag: %v", err)
	}
	if val := viper.GetString("output"); val != "json" {
		t.Errorf("expected viper to reflect flag value 'json', got %q", val)
	}
}

func TestRunCmd_RunSuccess(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("PVTR_LOGLEVEL", "")
	t.Cleanup(viper.Reset)
	origLogWriter := log.Writer()
	t.Cleanup(func() { log.SetOutput(origLogWriter) })

	origRunFn := runFn
	defer func() { runFn = origRunFn }()

	called := false
	runFn = func(ctx context.Context, c *CLI) int {
		called = true
		return 0
	}

	c := NewCLI("1.0.0", "hash", "date")
	c.rootCmd.SetArgs([]string{"run"})
	if err := c.rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing run: %v", err)
	}

	if !called {
		t.Error("expected runFn to be called")
	}
}

func TestRunCmd_PropagatesContext(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("PVTR_LOGLEVEL", "")
	t.Cleanup(viper.Reset)
	origLogWriter := log.Writer()
	t.Cleanup(func() { log.SetOutput(origLogWriter) })

	origRunFn := runFn
	defer func() { runFn = origRunFn }()

	ctx := context.WithValue(context.Background(), runContextKey{}, "test-marker")
	var receivedCtx context.Context

	runFn = func(ctx context.Context, c *CLI) int {
		receivedCtx = ctx
		return 0
	}

	c := NewCLI("1.0.0", "hash", "date")
	c.rootCmd.SetArgs([]string{"run"})
	if err := c.rootCmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("unexpected error executing run with context: %v", err)
	}

	if receivedCtx == nil {
		t.Fatal("expected runFn to receive non-nil context")
	}
	if val, ok := receivedCtx.Value(runContextKey{}).(string); !ok || val != "test-marker" {
		t.Errorf("expected context value 'test-marker', got %v", val)
	}
}

func TestRunCmd_NoArgsRejectsPositionalArgs(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("PVTR_LOGLEVEL", "")
	t.Cleanup(viper.Reset)
	origLogWriter := log.Writer()
	t.Cleanup(func() { log.SetOutput(origLogWriter) })

	origRunFn := runFn
	defer func() { runFn = origRunFn }()
	runFn = func(context.Context, *CLI) int {
		t.Error("runFn should not be called when positional args are rejected")
		return 0
	}

	c := NewCLI("1.0.0", "hash", "date")
	c.rootCmd.SetArgs([]string{"run", "unexpected-arg"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected run command with extra positional arguments to fail")
	}
	if !strings.Contains(err.Error(), `unknown command "unexpected-arg" for "pvtr run"`) {
		t.Errorf("expected error message to indicate unknown command violation, got: %v", err)
	}
}
