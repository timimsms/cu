package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show cu version information",
	Long:  `Display the version of cu along with build information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.FullVersion())
	},
}
