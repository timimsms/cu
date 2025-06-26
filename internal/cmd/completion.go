package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
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
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported shell type: %s\n", args[0])
			os.Exit(1)
		}
		
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate completion script: %v\n", err)
			os.Exit(1)
		}
	},
}