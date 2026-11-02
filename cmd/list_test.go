package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
)

func TestListCmd(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("HOME", t.TempDir())
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addListCmd()

	cmd, _, err := c.rootCmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("expected to find list command: %v", err)
	}
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("| Plugin")) {
		t.Errorf("expected plugin table header in output, got:\n%s", buf.String())
	}
}
