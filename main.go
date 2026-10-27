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
			if setting.Value == "true" {
				postfix = "-dev"
			}
		}
	}
	return
}

func setVCSInfo(settings []debug.BuildSetting) {
	commit, builtAt, postfix := parseVCSInfo(settings)
	if commit != "" {
		GitCommitHash = commit
	}
	if builtAt != "" {
		BuiltAt = builtAt
	}
	if postfix != "" {
		VersionPostfix = postfix
	}
}

// formatVersion returns the version string, defaulting to "edge" if empty and appending any postfix.
func formatVersion(version, postfix string) string {
	if version == "" {
		version = "edge"
	}
	return version + postfix
}

var executeCLI = func(version, commit, builtAt string) {
	cmd.NewCLI(version, commit, builtAt).Execute()
}

func runMain() {
	if Version == "" {
		info, ok := debug.ReadBuildInfo()
		if ok {
			setVCSInfo(info.Settings)
		}
	}
	Version = formatVersion(Version, VersionPostfix)
	executeCLI(Version, GitCommitHash, BuiltAt)
}

// main is the entry point for the pvtr application.
func main() {
	runMain()
}
