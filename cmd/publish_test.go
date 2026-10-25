package cmd

import (
	"bytes"
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
	if err := cmd.RunE(cmd, []string{}); err == nil {
		t.Fatal("expected publish to fail on an empty dist dir")
	}
}
