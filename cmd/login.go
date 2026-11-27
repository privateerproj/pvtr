package cmd

import (
	"github.com/spf13/cobra"

	"github.com/privateerproj/privateer-sdk/command/harness"
)

func (c *CLI) addLoginCmds() {
	loginCmd := harness.GetLoginCmd(func() harness.Writer { return c.writer })
	loginCmd.Args = cobra.NoArgs
	c.rootCmd.AddCommand(loginCmd)

	logoutCmd := harness.GetLogoutCmd(func() harness.Writer { return c.writer })
	logoutCmd.Args = cobra.NoArgs
	c.rootCmd.AddCommand(logoutCmd)
}
