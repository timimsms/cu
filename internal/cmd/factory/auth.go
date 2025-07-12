package factory

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// AuthCommand implements the auth command with dependency injection
type AuthCommand struct {
	*base.Command
	subcommands map[string]func(context.Context, []string) error
	
	// Input/output dependencies for testing
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	
	// Flags
	token     string
	workspace string
}

// createAuthCommand creates a new auth command
func (f *Factory) createAuthCommand() interfaces.Command {
	cmd := &AuthCommand{
		Command: &base.Command{
			Use:   "auth",
			Short: "Manage authentication with ClickUp",
			Long:  `Authenticate cu with ClickUp API using personal tokens or OAuth.`,
			API:    f.api,
			Auth:   f.auth,
			Output: f.output,
			Config: f.config,
		},
		subcommands: make(map[string]func(context.Context, []string) error),
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
	}

	// Register subcommands
	cmd.subcommands["login"] = cmd.runLogin
	cmd.subcommands["status"] = cmd.runStatus
	cmd.subcommands["logout"] = cmd.runLogout

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the auth command
func (c *AuthCommand) run(ctx context.Context, args []string) error {
	// Auth command requires a subcommand
	if len(args) == 0 {
		return fmt.Errorf("no subcommand specified. Available subcommands: login, status, logout")
	}

	subcommand := args[0]
	handler, exists := c.subcommands[subcommand]
	if !exists {
		return fmt.Errorf("unknown subcommand: %s. Available subcommands: login, status, logout", subcommand)
	}

	// Execute subcommand with remaining args
	return handler(ctx, args[1:])
}

// runLogin executes the auth login subcommand
func (c *AuthCommand) runLogin(ctx context.Context, args []string) error {
	// Ensure Auth manager is available
	if c.Auth == nil {
		return fmt.Errorf("auth manager not initialized")
	}

	// If token is provided via flag, use it
	if c.token != "" {
		authToken := &auth.Token{
			Value:     c.token,
			Workspace: c.workspace,
		}

		if err := c.Auth.SaveToken(c.workspace, authToken); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		c.Output.PrintSuccess("Successfully authenticated!")
		return nil
	}

	// Interactive authentication
	c.Output.PrintInfo("To authenticate, you'll need a ClickUp personal API token.")
	c.Output.PrintInfo("You can create one at: https://app.clickup.com/settings/apps")
	fmt.Fprintln(c.stdout)

	reader := bufio.NewReader(c.stdin)
	fmt.Fprint(c.stdout, "Enter your ClickUp API token: ")
	tokenInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	tokenInput = strings.TrimSpace(tokenInput)
	if tokenInput == "" {
		return fmt.Errorf("token cannot be empty")
	}

	authToken := &auth.Token{
		Value:     tokenInput,
		Workspace: c.workspace,
	}

	if err := c.Auth.SaveToken(c.workspace, authToken); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Save workspace as default if it's the first one
	if c.workspace != "" && c.workspace != auth.DefaultWorkspace {
		c.Config.Set("default_workspace", c.workspace)
		if saver, ok := c.Config.(interface{ Save() error }); ok {
			if err := saver.Save(); err != nil {
				// Log warning but don't fail - the auth is already saved
				c.Output.PrintWarning(fmt.Sprintf("failed to save default workspace: %v", err))
			}
		}
	}

	fmt.Fprintln(c.stdout)
	c.Output.PrintSuccess("Successfully authenticated!")
	c.Output.PrintInfo("You can now use cu commands to interact with ClickUp.")
	return nil
}

// runStatus executes the auth status subcommand
func (c *AuthCommand) runStatus(ctx context.Context, args []string) error {
	// Ensure Auth manager is available
	if c.Auth == nil {
		return fmt.Errorf("auth manager not initialized")
	}

	workspace := c.Config.GetString("default_workspace")
	if workspace == "" {
		workspace = auth.DefaultWorkspace
	}

	token, err := c.Auth.GetToken(workspace)
	if err != nil {
		c.Output.PrintInfo("Not authenticated")
		fmt.Fprintln(c.stdout)
		c.Output.PrintInfo("Run 'cu auth login' to authenticate")
		return fmt.Errorf("not authenticated")
	}

	c.Output.PrintInfo("Authenticated")
	c.Output.PrintInfo(fmt.Sprintf("Workspace: %s", workspace))
	if token.Email != "" {
		c.Output.PrintInfo(fmt.Sprintf("Email: %s", token.Email))
	}
	fmt.Fprintln(c.stdout)
	c.Output.PrintInfo("Token stored securely in system keychain")
	return nil
}

// runLogout executes the auth logout subcommand
func (c *AuthCommand) runLogout(ctx context.Context, args []string) error {
	// Ensure Auth manager is available
	if c.Auth == nil {
		return fmt.Errorf("auth manager not initialized")
	}

	workspace := c.workspace
	if workspace == "" {
		workspace = c.Config.GetString("default_workspace")
		if workspace == "" {
			workspace = auth.DefaultWorkspace
		}
	}

	if err := c.Auth.DeleteToken(workspace); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	c.Output.PrintSuccess(fmt.Sprintf("Successfully logged out from workspace: %s", workspace))
	return nil
}

// GetCobraCommand returns the cobra command with subcommands
func (c *AuthCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()

	// Add login subcommand
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with ClickUp",
		Long:  `Authenticate with ClickUp using a personal API token or OAuth device flow.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.token, _ = cmd.Flags().GetString("token")
			c.workspace, _ = cmd.Flags().GetString("workspace")
			
			return c.runLogin(cmd.Context(), args)
		},
	}

	// Add status subcommand
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  `Display the current authentication status and user information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runStatus(cmd.Context(), args)
		},
	}

	// Add logout subcommand
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out from ClickUp",
		Long:  `Remove stored authentication credentials.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.workspace, _ = cmd.Flags().GetString("workspace")
			
			return c.runLogout(cmd.Context(), args)
		},
	}

	// Add flags to login subcommand
	loginCmd.Flags().StringP("token", "t", "", "Personal API token")
	loginCmd.Flags().StringP("workspace", "w", "", "Workspace name")

	// Add flags to logout subcommand
	logoutCmd.Flags().StringP("workspace", "w", "", "Workspace to logout from")

	cmd.AddCommand(loginCmd, statusCmd, logoutCmd)

	return cmd
}

// SetStdin sets the stdin for testing
func (c *AuthCommand) SetStdin(stdin io.Reader) {
	c.stdin = stdin
}

// SetStdout sets the stdout for testing
func (c *AuthCommand) SetStdout(stdout io.Writer) {
	c.stdout = stdout
}

// SetStderr sets the stderr for testing
func (c *AuthCommand) SetStderr(stderr io.Writer) {
	c.stderr = stderr
}