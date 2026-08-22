package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current semantic version of AIO-PANEL
	Version = "0.1.0-dev"

	// GitCommit is set dynamically during build time (-ldflags)
	GitCommit = "HEAD"

	// BuildDate is set dynamically during build time (-ldflags)
	BuildDate = "unknown"
)

// String returns a formatted version string
func String() string {
	return fmt.Sprintf("AIO-PANEL v%s (%s) built on %s [%s/%s]",
		Version, GitCommit, BuildDate, runtime.GOOS, runtime.GOARCH)
}
