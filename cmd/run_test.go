package cmd

import (
	"os"
	"path/filepath"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/spf13/viper"

	"github.com/privateerproj/privateer-sdk/command"
	"github.com/privateerproj/privateer-sdk/command/harness"
)

// REV-740: with any plugin installed, a missing or empty config made
// `pvtr run` execute nothing and exit 0, which reads as "every control passed".
func TestRunFn_RefusesConfigThatRunsNothing(t *testing.T) {
	// One installed plugin that no test config requests.
	binDir := t.TempDir()
	manifest := `{"plugins":{"local/example@local":{"name":"local/example","version":"local","binaryPath":"local/example"}}}`
	if err := os.WriteFile(filepath.Join(binDir, "plugins.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgDir := t.TempDir()

	for _, tc := range []struct {
		name   string
		config string // file contents; ignored when absent
		absent bool
		want   int
	}{
		{name: "missing config file", absent: true, want: command.BadUsage},
		{name: "empty config file", config: "", want: command.NoTests},
		{name: "config without services", config: "loglevel: error\n", want: command.NoTests},
		{name: "services block with no entries", config: "services:\n", want: command.NoTests},
		// Still refused, by the SDK: the check must not mask its error.
		{name: "service requests an uninstalled plugin", config: "services:\n  repo:\n    plugin: other\n", want: command.BadUsage},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(viper.Reset)
			path := filepath.Join(cfgDir, tc.name+".yml")
			if !tc.absent {
				if err := os.WriteFile(path, []byte(tc.config), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			viper.Set("config", path)
			command.ReadConfig()
			viper.Set("binaries-path", binDir)

			c := NewCLI("1.0.0", "abc123", "2026-01-01T00:00:00Z")
			c.logger = hclog.NewNullLogger()
			if got := runFn(c); got != tc.want {
				t.Errorf("exit code = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCheckRunnable_AllowsRequestedPlugin(t *testing.T) {
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("services:\n  repo:\n    plugin: example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)

	plugins := func() []*harness.PluginPkg {
		return []*harness.PluginPkg{{Name: "unrelated", Installed: true}, {Name: "example", Requested: true}}
	}
	if code, msg := checkRunnable(plugins); code != command.TestPass {
		t.Errorf("checkRunnable = %d (%s), want %d", code, msg, command.TestPass)
	}
}
