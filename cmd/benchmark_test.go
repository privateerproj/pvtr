package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestBenchmarkCmd_MissingTargetFails(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addBenchmarkCmd()

	c.rootCmd.SetArgs([]string{"benchmark", "dummy-binary"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected benchmark to fail without --target")
	}
	if !strings.Contains(err.Error(), "--target is required") {
		t.Errorf("expected error to mention '--target is required', got: %v", err)
	}
}

func TestBenchmarkCmd_MissingBinaryFails(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addBenchmarkCmd()

	c.rootCmd.SetArgs([]string{"benchmark"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected benchmark to fail without plugin binary argument")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s), received 0") {
		t.Errorf("expected error to mention accepts 1 arg(s), got: %v", err)
	}
}
