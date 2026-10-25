package main

import (
	"runtime/debug"
	"testing"
)

func TestParseVCSInfo(t *testing.T) {
	settings := []debug.BuildSetting{
		{Key: "vcs.revision", Value: "1234567890abcdef"},
		{Key: "vcs.time", Value: "2026-10-25T10:00:00Z"},
		{Key: "vcs.modified", Value: "true"},
	}

	commit, builtAt, postfix := parseVCSInfo(settings)

	if commit != "1234567890abcdef" {
		t.Errorf("expected commit '1234567890abcdef', got %q", commit)
	}
	if builtAt != "2026-10-25T10:00:00Z" {
		t.Errorf("expected builtAt '2026-10-25T10:00:00Z', got %q", builtAt)
	}
	if postfix != "-dev" {
		t.Errorf("expected postfix '-dev', got %q", postfix)
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
	origGitCommitHash := GitCommitHash
	origBuiltAt := BuiltAt
	origVersionPostfix := VersionPostfix
	defer func() {
		GitCommitHash = origGitCommitHash
		BuiltAt = origBuiltAt
		VersionPostfix = origVersionPostfix
	}()

	setVCSInfo()
}
