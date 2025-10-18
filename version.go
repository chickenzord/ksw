package main

import (
	"fmt"
	"runtime"
)

var (
	// Version is set via ldflags during build
	Version = "dev"
	// GitCommit is set via ldflags during build
	GitCommit = "unknown"
	// BuildDate is set via ldflags during build
	BuildDate = "unknown"
)

// VersionInfo holds version information
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetVersionInfo returns version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i VersionInfo) String() string {
	return fmt.Sprintf("portosync %s (%s) built on %s with %s for %s",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.Platform)
}
