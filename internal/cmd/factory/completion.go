package factory

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// CompletionCommand implements the completion command using dependency injection
type CompletionCommand struct {
	*base.Command
	rootCmd *cobra.Command
}

// createCompletionCommand creates a new completion command
func (f *Factory) createCompletionCommand() interfaces.Command {
	cmd := &CompletionCommand{
		Command: &base.Command{
			Use:   "completion [bash|zsh|fish|powershell]",
			Short: "Generate shell completion script",
			Long: `Generate a shell completion script for cu.

To load completions:

Bash:
  $ source <(cu completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ cu completion bash > /etc/bash_completion.d/cu
  # macOS:
  $ cu completion bash > $(brew --prefix)/etc/bash_completion.d/cu

Zsh:
  $ source <(cu completion zsh)

  # To load completions for each session, execute once:
  $ cu completion zsh > "${fpath[1]}/_cu"

Fish:
  $ cu completion fish | source

  # To load completions for each session, execute once:
  $ cu completion fish > ~/.config/fish/completions/cu.fish

PowerShell:
  PS> cu completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> cu completion powershell > cu.ps1
  # and source this file from your PowerShell profile.
`,
			Output: f.output,
			// Completion command doesn't need API, Auth, or Config
		},
		// Note: rootCmd will be set when integrating with main app
	}

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the completion command
func (c *CompletionCommand) run(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("exactly one argument required: shell type")
	}

	// For testing, use stdout directly. In production, this will be handled by the output formatter
	var writer io.Writer = os.Stdout
	if c.Output != nil {
		// Allow tests to capture output
		if pw, ok := c.Output.(io.Writer); ok {
			writer = pw
		}
	}

	// Get the root command - in tests this might be nil
	rootCmd := c.rootCmd
	if rootCmd == nil {
		// Try to get from cobra command
		if cobraCmd := c.Command.GetCobraCommand(); cobraCmd != nil {
			rootCmd = cobraCmd.Root()
		}
	}
	if rootCmd == nil {
		return fmt.Errorf("root command not available")
	}

	var err error
	switch args[0] {
	case "bash":
		err = rootCmd.GenBashCompletion(writer)
	case "zsh":
		err = rootCmd.GenZshCompletion(writer)
	case "fish":
		err = rootCmd.GenFishCompletion(writer, true)
	case "powershell":
		err = rootCmd.GenPowerShellCompletionWithDesc(writer)
	default:
		return fmt.Errorf("unsupported shell type: %s", args[0])
	}

	if err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	return nil
}

// SetRootCommand sets the root command for completion generation
func (c *CompletionCommand) SetRootCommand(rootCmd *cobra.Command) {
	c.rootCmd = rootCmd
}

// GetCobraCommand returns the cobra command with completion-specific settings
func (c *CompletionCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()
	
	// Apply completion-specific settings
	cmd.DisableFlagsInUseLine = true
	cmd.ValidArgs = []string{"bash", "zsh", "fish", "powershell"}
	cmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	
	return cmd
}