// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package version provides build-time version information for all
// straitKubegateway binaries.
package version

import (
	"fmt"
	"runtime"
)

// These variables are populated at link time via -ldflags.
var (
	// Version is the semantic version string (e.g. "v0.1.0").
	Version = "dev"
	// GitCommit is the short Git commit hash.
	GitCommit = "unknown"
	// BuildDate is the RFC3339 build timestamp.
	BuildDate = "unknown"
)

// Info holds the full version information for a binary.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the current version information.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a human-readable version string.
func (i Info) String() string {
	return fmt.Sprintf("straitKubegateway %s (commit: %s, built: %s, go: %s, %s/%s)",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.OS, i.Arch)
}
