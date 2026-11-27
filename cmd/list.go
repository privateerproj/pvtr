package cmd

import (
	"github.com/spf13/cobra"

	"github.com/privateerproj/privateer-sdk/command/harness"
)

func (c *CLI) addListCmd() {
	listCmd := harness.GetListCmd(func() harness.Writer { return c.writer })
	listCmd.Args = cobra.NoArgs
	c.rootCmd.AddCommand(listCmd)
}
