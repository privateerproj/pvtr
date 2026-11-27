package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPublishCmd_EmptyDistFails(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addPublishCmd()

	cmd, _, err := c.rootCmd.Find([]string{"publish"})
	if err != nil {
		t.Fatalf("expected to find publish command: %v", err)
	}
	if err := cmd.Flags().Set("dist", t.TempDir()); err != nil {
		t.Fatal(err)
	}
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected publish to fail on an empty dist dir")
	}
	if !strings.Contains(err.Error(), "loading GoReleaser build") {
		t.Errorf("expected error to mention 'loading GoReleaser build', got: %v", err)
	}
}

func TestPublishCmd_NoArgsRejectsPositionalArgs(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addPublishCmd()

	c.rootCmd.SetArgs([]string{"publish", "unexpected-arg"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected publish command with extra positional arguments to fail")
	}
	if !strings.Contains(err.Error(), `unknown command "unexpected-arg" for "pvtr publish"`) {
		t.Errorf("expected error message to indicate unknown command violation, got: %v", err)
	}
}
