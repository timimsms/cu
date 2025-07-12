package factory

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
	"github.com/tim/cu/internal/version"
)

// RootCommand implements the root command with dependency injection
type RootCommand struct {
	*base.Command
	factory         *Factory
	subcommands     []interfaces.Command
	cfgFile         string
	debug           bool
	outputFormat    string
	rootCobraCmd    *cobra.Command
}

// NewRootCommand creates a new root command with the factory
func NewRootCommand(factory *Factory) (*RootCommand, error) {
	cmd := &RootCommand{
		Command: &base.Command{
			Use:   "cu",
			Short: "A GitHub CLI-inspired command-line interface for ClickUp",
			Long: `cu is a command-line interface for ClickUp that provides GitHub CLI-like
functionality for managing tasks, lists, spaces, and other ClickUp resources.

It allows developers and teams to interact with ClickUp directly from the terminal,
enabling efficient task management and seamless integration with development workflows.`,
			Output: factory.output,
			Config: factory.config,
		},
		factory:     factory,
		subcommands: make([]interfaces.Command, 0),
	}

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	// Initialize all subcommands
	if err := cmd.initSubcommands(); err != nil {
		return nil, fmt.Errorf("failed to initialize subcommands: %w", err)
	}

	return cmd, nil
}

// run handles the root command execution
func (c *RootCommand) run(ctx context.Context, args []string) error {
	// If no args provided, show help
	if len(args) == 0 && c.rootCobraCmd != nil {
		return c.rootCobraCmd.Help()
	}
	return nil
}

// initSubcommands initializes all subcommands
func (c *RootCommand) initSubcommands() error {
	// List of commands to create
	commandNames := []string{
		"auth",
		"config", 
		"completion",
		"version",
		"interactive",
		"task",
		"list",
		"space",
		// Add other commands as they are refactored
	}

	// Create each command
	for _, name := range commandNames {
		cmd, err := c.factory.CreateCommand(name)
		if err != nil {
			// Skip commands that aren't implemented yet
			if err.Error() == fmt.Sprintf("unknown command: %s", name) {
				continue
			}
			return fmt.Errorf("failed to create %s command: %w", name, err)
		}
		if cmd != nil {
			c.subcommands = append(c.subcommands, cmd)
		}
	}

	return nil
}

// GetCobraCommand returns the cobra command with all subcommands configured
func (c *RootCommand) GetCobraCommand() *cobra.Command {
	if c.rootCobraCmd != nil {
		return c.rootCobraCmd
	}

	cmd := &cobra.Command{
		Use:   c.Use,
		Short: c.Short,
		Long:  c.Long,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize configuration if needed
			if c.Config != nil {
				// Config is already injected, no need to initialize from file
				// This allows for better testing
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(cmd.Context(), args)
		},
	}

	// Add persistent flags
	cmd.PersistentFlags().StringVar(&c.cfgFile, "config", "", "config file (default is $HOME/.config/cu/config.yml)")
	cmd.PersistentFlags().BoolVar(&c.debug, "debug", false, "enable debug mode")
	cmd.PersistentFlags().StringVarP(&c.outputFormat, "output", "o", "table", "output format (table|json|yaml|csv)")

	// Set version
	cmd.Version = version.Version
	cmd.SetVersionTemplate(version.FullVersion())

	// Add all subcommands
	for _, subcmd := range c.subcommands {
		if subcmd != nil {
			cmd.AddCommand(subcmd.GetCobraCommand())
		}
	}

	// Store reference for later use
	c.rootCobraCmd = cmd

	return cmd
}

// Execute runs the root command
func (c *RootCommand) Execute() error {
	cmd := c.GetCobraCommand()
	return cmd.Execute()
}

// AddCommand adds a subcommand to the root command
func (c *RootCommand) AddCommand(cmd interfaces.Command) {
	c.subcommands = append(c.subcommands, cmd)
	
	// If cobra command is already created, add it directly
	if c.rootCobraCmd != nil && cmd != nil {
		c.rootCobraCmd.AddCommand(cmd.GetCobraCommand())
	}
}

// GetFactory returns the command factory
func (c *RootCommand) GetFactory() *Factory {
	return c.factory
}