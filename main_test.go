package main

import (
	"runtime/debug"
	"testing"
)

func TestParseVCSInfo(t *testing.T) {
	settingsModified := []debug.BuildSetting{
		{Key: "vcs.revision", Value: "1234567890abcdef"},
		{Key: "vcs.time", Value: "2026-10-25T10:00:00Z"},
		{Key: "vcs.modified", Value: "true"},
	}

	commit, builtAt, postfix := parseVCSInfo(settingsModified)

	if commit != "1234567890abcdef" {
		t.Errorf("expected commit '1234567890abcdef', got %q", commit)
	}
	if builtAt != "2026-10-25T10:00:00Z" {
		t.Errorf("expected builtAt '2026-10-25T10:00:00Z', got %q", builtAt)
	}
	if postfix != "-dev" {
		t.Errorf("expected postfix '-dev', got %q", postfix)
	}

	settingsClean := []debug.BuildSetting{
		{Key: "vcs.revision", Value: "abcdef1234567890"},
		{Key: "vcs.time", Value: "2026-10-25T12:00:00Z"},
		{Key: "vcs.modified", Value: "false"},
	}

	commit, builtAt, postfix = parseVCSInfo(settingsClean)
	if commit != "abcdef1234567890" {
		t.Errorf("expected commit 'abcdef1234567890', got %q", commit)
	}
	if builtAt != "2026-10-25T12:00:00Z" {
		t.Errorf("expected builtAt '2026-10-25T12:00:00Z', got %q", builtAt)
	}
	if postfix != "" {
		t.Errorf("expected empty postfix for clean build, got %q", postfix)
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		postfix  string
		expected string
	}{
		{
			name:     "empty version defaults to edge",
			version:  "",
			postfix:  "-dev",
			expected: "edge-dev",
		},
		{
			name:     "custom version with no postfix",
			version:  "1.0.0",
			postfix:  "",
			expected: "1.0.0",
		},
		{
			name:     "custom version with release candidate postfix",
			version:  "2.0.0",
			postfix:  "-rc1",
			expected: "2.0.0-rc1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatVersion(tc.version, tc.postfix)
			if result != tc.expected {
				t.Errorf("formatVersion(%q, %q) = %q, expected %q", tc.version, tc.postfix, result, tc.expected)
			}
		})
	}
}

func TestSetVCSInfo(t *testing.T) {
	origVersion := Version
	origGitCommitHash := GitCommitHash
	origBuiltAt := BuiltAt
	origVersionPostfix := VersionPostfix
	defer func() {
		Version = origVersion
		GitCommitHash = origGitCommitHash
		BuiltAt = origBuiltAt
		VersionPostfix = origVersionPostfix
	}()

	// 1. Unset fields are populated from VCS when empty
	Version = ""
	GitCommitHash = ""
	BuiltAt = ""
	VersionPostfix = ""

	settings := []debug.BuildSetting{
		{Key: "vcs.revision", Value: "new-hash"},
		{Key: "vcs.time", Value: "new-time"},
		{Key: "vcs.modified", Value: "true"},
	}
	setVCSInfo(settings)

	if GitCommitHash != "new-hash" {
		t.Errorf("expected GitCommitHash 'new-hash', got %q", GitCommitHash)
	}
	if BuiltAt != "new-time" {
		t.Errorf("expected BuiltAt 'new-time', got %q", BuiltAt)
	}
	if VersionPostfix != "-dev" {
		t.Errorf("expected VersionPostfix '-dev', got %q", VersionPostfix)
	}

	// 2. Explicit fields (e.g. from ldflags) are preserved and not overwritten
	Version = "1.0.0"
	GitCommitHash = "explicit-hash"
	BuiltAt = "explicit-time"
	VersionPostfix = "-custom"

	setVCSInfo(settings)

	if GitCommitHash != "explicit-hash" {
		t.Errorf("expected GitCommitHash 'explicit-hash' to be preserved, got %q", GitCommitHash)
	}
	if BuiltAt != "explicit-time" {
		t.Errorf("expected BuiltAt 'explicit-time' to be preserved, got %q", BuiltAt)
	}
	if VersionPostfix != "-custom" {
		t.Errorf("expected VersionPostfix '-custom' to be preserved, got %q", VersionPostfix)
	}

	// 3. Explicit Version does not get -dev appended on dirty tree
	Version = "1.0.0"
	VersionPostfix = ""
	setVCSInfo(settings)
	if VersionPostfix != "" {
		t.Errorf("expected VersionPostfix to remain empty for explicit Version, got %q", VersionPostfix)
	}
}

func TestRunMain(t *testing.T) {
	origExecuteCLI := executeCLI
	origReadBuildInfo := readBuildInfo
	origVersion := Version
	origVersionPostfix := VersionPostfix
	origGitCommitHash := GitCommitHash
	origBuiltAt := BuiltAt
	defer func() {
		executeCLI = origExecuteCLI
		readBuildInfo = origReadBuildInfo
		Version = origVersion
		VersionPostfix = origVersionPostfix
		GitCommitHash = origGitCommitHash
		BuiltAt = origBuiltAt
	}()

	var calledVersion, calledCommit, calledBuiltAt string
	executeCLI = func(version, commit, builtAt string) {
		calledVersion = version
		calledCommit = commit
		calledBuiltAt = builtAt
	}

	t.Run("explicit version", func(t *testing.T) {
		Version = "1.2.3"
		VersionPostfix = "-custom"
		GitCommitHash = "commithash"
		BuiltAt = "buildtime"

		runMain()

		if calledVersion != "1.2.3-custom" {
			t.Errorf("expected calledVersion '1.2.3-custom', got %q", calledVersion)
		}
		if calledCommit != "commithash" {
			t.Errorf("expected calledCommit 'commithash', got %q", calledCommit)
		}
		if calledBuiltAt != "buildtime" {
			t.Errorf("expected calledBuiltAt 'buildtime', got %q", calledBuiltAt)
		}
	})

	t.Run("empty version with build info", func(t *testing.T) {
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "mock-hash"},
					{Key: "vcs.time", Value: "mock-time"},
					{Key: "vcs.modified", Value: "true"},
				},
			}, true
		}

		Version = ""
		VersionPostfix = ""
		GitCommitHash = ""
		BuiltAt = ""

		runMain()

		if calledVersion != "edge-dev" {
			t.Errorf("expected calledVersion 'edge-dev' for empty Version, got %q", calledVersion)
		}
		if calledCommit != "mock-hash" {
			t.Errorf("expected calledCommit 'mock-hash', got %q", calledCommit)
		}
		if calledBuiltAt != "mock-time" {
			t.Errorf("expected calledBuiltAt 'mock-time', got %q", calledBuiltAt)
		}
	})

	t.Run("explicit version with fallback build info", func(t *testing.T) {
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "mock-hash"},
					{Key: "vcs.time", Value: "mock-time"},
				},
			}, true
		}

		Version = "2.0.0"
		VersionPostfix = ""
		GitCommitHash = ""
		BuiltAt = ""

		runMain()

		if calledVersion != "2.0.0" {
			t.Errorf("expected calledVersion '2.0.0', got %q", calledVersion)
		}
		if calledCommit != "mock-hash" {
			t.Errorf("expected calledCommit 'mock-hash', got %q", calledCommit)
		}
		if calledBuiltAt != "mock-time" {
			t.Errorf("expected calledBuiltAt 'mock-time', got %q", calledBuiltAt)
		}
	})

	t.Run("partial ldflags preserves explicit commit and fills build time", func(t *testing.T) {
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "full-40-char-vcs-revision"},
					{Key: "vcs.time", Value: "mock-time"},
				},
			}, true
		}

		Version = "1.0.0"
		VersionPostfix = ""
		GitCommitHash = "abc1234"
		BuiltAt = ""

		runMain()

		if calledVersion != "1.0.0" {
			t.Errorf("expected calledVersion '1.0.0', got %q", calledVersion)
		}
		if calledCommit != "abc1234" {
			t.Errorf("expected calledCommit 'abc1234' to be preserved, got %q", calledCommit)
		}
		if calledBuiltAt != "mock-time" {
			t.Errorf("expected calledBuiltAt 'mock-time', got %q", calledBuiltAt)
		}
	})

	t.Run("explicit version with dirty tree does not append dev postfix", func(t *testing.T) {
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "mock-hash"},
					{Key: "vcs.time", Value: "mock-time"},
					{Key: "vcs.modified", Value: "true"},
				},
			}, true
		}

		Version = "0.2.0"
		VersionPostfix = ""
		GitCommitHash = ""
		BuiltAt = ""

		runMain()

		if calledVersion != "0.2.0" {
			t.Errorf("expected calledVersion '0.2.0' without -dev, got %q", calledVersion)
		}
		if calledCommit != "mock-hash" {
			t.Errorf("expected calledCommit 'mock-hash', got %q", calledCommit)
		}
		if calledBuiltAt != "mock-time" {
			t.Errorf("expected calledBuiltAt 'mock-time', got %q", calledBuiltAt)
		}
	})
}
