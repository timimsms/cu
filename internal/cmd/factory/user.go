package factory

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// UserCommand implements the user command with dependency injection
type UserCommand struct {
	*base.Command
	subcommands map[string]func(context.Context, []string) error
}

// createUserCommand creates a new user command
func (f *Factory) createUserCommand() interfaces.Command {
	cmd := &UserCommand{
		Command: &base.Command{
			Use:   "user",
			Short: "Manage users",
			Long:  `View and manage workspace users.`,
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

// run executes the user command
func (c *UserCommand) run(ctx context.Context, args []string) error {
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

// runList executes the user list subcommand
func (c *UserCommand) runList(ctx context.Context, args []string) error {
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

	// Get workspace members
	users, err := c.API.GetWorkspaceMembers(ctx, workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to get workspace members: %w", err)
	}

	// Format output
	format := c.Config.GetString("output")
	if format == "" {
		format = "table"
	}

	if format == "table" {
		// Prepare table data
		type userRow struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		}

		var rows []userRow
		for _, user := range users {
			// Handle role conversion based on the actual user structure
			roleStr := ""
			if user.Role != nil {
				roleStr = fmt.Sprintf("%d", *user.Role)
			}

			row := userRow{
				ID:       fmt.Sprintf("%d", user.User.ID),
				Username: user.User.Username,
				Email:    user.User.Email,
				Role:     roleStr,
			}
			rows = append(rows, row)
		}

		return c.Output.Print(rows)
	}

	// For other formats, output raw user data
	return c.Output.Print(users)
}

// GetCobraCommand returns the cobra command with subcommands
func (c *UserCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()

	// Add list subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List workspace users",
		Long:  `List all users in your ClickUp workspace.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runList(cmd.Context(), args)
		},
	}

	cmd.AddCommand(listCmd)

	return cmd
}