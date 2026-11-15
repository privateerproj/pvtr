package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *CLI) addEnvCmd() {
	envCmd := &cobra.Command{
		Use:     "env",
		Aliases: []string{"info"},
		Short:   "Display runtime environment details.",
		Long:    `Display the binary path, config location, plugins directory, installed plugins, version, and build information.`,
		Run: func(cmd *cobra.Command, args []string) {
			binaryPath, err := os.Executable()
			if err != nil {
				binaryPath = "unknown"
			}

			configFile := viper.ConfigFileUsed()
			var configStatus string
			if configFile == "" {
				configStatus = "none"
			} else if _, err := os.Stat(configFile); err == nil {
				configStatus = fmt.Sprintf("%s (found)", configFile)
			} else {
				configStatus = fmt.Sprintf("%s (not found)", configFile)
			}

			pluginsDir := viper.GetString("binaries-path")
			pluginNames := discoverPluginNames(c.logger, pluginsDir)

			_, _ = fmt.Fprintf(c.writer, "Binary:\t%s\n", binaryPath)
			_, _ = fmt.Fprintf(c.writer, "Config:\t%s\n", configStatus)
			_, _ = fmt.Fprintf(c.writer, "Plugins Dir:\t%s\n", pluginsDir)
			_, _ = fmt.Fprintf(c.writer, "Plugins:\t%s\n", pluginNames)
			_, _ = fmt.Fprintf(c.writer, "Version:\t%s\n", c.buildVersion)
			_, _ = fmt.Fprintf(c.writer, "Commit:\t%s\n", c.buildGitCommitHash)
			_, _ = fmt.Fprintf(c.writer, "Build Time:\t%s\n", c.buildTime)
			_, _ = fmt.Fprintf(c.writer, "Go Version:\t%s\n", runtime.Version())
			_, _ = fmt.Fprintf(c.writer, "OS/Arch:\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
			if err := c.writer.Flush(); err != nil {
				log.Printf("Error flushing writer: %v", err)
			}
		},
	}
	c.rootCmd.AddCommand(envCmd)
}

// pluginManifestFile is the manifest the SDK writes to binaries-path to track
// installed plugins (the same source `pvtr list --installed` reads).
const pluginManifestFile = "plugins.json"

type manifestPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// discoverPluginNames lists installed plugins from {pluginsDir}/plugins.json as
// "name@version", sorted. Plugins installed from grc.store live at nested
// paths, so the manifest — not a directory scan — is the source of truth.
// Returns "none" when the manifest is missing, empty, or unreadable.
func discoverPluginNames(logger interface{ Error(string, ...interface{}) }, pluginsDir string) string {
	plugins, err := loadManifestPlugins(pluginsDir)
	if err != nil {
		logger.Error("error reading plugin manifest", "dir", pluginsDir, "error", err)
		return "none"
	}
	seen := make(map[string]bool)
	var names []string
	for _, p := range plugins {
		if p.Name == "" {
			continue
		}
		entry := p.Name
		if p.Version != "" {
			entry += "@" + p.Version
		}
		if seen[entry] {
			continue
		}
		seen[entry] = true
		names = append(names, entry)
	}
	if len(names) == 0 {
		return "none"
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// loadManifestPlugins reads the manifest, accepting both the current map
// format and the legacy array format. A missing file yields no plugins.
// Mirrors privateer-sdk internal/manifest.Load, which pvtr cannot import; keep in
// sync.
func loadManifestPlugins(pluginsDir string) ([]manifestPlugin, error) {
	data, err := os.ReadFile(filepath.Join(pluginsDir, pluginManifestFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var current struct {
		Plugins map[string]manifestPlugin `json:"plugins"`
	}
	if err := json.Unmarshal(data, &current); err == nil {
		out := make([]manifestPlugin, 0, len(current.Plugins))
		for _, p := range current.Plugins {
			out = append(out, p)
		}
		return out, nil
	}
	var legacy struct {
		Plugins []manifestPlugin `json:"plugins"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", pluginManifestFile, err)
	}
	return legacy.Plugins, nil
}
