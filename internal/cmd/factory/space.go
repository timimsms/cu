package factory

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// SpaceCommand implements the space command with dependency injection
type SpaceCommand struct {
	*base.Command
	subcommands map[string]func(context.Context, []string) error
}

// createSpaceCommand creates a new space command
func (f *Factory) createSpaceCommand() interfaces.Command {
	cmd := &SpaceCommand{
		Command: &base.Command{
			Use:   "space",
			Short: "Manage spaces",
			Long:  `View and manage ClickUp spaces within your workspace.`,
			API:    f.api,
			Auth:   f.auth,
			Output: f.output,
			Config: f.config,
		},
		subcommands: make(map[string]func(context.Context, []string) error),
	}

	// Register subcommands
	cmd.subcommands["list"] = cmd.runList

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the space command
func (c *SpaceCommand) run(ctx context.Context, args []string) error {
	// If no subcommand, default to list
	if len(args) == 0 {
		return c.runList(ctx, args)
	}

	subcommand := args[0]
	handler, exists := c.subcommands[subcommand]
	if !exists {
		return fmt.Errorf("unknown subcommand: %s. Available subcommands: list", subcommand)
	}

	// Execute subcommand with remaining args
	return handler(ctx, args[1:])
}

// runList executes the space list subcommand
func (c *SpaceCommand) runList(ctx context.Context, args []string) error {
	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Get workspaces first
	workspaces, err := c.API.GetWorkspaces(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		return fmt.Errorf("no workspaces found")
	}

	// For now, use the first workspace
	// TODO: Add workspace selection support
	workspace := workspaces[0]

	// Get spaces from the workspace
	spaces, err := c.API.GetSpaces(ctx, workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to get spaces: %w", err)
	}

	// Format output
	format := c.Config.GetString("output")
	if format == "" {
		format = "table"
	}

	if format == "table" {
		// Prepare table data
		type spaceRow struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Private  bool   `json:"private"`
			Archived bool   `json:"archived"`
		}

		var rows []spaceRow
		for _, space := range spaces {
			row := spaceRow{
				ID:       space.ID,
				Name:     space.Name,
				Private:  space.Private,
				Archived: space.Archived,
			}
			rows = append(rows, row)
		}

		return c.Output.Print(rows)
	}

	// For other formats, output raw space data
	return c.Output.Print(spaces)
}

// GetCobraCommand returns the cobra command with subcommands
func (c *SpaceCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()

	// Add list subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all spaces",
		Long:  `List all spaces in your ClickUp workspace.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runList(cmd.Context(), args)
		},
	}

	cmd.AddCommand(listCmd)

	return cmd
}