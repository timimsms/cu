package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the CLI
	Version = "dev"
	// Commit is the git commit hash
	Commit = "none"
	// Date is the build date
	Date = "unknown"
	// BuiltBy indicates who built the binary
	BuiltBy = "unknown"
)

// FullVersion returns the full version string
func FullVersion() string {
	return fmt.Sprintf(`cu version %s

Build Details:
  Commit:  %s
  Date:    %s
  Built by: %s
  
System:
  OS/Arch: %s/%s
  Go:      %s`,
		Version, Commit, Date, BuiltBy,
		runtime.GOOS, runtime.GOARCH, runtime.Version())
}
