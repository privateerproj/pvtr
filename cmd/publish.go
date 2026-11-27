package cmd

import (
	"github.com/spf13/cobra"

	"github.com/privateerproj/privateer-sdk/command/harness"
)

func (c *CLI) addPublishCmd() {
	publishCmd := harness.GetPublishCmd(func() harness.Writer { return c.writer })
	publishCmd.Args = cobra.NoArgs
	c.rootCmd.AddCommand(publishCmd)
}
