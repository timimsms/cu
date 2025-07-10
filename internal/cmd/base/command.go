package base

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/interfaces"
)

// Command provides base functionality for all commands
type Command struct {
	// Dependencies
	API    interfaces.APIClient
	Auth   interfaces.AuthManager
	Output interfaces.OutputFormatter
	Config interfaces.ConfigProvider

	// Command metadata
	Use   string
	Short string
	Long  string

	// Cobra command
	cmd *cobra.Command

	// Execution function
	RunFunc func(ctx context.Context, args []string) error
}

// Setup initializes the command
func (c *Command) Setup() {
	c.cmd = &cobra.Command{
		Use:   c.Use,
		Short: c.Short,
		Long:  c.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with command
			ctx := context.WithValue(cmd.Context(), "command", cmd)
			
			// Check authentication if needed
			if c.requiresAuth() && !c.isAuthenticated() {
				return fmt.Errorf("not authenticated. Please run 'cu auth login' first")
			}

			// Execute the actual command logic
			if c.RunFunc != nil {
				return c.RunFunc(ctx, args)
			}
			
			return fmt.Errorf("command not implemented")
		},
	}
}

// GetCobraCommand returns the underlying cobra command
func (c *Command) GetCobraCommand() *cobra.Command {
	if c.cmd == nil {
		c.Setup()
	}
	return c.cmd
}

// Execute runs the command
func (c *Command) Execute(ctx context.Context, args []string) error {
	if c.RunFunc != nil {
		return c.RunFunc(ctx, args)
	}
	return fmt.Errorf("command not implemented")
}

// AddFlag adds a flag to the command
func (c *Command) AddFlag(name, shorthand, defaultValue, usage string) {
	if c.cmd == nil {
		c.Setup()
	}
	c.cmd.Flags().StringP(name, shorthand, defaultValue, usage)
}

// AddBoolFlag adds a boolean flag to the command
func (c *Command) AddBoolFlag(name, shorthand string, defaultValue bool, usage string) {
	if c.cmd == nil {
		c.Setup()
	}
	c.cmd.Flags().BoolP(name, shorthand, defaultValue, usage)
}

// GetFlag retrieves a flag value
func (c *Command) GetFlag(name string) (string, error) {
	if c.cmd == nil {
		return "", fmt.Errorf("command not initialized")
	}
	return c.cmd.Flags().GetString(name)
}

// GetBoolFlag retrieves a boolean flag value
func (c *Command) GetBoolFlag(name string) (bool, error) {
	if c.cmd == nil {
		return false, fmt.Errorf("command not initialized")
	}
	return c.cmd.Flags().GetBool(name)
}

// requiresAuth determines if the command requires authentication
func (c *Command) requiresAuth() bool {
	// Commands that don't require auth
	noAuthCommands := map[string]bool{
		"version":    true,
		"help":       true,
		"completion": true,
		"auth":       true,
	}
	
	return !noAuthCommands[c.Use]
}

// isAuthenticated checks if the user is authenticated
func (c *Command) isAuthenticated() bool {
	if c.Auth == nil {
		return false
	}
	
	workspace := c.Config.GetString("workspace")
	if workspace == "" {
		workspace = "default"
	}
	
	return c.Auth.IsAuthenticated(workspace)
}