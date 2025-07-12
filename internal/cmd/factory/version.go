package factory

import (
	"context"
	"fmt"
	"runtime"

	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
	"github.com/tim/cu/internal/version"
)

// VersionCommand implements the version command using dependency injection
type VersionCommand struct {
	*base.Command
}

// createVersionCommand creates a new version command
func (f *Factory) createVersionCommand() interfaces.Command {
	cmd := &VersionCommand{
		Command: &base.Command{
			Use:    "version",
			Short:  "Show cu version information",
			Long:   `Display the version of cu along with build information.`,
			Output: f.output,
			Config: f.config,
			// Version command doesn't need API or Auth
		},
	}

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the version command
func (c *VersionCommand) run(ctx context.Context, args []string) error {
	// Get version information
	versionInfo := version.FullVersion()

	// Check if we should output JSON or other formats
	format := c.Config.GetString("output")
	
	switch format {
	case "json":
		// Output structured version data
		data := map[string]string{
			"version":   version.Version,
			"commit":    version.Commit,
			"date":      version.Date,
			"builtBy":   version.BuiltBy,
			"goVersion": runtime.Version(),
			"platform":  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		}
		return c.Output.Print(data)
	case "yaml":
		// Output structured version data
		data := map[string]string{
			"version":   version.Version,
			"commit":    version.Commit,
			"date":      version.Date,
			"builtBy":   version.BuiltBy,
			"goVersion": runtime.Version(),
			"platform":  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		}
		return c.Output.Print(data)
	default:
		// Default text output
		c.Output.PrintInfo(versionInfo)
		return nil
	}
}