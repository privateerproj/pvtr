package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoginCmd_FailsWithInvalidHub(t *testing.T) {
	t.Setenv("PVTR_HUB_URL", "http://127.0.0.1:0")
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addLoginCmds()

	c.rootCmd.SetArgs([]string{"login"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected login to fail with unreachable hub URL")
	}
	if !strings.Contains(err.Error(), "hub discovery") {
		t.Errorf("expected error to mention 'hub discovery', got: %v", err)
	}
}

func TestLogoutCmd_FailsWithInvalidHub(t *testing.T) {
	t.Setenv("PVTR_HUB_URL", "http://127.0.0.1:0")
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addLoginCmds()

	c.rootCmd.SetArgs([]string{"logout"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected logout to fail with unreachable hub URL")
	}
	if !strings.Contains(err.Error(), "hub discovery") {
		t.Errorf("expected error to mention 'hub discovery', got: %v", err)
	}
}

func TestLoginCmd_NoArgsRejectsPositionalArgs(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addLoginCmds()

	c.rootCmd.SetArgs([]string{"login", "unexpected-arg"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected login command with extra positional arguments to fail")
	}
	if !strings.Contains(err.Error(), `unknown command "unexpected-arg" for "pvtr login"`) {
		t.Errorf("expected error message to indicate unknown command violation, got: %v", err)
	}
}

func TestLogoutCmd_NoArgsRejectsPositionalArgs(t *testing.T) {
	var buf bytes.Buffer
	c := newTestCLI(&buf, "1.0.0", "hash", "date")
	c.addLoginCmds()

	c.rootCmd.SetArgs([]string{"logout", "unexpected-arg"})
	err := c.rootCmd.Execute()
	if err == nil {
		t.Fatal("expected logout command with extra positional arguments to fail")
	}
	if !strings.Contains(err.Error(), `unknown command "unexpected-arg" for "pvtr logout"`) {
		t.Errorf("expected error message to indicate unknown command violation, got: %v", err)
	}
}
