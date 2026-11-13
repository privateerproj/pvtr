package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/privateerproj/privateer-sdk/command"
	"github.com/privateerproj/privateer-sdk/command/harness"
)

// runFn is the function used by runCmd to execute plugins. It's a package-level
// variable so tests can swap it out to avoid actually running plugins.
var runFn = func(c *CLI) int {
	if code, msg := checkRunnable(harness.GetPlugins); code != command.TestPass {
		c.logger.Error(msg)
		return code
	}
	c.setupCloseHandler()
	return harness.Run(context.Background(), c.writer, c.logger, harness.GetPlugins)
}

func (c *CLI) addRunCmd() {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run plugins that have been specified in the config.",
		Long: `
When everything is battoned down, it is time to run forth.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			c.logger.Trace("run called")
			exitCode := runFn(c)
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}
	// Execution-specific flags (output, write-directory, etc.) live on the run
	// command rather than the shared root, so their shorthands (e.g. -o) do not
	// collide with sibling subcommands like generate-plugin's --output-dir.
	command.SetRunFlags(runCmd)
	c.rootCmd.AddCommand(runCmd)
}

// checkRunnable refuses a run that cannot evaluate anything, so a missing,
// empty, or service-less config never exits 0 as if every control passed.
// The SDK only reports NoTests when no plugin is installed at all; with any
// plugin installed but none requested it runs nothing and returns 0.
func checkRunnable(getPlugins func() []*harness.PluginPkg) (int, string) {
	// persistentPreRun's ReadConfig only logs a read failure, so check again.
	if err := viper.ReadInConfig(); err != nil {
		return command.BadUsage, fmt.Sprintf("could not read config file: %s", err)
	}
	for _, p := range getPlugins() {
		if p.Requested {
			return command.TestPass, ""
		}
	}
	return command.NoTests, fmt.Sprintf("config file %s requests no plugins: define at least one target with a plugin under 'targets'", viper.ConfigFileUsed())
}
