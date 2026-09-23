package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/timimsms/cu/internal/output"
	"github.com/timimsms/cu/internal/version"
)

// versionInfo is the machine-readable form of the build details that
// version.FullVersion renders for humans.
type versionInfo struct {
	Version   string `json:"version" yaml:"version"`
	Commit    string `json:"commit" yaml:"commit"`
	Date      string `json:"date" yaml:"date"`
	BuiltBy   string `json:"built_by" yaml:"built_by"`
	GoVersion string `json:"go_version" yaml:"go_version"`
	OS        string `json:"os" yaml:"os"`
	Arch      string `json:"arch" yaml:"arch"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show cu version information",
	Long:  `Display the version of cu along with build information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if outputFormat == "table" {
			fmt.Println(version.FullVersion())
			return nil
		}

		return output.Format(outputFormat, versionInfo{
			Version:   version.Version,
			Commit:    version.Commit,
			Date:      version.Date,
			BuiltBy:   version.BuiltBy,
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		})
	},
}
