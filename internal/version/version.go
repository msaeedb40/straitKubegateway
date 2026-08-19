// Package version provides build-time and runtime version information
// for straitKubegateway components.
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current semantic version, populated at link time.
	Version = "v1.0.0-dev"

	// GitCommit is the git commit SHA, populated at link time.
	GitCommit = "unknown"

	// BuildDate is the RFC3339 build date, populated at link time.
	BuildDate = "unknown"
)

// Info holds version and runtime details.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
}

// Get returns the full version info struct.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a human-readable summary of the version info.
func (i Info) String() string {
	return fmt.Sprintf("straitKubegateway %s (%s) built at %s on %s using %s",
		i.Version, i.GitCommit, i.BuildDate, i.Platform, i.GoVersion)
}
