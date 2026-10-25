// Package main provides the entry point for the pvtr command-line tool.
// pvtr is a validation framework that executes plugins to perform
// compliance and security evaluations.
package main

import (
	"runtime/debug"

	"github.com/privateerproj/pvtr/cmd"
)

var (
	// Version is the version string that will be replaced at build time by the associated tag.
	Version string
	// VersionPostfix is a marker for the version such as "dev", "beta", "rc", etc.
	// This is appended to the version string if it is not empty.
	VersionPostfix string
	// GitCommitHash is the git commit hash at build time.
	GitCommitHash string
	// BuiltAt is the actual build datetime string.
	BuiltAt string
)

func parseVCSInfo(settings []debug.BuildSetting) (commit, builtAt, postfix string) {
	for _, setting := range settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.time":
			builtAt = setting.Value
		case "vcs.modified":
			postfix = "-dev"
		}
	}
	return
}

func setVCSInfo() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	GitCommitHash, BuiltAt, VersionPostfix = parseVCSInfo(info.Settings)
}

// formatVersion returns the version string, defaulting to "edge" if empty and appending any postfix.
func formatVersion(version, postfix string) string {
	if version == "" {
		version = "edge"
	}
	return version + postfix
}

// main is the entry point for the pvtr application.
// It formats the version string with any postfix and executes the root command.
func main() {
	if Version == "" {
		setVCSInfo()
	}
	Version = formatVersion(Version, VersionPostfix)
	cmd.NewCLI(Version, GitCommitHash, BuiltAt).Execute()
}
