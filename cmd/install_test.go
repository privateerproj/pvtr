package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestInstallCmd_NoArgsFails(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addInstallCmd()

	c.rootCmd.SetArgs([]string{"install"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected install to fail without arguments or flags")
	}
	if !strings.Contains(err.Error(), "a plugin coordinate <namespace>/<plugin_id> is required") {
		t.Errorf("expected error to mention plugin coordinate required, got: %v", err)
	}
}

func TestInstallCmd_LocalNonExistentPathFails(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addInstallCmd()

	c.rootCmd.SetArgs([]string{"install", "--local", "/nonexistent/path/to/binary"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected install --local with non-existent path to fail")
	}
	if !strings.Contains(err.Error(), "cannot access") {
		t.Errorf("expected error to mention 'cannot access', got: %v", err)
	}
}
